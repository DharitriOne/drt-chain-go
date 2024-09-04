package transactionsfee

import (
	"math/big"

	"github.com/DharitriOne/drt-chain-core-go/data"
	"github.com/DharitriOne/drt-chain-core-go/data/transaction"
	datafield "github.com/DharitriOne/drt-chain-vm-common-go/parsers/dataField"
)

// FeesProcessorHandler defines the interface for the transaction fees processor
type FeesProcessorHandler interface {
	ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64
	IsInterfaceNil() bool
}

type transactionGetter interface {
	GetTxByHash(txHash []byte) (*transaction.Transaction, error)
}

type dataFieldParser interface {
	Parse(dataField []byte, sender, receiver []byte, numOfShards uint32) *datafield.ResponseParseData
}
