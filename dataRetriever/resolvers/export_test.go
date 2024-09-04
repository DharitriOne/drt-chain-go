package resolvers

import (
	"github.com/DharitriOne/drt-chain-go/dataRetriever"
	"github.com/DharitriOne/drt-chain-go/p2p"
)

// EpochHandler -
func (hdrRes *HeaderResolver) EpochHandler() dataRetriever.EpochHandler {
	return hdrRes.epochHandler
}

// ResolveMultipleHashes -
func (tnRes *TrieNodeResolver) ResolveMultipleHashes(hashesBuff []byte, message p2p.MessageP2P, source p2p.MessageHandler) error {
	return tnRes.resolveMultipleHashes(hashesBuff, message, source)
}
