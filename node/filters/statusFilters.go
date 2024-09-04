package filters

import (
	"bytes"
	"strings"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data/api"
	"github.com/DharitriOne/drt-chain-core-go/data/block"
	"github.com/DharitriOne/drt-chain-core-go/data/transaction"
)

type statusFilters struct {
	selfShardID uint32
}

// NewStatusFilters will create a new instance of a statusFilters
func NewStatusFilters(selfShardID uint32) *statusFilters {
	return &statusFilters{
		selfShardID: selfShardID,
	}
}

// SetStatusIfIsFailedDCDTTransfer will set the status if the provided transaction if a failed DCDT transfer
func (sf *statusFilters) SetStatusIfIsFailedDCDTTransfer(tx *transaction.ApiTransactionResult) {
	if len(tx.SmartContractResults) < 1 {
		return
	}

	isCrossShardTxDestMe := tx.SourceShard != tx.DestinationShard && sf.selfShardID == tx.DestinationShard
	if !isCrossShardTxDestMe {
		return
	}

	if !isDCDTTransfer(tx) {
		return
	}

	for _, scr := range tx.SmartContractResults {
		setStatusBasedOnSCRDataAndNonce(tx, []byte(scr.Data), scr.Nonce)
	}
}

// ApplyStatusFilters will apply status filters on the provided miniblocks
func (sf *statusFilters) ApplyStatusFilters(miniblocks []*api.MiniBlock) {
	for _, mb := range miniblocks {
		if mb.Type != block.TxBlock.String() {
			continue
		}

		isNotCrossShardDestinationMe := mb.SourceShard == mb.DestinationShard || mb.DestinationShard != sf.selfShardID
		if isNotCrossShardDestinationMe {
			continue
		}

		iterateMiniblockTxsForDCDTTransfer(mb, miniblocks)
	}
}

func iterateMiniblockTxsForDCDTTransfer(miniblock *api.MiniBlock, miniblocks []*api.MiniBlock) {
	for _, tx := range miniblock.Transactions {
		if !isDCDTTransfer(tx) {
			continue
		}

		searchUnsignedTransaction(tx, miniblocks)
	}
}

func searchUnsignedTransaction(tx *transaction.ApiTransactionResult, miniblocks []*api.MiniBlock) {
	for _, mb := range miniblocks {
		if mb.Type != block.SmartContractResultBlock.String() {
			continue
		}

		shouldCheckTransaction := mb.DestinationShard == tx.SourceShard && mb.SourceShard == tx.DestinationShard
		if shouldCheckTransaction {
			tryToSetStatusOfDCDTTransfer(tx, mb)
		}
	}
}

func tryToSetStatusOfDCDTTransfer(tx *transaction.ApiTransactionResult, miniblock *api.MiniBlock) {
	for _, unsignedTx := range miniblock.Transactions {
		if unsignedTx.OriginalTransactionHash != tx.Hash {
			continue
		}

		setStatusBasedOnSCRDataAndNonce(tx, unsignedTx.Data, unsignedTx.Nonce)
	}
}

func setStatusBasedOnSCRDataAndNonce(tx *transaction.ApiTransactionResult, scrDataField []byte, scrNonce uint64) {
	isSCRWithRefund := bytes.HasPrefix(scrDataField, tx.Data) && scrNonce == tx.Nonce
	if isSCRWithRefund {
		tx.Status = transaction.TxStatusFail
		return
	}
}

func isDCDTTransfer(tx *transaction.ApiTransactionResult) bool {
	return strings.HasPrefix(string(tx.Data), core.BuiltInFunctionDCDTTransfer)
}
