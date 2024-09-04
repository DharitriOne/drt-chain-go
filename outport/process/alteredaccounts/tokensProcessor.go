package alteredaccounts

import (
	"math/big"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data"
	outportcore "github.com/DharitriOne/drt-chain-core-go/data/outport"
	"github.com/DharitriOne/drt-chain-go/sharding"
)

const (
	idxTokenIDInTopics         = 0
	idxTokenNonceInTopics      = 1
	idxReceiverAddressInTopics = 3
	minTopicsMultiTransfer     = 4
)

type tokensProcessor struct {
	shardCoordinator sharding.Coordinator
	tokensIdentifier map[string]struct{}
}

func newTokensProcessor(shardCoordinator sharding.Coordinator) *tokensProcessor {
	return &tokensProcessor{
		tokensIdentifier: map[string]struct{}{
			core.BuiltInFunctionDCDTTransfer:         {},
			core.BuiltInFunctionDCDTBurn:             {},
			core.BuiltInFunctionDCDTLocalMint:        {},
			core.BuiltInFunctionDCDTLocalBurn:        {},
			core.BuiltInFunctionDCDTWipe:             {},
			core.BuiltInFunctionMultiDCDTNFTTransfer: {},
			core.BuiltInFunctionDCDTNFTTransfer:      {},
			core.BuiltInFunctionDCDTNFTBurn:          {},
			core.BuiltInFunctionDCDTNFTAddQuantity:   {},
			core.BuiltInFunctionDCDTNFTCreate:        {},
			core.BuiltInFunctionDCDTFreeze:           {},
			core.BuiltInFunctionDCDTUnFreeze:         {},
		},
		shardCoordinator: shardCoordinator,
	}
}

func (tp *tokensProcessor) extractDCDTAccounts(
	txPool *outportcore.TransactionPool,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) error {
	for _, txLog := range txPool.Logs {
		for _, event := range txLog.Log.Events {
			tp.processEvent(event, markedAlteredAccounts)
		}
	}

	return nil
}

func (tp *tokensProcessor) processEvent(
	event data.EventHandler,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) {
	eventIdentifier := string(event.GetIdentifier())
	_, isDCDT := tp.tokensIdentifier[eventIdentifier]
	if !isDCDT {
		return
	}

	topics := event.GetTopics()
	if len(topics) < idxTokenNonceInTopics+1 {
		return
	}

	isMultiTransferEvent := eventIdentifier == core.BuiltInFunctionMultiDCDTNFTTransfer
	if isMultiTransferEvent {
		tp.processMultiTransferEvent(event, markedAlteredAccounts)
		return
	}

	nonce := topics[idxTokenNonceInTopics]
	nonceBigInt := big.NewInt(0).SetBytes(nonce)
	tp.extractDcdtData(event, nonceBigInt, markedAlteredAccounts)
}

func (tp *tokensProcessor) extractDcdtData(
	event data.EventHandler,
	nonce *big.Int,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) {
	address := event.GetAddress()
	topics := event.GetTopics()
	if len(topics) == 0 {
		return
	}

	identifier := string(event.GetIdentifier())
	isNFTCreate := identifier == core.BuiltInFunctionDCDTNFTCreate
	tokenID := topics[idxTokenIDInTopics]
	tp.processDcdtDataForAddress(address, nonce, string(tokenID), markedAlteredAccounts, isNFTCreate)

	// in case of dcdt transfer, nft transfer, wipe or multi dcdt transfers, the 3rd index of the topics contains the destination address
	eventShouldContainReceiverAddress := identifier == core.BuiltInFunctionDCDTTransfer ||
		identifier == core.BuiltInFunctionDCDTNFTTransfer ||
		identifier == core.BuiltInFunctionDCDTWipe ||
		identifier == core.BuiltInFunctionMultiDCDTNFTTransfer ||
		identifier == core.BuiltInFunctionDCDTFreeze ||
		identifier == core.BuiltInFunctionDCDTUnFreeze

	if eventShouldContainReceiverAddress && len(topics) > idxReceiverAddressInTopics {
		destinationAddress := topics[idxReceiverAddressInTopics]
		tp.processDcdtDataForAddress(destinationAddress, nonce, string(tokenID), markedAlteredAccounts, false)
	}
}

func (tp *tokensProcessor) processMultiTransferEvent(event data.EventHandler, markedAlteredAccounts map[string]*markedAlteredAccount) {
	topics := event.GetTopics()
	// MultiDCDTNFTTransfer event
	// N = len(topics)
	// i := 0; i < N-1; i+=3
	// {
	// 		topics[i] --- token identifier
	// 		topics[i+1] --- token nonce
	// 		topics[i+2] --- transferred value
	// }
	// topics[N-1]   --- destination address
	numOfTopics := len(topics)
	if numOfTopics < minTopicsMultiTransfer || numOfTopics%3 != 1 {
		log.Warn("tokensProcessor.processMultiTransferEvent: wrong number of topics", "got", numOfTopics)
		return
	}

	address := event.GetAddress()

	destinationAddress := topics[numOfTopics-1]
	for i := 0; i < numOfTopics-1; i += 3 {
		tokenID := topics[i]
		nonceBigInt := big.NewInt(0).SetBytes(topics[i+1])
		// process event for the sender address
		tp.processDcdtDataForAddress(address, nonceBigInt, string(tokenID), markedAlteredAccounts, false)

		// process event for the destination address
		tp.processDcdtDataForAddress(destinationAddress, nonceBigInt, string(tokenID), markedAlteredAccounts, false)
	}
}

func (tp *tokensProcessor) processDcdtDataForAddress(
	address []byte,
	nonce *big.Int,
	tokenID string,
	markedAlteredAccounts map[string]*markedAlteredAccount,
	isNFTCreate bool,
) {
	if !tp.isSameShard(address) {
		return
	}

	addressStr := string(address)
	markedAccount, exists := markedAlteredAccounts[addressStr]
	if !exists {
		markedAccount = &markedAlteredAccount{}
		markedAlteredAccounts[addressStr] = markedAccount
	}

	if markedAccount.tokens == nil {
		markedAccount.tokens = make(map[string]*markedAlteredAccountToken)
	}

	tokenKey := tokenID + string(nonce.Bytes())
	_, alreadyExists := markedAccount.tokens[tokenKey]
	if alreadyExists {
		markedAccount.tokens[tokenKey].isNFTCreate = markedAccount.tokens[tokenKey].isNFTCreate || isNFTCreate
		return
	}

	markedAccount.tokens[tokenKey] = &markedAlteredAccountToken{
		identifier:  tokenID,
		nonce:       nonce.Uint64(),
		isNFTCreate: isNFTCreate,
	}
}

func (tp *tokensProcessor) isSameShard(address []byte) bool {
	return tp.shardCoordinator.SelfId() == tp.shardCoordinator.ComputeId(address)
}
