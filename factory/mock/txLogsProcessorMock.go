package mock

import (
	"github.com/DharitriOne/drt-chain-core-go/data"
	"github.com/DharitriOne/drt-chain-go/process"
)

var _ process.TransactionLogProcessorDatabase = (*TxLogProcessorMock)(nil)

// TxLogProcessorMock -
type TxLogProcessorMock struct {
}

// GetLogFromCache -
func (t *TxLogProcessorMock) GetLogFromCache(_ []byte) (*data.LogData, bool) {
	return &data.LogData{}, false
}

// EnableLogToBeSavedInCache -
func (t *TxLogProcessorMock) EnableLogToBeSavedInCache() {
}

// Clean -
func (t *TxLogProcessorMock) Clean() {
}

// IsInterfaceNil -
func (t *TxLogProcessorMock) IsInterfaceNil() bool {
	return t == nil
}
