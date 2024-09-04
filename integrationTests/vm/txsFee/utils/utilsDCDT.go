package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data/dcdt"
	"github.com/DharitriOne/drt-chain-core-go/data/transaction"
	"github.com/DharitriOne/drt-chain-go/integrationTests"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm"
	"github.com/DharitriOne/drt-chain-go/state"
	"github.com/DharitriOne/drt-chain-go/testscommon/txDataBuilder"
	"github.com/stretchr/testify/require"
)

// CreateAccountWithDCDTBalance -
func CreateAccountWithDCDTBalance(
	t *testing.T,
	accnts state.AccountsAdapter,
	pubKey []byte,
	rewaValue *big.Int,
	tokenIdentifier []byte,
	dcdtNonce uint64,
	dcdtValue *big.Int,
) {
	account, err := accnts.LoadAccount(pubKey)
	require.Nil(t, err)

	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	userAccount.IncreaseNonce(0)
	err = userAccount.AddToBalance(rewaValue)
	require.Nil(t, err)

	dcdtData := &dcdt.DCDigitalToken{
		Value:      dcdtValue,
		Properties: []byte{},
	}
	if dcdtNonce > 0 {
		dcdtData.TokenMetaData = &dcdt.MetaData{
			Name:    []byte(fmt.Sprintf("Token %d", dcdtNonce)),
			URIs:    [][]byte{[]byte(fmt.Sprintf("URI for token %d", dcdtNonce))},
			Creator: pubKey,
			Nonce:   dcdtNonce,
		}
	}

	dcdtDataBytes, err := integrationTests.TestMarshalizer.Marshal(dcdtData)
	require.Nil(t, err)

	key := append([]byte(core.ProtectedKeyPrefix), []byte(core.DCDTKeyIdentifier)...)
	key = append(key, tokenIdentifier...)
	if dcdtNonce > 0 {
		key = append(key, big.NewInt(0).SetUint64(dcdtNonce).Bytes()...)
	}

	err = userAccount.SaveKeyValue(key, dcdtDataBytes)
	require.Nil(t, err)

	err = accnts.SaveAccount(account)
	require.Nil(t, err)

	saveNewTokenOnSystemAccount(t, accnts, key, dcdtData)

	_, err = accnts.Commit()
	require.Nil(t, err)
}

// CreateAccountWithNFT -
func CreateAccountWithNFT(
	t *testing.T,
	accnts state.AccountsAdapter,
	pubKey []byte,
	rewaValue *big.Int,
	tokenIdentifier []byte,
	attributes []byte,
) {
	account, err := accnts.LoadAccount(pubKey)
	require.Nil(t, err)

	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	userAccount.IncreaseNonce(0)
	err = userAccount.AddToBalance(rewaValue)
	require.Nil(t, err)

	dcdtData := &dcdt.DCDigitalToken{
		Value:      big.NewInt(1),
		Properties: []byte{},
		TokenMetaData: &dcdt.MetaData{
			Nonce:      1,
			Attributes: attributes,
		},
	}

	dcdtDataBytes, err := integrationTests.TestMarshalizer.Marshal(dcdtData)
	require.Nil(t, err)

	key := append([]byte(core.ProtectedKeyPrefix), []byte(core.DCDTKeyIdentifier)...)
	key = append(key, tokenIdentifier...)
	key = append(key, big.NewInt(0).SetUint64(1).Bytes()...)

	err = userAccount.SaveKeyValue(key, dcdtDataBytes)
	require.Nil(t, err)

	err = accnts.SaveAccount(account)
	require.Nil(t, err)

	saveNewTokenOnSystemAccount(t, accnts, key, dcdtData)

	_, err = accnts.Commit()
	require.Nil(t, err)
}

