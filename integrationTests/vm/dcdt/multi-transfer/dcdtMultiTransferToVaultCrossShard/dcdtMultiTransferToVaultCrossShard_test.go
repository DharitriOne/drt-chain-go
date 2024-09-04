package dcdtMultiTransferToVaultCrossShard

import (
	"testing"

	multitransfer "github.com/DharitriOne/drt-chain-go/integrationTests/vm/dcdt/multi-transfer"
)

func TestDCDTMultiTransferToVaultCrossShard(t *testing.T) {
	multitransfer.DcdtMultiTransferToVault(t, true, "../../testdata/vaultV2.wasm")
}
