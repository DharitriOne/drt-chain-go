package nodesCoordinator

import (
	"github.com/DharitriOne/drt-chain-core-go/data/endProcess"
	"github.com/DharitriOne/drt-chain-core-go/hashing"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/epochStart"
	"github.com/DharitriOne/drt-chain-go/storage"
)

// ArgNodesCoordinator holds all dependencies required by the nodes coordinator in order to create new instances
type ArgNodesCoordinator struct {
	ShardConsensusGroupSize         int
	MetaConsensusGroupSize          int
	Marshalizer                     marshal.Marshalizer
	Hasher                          hashing.Hasher
	Shuffler                        NodesShuffler
	EpochStartNotifier              EpochStartEventNotifier
	BootStorer                      storage.Storer
	ShardIDAsObserver               uint32
	NbShards                        uint32
	EligibleNodes                   map[uint32][]Validator
	WaitingNodes                    map[uint32][]Validator
	SelfPublicKey                   []byte
	Epoch                           uint32
	StartEpoch                      uint32
	ConsensusGroupCache             Cacher
	ShuffledOutHandler              ShuffledOutHandler
	ChanStopNode                    chan endProcess.ArgEndProcess
	NodeTypeProvider                NodeTypeProviderHandler
	IsFullArchive                   bool
	EnableEpochsHandler             common.EnableEpochsHandler
	ValidatorInfoCacher             epochStart.ValidatorInfoCacher
	GenesisNodesSetupHandler        GenesisNodesSetupHandler
	NodesCoordinatorRegistryFactory NodesCoordinatorRegistryFactory
}
