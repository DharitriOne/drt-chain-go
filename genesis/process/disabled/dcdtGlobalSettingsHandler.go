package disabled

// DCDTGlobalSettingsHandler implements the DCDTGlobalSettingsHandler interface but does nothing as it is disabled
type DCDTGlobalSettingsHandler struct {
}

// IsPaused is disabled
func (e *DCDTGlobalSettingsHandler) IsPaused(_ []byte) bool {
	return false
}

// IsLimitedTransfer is disabled
func (e *DCDTGlobalSettingsHandler) IsLimitedTransfer(_ []byte) bool {
	return false
}

// IsInterfaceNil return true if underlying object is nil
func (e *DCDTGlobalSettingsHandler) IsInterfaceNil() bool {
	return e == nil
}
