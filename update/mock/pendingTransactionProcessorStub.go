package mock

import (
	"github.com/DharitriOne/drt-chain-core-go/data/block"
	"github.com/DharitriOne/drt-chain-go/update"
)

// PendingTransactionProcessorStub -
type PendingTransactionProcessorStub struct {
	ProcessTransactionsDstMeCalled func(mbInfo *update.MbInfo) (*block.MiniBlock, error)
	RootHashCalled                 func() ([]byte, error)
	CommitCalled                   func() ([]byte, error)
}

// ProcessTransactionsDstMe -
func (ptps *PendingTransactionProcessorStub) ProcessTransactionsDstMe(mbInfo *update.MbInfo) (*block.MiniBlock, error) {
	if ptps.ProcessTransactionsDstMeCalled != nil {
		return ptps.ProcessTransactionsDstMeCalled(mbInfo)
	}
	return nil, nil
}

// RootHash -
func (ptps *PendingTransactionProcessorStub) RootHash() ([]byte, error) {
	if ptps.RootHashCalled != nil {
		return ptps.RootHashCalled()
	}
	return nil, nil
}

// Commit -
func (ptps *PendingTransactionProcessorStub) Commit() ([]byte, error) {
	if ptps.CommitCalled != nil {
		return ptps.CommitCalled()
	}
	return nil, nil
}

// IsInterfaceNil -
func (ptps *PendingTransactionProcessorStub) IsInterfaceNil() bool {
	return ptps == nil
}
