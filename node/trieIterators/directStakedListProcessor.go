package trieIterators

import (
	"context"
	"math/big"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data/api"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/common/errChan"
	"github.com/DharitriOne/drt-chain-go/state"
	"github.com/DharitriOne/drt-chain-go/vm"
	logger "github.com/DharitriOne/drt-chain-logger-go"
)

var log = logger.GetOrCreate("node/trieIterators")

type directStakedListProcessor struct {
	*commonStakingProcessor
	publicKeyConverter core.PubkeyConverter
}

var metachainIdentifier = []byte{255}

// NewDirectStakedListProcessor will create a new instance of stakedValuesProc
func NewDirectStakedListProcessor(arg ArgTrieIteratorProcessor) (*directStakedListProcessor, error) {
	err := checkArguments(arg)
	if err != nil {
		return nil, err
	}

	return &directStakedListProcessor{
		commonStakingProcessor: &commonStakingProcessor{
			queryService: arg.QueryService,
			accounts:     arg.Accounts,
		},
		publicKeyConverter: arg.PublicKeyConverter,
	}, nil
}

// GetDirectStakedList will return the list for the direct staked addresses
func (dslp *directStakedListProcessor) GetDirectStakedList(ctx context.Context) ([]*api.DirectStakedValue, error) {
	dslp.accounts.Lock()
	defer dslp.accounts.Unlock()

	validatorAccount, err := dslp.getAccount(vm.ValidatorSCAddress)
	if err != nil {
		return nil, err
	}

	return dslp.getAllStakedAccounts(validatorAccount, ctx)
}

func (dslp *directStakedListProcessor) getAllStakedAccounts(validatorAccount state.UserAccountHandler, ctx context.Context) ([]*api.DirectStakedValue, error) {
	chLeaves := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err := validatorAccount.GetAllLeaves(chLeaves, ctx)
	if err != nil {
		return nil, err
	}

	stakedAccounts := make([]*api.DirectStakedValue, 0)
	for leaf := range chLeaves.LeavesChan {
		leafKey := leaf.Key()
		if len(leafKey) != dslp.publicKeyConverter.Len() {
			continue
		}
		if core.IsSmartContractOnMetachain(metachainIdentifier, leafKey) {
			continue
		}

		info, errGet := dslp.getValidatorInfoFromSC(leafKey)
		if errGet != nil {
			continue
		}

		baseStaked := big.NewInt(0).Set(info.totalStakedValue)
		baseStaked.Sub(baseStaked, info.topUpValue)
		encodedLeafKey := dslp.publicKeyConverter.SilentEncode(leafKey, log)

		val := &api.DirectStakedValue{
			Address:    encodedLeafKey,
			BaseStaked: baseStaked.String(),
			TopUp:      info.topUpValue.String(),
			Total:      info.totalStakedValue.String(),
		}

		stakedAccounts = append(stakedAccounts, val)
	}

	err = chLeaves.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return nil, err
	}

	if common.IsContextDone(ctx) {
		return nil, ErrTrieOperationsTimeout
	}

	return stakedAccounts, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dslp *directStakedListProcessor) IsInterfaceNil() bool {
	return dslp == nil
}
