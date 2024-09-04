package factory

import (
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/core/check"
	"github.com/DharitriOne/drt-chain-go/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieNodeFactory(t *testing.T) {
	t.Parallel()

	tnf := NewTrieNodeFactory()
	assert.False(t, check.IfNil(tnf))
}

func TestTrieNodeFactory_CreateEmpty(t *testing.T) {
	t.Parallel()

	tnf := NewTrieNodeFactory()

	emptyInterceptedNode := tnf.CreateEmpty()
	n, ok := emptyInterceptedNode.(*trie.InterceptedTrieNode)
	assert.True(t, ok)
	assert.NotNil(t, n)
}
