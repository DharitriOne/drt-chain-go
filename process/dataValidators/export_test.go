package dataValidators

import (
	"github.com/DharitriOne/drt-chain-go/process"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
)

// CheckAccount -
func (txv *txValidator) CheckAccount(
	interceptedTx process.InterceptedTransactionHandler,
	accountHandler vmcommon.AccountHandler,
) error {
	return txv.checkAccount(interceptedTx, accountHandler)
}

// GetTxData -
func GetTxData(interceptedTx process.InterceptedTransactionHandler) ([]byte, error) {
	return getTxData(interceptedTx)
}
