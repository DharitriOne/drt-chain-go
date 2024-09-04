package factory

import (
	"fmt"
	"sync"
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-go/node/mock"
	"github.com/DharitriOne/drt-chain-go/node/trieIterators"
	"github.com/DharitriOne/drt-chain-go/testscommon"
	stateMock "github.com/DharitriOne/drt-chain-go/testscommon/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDirectStakedListHandlerHandler_Disabled(t *testing.T) {
	t.Parallel()

	args := trieIterators.ArgTrieIteratorProcessor{
		ShardID: 0,
	}

	directStakedListHandler, err := CreateDirectStakedListHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*disabled.directStakedListProcessor", fmt.Sprintf("%T", directStakedListHandler))
}

func TestCreateDirectStakedListHandlerHandler_DirectStakedListProcessor(t *testing.T) {
	t.Parallel()

	args := trieIterators.ArgTrieIteratorProcessor{
		ShardID: core.MetachainShardId,
		Accounts: &trieIterators.AccountsWrapper{
			Mutex:           &sync.Mutex{},
			AccountsAdapter: &stateMock.AccountsStub{},
		},
		PublicKeyConverter: &testscommon.PubkeyConverterMock{},
		QueryService:       &mock.SCQueryServiceStub{},
	}

	directStakedListHandler, err := CreateDirectStakedListHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*trieIterators.directStakedListProcessor", fmt.Sprintf("%T", directStakedListHandler))
}
