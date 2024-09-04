package interceptorscontainer

import (
	crypto "github.com/DharitriOne/drt-chain-crypto-go"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/dataRetriever"
	"github.com/DharitriOne/drt-chain-go/heartbeat"
	"github.com/DharitriOne/drt-chain-go/process"
	"github.com/DharitriOne/drt-chain-go/sharding"
	"github.com/DharitriOne/drt-chain-go/sharding/nodesCoordinator"
	"github.com/DharitriOne/drt-chain-go/state"
)

// CommonInterceptorsContainerFactoryArgs holds the arguments needed for the metachain/shard interceptors factories
type CommonInterceptorsContainerFactoryArgs struct {
	CoreComponents               process.CoreComponentsHolder
	CryptoComponents             process.CryptoComponentsHolder
	Accounts                     state.AccountsAdapter
	ShardCoordinator             sharding.Coordinator
	NodesCoordinator             nodesCoordinator.NodesCoordinator
	MainMessenger                process.TopicHandler
	FullArchiveMessenger         process.TopicHandler
	Store                        dataRetriever.StorageService
	DataPool                     dataRetriever.PoolsHolder
	MaxTxNonceDeltaAllowed       int
	TxFeeHandler                 process.FeeHandler
	BlockBlackList               process.TimeCacher
	HeaderSigVerifier            process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier      process.HeaderIntegrityVerifier
	ValidityAttester             process.ValidityAttester
	EpochStartTrigger            process.EpochStartTriggerHandler
	WhiteListHandler             process.WhiteListHandler
	WhiteListerVerifiedTxs       process.WhiteListHandler
	AntifloodHandler             process.P2PAntifloodHandler
	ArgumentsParser              process.ArgumentsParser
	PreferredPeersHolder         process.PreferredPeersHolderHandler
	SizeCheckDelta               uint32
	RequestHandler               process.RequestHandler
	PeerSignatureHandler         crypto.PeerSignatureHandler
	SignaturesHandler            process.SignaturesHandler
	HeartbeatExpiryTimespanInSec int64
	MainPeerShardMapper          process.PeerShardMapper
	FullArchivePeerShardMapper   process.PeerShardMapper
	HardforkTrigger              heartbeat.HardforkTrigger
	NodeOperationMode            common.NodeOperation
}
