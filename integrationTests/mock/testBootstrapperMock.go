package mock

import (
	"github.com/DharitriOne/drt-chain-go/process"
	"github.com/DharitriOne/drt-chain-go/process/sync/disabled"
)

type testBootstrapperMock struct {
	process.Bootstrapper
}

// NewTestBootstrapperMock -
func NewTestBootstrapperMock() *testBootstrapperMock {
	return &testBootstrapperMock{
		Bootstrapper: disabled.NewDisabledBootstrapper(),
	}
}

// RollBack -
func (tbm *testBootstrapperMock) RollBack(_ bool) error {
	return nil
}

// SetProbableHighestNonce -
func (tbm *testBootstrapperMock) SetProbableHighestNonce(_ uint64) {
}
