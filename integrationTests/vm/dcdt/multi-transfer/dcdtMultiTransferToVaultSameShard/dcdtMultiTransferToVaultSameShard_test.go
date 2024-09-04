package dcdtMultiTransferToVaultSameShard

import (
	"testing"

	multitransfer "github.com/DharitriOne/drt-chain-go/integrationTests/vm/dcdt/multi-transfer"
)

func TestDCDTMultiTransferToVaultSameShard(t *testing.T) {
	multitransfer.DcdtMultiTransferToVault(t, false, "../../testdata/vaultV2.wasm")
}
