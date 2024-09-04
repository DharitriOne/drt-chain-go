package factory

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/core/check"
	"github.com/DharitriOne/drt-chain-core-go/data/typeConverters"
	"github.com/DharitriOne/drt-chain-core-go/hashing"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/dataRetriever"
	"github.com/DharitriOne/drt-chain-go/dblookupext"
	"github.com/DharitriOne/drt-chain-go/dblookupext/disabled"
	"github.com/DharitriOne/drt-chain-go/dblookupext/dcdtSupply"
	"github.com/DharitriOne/drt-chain-go/process"
)

// ArgsHistoryRepositoryFactory holds all dependencies required by the history processor factory in order to create
// new instances
type ArgsHistoryRepositoryFactory struct {
	SelfShardID              uint32
	Config                   config.DbLookupExtensionsConfig
	Store                    dataRetriever.StorageService
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

type historyRepositoryFactory struct {
	selfShardID              uint32
	dbLookupExtensionsConfig config.DbLookupExtensionsConfig
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	uInt64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

// NewHistoryRepositoryFactory creates an instance of historyRepositoryFactory
func NewHistoryRepositoryFactory(args *ArgsHistoryRepositoryFactory) (dblookupext.HistoryRepositoryFactory, error) {
	if check.IfNil(args.Marshalizer) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, core.ErrNilHasher
	}
	if check.IfNil(args.Store) {
		return nil, core.ErrNilStore
	}
	if check.IfNil(args.Uint64ByteSliceConverter) {
		return nil, process.ErrNilUint64Converter
	}

	return &historyRepositoryFactory{
		selfShardID:              args.SelfShardID,
		dbLookupExtensionsConfig: args.Config,
		store:                    args.Store,
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		uInt64ByteSliceConverter: args.Uint64ByteSliceConverter,
	}, nil
}

// Create creates instances of HistoryRepository
func (hpf *historyRepositoryFactory) Create() (dblookupext.HistoryRepository, error) {
	if !hpf.dbLookupExtensionsConfig.Enabled {
		return disabled.NewNilHistoryRepository()
	}

	dcdtSuppliesStorer, err := hpf.store.GetStorer(dataRetriever.DCDTSuppliesUnit)
	if err != nil {
		return nil, err
	}

	txLogsStorer, err := hpf.store.GetStorer(dataRetriever.TxLogsUnit)
	if err != nil {
		return nil, err
	}

	dcdtSuppliesHandler, err := dcdtSupply.NewSuppliesProcessor(
		hpf.marshalizer,
		dcdtSuppliesStorer,
		txLogsStorer,
	)
	if err != nil {
		return nil, err
	}

	roundHdrHashDataStorer, err := hpf.store.GetStorer(dataRetriever.RoundHdrHashDataUnit)
	if err != nil {
		return nil, err
	}

	miniblocksMetadataStorer, err := hpf.store.GetStorer(dataRetriever.MiniblocksMetadataUnit)
	if err != nil {
		return nil, err
	}

	epochByHashStorer, err := hpf.store.GetStorer(dataRetriever.EpochByHashUnit)
	if err != nil {
		return nil, err
	}

	miniblockHashByTxHashStorer, err := hpf.store.GetStorer(dataRetriever.MiniblockHashByTxHashUnit)
	if err != nil {
		return nil, err
	}

	resultsHashesByTxHashStorer, err := hpf.store.GetStorer(dataRetriever.ResultsHashesByTxHashUnit)
	if err != nil {
		return nil, err
	}

	historyRepArgs := dblookupext.HistoryRepositoryArguments{
		SelfShardID:                 hpf.selfShardID,
		Hasher:                      hpf.hasher,
		Marshalizer:                 hpf.marshalizer,
		BlockHashByRound:            roundHdrHashDataStorer,
		Uint64ByteSliceConverter:    hpf.uInt64ByteSliceConverter,
		MiniblocksMetadataStorer:    miniblocksMetadataStorer,
		EpochByHashStorer:           epochByHashStorer,
		MiniblockHashByTxHashStorer: miniblockHashByTxHashStorer,
		EventsHashesByTxHashStorer:  resultsHashesByTxHashStorer,
		DCDTSuppliesHandler:         dcdtSuppliesHandler,
	}
	return dblookupext.NewHistoryRepository(historyRepArgs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hpf *historyRepositoryFactory) IsInterfaceNil() bool {
	return hpf == nil
}
