package localFuncs

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-go/integrationTests"
	dcdtCommon "github.com/DharitriOne/drt-chain-go/integrationTests/vm/dcdt"
	"github.com/DharitriOne/drt-chain-go/testscommon/txDataBuilder"
	"github.com/stretchr/testify/assert"
)

func TestDCDTLocalMintAndBurnFromSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := dcdtCommon.CreateNodesAndPrepareBalances(1)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	scAddress := dcdtCommon.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../testdata/local-dcdt-and-nft.wasm")

	dcdtLocalMintAndBurnFromSCRunTestsAndAsserts(t, nodes, nodes[0].OwnAccount, scAddress, idxProposers, nonce, round)
}

func dcdtLocalMintAndBurnFromSCRunTestsAndAsserts(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	ownerWallet *integrationTests.TestWalletAccount,
	scAddress []byte,
	idxProposers []int,
	nonce uint64,
	round uint64,
) {
	tokenIdentifier := dcdtCommon.PrepareFungibleTokensWithLocalBurnAndMintWithIssuerAccount(t, nodes, ownerWallet, scAddress, idxProposers, &nonce, &round)

	txData := []byte("localMint" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(100).Bytes()))
	integrationTests.CreateAndSendTransactionWithSenderAccount(
		nodes[0],
		nodes,
		big.NewInt(0),
		ownerWallet,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)
	integrationTests.CreateAndSendTransactionWithSenderAccount(
		nodes[0],
		nodes,
		big.NewInt(0),
		ownerWallet,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 2
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, scAddress, nodes, []byte(tokenIdentifier), 0, 200)

	txData = []byte("localBurn" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(50).Bytes()))
	integrationTests.CreateAndSendTransactionWithSenderAccount(
		nodes[0],
		nodes,
		big.NewInt(0),
		ownerWallet,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)
	integrationTests.CreateAndSendTransactionWithSenderAccount(
		nodes[0],
		nodes,
		big.NewInt(0),
		ownerWallet,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, scAddress, nodes, []byte(tokenIdentifier), 0, 100)
}

func TestDCDTSetRolesAndLocalMintAndBurnFromSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := dcdtCommon.CreateNodesAndPrepareBalances(1)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	scAddress := dcdtCommon.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../testdata/local-dcdt-and-nft.wasm")

	issuePrice := big.NewInt(1000)
	txData := []byte("issueFungibleToken" + "@" + hex.EncodeToString([]byte("TOKEN")) +
		"@" + hex.EncodeToString([]byte("TKR")) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		issuePrice,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte("TKR")))
	txData = []byte("setLocalRoles" + "@" + hex.EncodeToString(scAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + "01" + "@" + "02")
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	txData = []byte("localMint" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(100).Bytes()))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, scAddress, nodes, []byte(tokenIdentifier), 0, 201)

	txData = []byte("localBurn" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(50).Bytes()))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, scAddress, nodes, []byte(tokenIdentifier), 0, 101)
}

func TestDCDTSetTransferRoles(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := dcdtCommon.CreateNodesAndPrepareBalances(2)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	scAddress := dcdtCommon.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../testdata/use-module.wasm")
	nrRoundsToPropagateMultiShard := 12
	tokenIdentifier := dcdtCommon.PrepareFungibleTokensWithLocalBurnAndMint(t, nodes, scAddress, idxProposers, &nonce, &round)

	dcdtCommon.SetRoles(nodes, scAddress, []byte(tokenIdentifier), [][]byte{[]byte(core.DCDTRoleTransfer)})

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	destAddress := nodes[1].OwnAccount.Address

	amount := int64(100)
	txData := txDataBuilder.NewBuilder()
	txData.Clear().TransferDCDT(tokenIdentifier, amount).Str("forwardPayments").Bytes(destAddress).Str("fund")

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 10, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, destAddress, nodes, []byte(tokenIdentifier), 0, amount)
}

func TestDCDTSetTransferRolesForwardAsyncCallFailsIntra(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testDCDTWithTransferRoleAndForwarder(t, 1)
}

func TestDCDTSetTransferRolesForwardAsyncCallFailsCross(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testDCDTWithTransferRoleAndForwarder(t, 2)
}

func testDCDTWithTransferRoleAndForwarder(t *testing.T, numShards int) {
	nodes, idxProposers := dcdtCommon.CreateNodesAndPrepareBalances(numShards)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	scAddressA := dcdtCommon.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../testdata/use-module.wasm")
	scAddressB := dcdtCommon.DeployNonPayableSmartContractFromNode(t, nodes, 1, idxProposers, &nonce, &round, "../testdata/use-module.wasm")
	nrRoundsToPropagateMultiShard := 12
	tokenIdentifier := dcdtCommon.PrepareFungibleTokensWithLocalBurnAndMint(t, nodes, scAddressA, idxProposers, &nonce, &round)

	dcdtCommon.SetRoles(nodes, scAddressA, []byte(tokenIdentifier), [][]byte{[]byte(core.DCDTRoleTransfer)})

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	amount := int64(100)
	txData := txDataBuilder.NewBuilder()
	txData.Clear().TransferDCDT(tokenIdentifier, amount).Str("forwardPayments").Bytes(scAddressB).Str("depositTokensForAction").Str("fund")

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddressA,
		txData.ToString(),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 15, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, scAddressB, nodes, []byte(tokenIdentifier), 0, 0)
	dcdtCommon.CheckAddressHasTokens(t, scAddressA, nodes, []byte(tokenIdentifier), 0, 0)
	dcdtCommon.CheckAddressHasTokens(t, nodes[0].OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, amount)
}

func TestAsyncCallsAndCallBacksArgumentsIntra(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testAsyncCallAndCallBacksArguments(t, 1)
}

func TestAsyncCallsAndCallBacksArgumentsCross(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testAsyncCallAndCallBacksArguments(t, 2)
}

func testAsyncCallAndCallBacksArguments(t *testing.T, numShards int) {
	nodes, idxProposers := dcdtCommon.CreateNodesAndPrepareBalances(numShards)
	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	scAddressA := dcdtCommon.DeployNonPayableSmartContractFromNode(t, nodes, 0, idxProposers, &nonce, &round, "forwarder.wasm")
	scAddressB := dcdtCommon.DeployNonPayableSmartContractFromNode(t, nodes, 1, idxProposers, &nonce, &round, "vault.wasm")

	txData := txDataBuilder.NewBuilder()
	txData.Clear().Func("echo_args_async").Bytes(scAddressB).Str("AA").Str("BB")

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddressA,
		txData.ToString(),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 15, nonce, round, idxProposers)
	time.Sleep(time.Second)

	callbackArgs := append([]byte("success"), []byte{0}...)
	callbackArgs = append(callbackArgs, []byte("AABB")...)
	checkDataFromAccountAndKey(t, nodes, scAddressA, []byte("callbackStorage"), callbackArgs)

	txData.Clear().Func("echo_args_async").Bytes(scAddressB)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddressA,
		txData.ToString(),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 15, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkDataFromAccountAndKey(t, nodes, scAddressA, []byte("callbackStorage"), append([]byte("success"), []byte{0}...))
}

func checkDataFromAccountAndKey(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	address []byte,
	key []byte,
	expectedData []byte,
) {
	userAcc := dcdtCommon.GetUserAccountWithAddress(t, address, nodes)
	val, _, _ := userAcc.RetrieveValue(key)
	assert.Equal(t, expectedData, val)
}
