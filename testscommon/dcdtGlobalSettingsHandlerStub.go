package testscommon

// DCDTGlobalSettingsHandlerStub -
type DCDTGlobalSettingsHandlerStub struct {
	IsPausedCalled          func(dcdtTokenKey []byte) bool
	IsLimitedTransferCalled func(dcdtTokenKey []byte) bool
}

// IsPaused -
func (e *DCDTGlobalSettingsHandlerStub) IsPaused(dcdtTokenKey []byte) bool {
	if e.IsPausedCalled != nil {
		return e.IsPausedCalled(dcdtTokenKey)
	}
	return false
}

// IsLimitedTransfer -
func (e *DCDTGlobalSettingsHandlerStub) IsLimitedTransfer(dcdtTokenKey []byte) bool {
	if e.IsLimitedTransferCalled != nil {
		return e.IsLimitedTransferCalled(dcdtTokenKey)
	}
	return false
}

// IsInterfaceNil -
func (e *DCDTGlobalSettingsHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
