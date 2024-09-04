package txsFee

import (
	"math/big"
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/DharitriOne/drt-chain-go/process"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestDCDTLocalMintShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")

	rewaBalance := big.NewInt(100000000)
	dcdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	roles := [][]byte{[]byte(core.DCDTRoleLocalMint), []byte(core.DCDTRoleLocalBurn)}
	utils.CreateAccountWithDCDTBalanceAndRoles(t, testContext.Accounts, sndAddr, rewaBalance, token, 0, dcdtBalance, roles)

	gasLimit := uint64(40)
	tx := utils.CreateDCDTLocalMintTx(0, sndAddr, sndAddr, token, big.NewInt(100), gasPrice, gasLimit)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalanceSnd := big.NewInt(100000100)
	utils.CheckDCDTBalance(t, testContext, sndAddr, token, expectedBalanceSnd)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(370), accumulatedFees)
}

func TestDCDTLocalMintNotAllowedShouldErr(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")

	rewaBalance := big.NewInt(100000000)
	dcdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithDCDTBalance(t, testContext.Accounts, sndAddr, rewaBalance, token, 0, dcdtBalance)

	gasLimit := uint64(40)
	tx := utils.CreateDCDTLocalMintTx(0, sndAddr, sndAddr, token, big.NewInt(100), gasPrice, gasLimit)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, process.ErrFailedTransaction, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalanceSnd := big.NewInt(100000000)
	utils.CheckDCDTBalance(t, testContext, sndAddr, token, expectedBalanceSnd)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(400), accumulatedFees)
}
