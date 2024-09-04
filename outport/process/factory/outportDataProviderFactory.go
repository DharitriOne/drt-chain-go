package factory

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/hashing"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/outport"
	"github.com/DharitriOne/drt-chain-go/outport/process"
	"github.com/DharitriOne/drt-chain-go/outport/process/alteredaccounts"
	"github.com/DharitriOne/drt-chain-go/outport/process/disabled"
	"github.com/DharitriOne/drt-chain-go/outport/process/transactionsfee"
	processTxs "github.com/DharitriOne/drt-chain-go/process"
	"github.com/DharitriOne/drt-chain-go/sharding"
	"github.com/DharitriOne/drt-chain-go/sharding/nodesCoordinator"
	"github.com/DharitriOne/drt-chain-go/state"
	"github.com/DharitriOne/drt-chain-go/storage"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
)

// ArgOutportDataProviderFactory holds the arguments needed for creating a new instance of outport.DataProviderOutport
type ArgOutportDataProviderFactory struct {
	IsImportDBMode         bool
	HasDrivers             bool
	AddressConverter       core.PubkeyConverter
	AccountsDB             state.AccountsAdapter
	Marshaller             marshal.Marshalizer
	DcdtDataStorageHandler vmcommon.DCDTNFTStorageHandler
	TransactionsStorer     storage.Storer
	ShardCoordinator       sharding.Coordinator
	TxCoordinator          processTxs.TransactionCoordinator
	NodesCoordinator       nodesCoordinator.NodesCoordinator
	GasConsumedProvider    process.GasConsumedProvider
	EconomicsData          process.EconomicsDataHandler
	Hasher                 hashing.Hasher
	MbsStorer              storage.Storer
	EnableEpochsHandler    common.EnableEpochsHandler
	ExecutionOrderGetter   common.ExecutionOrderGetter
}

// CreateOutportDataProvider will create a new instance of outport.DataProviderOutport
func CreateOutportDataProvider(arg ArgOutportDataProviderFactory) (outport.DataProviderOutport, error) {
	if !arg.HasDrivers {
		return disabled.NewDisabledOutportDataProvider(), nil
	}

	err := checkArgOutportDataProviderFactory(arg)
	if err != nil {
		return nil, err
	}

	alteredAccountsProvider, err := alteredaccounts.NewAlteredAccountsProvider(alteredaccounts.ArgsAlteredAccountsProvider{
		ShardCoordinator:       arg.ShardCoordinator,
		AddressConverter:       arg.AddressConverter,
		AccountsDB:             arg.AccountsDB,
		DcdtDataStorageHandler: arg.DcdtDataStorageHandler,
	})
	if err != nil {
		return nil, err
	}

	transactionsFeeProc, err := transactionsfee.NewTransactionsFeeProcessor(transactionsfee.ArgTransactionsFeeProcessor{
		Marshaller:         arg.Marshaller,
		TransactionsStorer: arg.TransactionsStorer,
		ShardCoordinator:   arg.ShardCoordinator,
		TxFeeCalculator:    arg.EconomicsData,
		PubKeyConverter:    arg.AddressConverter,
	})
	if err != nil {
		return nil, err
	}

	return process.NewOutportDataProvider(process.ArgOutportDataProvider{
		IsImportDBMode:           arg.IsImportDBMode,
		ShardCoordinator:         arg.ShardCoordinator,
		AlteredAccountsProvider:  alteredAccountsProvider,
		TransactionsFeeProcessor: transactionsFeeProc,
		TxCoordinator:            arg.TxCoordinator,
		NodesCoordinator:         arg.NodesCoordinator,
		GasConsumedProvider:      arg.GasConsumedProvider,
		EconomicsData:            arg.EconomicsData,
		ExecutionOrderHandler:    arg.ExecutionOrderGetter,
		Hasher:                   arg.Hasher,
		Marshaller:               arg.Marshaller,
	})
}
