package storage

import (
	"github.com/DharitriOne/drt-chain-go/common/statistics/disabled"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/dataRetriever"
	"github.com/DharitriOne/drt-chain-go/genesis/mock"
	"github.com/DharitriOne/drt-chain-go/testscommon"
	"github.com/DharitriOne/drt-chain-go/testscommon/hashingMocks"
	"github.com/DharitriOne/drt-chain-go/trie"
)

// GetStorageManagerArgs returns mock args for trie storage manager creation
func GetStorageManagerArgs() trie.NewTrieStorageManagerArgs {
	return trie.NewTrieStorageManagerArgs{
		MainStorer:  testscommon.NewSnapshotPruningStorerMock(),
		Marshalizer: &mock.MarshalizerMock{},
		Hasher:      &hashingMocks.HasherMock{},
		GeneralConfig: config.TrieStorageManagerConfig{
			PruningBufferLen:      1000,
			SnapshotsBufferLen:    10,
			SnapshotsGoroutineNum: 2,
		},
		IdleProvider:   &testscommon.ProcessStatusHandlerStub{},
		Identifier:     dataRetriever.UserAccountsUnit.String(),
		StatsCollector: disabled.NewStateStatistics(),
	}
}

// GetStorageManagerOptions returns default options for trie storage manager creation
func GetStorageManagerOptions() trie.StorageManagerOptions {
	return trie.StorageManagerOptions{
		PruningEnabled:   true,
		SnapshotsEnabled: true,
	}
}
