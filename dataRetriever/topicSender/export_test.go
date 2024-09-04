package topicsender

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
)

func MakeDiffList(
	allConnectedPeers []core.PeerID,
	excludedConnectedPeers []core.PeerID,
) []core.PeerID {

	return makeDiffList(allConnectedPeers, excludedConnectedPeers)
}

func (dplc *diffPeerListCreator) MainTopic() string {
	return dplc.mainTopic
}

func (dplc *diffPeerListCreator) ExcludedPeersOnTopic() string {
	return dplc.excludePeersFromTopic
}
