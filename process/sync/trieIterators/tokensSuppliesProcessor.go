package trieIterators

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/core/check"
	"github.com/DharitriOne/drt-chain-core-go/data/dcdt"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/common/errChan"
	"github.com/DharitriOne/drt-chain-go/dataRetriever"
	"github.com/DharitriOne/drt-chain-go/dblookupext/dcdtSupply"
	"github.com/DharitriOne/drt-chain-go/state"
)

type tokensSuppliesProcessor struct {
	storageService dataRetriever.StorageService
	marshaller     marshal.Marshalizer
	tokensSupplies map[string]*big.Int
}

// ArgsTokensSuppliesProcessor is the arguments struct for NewTokensSuppliesProcessor
type ArgsTokensSuppliesProcessor struct {
	StorageService dataRetriever.StorageService
	Marshaller     marshal.Marshalizer
}

// NewTokensSuppliesProcessor returns a new instance of tokensSuppliesProcessor
func NewTokensSuppliesProcessor(args ArgsTokensSuppliesProcessor) (*tokensSuppliesProcessor, error) {
	if check.IfNil(args.StorageService) {
		return nil, errNilStorageService
	}
	if check.IfNil(args.Marshaller) {
		return nil, errNilMarshaller
	}

	return &tokensSuppliesProcessor{
		storageService: args.StorageService,
		marshaller:     args.Marshaller,
		tokensSupplies: make(map[string]*big.Int),
	}, nil
}

// HandleTrieAccountIteration is the handler for the trie account iteration
// note that this function is not concurrent safe
func (t *tokensSuppliesProcessor) HandleTrieAccountIteration(userAccount state.UserAccountHandler) error {
	if check.IfNil(userAccount) {
		return errNilUserAccount
	}
	if bytes.Equal(core.SystemAccountAddress, userAccount.AddressBytes()) {
		log.Debug("repopulate tokens supplies: skipping system account address")
		return nil
	}

	rh := userAccount.GetRootHash()
	isValidRootHashToIterateFor := len(rh) > 0 && !bytes.Equal(rh, make([]byte, len(rh)))
	if !isValidRootHashToIterateFor {
		return nil
	}

	dataTrieChan := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}

	errDataTrieGet := userAccount.GetAllLeaves(dataTrieChan, context.Background())
	if errDataTrieGet != nil {
		return fmt.Errorf("%w while getting all leaves for root hash %s", errDataTrieGet, hex.EncodeToString(rh))
	}

	log.Trace("extractTokensSupplies - parsing account", "address", userAccount.AddressBytes())
	dcdtPrefix := []byte(core.ProtectedKeyPrefix + core.DCDTKeyIdentifier)
	for userLeaf := range dataTrieChan.LeavesChan {
		if !bytes.HasPrefix(userLeaf.Key(), dcdtPrefix) {
			continue
		}

		tokenKey := userLeaf.Key()
		lenDCDTPrefix := len(dcdtPrefix)
		value := userLeaf.Value()

		var esToken dcdt.DCDigitalToken
		err := t.marshaller.Unmarshal(&esToken, value)
		if err != nil {
			return fmt.Errorf("%w while unmarshaling the token with key %s", err, hex.EncodeToString(tokenKey))
		}

		tokenName := string(tokenKey)[lenDCDTPrefix:]
		tokenID, nonce := common.ExtractTokenIDAndNonceFromTokenStorageKey([]byte(tokenName))
		t.addToBalance(tokenID, nonce, esToken.Value)
	}

	err := dataTrieChan.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return fmt.Errorf("%w while parsing errors from the trie iteration", err)
	}

	return nil
}

func (t *tokensSuppliesProcessor) addToBalance(tokenID []byte, nonce uint64, value *big.Int) {
	tokenIDStr := string(tokenID)
	if nonce > 0 {
		t.putInSuppliesMap(string(tokenID), value) // put for collection as well
		nonceStr := hex.EncodeToString(big.NewInt(int64(nonce)).Bytes())
		tokenIDStr += fmt.Sprintf("-%s", nonceStr)
	}

	t.putInSuppliesMap(tokenIDStr, value)
}

func (t *tokensSuppliesProcessor) putInSuppliesMap(id string, value *big.Int) {
	currentValue, found := t.tokensSupplies[id]
	if !found {
		t.tokensSupplies[id] = value
		return
	}

	currentValue = big.NewInt(0).Add(currentValue, value)
	t.tokensSupplies[id] = currentValue
}

// SaveSupplies will store the recomputed tokens supplies into the database
// note that this function is not concurrent safe
func (t *tokensSuppliesProcessor) SaveSupplies() error {
	suppliesStorer, err := t.storageService.GetStorer(dataRetriever.DCDTSuppliesUnit)
	if err != nil {
		return err
	}

	for tokenName, supply := range t.tokensSupplies {
		log.Trace("repopulate tokens supplies", "token", tokenName, "supply", supply.String())
		supplyObj := &dcdtSupply.SupplyDCDT{
			Supply:           supply,
			RecomputedSupply: true,
		}
		supplyObjBytes, err := t.marshaller.Marshal(supplyObj)
		if err != nil {
			return err
		}

		err = suppliesStorer.Put([]byte(tokenName), supplyObjBytes)
		if err != nil {
			return fmt.Errorf("%w while saving recomputed supply of the token %s", err, tokenName)
		}
	}

	log.Debug("finished the repopulation of the tokens supplies", "num tokens", len(t.tokensSupplies))

	return nil
}