func saveNewTokenOnSystemAccount(t *testing.T, accnts state.AccountsAdapter, tokenKey []byte, dcdtData *dcdt.DCDigitalToken) {
	dcdtDataOnSystemAcc := dcdtData
	dcdtDataOnSystemAcc.Properties = nil
	dcdtDataOnSystemAcc.Reserved = []byte{1}
	dcdtDataOnSystemAcc.Value.Set(dcdtData.Value)

	dcdtDataBytes, err := integrationTests.TestMarshalizer.Marshal(dcdtData)
	require.Nil(t, err)

	sysAccount, err := accnts.LoadAccount(core.SystemAccountAddress)
	require.Nil(t, err)

	sysUserAccount, ok := sysAccount.(state.UserAccountHandler)
	require.True(t, ok)

	err = sysUserAccount.SaveKeyValue(tokenKey, dcdtDataBytes)
	require.Nil(t, err)

	err = accnts.SaveAccount(sysAccount)
	require.Nil(t, err)
}

// CreateAccountWithDCDTBalanceAndRoles -
func CreateAccountWithDCDTBalanceAndRoles(
	t *testing.T,
	accnts state.AccountsAdapter,
	pubKey []byte,
	rewaValue *big.Int,
	tokenIdentifier []byte,
	dcdtNonce uint64,
	dcdtValue *big.Int,
	roles [][]byte,
) {
	CreateAccountWithDCDTBalance(t, accnts, pubKey, rewaValue, tokenIdentifier, dcdtNonce, dcdtValue)
	SetDCDTRoles(t, accnts, pubKey, tokenIdentifier, roles)
}

// SetDCDTRoles -
func SetDCDTRoles(
	t *testing.T,
	accnts state.AccountsAdapter,
	pubKey []byte,
	tokenIdentifier []byte,
	roles [][]byte,
) {
	account, err := accnts.LoadAccount(pubKey)
	require.Nil(t, err)

	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	key := append([]byte(core.ProtectedKeyPrefix), append([]byte(core.DCDTRoleIdentifier), []byte(core.DCDTKeyIdentifier)...)...)
	key = append(key, tokenIdentifier...)

	if len(roles) == 0 {
		err = userAccount.SaveKeyValue(key, []byte{})
		require.Nil(t, err)

		return
	}

	rolesData := &dcdt.DCDTRoles{
		Roles: roles,
	}

	rolesDataBytes, err := integrationTests.TestMarshalizer.Marshal(rolesData)
	require.Nil(t, err)

	err = userAccount.SaveKeyValue(key, rolesDataBytes)
	require.Nil(t, err)

	err = accnts.SaveAccount(account)
	require.Nil(t, err)

	_, err = accnts.Commit()
	require.Nil(t, err)
}

// SetLastNFTNonce -
func SetLastNFTNonce(
	t *testing.T,
	accnts state.AccountsAdapter,
	pubKey []byte,
	tokenIdentifier []byte,
	lastNonce uint64,
) {
	account, err := accnts.LoadAccount(pubKey)
	require.Nil(t, err)

	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	key := append([]byte(core.ProtectedKeyPrefix), []byte(core.DCDTNFTLatestNonceIdentifier)...)
	key = append(key, tokenIdentifier...)

	err = userAccount.SaveKeyValue(key, big.NewInt(int64(lastNonce)).Bytes())
	require.Nil(t, err)

	err = accnts.SaveAccount(account)
	require.Nil(t, err)

	_, err = accnts.Commit()
	require.Nil(t, err)
}

// CreateDCDTTransferTx -
func CreateDCDTTransferTx(nonce uint64, sndAddr, rcvAddr []byte, tokenIdentifier []byte, dcdtValue *big.Int, gasPrice, gasLimit uint64) *transaction.Transaction {
	hexEncodedToken := hex.EncodeToString(tokenIdentifier)
	dcdtValueEncoded := hex.EncodeToString(dcdtValue.Bytes())
	txDataField := bytes.Join([][]byte{[]byte(core.BuiltInFunctionDCDTTransfer), []byte(hexEncodedToken), []byte(dcdtValueEncoded)}, []byte("@"))

	return &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  sndAddr,
		RcvAddr:  rcvAddr,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txDataField,
		Value:    big.NewInt(0),
	}
}

