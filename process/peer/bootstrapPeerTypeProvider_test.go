package peer

import (
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/core/check"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/state"
	"github.com/stretchr/testify/assert"
)

func TestNewBootstrapPeerTypeProvider(t *testing.T) {
	t.Parallel()

	peerTypeProvider := NewBootstrapPeerTypeProvider()
	assert.False(t, check.IfNil(peerTypeProvider))
	assert.Equal(t, make([]*state.PeerTypeInfo, 0), peerTypeProvider.GetAllPeerTypeInfos())
	peerType, shard, err := peerTypeProvider.ComputeForPubKey(nil)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), shard)
	assert.Equal(t, common.ObserverList, peerType)
}
