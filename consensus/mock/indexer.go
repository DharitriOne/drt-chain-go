package mock

import "github.com/DharitriOne/drt-chain-core-go/data/outport"

// IndexerMock is a mock implementation fot the Indexer interface
type IndexerMock struct{}

// SaveRoundsInfo -
func (im *IndexerMock) SaveRoundsInfo(_ []*outport.RoundInfo) {
}

// IsInterfaceNil -
func (im *IndexerMock) IsInterfaceNil() bool {
	return false
}
