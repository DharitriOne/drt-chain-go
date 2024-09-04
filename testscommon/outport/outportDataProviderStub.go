package outport

import (
	outportcore "github.com/DharitriOne/drt-chain-core-go/data/outport"
	"github.com/DharitriOne/drt-chain-go/outport/process"
)

// OutportDataProviderStub -
type OutportDataProviderStub struct {
	PrepareOutportSaveBlockDataCalled func(
		arg process.ArgPrepareOutportSaveBlockData,
	) (*outportcore.OutportBlockWithHeaderAndBody, error)
}

// PrepareOutportSaveBlockData -
func (a *OutportDataProviderStub) PrepareOutportSaveBlockData(
	arg process.ArgPrepareOutportSaveBlockData,
) (*outportcore.OutportBlockWithHeaderAndBody, error) {
	if a.PrepareOutportSaveBlockDataCalled != nil {
		return a.PrepareOutportSaveBlockDataCalled(arg)
	}

	return nil, nil
}

// IsInterfaceNil -
func (a *OutportDataProviderStub) IsInterfaceNil() bool {
	return a == nil
}
