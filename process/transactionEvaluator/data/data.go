package data

import (
	"github.com/DharitriOne/drt-chain-core-go/data/transaction"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
)

// SimulationResultsWithVMOutput is the data transfer object which will hold results for simulation a transaction's execution
type SimulationResultsWithVMOutput struct {
	transaction.SimulationResults
	VMOutput *vmcommon.VMOutput `json:"-"`
}