// TransferDCDTData -
type TransferDCDTData struct {
	Token []byte
	Nonce uint64
	Value *big.Int
}

// CreateMultiTransferTX -
func CreateMultiTransferTX(nonce uint64, sender, dest []byte, gasPrice, gasLimit uint64, tds ...*TransferDCDTData) *transaction.Transaction {
	numTransfers := len(tds)
	encodedReceiver := hex.EncodeToString(dest)
	hexEncodedNumTransfers := hex.EncodeToString(big.NewInt(int64(numTransfers)).Bytes())

	txDataField := []byte(strings.Join([]string{core.BuiltInFunctionMultiDCDTNFTTransfer, encodedReceiver, hexEncodedNumTransfers}, "@"))
	for _, td := range tds {
		hexEncodedToken := hex.EncodeToString(td.Token)
		dcdtValueEncoded := hex.EncodeToString(td.Value.Bytes())
		hexEncodedNonce := "00"
		if td.Nonce != 0 {
			hexEncodedNonce = hex.EncodeToString(big.NewInt(int64(td.Nonce)).Bytes())
		}

		txDataField = []byte(strings.Join([]string{string(txDataField), hexEncodedToken, hexEncodedNonce, dcdtValueEncoded}, "@"))
	}

	return &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  sender,
		RcvAddr:  sender,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txDataField,
		Value:    big.NewInt(0),
	}
}

// CreateDCDTNFTTransferTx -
func CreateDCDTNFTTransferTx(
	nonce uint64,
	sndAddr []byte,
	rcvAddr []byte,
	tokenIdentifier []byte,
	dcdtNonce uint64,
	dcdtValue *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	endpointName string,
	arguments ...[]byte) *transaction.Transaction {

	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionDCDTNFTTransfer)
	txData.Bytes(tokenIdentifier)
	txData.Int64(int64(dcdtNonce))
	txData.BigInt(dcdtValue)
	txData.Bytes(rcvAddr)

	if len(endpointName) > 0 {
		txData.Str(endpointName)

		for _, arg := range arguments {
			txData.Bytes(arg)
		}
	}

	return &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  sndAddr,
		RcvAddr:  sndAddr, // receiver = sender for DCDTNFTTransfer
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txData.ToBytes(),
		Value:    big.NewInt(0),
	}
}

// CheckDCDTBalance -
func CheckDCDTBalance(t *testing.T, testContext *vm.VMTestContext, addr []byte, tokenIdentifier []byte, expectedBalance *big.Int) {
	checkDcdtBalance(t, testContext, addr, tokenIdentifier, 0, expectedBalance)
}

// CheckDCDTNFTBalance -
func CheckDCDTNFTBalance(tb testing.TB, testContext *vm.VMTestContext, addr []byte, tokenIdentifier []byte, dcdtNonce uint64, expectedBalance *big.Int) {
	checkDcdtBalance(tb, testContext, addr, tokenIdentifier, dcdtNonce, expectedBalance)
}

// CreateDCDTLocalBurnTx -
func CreateDCDTLocalBurnTx(nonce uint64, sndAddr, rcvAddr []byte, tokenIdentifier []byte, dcdtValue *big.Int, gasPrice, gasLimit uint64) *transaction.Transaction {
	hexEncodedToken := hex.EncodeToString(tokenIdentifier)
	dcdtValueEncoded := hex.EncodeToString(dcdtValue.Bytes())
	txDataField := bytes.Join([][]byte{[]byte(core.BuiltInFunctionDCDTLocalBurn), []byte(hexEncodedToken), []byte(dcdtValueEncoded)}, []byte("@"))

	return &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  sndAddr,
		RcvAddr:  rcvAddr,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txDataField,
		Value:    big.NewInt(0),
	}
}

