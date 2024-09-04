package storagerequesterscontainer

import (
	"github.com/DharitriOne/drt-chain-core-go/data/endProcess"
	"github.com/DharitriOne/drt-chain-core-go/data/typeConverters"
	"github.com/DharitriOne/drt-chain-core-go/hashing"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/dataRetriever"
	"github.com/DharitriOne/drt-chain-go/p2p"
	"github.com/DharitriOne/drt-chain-go/sharding"
)

// FactoryArgs will hold the arguments for RequestersContainerFactory for both shard and meta
type FactoryArgs struct {
	GeneralConfig            config.Config
	ShardIDForTries          uint32
	ChainID                  string
	WorkingDirectory         string
	Hasher                   hashing.Hasher
	ShardCoordinator         sharding.Coordinator
	Messenger                p2p.Messenger
	Store                    dataRetriever.StorageService
	Marshalizer              marshal.Marshalizer
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	DataPacker               dataRetriever.DataPacker
	ManualEpochStartNotifier dataRetriever.ManualEpochStartNotifier
	ChanGracefullyClose      chan endProcess.ArgEndProcess
	EnableEpochsHandler      common.EnableEpochsHandler
	StateStatsHandler        common.StateStatisticsHandler
}
