package txsFee

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data/smartContractResult"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/DharitriOne/drt-chain-go/process"
	"github.com/DharitriOne/drt-chain-go/sharding"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestDCDTTransferShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	rewaBalance := big.NewInt(100000000)
	dcdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithDCDTBalance(t, testContext.Accounts, sndAddr, rewaBalance, token, 0, dcdtBalance)

	gasLimit := uint64(40)
	tx := utils.CreateDCDTTransferTx(0, sndAddr, rcvAddr, token, big.NewInt(100), gasPrice, gasLimit)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalanceSnd := big.NewInt(99999900)
	utils.CheckDCDTBalance(t, testContext, sndAddr, token, expectedBalanceSnd)

	expectedReceiverBalance := big.NewInt(100)
	utils.CheckDCDTBalance(t, testContext, rcvAddr, token, expectedReceiverBalance)

	expectedREWABalance := big.NewInt(99999640)
	utils.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedREWABalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(360), accumulatedFees)
}

func TestDCDTTransferShouldWorkToMuchGasShouldConsumeAllGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	rewaBalance := big.NewInt(100000000)
	dcdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithDCDTBalance(t, testContext.Accounts, sndAddr, rewaBalance, token, 0, dcdtBalance)

	gasLimit := uint64(1000)
	tx := utils.CreateDCDTTransferTx(0, sndAddr, rcvAddr, token, big.NewInt(100), gasPrice, gasLimit)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalanceSnd := big.NewInt(99999900)
	utils.CheckDCDTBalance(t, testContext, sndAddr, token, expectedBalanceSnd)

	expectedReceiverBalance := big.NewInt(100)
	utils.CheckDCDTBalance(t, testContext, rcvAddr, token, expectedReceiverBalance)

	expectedREWABalance := big.NewInt(99990000)
	utils.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedREWABalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(10000), accumulatedFees)
}

func TestDCDTTransferInvalidDCDTValueShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	rewaBalance := big.NewInt(100000000)
	dcdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithDCDTBalance(t, testContext.Accounts, sndAddr, rewaBalance, token, 0, dcdtBalance)

	gasLimit := uint64(1000)
	tx := utils.CreateDCDTTransferTx(0, sndAddr, rcvAddr, token, big.NewInt(100000000+1), gasPrice, gasLimit)

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, process.ErrFailedTransaction, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckDCDTBalance(t, testContext, sndAddr, token, rewaBalance)

	expectedReceiverBalance := big.NewInt(0)
	utils.CheckDCDTBalance(t, testContext, rcvAddr, token, expectedReceiverBalance)

	expectedREWABalance := big.NewInt(99990000)
	utils.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedREWABalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(10000), accumulatedFees)
}

func TestDCDTTransferCallBackOnErrorShouldNotGenerateSCRsFurther(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardC, _ := sharding.NewMultiShardCoordinator(2, 0)
	testContext, err := vm.CreatePreparedTxProcessorWithVMsWithShardCoordinator(config.EnableEpochs{}, shardC)
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := bytes.Repeat([]byte{0}, 32)
	sndAddr[12] = 1
	sndAddr[31] = 1
	rcvAddr := bytes.Repeat([]byte{0}, 32)
	rcvAddr[12] = 1

	rewaBalance := big.NewInt(100000000)
	dcdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithDCDTBalance(t, testContext.Accounts, sndAddr, rewaBalance, token, 0, dcdtBalance)

	hexEncodedToken := hex.EncodeToString(token)
	dcdtValueEncoded := hex.EncodeToString(big.NewInt(100).Bytes())
	txDataField := bytes.Join([][]byte{[]byte(core.BuiltInFunctionDCDTTransfer), []byte(hexEncodedToken), []byte(dcdtValueEncoded), []byte(hex.EncodeToString([]byte("something")))}, []byte("@"))
	scr := &smartContractResult.SmartContractResult{
		Nonce:         0,
		Value:         big.NewInt(0),
		RcvAddr:       rcvAddr,
		SndAddr:       sndAddr,
		Data:          txDataField,
		CallType:      0,
		ReturnMessage: []byte("something"),
	}
	retCode, err := testContext.ScProcessor.ProcessSmartContractResult(scr)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedReceiverBalance := big.NewInt(100)
	utils.CheckDCDTBalance(t, testContext, rcvAddr, token, expectedReceiverBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(0), accumulatedFees)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.Equal(t, 0, len(intermediateTxs))
}
