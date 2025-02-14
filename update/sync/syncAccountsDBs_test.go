package sync

import (
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/data/block"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/state"
	stateMock "github.com/DharitriOne/drt-chain-go/testscommon/state"
	trieMock "github.com/DharitriOne/drt-chain-go/testscommon/trie"
	"github.com/DharitriOne/drt-chain-go/update"
	"github.com/DharitriOne/drt-chain-go/update/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSyncState_NilTrieSyncersShouldErr(t *testing.T) {
	t.Parallel()

	args := ArgsNewSyncAccountsDBsHandler{
		AccountsDBsSyncers: nil,
		ActiveAccountsDBs:  nil,
	}

	triesSyncHandler, err := NewSyncAccountsDBsHandler(args)
	require.Nil(t, triesSyncHandler)
	require.Equal(t, update.ErrNilAccountsDBSyncContainer, err)
}

func TestNewSyncState(t *testing.T) {
	t.Parallel()

	args := ArgsNewSyncAccountsDBsHandler{
		AccountsDBsSyncers: &mock.AccountsDBSyncersStub{
			GetCalled: func(key string) (syncer update.AccountsDBSyncer, err error) {
				return &mock.AccountsDBSyncerStub{}, nil
			},
		},
		ActiveAccountsDBs: make(map[state.AccountsDbIdentifier]state.AccountsAdapter),
	}

	args.ActiveAccountsDBs[state.UserAccountsState] = &stateMock.AccountsStub{
		RecreateAllTriesCalled: func(rootHash []byte) (map[string]common.Trie, error) {
			tries := make(map[string]common.Trie)
			tries[string(rootHash)] = &trieMock.TrieStub{}
			return tries, nil
		},
	}

	args.ActiveAccountsDBs[state.PeerAccountsState] = &stateMock.AccountsStub{
		RecreateAllTriesCalled: func(rootHash []byte) (map[string]common.Trie, error) {
			tries := make(map[string]common.Trie)
			tries[string(rootHash)] = &trieMock.TrieStub{}
			return tries, nil
		},
	}

	triesSyncHandler, err := NewSyncAccountsDBsHandler(args)
	require.Nil(t, err)

	metaBlock := &block.MetaBlock{
		Nonce: 1, Epoch: 1, RootHash: []byte("metaRootHash"),
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardID: 0, RootHash: []byte("shardDataRootHash")},
			},
		},
	}

	err = triesSyncHandler.SyncTriesFrom(metaBlock)
	require.Nil(t, err)

	tries, err := triesSyncHandler.GetTries()
	assert.NotNil(t, tries)
	assert.Nil(t, err)
}
