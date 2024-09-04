package mock

import (
	"github.com/DharitriOne/drt-chain-core-go/data/batch"
	"github.com/DharitriOne/drt-chain-go/process"
)

// ChunkProcessorStub -
type ChunkProcessorStub struct {
	CheckBatchCalled func(b *batch.Batch, w process.WhiteListHandler) (process.CheckedChunkResult, error)
	CloseCalled      func() error
}

// CheckBatch -
func (c *ChunkProcessorStub) CheckBatch(b *batch.Batch, w process.WhiteListHandler) (process.CheckedChunkResult, error) {
	if c.CheckBatchCalled != nil {
		return c.CheckBatchCalled(b, w)
	}

	return process.CheckedChunkResult{}, nil
}

// Close -
func (c *ChunkProcessorStub) Close() error {
	if c.CloseCalled != nil {
		return c.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (c *ChunkProcessorStub) IsInterfaceNil() bool {
	return c == nil
}
