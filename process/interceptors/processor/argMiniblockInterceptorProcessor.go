package processor

import (
	"github.com/DharitriOne/drt-chain-core-go/hashing"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/DharitriOne/drt-chain-go/process"
	"github.com/DharitriOne/drt-chain-go/sharding"
	"github.com/DharitriOne/drt-chain-go/storage"
)

// ArgMiniblockInterceptorProcessor is the argument for the interceptor processor used for miniblocks
type ArgMiniblockInterceptorProcessor struct {
	MiniblockCache   storage.Cacher
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
	WhiteListHandler process.WhiteListHandler
}
