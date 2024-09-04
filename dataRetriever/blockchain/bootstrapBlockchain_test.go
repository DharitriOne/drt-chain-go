package blockchain

import (
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/core/check"
	"github.com/DharitriOne/drt-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewBootstrapBlockchain(t *testing.T) {
	t.Parallel()

	blockchain := NewBootstrapBlockchain()
	assert.False(t, check.IfNil(blockchain))
	providedHeaderHandler := &testscommon.HeaderHandlerStub{}
	assert.Nil(t, blockchain.SetCurrentBlockHeaderAndRootHash(providedHeaderHandler, nil))
	assert.Equal(t, providedHeaderHandler, blockchain.GetCurrentBlockHeader())
}
