package outport

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data/block"
)

// BlockContainerStub -
type BlockContainerStub struct {
	GetCalled func(headerType core.HeaderType) (block.EmptyBlockCreator, error)
}

// Get -
func (bcs *BlockContainerStub) Get(headerType core.HeaderType) (block.EmptyBlockCreator, error) {
	if bcs.GetCalled != nil {
		return bcs.GetCalled(headerType)
	}

	return nil, nil
}
