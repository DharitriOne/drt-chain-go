package mock

import (
	"github.com/DharitriOne/drt-chain-core-go/data"
	"github.com/DharitriOne/drt-chain-core-go/data/block"
)

// ValidatorInfoSyncerStub -
type ValidatorInfoSyncerStub struct {
}

// SyncMiniBlocks -
func (vip *ValidatorInfoSyncerStub) SyncMiniBlocks(*block.MetaBlock) ([][]byte, data.BodyHandler, error) {
	return nil, nil, nil
}

// IsInterfaceNil -
func (vip *ValidatorInfoSyncerStub) IsInterfaceNil() bool {
	return vip == nil
}
