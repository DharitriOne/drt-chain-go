package multiShard

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data/dcdt"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/DharitriOne/drt-chain-go/state"
	"github.com/DharitriOne/drt-chain-go/testscommon/integrationtests"
	"github.com/DharitriOne/drt-chain-go/vm/systemSmartContracts"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestSystemAccountLiquidityAfterCrossShardTransferAndBurn(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	tokenID := []byte("MYNFT")
	sh0Addr := []byte("12345678901234567890123456789010")
	sh1Addr := []byte("12345678901234567890123456789011")
	sh0Context, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer sh0Context.Close()

	sh1Context, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer sh1Context.Close()
	_, _ = vm.CreateAccount(sh1Context.Accounts, sh1Addr, 0, big.NewInt(1000000000))

	// create the nft and ensure that it exists on the account's trie and the liquidity is set on the system account
	utils.CreateAccountWithDCDTBalance(t, sh0Context.Accounts, sh0Addr, big.NewInt(100000000), tokenID, 1, big.NewInt(1))
	utils.CheckDCDTNFTBalance(t, sh0Context, sh0Addr, tokenID, 1, big.NewInt(1))
	utils.CheckDCDTNFTBalance(t, sh0Context, core.SystemAccountAddress, tokenID, 1, big.NewInt(1))

	sh0Accnt, _ := sh0Context.Accounts.LoadAccount(sh0Addr)
	crossShardTransferTx := utils.CreateDCDTNFTTransferTx(sh0Accnt.GetNonce(), sh0Addr, sh1Addr, tokenID, 1, big.NewInt(1), 10, 1000000, "")
	retCode, err := sh0Context.TxProcessor.ProcessTransaction(crossShardTransferTx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.NoError(t, err)

	scrs := sh0Context.GetIntermediateTransactions(t)

	// check the balances after the transfer, as well as the liquidity
	utils.ProcessSCRResult(t, sh1Context, scrs[0], vmcommon.Ok, nil)
	utils.CheckDCDTNFTBalance(t, sh0Context, sh0Addr, tokenID, 1, big.NewInt(0))
	utils.CheckDCDTNFTBalance(t, sh0Context, core.SystemAccountAddress, tokenID, 1, big.NewInt(0))
	utils.CheckDCDTNFTBalance(t, sh1Context, sh1Addr, tokenID, 1, big.NewInt(1))
	utils.CheckDCDTNFTBalance(t, sh1Context, core.SystemAccountAddress, tokenID, 1, big.NewInt(1))

	// set roles and burn the NFT on shard 1
	utils.SetDCDTRoles(t, sh1Context.Accounts, sh1Addr, tokenID, [][]byte{[]byte(core.DCDTRoleNFTBurn)})

	tx := utils.CreateDCDTNFTBurnTx(0, sh1Addr, sh1Addr, tokenID, 1, big.NewInt(1), 10, 100000)
	retCode, err = sh1Context.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	// ensure that the token is burnt from all addresses and system account liquidity
	utils.CheckDCDTNFTBalance(t, sh0Context, sh0Addr, tokenID, 1, big.NewInt(0))
	utils.CheckDCDTNFTBalance(t, sh0Context, core.SystemAccountAddress, tokenID, 1, big.NewInt(0))
	utils.CheckDCDTNFTBalance(t, sh1Context, sh1Addr, tokenID, 1, big.NewInt(0))
	utils.CheckDCDTNFTBalance(t, sh1Context, core.SystemAccountAddress, tokenID, 1, big.NewInt(0))
}

func TestSystemAccountLiquidityAfterNFTWipe(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	tokenID := []byte("MYNFT-0a0a0a")
	sh0Addr := bytes.Repeat([]byte{1}, 31)
	sh0Addr = append(sh0Addr, 0)
	sh0Context, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer sh0Context.Close()

	metaContext, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, config.EnableEpochs{})
	require.Nil(t, err)
	defer metaContext.Close()

	// create the nft and ensure that it exists on the account's trie and the liquidity is set on the system account
	utils.CreateAccountWithDCDTBalance(t, sh0Context.Accounts, sh0Addr, big.NewInt(10000000000000), tokenID, 1, big.NewInt(1))
	utils.CheckDCDTNFTBalance(t, sh0Context, sh0Addr, tokenID, 1, big.NewInt(1))
	utils.CheckDCDTNFTBalance(t, sh0Context, core.SystemAccountAddress, tokenID, 1, big.NewInt(1))

	addTokenInMeta(t, metaContext, tokenID, &systemSmartContracts.DCDTDataV2{
		OwnerAddress: sh0Addr,
		CanFreeze:    true,
		CanWipe:      true,
	})
	sh0Accnt, _ := sh0Context.Accounts.LoadAccount(sh0Addr)
	freezeTx, wipeTx := utils.CreateNFTSingleFreezeAndWipeTxs(sh0Accnt.GetNonce(), sh0Addr, sh0Addr, tokenID, 1, 10, 1000000000)
	retCode, err := metaContext.TxProcessor.ProcessTransaction(freezeTx)
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, retCode)

	scrs := metaContext.GetIntermediateTransactions(t)
	utils.ProcessSCRResult(t, sh0Context, scrs[0], vmcommon.Ok, nil)

	retCode, err = metaContext.TxProcessor.ProcessTransaction(wipeTx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.NoError(t, err)

	scrs = metaContext.GetIntermediateTransactions(t)
	utils.ProcessSCRResult(t, sh0Context, scrs[1], vmcommon.Ok, nil)

	_, _ = sh0Context.Accounts.Commit()

	// ensure that there is no liquidity left in the account or the system account

	checkDcdtBalanceInAccountStorage(t, sh0Context, sh0Addr, tokenID, 1, big.NewInt(0))
	utils.CheckDCDTNFTBalance(t, sh0Context, core.SystemAccountAddress, tokenID, 1, big.NewInt(0))
}

