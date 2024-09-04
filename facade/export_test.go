package facade

import (
	"github.com/DharitriOne/drt-chain-go/ntp"
)

// GetSyncer returns the current syncer
func (nf *nodeFacade) GetSyncer() ntp.SyncTimer {
	return nf.syncer
}
