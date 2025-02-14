package mock

import (
	"github.com/DharitriOne/drt-chain-core-go/data"
	"github.com/DharitriOne/drt-chain-go/process"
)

// ValidityAttesterStub -
type ValidityAttesterStub struct {
	CheckBlockAgainstRoundHandlerCalled func(headerHandler data.HeaderHandler) error
	CheckBlockAgainstFinalCalled        func(headerHandler data.HeaderHandler) error
	CheckBlockAgainstWhitelistCalled    func(interceptedData process.InterceptedData) bool
}

// CheckBlockAgainstRoundHandler -
func (vas *ValidityAttesterStub) CheckBlockAgainstRoundHandler(headerHandler data.HeaderHandler) error {
	if vas.CheckBlockAgainstRoundHandlerCalled != nil {
		return vas.CheckBlockAgainstRoundHandlerCalled(headerHandler)
	}

	return nil
}

// CheckBlockAgainstFinal -
func (vas *ValidityAttesterStub) CheckBlockAgainstFinal(headerHandler data.HeaderHandler) error {
	if vas.CheckBlockAgainstFinalCalled != nil {
		return vas.CheckBlockAgainstFinalCalled(headerHandler)
	}

	return nil
}

// CheckBlockAgainstWhitelist -
func (vas *ValidityAttesterStub) CheckBlockAgainstWhitelist(interceptedData process.InterceptedData) bool {
	if vas.CheckBlockAgainstWhitelistCalled != nil {
		return vas.CheckBlockAgainstWhitelistCalled(interceptedData)
	}

	return false
}

// IsInterfaceNil -
func (vas *ValidityAttesterStub) IsInterfaceNil() bool {
	return vas == nil
}