func TestSystemAccountLiquidityAfterSFTWipe(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	tokenID := []byte("MYSFT-0a0a0a")
	sh0Addr := bytes.Repeat([]byte{1}, 31)
	sh0Addr = append(sh0Addr, 0)
	sh0Context, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer sh0Context.Close()

	metaContext, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(core.MetachainShardId, config.EnableEpochs{})
	require.Nil(t, err)
	defer metaContext.Close()

	// create the nft and ensure that it exists on the account's trie and the liquidity is set on the system account
	utils.CreateAccountWithDCDTBalance(t, sh0Context.Accounts, sh0Addr, big.NewInt(10000000000000), tokenID, 1, big.NewInt(10))
	utils.CheckDCDTNFTBalance(t, sh0Context, sh0Addr, tokenID, 1, big.NewInt(10))
	utils.CheckDCDTNFTBalance(t, sh0Context, core.SystemAccountAddress, tokenID, 1, big.NewInt(10))

	addTokenInMeta(t, metaContext, tokenID, &systemSmartContracts.DCDTDataV2{
		OwnerAddress: sh0Addr,
		CanFreeze:    true,
		CanWipe:      true,
	})
	sh0Accnt, _ := sh0Context.Accounts.LoadAccount(sh0Addr)
	freezeTx, wipeTx := utils.CreateNFTSingleFreezeAndWipeTxs(sh0Accnt.GetNonce(), sh0Addr, sh0Addr, tokenID, 1, 10, 1000000000)
	retCode, err := metaContext.TxProcessor.ProcessTransaction(freezeTx)
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, retCode)

	scrs := metaContext.GetIntermediateTransactions(t)
	utils.ProcessSCRResult(t, sh0Context, scrs[0], vmcommon.Ok, nil)

	retCode, err = metaContext.TxProcessor.ProcessTransaction(wipeTx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.NoError(t, err)

	scrs = metaContext.GetIntermediateTransactions(t)
	utils.ProcessSCRResult(t, sh0Context, scrs[1], vmcommon.Ok, nil)

	// ensure that there is no liquidity left in the account or the system account
	checkDcdtBalanceInAccountStorage(t, sh0Context, sh0Addr, tokenID, 1, big.NewInt(0))
	utils.CheckDCDTNFTBalance(t, sh0Context, core.SystemAccountAddress, tokenID, 1, big.NewInt(0))
}

func addTokenInMeta(t *testing.T, metaContext *vm.VMTestContext, tokenID []byte, token *systemSmartContracts.DCDTDataV2) {
	marshaledData, err := integrationtests.TestMarshalizer.Marshal(token)
	require.NoError(t, err)

	dcdtAccnt, err := metaContext.Accounts.LoadAccount(core.DCDTSCAddress)
	require.NoError(t, err)

	dcdtAccntUser, ok := dcdtAccnt.(state.UserAccountHandler)
	require.True(t, ok)

	err = dcdtAccntUser.SaveKeyValue(tokenID, marshaledData)
	require.NoError(t, err)

	err = metaContext.Accounts.SaveAccount(dcdtAccntUser)
	require.NoError(t, err)
	_, err = metaContext.Accounts.Commit()
	require.NoError(t, err)
}

func checkDcdtBalanceInAccountStorage(t *testing.T, context *vm.VMTestContext, address []byte, tokenID []byte, nonce uint64, expectedValue *big.Int) {
	res, err := context.Accounts.LoadAccount(address)
	require.NoError(t, err)

	userAcc, ok := res.(state.UserAccountHandler)
	require.True(t, ok)

	dcdtTokenKey := []byte(core.ProtectedKeyPrefix + core.DCDTKeyIdentifier + string(tokenID))
	dcdtTokenKey = append(dcdtTokenKey, big.NewInt(int64(nonce)).Bytes()...)
	tokenBytes, _, err := userAcc.RetrieveValue(dcdtTokenKey)
	require.NoError(t, err)

	esToken := dcdt.DCDigitalToken{}
	err = integrationtests.TestMarshalizer.Unmarshal(&esToken, tokenBytes)
	require.NoError(t, err)

	if expectedValue.Uint64() == 0 {
		require.Nil(t, esToken.Value)
	} else {
		require.Equal(t, expectedValue, esToken.Value)
	}
}
