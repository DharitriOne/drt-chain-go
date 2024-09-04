package integrationTests

import (
	"fmt"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data/endProcess"
	"github.com/DharitriOne/drt-chain-core-go/hashing"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/integrationTests/mock"
	"github.com/DharitriOne/drt-chain-go/sharding"
	"github.com/DharitriOne/drt-chain-go/sharding/nodesCoordinator"
	"github.com/DharitriOne/drt-chain-go/storage"
	"github.com/DharitriOne/drt-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/DharitriOne/drt-chain-go/testscommon/genesisMocks"
	"github.com/DharitriOne/drt-chain-go/testscommon/nodeTypeProviderMock"
	vic "github.com/DharitriOne/drt-chain-go/testscommon/validatorInfoCacher"
)

// ArgIndexHashedNodesCoordinatorFactory -
type ArgIndexHashedNodesCoordinatorFactory struct {
	nodesPerShard           int
	nbMetaNodes             int
	shardConsensusGroupSize int
	metaConsensusGroupSize  int
	shardId                 uint32
	nbShards                int
	validatorsMap           map[uint32][]nodesCoordinator.Validator
	waitingMap              map[uint32][]nodesCoordinator.Validator
	keyIndex                int
	cp                      *CryptoParams
	epochStartSubscriber    nodesCoordinator.EpochStartEventNotifier
	hasher                  hashing.Hasher
	consensusGroupCache     nodesCoordinator.Cacher
	bootStorer              storage.Storer
}

// IndexHashedNodesCoordinatorFactory -
type IndexHashedNodesCoordinatorFactory struct {
}

// CreateNodesCoordinator -
func (tpn *IndexHashedNodesCoordinatorFactory) CreateNodesCoordinator(arg ArgIndexHashedNodesCoordinatorFactory) nodesCoordinator.NodesCoordinator {

	keys := arg.cp.NodesKeys[arg.shardId][arg.keyIndex]
	pubKeyBytes, _ := keys.MainKey.Pk.ToByteArray()

	nodeShufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           uint32(arg.nodesPerShard),
		NodesMeta:            uint32(arg.nbMetaNodes),
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler:  &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}

	nodeShuffler, _ := nodesCoordinator.NewHashValidatorsShuffler(nodeShufflerArgs)
	nodesCoordinatorRegistryFactory, _ := nodesCoordinator.NewNodesCoordinatorRegistryFactory(
		TestMarshalizer,
		StakingV4Step2EnableEpoch,
	)
	argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
		ShardConsensusGroupSize: arg.shardConsensusGroupSize,
		MetaConsensusGroupSize:  arg.metaConsensusGroupSize,
		Marshalizer:             TestMarshalizer,
		Hasher:                  arg.hasher,
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      arg.epochStartSubscriber,
		ShardIDAsObserver:       arg.shardId,
		NbShards:                uint32(arg.nbShards),
		EligibleNodes:           arg.validatorsMap,
		WaitingNodes:            arg.waitingMap,
		SelfPublicKey:           pubKeyBytes,
		ConsensusGroupCache:     arg.consensusGroupCache,
		BootStorer:              arg.bootStorer,
		ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		ChanStopNode:            endProcess.GetDummyEndProcessChannel(),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
				if flag == common.RefactorPeersMiniBlocksFlag || flag == common.StakingV4Step2Flag {
					return UnreachableEpoch
				}
				return 0
			},
		},
		ValidatorInfoCacher:             &vic.ValidatorInfoCacherStub{},
		GenesisNodesSetupHandler:        &genesisMocks.NodesSetupStub{},
		NodesCoordinatorRegistryFactory: nodesCoordinatorRegistryFactory,
	}
	nodesCoord, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
	if err != nil {
		fmt.Println("Error creating node coordinator")
	}

	return nodesCoord
}

// IndexHashedNodesCoordinatorWithRaterFactory -
type IndexHashedNodesCoordinatorWithRaterFactory struct {
	sharding.PeerAccountListAndRatingHandler
}

// CreateNodesCoordinator is used for creating a nodes coordinator in the integration tests
// based on the provided parameters
func (ihncrf *IndexHashedNodesCoordinatorWithRaterFactory) CreateNodesCoordinator(
	arg ArgIndexHashedNodesCoordinatorFactory,
) nodesCoordinator.NodesCoordinator {
	keys := arg.cp.NodesKeys[arg.shardId][arg.keyIndex]
	pubKeyBytes, _ := keys.MainKey.Pk.ToByteArray()

	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           uint32(arg.nodesPerShard),
		NodesMeta:            uint32(arg.nbMetaNodes),
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler:  &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}

	nodeShuffler, _ := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	nodesCoordinatorRegistryFactory, _ := nodesCoordinator.NewNodesCoordinatorRegistryFactory(
		TestMarshalizer,
		StakingV4Step2EnableEpoch,
	)
	argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
		ShardConsensusGroupSize: arg.shardConsensusGroupSize,
		MetaConsensusGroupSize:  arg.metaConsensusGroupSize,
		Marshalizer:             TestMarshalizer,
		Hasher:                  arg.hasher,
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      arg.epochStartSubscriber,
		ShardIDAsObserver:       arg.shardId,
		NbShards:                uint32(arg.nbShards),
		EligibleNodes:           arg.validatorsMap,
		WaitingNodes:            arg.waitingMap,
		SelfPublicKey:           pubKeyBytes,
		ConsensusGroupCache:     arg.consensusGroupCache,
		BootStorer:              arg.bootStorer,
		ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		ChanStopNode:            endProcess.GetDummyEndProcessChannel(),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
				if flag == common.RefactorPeersMiniBlocksFlag {
					return UnreachableEpoch
				}
				return 0
			},
		},
		ValidatorInfoCacher:             &vic.ValidatorInfoCacherStub{},
		GenesisNodesSetupHandler:        &genesisMocks.NodesSetupStub{},
		NodesCoordinatorRegistryFactory: nodesCoordinatorRegistryFactory,
	}

	baseCoordinator, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
	if err != nil {
		log.Debug("Error creating node coordinator")
	}

	nodesCoord, err := nodesCoordinator.NewIndexHashedNodesCoordinatorWithRater(baseCoordinator, ihncrf.PeerAccountListAndRatingHandler)
	if err != nil {
		log.Debug("Error creating node coordinator")
	}

	return &NodesWithRater{
		NodesCoordinator: nodesCoord,
		rater:            ihncrf.PeerAccountListAndRatingHandler,
	}
}

// NodesWithRater -
type NodesWithRater struct {
	nodesCoordinator.NodesCoordinator
	rater sharding.PeerAccountListAndRatingHandler
}

// IsInterfaceNil -
func (nwr *NodesWithRater) IsInterfaceNil() bool {
	return nwr == nil
}
