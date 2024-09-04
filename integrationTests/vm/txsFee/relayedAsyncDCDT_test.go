package txsFee

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/integrationTests"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestRelayedAsyncDCDTCallShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	})
	require.Nil(t, err)
	defer testContext.Close()

	localRewaBalance := big.NewInt(100000000)
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, localRewaBalance)

	// create an address with DCDT token
	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789012")

	localDcdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithDCDTBalance(t, testContext.Accounts, sndAddr, big.NewInt(0), token, 0, localDcdtBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, localRewaBalance)

	// deploy 2 contracts
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(50000)

	argsSecond := [][]byte{[]byte(hex.EncodeToString(token))}
	secondSCAddress := utils.DoDeploySecond(t, testContext, "../dcdt/testdata/second-contract.wasm", ownerAccount, gasPrice, deployGasLimit, argsSecond, big.NewInt(0))

	args := [][]byte{[]byte(hex.EncodeToString(token)), []byte(hex.EncodeToString(secondSCAddress))}
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	firstSCAddress := utils.DoDeploySecond(t, testContext, "../dcdt/testdata/first-contract.wasm", ownerAccount, gasPrice, deployGasLimit, args, big.NewInt(0))

	testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	gasLimit := uint64(500000)
	innerTx := utils.CreateDCDTTransferTx(0, sndAddr, firstSCAddress, token, big.NewInt(5000), gasPrice, gasLimit)
	innerTx.Data = []byte(string(innerTx.Data) + "@" + hex.EncodeToString([]byte("transferToSecondContractHalf")))

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckDCDTBalance(t, testContext, firstSCAddress, token, big.NewInt(2500))
	utils.CheckDCDTBalance(t, testContext, secondSCAddress, token, big.NewInt(2500))

	expectedSenderBalance := big.NewInt(94996430)
	utils.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedSenderBalance)

	expectedAccumulatedFees := big.NewInt(5003570)
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, expectedAccumulatedFees, accumulatedFees)
}

func TestRelayedAsyncDCDTCall_InvalidCallFirstContract(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	localRewaBalance := big.NewInt(100000000)
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, localRewaBalance)

	// create an address with DCDT token
	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789012")

	localDcdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithDCDTBalance(t, testContext.Accounts, sndAddr, big.NewInt(0), token, 0, localDcdtBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, localRewaBalance)

	// deploy 2 contracts
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(50000)

	argsSecond := [][]byte{[]byte(hex.EncodeToString(token))}
	secondSCAddress := utils.DoDeploySecond(t, testContext, "../dcdt/testdata/second-contract.wasm", ownerAccount, gasPrice, deployGasLimit, argsSecond, big.NewInt(0))

	args := [][]byte{[]byte(hex.EncodeToString(token)), []byte(hex.EncodeToString(secondSCAddress))}
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	firstSCAddress := utils.DoDeploySecond(t, testContext, "../dcdt/testdata/first-contract.wasm", ownerAccount, gasPrice, deployGasLimit, args, big.NewInt(0))

	testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	gasLimit := uint64(500000)
	innerTx := utils.CreateDCDTTransferTx(0, sndAddr, firstSCAddress, token, big.NewInt(5000), gasPrice, gasLimit)
	innerTx.Data = []byte(string(innerTx.Data) + "@" + hex.EncodeToString([]byte("transferToSecondContractRejected")))

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckDCDTBalance(t, testContext, firstSCAddress, token, big.NewInt(5000))
	utils.CheckDCDTBalance(t, testContext, secondSCAddress, token, big.NewInt(0))

	expectedSenderBalance := big.NewInt(95996260)
	utils.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedSenderBalance)

	expectedAccumulatedFees := big.NewInt(4003740)
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, expectedAccumulatedFees, accumulatedFees)
}

func TestRelayedAsyncDCDTCall_InvalidOutOfGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	localRewaBalance := big.NewInt(100000000)
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, localRewaBalance)

	// create an address with DCDT token
	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789012")

	localDcdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithDCDTBalance(t, testContext.Accounts, sndAddr, big.NewInt(0), token, 0, localDcdtBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, localRewaBalance)

	// deploy 2 contracts
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(50000)

	argsSecond := [][]byte{[]byte(hex.EncodeToString(token))}
	secondSCAddress := utils.DoDeploySecond(t, testContext, "../dcdt/testdata/second-contract.wasm", ownerAccount, gasPrice, deployGasLimit, argsSecond, big.NewInt(0))

	args := [][]byte{[]byte(hex.EncodeToString(token)), []byte(hex.EncodeToString(secondSCAddress))}
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	firstSCAddress := utils.DoDeploySecond(t, testContext, "../dcdt/testdata/first-contract.wasm", ownerAccount, gasPrice, deployGasLimit, args, big.NewInt(0))

	testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	gasLimit := uint64(2000)
	innerTx := utils.CreateDCDTTransferTx(0, sndAddr, firstSCAddress, token, big.NewInt(5000), gasPrice, gasLimit)
	innerTx.Data = []byte(string(innerTx.Data) + "@" + hex.EncodeToString([]byte("transferToSecondContractHalf")))

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	testContext.TxsLogsProcessor.Clean()
	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckDCDTBalance(t, testContext, firstSCAddress, token, big.NewInt(0))
	utils.CheckDCDTBalance(t, testContext, secondSCAddress, token, big.NewInt(0))

	expectedSenderBalance := big.NewInt(99976450)
	utils.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedSenderBalance)

	expectedAccumulatedFees := big.NewInt(23550)
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, expectedAccumulatedFees, accumulatedFees)
}
