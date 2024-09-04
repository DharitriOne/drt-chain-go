package components

import (
	"time"

	crypto "github.com/DharitriOne/drt-chain-crypto-go"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/dataRetriever"
	"github.com/DharitriOne/drt-chain-go/factory/mock"
	"github.com/DharitriOne/drt-chain-go/sharding"
	"github.com/DharitriOne/drt-chain-go/testscommon"
	"github.com/DharitriOne/drt-chain-go/testscommon/consensus"
	"github.com/DharitriOne/drt-chain-go/testscommon/cryptoMocks"
	dataRetrieverTests "github.com/DharitriOne/drt-chain-go/testscommon/dataRetriever"
	"github.com/DharitriOne/drt-chain-go/testscommon/economicsmocks"
	"github.com/DharitriOne/drt-chain-go/testscommon/enableEpochsHandlerMock"
	epochNotifierMock "github.com/DharitriOne/drt-chain-go/testscommon/epochNotifier"
	"github.com/DharitriOne/drt-chain-go/testscommon/factory"
	"github.com/DharitriOne/drt-chain-go/testscommon/genesisMocks"
	"github.com/DharitriOne/drt-chain-go/testscommon/marshallerMock"
	"github.com/DharitriOne/drt-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/DharitriOne/drt-chain-go/testscommon/p2pmocks"
	"github.com/DharitriOne/drt-chain-go/testscommon/shardingMocks"
	"github.com/DharitriOne/drt-chain-go/testscommon/stakingcommon"
	stateMock "github.com/DharitriOne/drt-chain-go/testscommon/state"
	"github.com/DharitriOne/drt-chain-go/testscommon/storage"
	"github.com/DharitriOne/drt-chain-go/testscommon/storageManager"
	trieMock "github.com/DharitriOne/drt-chain-go/testscommon/trie"
)