// CreateDCDTLocalMintTx -
func CreateDCDTLocalMintTx(nonce uint64, sndAddr, rcvAddr []byte, tokenIdentifier []byte, dcdtValue *big.Int, gasPrice, gasLimit uint64) *transaction.Transaction {
	hexEncodedToken := hex.EncodeToString(tokenIdentifier)
	dcdtValueEncoded := hex.EncodeToString(dcdtValue.Bytes())
	txDataField := bytes.Join([][]byte{[]byte(core.BuiltInFunctionDCDTLocalMint), []byte(hexEncodedToken), []byte(dcdtValueEncoded)}, []byte("@"))

	return &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  sndAddr,
		RcvAddr:  rcvAddr,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txDataField,
		Value:    big.NewInt(0),
	}
}

// CreateDCDTNFTBurnTx -
func CreateDCDTNFTBurnTx(nonce uint64, sndAddr, rcvAddr []byte, tokenIdentifier []byte, tokenNonce uint64, dcdtValue *big.Int, gasPrice, gasLimit uint64) *transaction.Transaction {
	hexEncodedToken := hex.EncodeToString(tokenIdentifier)
	hexEncodedNonce := hex.EncodeToString(big.NewInt(int64(tokenNonce)).Bytes())
	dcdtValueEncoded := hex.EncodeToString(dcdtValue.Bytes())
	txDataField := bytes.Join([][]byte{[]byte(core.BuiltInFunctionDCDTNFTBurn), []byte(hexEncodedToken), []byte(hexEncodedNonce), []byte(dcdtValueEncoded)}, []byte("@"))

	return &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  sndAddr,
		RcvAddr:  rcvAddr,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txDataField,
		Value:    big.NewInt(0),
	}
}

// CreateNFTSingleFreezeAndWipeTxs -
func CreateNFTSingleFreezeAndWipeTxs(nonce uint64, tokenManager, addressToFreeze []byte, tokenIdentifier []byte, tokenNonce uint64, gasPrice, gasLimit uint64) (*transaction.Transaction, *transaction.Transaction) {
	hexEncodedToken := hex.EncodeToString(tokenIdentifier)
	hexEncodedNonce := hex.EncodeToString(big.NewInt(int64(tokenNonce)).Bytes())
	addressToFreezeHex := hex.EncodeToString(addressToFreeze)

	txDataField := bytes.Join([][]byte{[]byte("freezeSingleNFT"), []byte(hexEncodedToken), []byte(hexEncodedNonce), []byte(addressToFreezeHex)}, []byte("@"))
	freezeTx := &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  tokenManager,
		RcvAddr:  core.DCDTSCAddress,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txDataField,
		Value:    big.NewInt(0),
	}

	txDataField = bytes.Join([][]byte{[]byte("wipeSingleNFT"), []byte(hexEncodedToken), []byte(hexEncodedNonce), []byte(addressToFreezeHex)}, []byte("@"))
	wipeTx := &transaction.Transaction{
		Nonce:    nonce + 1,
		SndAddr:  tokenManager,
		RcvAddr:  core.DCDTSCAddress,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txDataField,
		Value:    big.NewInt(0),
	}

	return freezeTx, wipeTx
}

func checkDcdtBalance(
	tb testing.TB,
	testContext *vm.VMTestContext,
	addr []byte,
	tokenIdentifier []byte,
	dcdtNonce uint64,
	expectedBalance *big.Int,
) {
	dcdtData, err := testContext.BlockchainHook.GetDCDTToken(addr, tokenIdentifier, dcdtNonce)
	require.Nil(tb, err)
	require.Equal(tb, expectedBalance, dcdtData.Value)
}

// CreateDCDTNFTUpdateAttributesTx -
func CreateDCDTNFTUpdateAttributesTx(
	nonce uint64,
	sndAddr []byte,
	tokenIdentifier []byte,
	gasPrice uint64,
	gasLimit uint64,
	newAttributes []byte,
) *transaction.Transaction {

	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionDCDTNFTUpdateAttributes)
	txData.Bytes(tokenIdentifier)
	txData.Int64(1)
	txData.Bytes(newAttributes)

	return &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  sndAddr,
		RcvAddr:  sndAddr, // receiver = sender for DCDTNFTUpdateAttributes
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txData.ToBytes(),
		Value:    big.NewInt(0),
	}
}
