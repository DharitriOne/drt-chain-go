package helpers

import (
	"github.com/DharitriOne/drt-chain-core-go/data"
	"github.com/DharitriOne/drt-chain-go/common"
)

// ComputeRandomnessForTxSorting returns the randomness for transactions sorting
func ComputeRandomnessForTxSorting(header data.HeaderHandler, enableEpochsHandler common.EnableEpochsHandler) []byte {
	if enableEpochsHandler.IsFlagEnabled(common.CurrentRandomnessOnSortingFlag) {
		return header.GetRandSeed()
	}

	return header.GetPrevRandSeed()
}