// GetDefaultCoreComponents -
func GetDefaultCoreComponents() *mock.CoreComponentsMock {
	return &mock.CoreComponentsMock{
		IntMarsh:            &marshallerMock.MarshalizerMock{},
		TxMarsh:             &marshallerMock.MarshalizerMock{},
		VmMarsh:             &marshallerMock.MarshalizerMock{},
		Hash:                &testscommon.HasherStub{},
		UInt64ByteSliceConv: testscommon.NewNonceHashConverterMock(),
		AddrPubKeyConv:      testscommon.NewPubkeyConverterMock(32),
		ValPubKeyConv:       testscommon.NewPubkeyConverterMock(32),
		PathHdl:             &testscommon.PathManagerStub{},
		ChainIdCalled: func() string {
			return "chainID"
		},
		MinTransactionVersionCalled: func() uint32 {
			return 1
		},
		WatchdogTimer:            &testscommon.WatchdogMock{},
		AlarmSch:                 &testscommon.AlarmSchedulerStub{},
		NtpSyncTimer:             &testscommon.SyncTimerStub{},
		RoundHandlerField:        &testscommon.RoundHandlerMock{},
		EconomicsHandler:         &economicsmocks.EconomicsHandlerStub{},
		RatingsConfig:            &testscommon.RatingsInfoMock{},
		RatingHandler:            &testscommon.RaterMock{},
		NodesConfig:              &genesisMocks.NodesSetupStub{},
		StartTime:                time.Time{},
		NodeTypeProviderField:    &nodeTypeProviderMock.NodeTypeProviderStub{},
		EpochChangeNotifier:      &epochNotifierMock.EpochNotifierStub{},
		EnableEpochsHandlerField: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
}

// GetDefaultCryptoComponents -
func GetDefaultCryptoComponents() *mock.CryptoComponentsMock {
	return &mock.CryptoComponentsMock{
		PubKey:                  &mock.PublicKeyMock{},
		PrivKey:                 &mock.PrivateKeyStub{},
		P2pPubKey:               &mock.PublicKeyMock{},
		P2pPrivKey:              mock.NewP2pPrivateKeyMock(),
		P2pSig:                  &mock.SinglesignMock{},
		PubKeyString:            "pubKey",
		PubKeyBytes:             []byte("pubKey"),
		BlockSig:                &mock.SinglesignMock{},
		TxSig:                   &mock.SinglesignMock{},
		MultiSigContainer:       cryptoMocks.NewMultiSignerContainerMock(&cryptoMocks.MultisignerMock{}),
		PeerSignHandler:         &mock.PeerSignatureHandler{},
		BlKeyGen:                &mock.KeyGenMock{},
		TxKeyGen:                &mock.KeyGenMock{},
		P2PKeyGen:               &mock.KeyGenMock{},
		MsgSigVerifier:          &testscommon.MessageSignVerifierMock{},
		SigHandler:              &consensus.SigningHandlerStub{},
		ManagedPeersHolderField: &testscommon.ManagedPeersHolderStub{},
	}
}

// GetDefaultNetworkComponents -
func GetDefaultNetworkComponents() *mock.NetworkComponentsMock {
	return &mock.NetworkComponentsMock{
		Messenger:       &p2pmocks.MessengerStub{},
		InputAntiFlood:  &mock.P2PAntifloodHandlerStub{},
		OutputAntiFlood: &mock.P2PAntifloodHandlerStub{},
		PeerBlackList:   &mock.PeerBlackListHandlerStub{},
	}
}

// GetDefaultStateComponents -
func GetDefaultStateComponents() *factory.StateComponentsMock {
	return &factory.StateComponentsMock{
		PeersAcc: &stateMock.AccountsStub{},
		Accounts: &stateMock.AccountsStub{},
		Tries:    &trieMock.TriesHolderStub{},
		StorageManagers: map[string]common.StorageManager{
			"0":                                     &storageManager.StorageManagerStub{},
			dataRetriever.UserAccountsUnit.String(): &storageManager.StorageManagerStub{},
			dataRetriever.PeerAccountsUnit.String(): &storageManager.StorageManagerStub{},
		},
		MissingNodesNotifier: &testscommon.MissingTrieNodesNotifierStub{},
	}
}

// GetDefaultDataComponents -
func GetDefaultDataComponents() *mock.DataComponentsMock {
	return &mock.DataComponentsMock{
		Blkc:              &testscommon.ChainHandlerStub{},
		Storage:           &storage.ChainStorerStub{},
		DataPool:          &dataRetrieverTests.PoolsHolderMock{},
		MiniBlockProvider: &mock.MiniBlocksProviderStub{},
	}
}

// GetDefaultProcessComponents -
func GetDefaultProcessComponents(shardCoordinator sharding.Coordinator) *mock.ProcessComponentsMock {
	return &mock.ProcessComponentsMock{
		NodesCoord:               &shardingMocks.NodesCoordinatorMock{},
		ShardCoord:               shardCoordinator,
		IntContainer:             &testscommon.InterceptorsContainerStub{},
		ResContainer:             &dataRetrieverTests.ResolversContainerStub{},
		ReqFinder:                &dataRetrieverTests.RequestersFinderStub{},
		RoundHandlerField:        &testscommon.RoundHandlerMock{},
		EpochTrigger:             &testscommon.EpochStartTriggerStub{},
		EpochNotifier:            &mock.EpochStartNotifierStub{},
		ForkDetect:               &mock.ForkDetectorMock{},
		BlockProcess:             &testscommon.BlockProcessorStub{},
		BlackListHdl:             &testscommon.TimeCacheStub{},
		BootSore:                 &mock.BootstrapStorerMock{},
		HeaderSigVerif:           &mock.HeaderSigVerifierStub{},
		HeaderIntegrVerif:        &mock.HeaderIntegrityVerifierStub{},
		ValidatorStatistics:      &testscommon.ValidatorStatisticsProcessorStub{},
		ValidatorProvider:        &stakingcommon.ValidatorsProviderStub{},
		BlockTrack:               &mock.BlockTrackerStub{},
		PendingMiniBlocksHdl:     &mock.PendingMiniBlocksHandlerStub{},
		ReqHandler:               &testscommon.RequestHandlerStub{},
		TxLogsProcess:            &mock.TxLogProcessorMock{},
		HeaderConstructValidator: &mock.HeaderValidatorStub{},
		MainPeerMapper:           &p2pmocks.NetworkShardingCollectorStub{},
		FullArchivePeerMapper:    &p2pmocks.NetworkShardingCollectorStub{},
		FallbackHdrValidator:     &testscommon.FallBackHeaderValidatorStub{},
		NodeRedundancyHandlerInternal: &mock.RedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return false
			},
			IsMainMachineActiveCalled: func() bool {
				return false
			},
			ObserverPrivateKeyCalled: func() crypto.PrivateKey {
				return &mock.PrivateKeyStub{}
			},
		},
		HardforkTriggerField: &testscommon.HardforkTriggerStub{},
	}
}
