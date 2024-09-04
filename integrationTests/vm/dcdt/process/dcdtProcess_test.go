package process

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/core/check"
	"github.com/DharitriOne/drt-chain-core-go/data/block"
	"github.com/DharitriOne/drt-chain-core-go/data/dcdt"
	"github.com/DharitriOne/drt-chain-core-go/data/smartContractResult"
	vmData "github.com/DharitriOne/drt-chain-core-go/data/vm"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/integrationTests"
	testVm "github.com/DharitriOne/drt-chain-go/integrationTests/vm"
	dcdtCommon "github.com/DharitriOne/drt-chain-go/integrationTests/vm/dcdt"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm/wasm"
	"github.com/DharitriOne/drt-chain-go/process"
	vmFactory "github.com/DharitriOne/drt-chain-go/process/factory"
	"github.com/DharitriOne/drt-chain-go/testscommon/txDataBuilder"
	"github.com/DharitriOne/drt-chain-go/vm"
	"github.com/DharitriOne/drt-chain-go/vm/systemSmartContracts"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
	vmcommonBuiltInFunctions "github.com/DharitriOne/drt-chain-vm-common-go/builtInFunctions"
	"github.com/stretchr/testify/require"
)

func TestDCDTIssueAndTransactionsOnMultiShardEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	enableEpochs := config.EnableEpochs{
		GlobalMintBurnDisableEpoch:                  integrationTests.UnreachableEpoch,
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch: integrationTests.UnreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:              integrationTests.UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch:        integrationTests.UnreachableEpoch,
	}
	nodes := integrationTests.CreateNodesWithEnableEpochs(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		enableEpochs,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := int64(10000000000)
	integrationTests.MintAllNodes(nodes, big.NewInt(initialVal))

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// send token issue

	initialSupply := int64(10000000000)
	ticker := "TCK"
	dcdtCommon.IssueTestToken(nodes, initialSupply, ticker)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	txData := txDataBuilder.NewBuilder()

	// send tx to other nodes
	valueToSend := int64(100)
	for _, node := range nodes[1:] {
		txData = txData.Clear().TransferDCDT(tokenIdentifier, valueToSend)
		integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), node.OwnAccount.Address, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	mintValue := int64(10000)
	txData = txData.Clear().Func("mint").Str(tokenIdentifier).Int64(mintValue)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.DCDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	txData.Clear().Func("freeze").Str(tokenIdentifier).Bytes(nodes[2].OwnAccount.Address)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.DCDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	finalSupply := initialSupply + mintValue
	for _, node := range nodes[1:] {
		dcdtCommon.CheckAddressHasTokens(t, node.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, valueToSend)
		finalSupply = finalSupply - valueToSend
	}

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, finalSupply)

	txData.Clear().BurnDCDT(tokenIdentifier, mintValue)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.DCDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	txData.Clear().Func("freeze").Str(tokenIdentifier).Bytes(nodes[1].OwnAccount.Address)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.DCDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	txData.Clear().Func("wipe").Str(tokenIdentifier).Bytes(nodes[2].OwnAccount.Address)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.DCDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	txData.Clear().Func("pause").Str(tokenIdentifier)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.DCDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)

	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtFrozenData := dcdtCommon.GetDCDTTokenData(t, nodes[1].OwnAccount.Address, nodes, []byte(tokenIdentifier), 0)
	dcdtUserMetaData := vmcommonBuiltInFunctions.DCDTUserMetadataFromBytes(dcdtFrozenData.Properties)
	require.True(t, dcdtUserMetaData.Frozen)

	wipedAcc := dcdtCommon.GetUserAccountWithAddress(t, nodes[2].OwnAccount.Address, nodes)
	tokenKey := []byte(core.ProtectedKeyPrefix + "dcdt" + tokenIdentifier)
	retrievedData, _, _ := wipedAcc.RetrieveValue(tokenKey)
	require.Equal(t, 0, len(retrievedData))

	systemSCAcc := dcdtCommon.GetUserAccountWithAddress(t, core.SystemAccountAddress, nodes)
	retrievedData, _, _ = systemSCAcc.RetrieveValue(tokenKey)
	dcdtGlobalMetaData := vmcommonBuiltInFunctions.DCDTGlobalMetadataFromBytes(retrievedData)
	require.True(t, dcdtGlobalMetaData.Paused)

	finalSupply = finalSupply - mintValue
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, finalSupply)

	dcdtSCAcc := dcdtCommon.GetUserAccountWithAddress(t, vm.DCDTSCAddress, nodes)
	retrievedData, _, _ = dcdtSCAcc.RetrieveValue([]byte(tokenIdentifier))
	tokenInSystemSC := &systemSmartContracts.DCDTDataV2{}
	_ = integrationTests.TestMarshalizer.Unmarshal(tokenInSystemSC, retrievedData)
	require.Zero(t, tokenInSystemSC.MintedValue.Cmp(big.NewInt(initialSupply+mintValue)))
	require.Zero(t, tokenInSystemSC.BurntValue.Cmp(big.NewInt(mintValue)))
	require.True(t, tokenInSystemSC.IsPaused)
}

func TestDCDTCallBurnOnANonBurnableToken(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	enableEpochs := config.EnableEpochs{
		GlobalMintBurnDisableEpoch:                  integrationTests.UnreachableEpoch,
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch: integrationTests.UnreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:              integrationTests.UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch:        integrationTests.UnreachableEpoch,
		MultiClaimOnDelegationEnableEpoch:           integrationTests.UnreachableEpoch,
	}

	nodes := integrationTests.CreateNodesWithEnableEpochs(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		enableEpochs,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	// send token issue
	ticker := "ALC"
	issuePrice := big.NewInt(1000)
	initialSupply := int64(10000000000)
	tokenIssuer := nodes[0]
	txData := txDataBuilder.NewBuilder()

	txData.Clear().IssueDCDT("aliceToken", ticker, initialSupply, 6)
	txData.CanFreeze(true).CanWipe(true).CanPause(true).CanMint(true).CanBurn(false)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, issuePrice, vm.DCDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	// send tx to other nodes
	valueToSend := int64(100)
	for _, node := range nodes[1:] {
		txData.Clear().TransferDCDT(tokenIdentifier, valueToSend)
		integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), node.OwnAccount.Address, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	finalSupply := initialSupply
	for _, node := range nodes[1:] {
		dcdtCommon.CheckAddressHasTokens(t, node.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, valueToSend)
		finalSupply = finalSupply - valueToSend
	}

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, finalSupply)

	burnValue := int64(77)
	txData.Clear().BurnDCDT(tokenIdentifier, burnValue)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.DCDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)

	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtSCAcc := dcdtCommon.GetUserAccountWithAddress(t, vm.DCDTSCAddress, nodes)
	retrievedData, _, _ := dcdtSCAcc.RetrieveValue([]byte(tokenIdentifier))
	tokenInSystemSC := &systemSmartContracts.DCDTDataV2{}
	_ = integrationTests.TestMarshalizer.Unmarshal(tokenInSystemSC, retrievedData)
	require.Equal(t, initialSupply, tokenInSystemSC.MintedValue.Int64())
	require.Zero(t, tokenInSystemSC.BurntValue.Int64())

	// if everything is ok, the caller should have received the amount of burnt tokens back because canBurn = false
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, finalSupply)
}

func TestDCDTIssueAndSelfTransferShouldNotChangeBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := int64(10000000000)
	integrationTests.MintAllNodes(nodes, big.NewInt(initialVal))

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	initialSupply := int64(10000000000)
	ticker := "TCK"
	dcdtCommon.IssueTestToken(nodes, initialSupply, ticker)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	txData := txDataBuilder.NewBuilder()

	valueToSend := int64(100)
	txData = txData.Clear().TransferDCDT(tokenIdentifier, valueToSend)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), nodes[0].OwnAccount.Address, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, nodes[0].OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)
}

func TestDCDTIssueFromASmartContractSimulated(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	metaNode := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})

	defer func() {
		metaNode.Close()
	}()

	txData := txDataBuilder.NewBuilder()

	ticker := "RBT"
	issuePrice := big.NewInt(1000)
	initialSupply := big.NewInt(10000000000)
	numDecimals := byte(6)

	txData.Clear().IssueDCDT("robertWhyNot", ticker, initialSupply.Int64(), numDecimals)
	txData.CanFreeze(true).CanWipe(true).CanPause(true).CanMint(true).CanBurn(true)
	txData.Bytes([]byte("callID")).Bytes([]byte("callerCallID")) // async args
	txData.Int64(1000)                                           // gas locked
	scr := &smartContractResult.SmartContractResult{
		Nonce:          0,
		Value:          issuePrice,
		RcvAddr:        vm.DCDTSCAddress,
		SndAddr:        metaNode.OwnAccount.Address,
		Data:           txData.ToBytes(),
		PrevTxHash:     []byte("hash"),
		OriginalTxHash: []byte("hash"),
		GasLimit:       10000000,
		GasPrice:       1,
		CallType:       vmData.AsynchronousCall,
		OriginalSender: metaNode.OwnAccount.Address,
	}

	returnCode, err := metaNode.ScProcessor.ProcessSmartContractResult(scr)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)

	interimProc, _ := metaNode.InterimProcContainer.Get(block.SmartContractResultBlock)
	mapCreatedSCRs := interimProc.GetAllCurrentFinishedTxs()

	require.Equal(t, len(mapCreatedSCRs), 2)
	foundTransfer := false
	for _, addedSCR := range mapCreatedSCRs {
		foundTransfer = foundTransfer || strings.Contains(string(addedSCR.GetData()), core.BuiltInFunctionDCDTTransfer)
	}
	require.True(t, foundTransfer)
}

func TestScSendsDcdtToUserWithMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	// send token issue
	initialSupply := int64(10000000000)
	ticker := "TCK"
	dcdtCommon.IssueTestToken(nodes, initialSupply, ticker)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	// deploy the smart contract

	vaultScCode := wasm.GetSCCode("../testdata/vault.wasm")
	vaultScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxData(vaultScCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(vaultScAddress)
	require.Nil(t, err)

	txData := txDataBuilder.NewBuilder()

	// feed funds to the vault
	valueToSendToSc := int64(1000)
	txData.Clear().TransferDCDT(tokenIdentifier, valueToSendToSc)
	txData.Str("accept_funds")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vaultScAddress, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply-valueToSendToSc)
	dcdtCommon.CheckAddressHasTokens(t, vaultScAddress, nodes, []byte(tokenIdentifier), 0, valueToSendToSc)

	// take them back, with a message
	valueToRequest := valueToSendToSc / 4
	txData.Clear().Func("retrieve_funds").Str(tokenIdentifier).Int64(0).Int64(valueToRequest)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vaultScAddress, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply-valueToSendToSc+valueToRequest)
	dcdtCommon.CheckAddressHasTokens(t, vaultScAddress, nodes, []byte(tokenIdentifier), 0, valueToSendToSc-valueToRequest)
}

func TestDCDTcallsSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	// send token issue

	initialSupply := int64(10000000000)
	ticker := "TCK"
	dcdtCommon.IssueTestToken(nodes, initialSupply, ticker)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	// send tx to other nodes
	txData := txDataBuilder.NewBuilder()
	valueToSend := int64(100)
	for _, node := range nodes[1:] {
		txData.Clear().TransferDCDT(tokenIdentifier, valueToSend)
		integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), node.OwnAccount.Address, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	numNodesWithoutIssuer := int64(len(nodes) - 1)
	issuerBalance := initialSupply - valueToSend*numNodesWithoutIssuer
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, issuerBalance)
	for i := 1; i < len(nodes); i++ {
		dcdtCommon.CheckAddressHasTokens(t, nodes[i].OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, valueToSend)
	}

	// deploy the smart contract
	scCode := wasm.GetSCCode("../testdata/crowdfunding-dcdt.wasm")
	scAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxData(scCode)+"@"+
			hex.EncodeToString(big.NewInt(1000).Bytes())+"@"+
			hex.EncodeToString(big.NewInt(1000).Bytes())+"@"+
			hex.EncodeToString([]byte(tokenIdentifier)),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(scAddress)
	require.Nil(t, err)

	// call sc with dcdt
	valueToSendToSc := int64(10)
	for _, node := range nodes {
		txData.Clear().TransferDCDT(tokenIdentifier, valueToSendToSc).Str("fund")
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), scAddress, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	scQuery1 := &process.SCQuery{
		ScAddress: scAddress,
		FuncName:  "getCurrentFunds",
		Arguments: [][]byte{},
	}
	vmOutput1, _, _ := nodes[0].SCQueryService.ExecuteQuery(scQuery1)
	require.Equal(t, big.NewInt(60).Bytes(), vmOutput1.ReturnData[0])

	nodesBalance := valueToSend - valueToSendToSc
	issuerBalance = issuerBalance - valueToSendToSc
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, issuerBalance)
	for i := 1; i < len(nodes); i++ {
		dcdtCommon.CheckAddressHasTokens(t, nodes[i].OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, nodesBalance)
	}
}

func TestScCallsScWithDcdtIntraShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	// send token issue
	initialSupply := int64(10000000000)
	ticker := "TCK"
	dcdtCommon.IssueTestToken(nodes, initialSupply, ticker)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	// deploy the smart contracts

	vaultCode := wasm.GetSCCode("../testdata/vault-0.41.2.wasm")
	vault, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxData(vaultCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(vault)
	require.Nil(t, err)

	forwarderCode := wasm.GetSCCode("../testdata/forwarder-raw.wasm")
	forwarder, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxData(forwarderCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(forwarder)
	require.Nil(t, err)

	txData := txDataBuilder.NewBuilder()

	// call forwarder with dcdt, and forwarder automatically calls second sc
	valueToSendToSc := int64(1000)
	txData.TransferDCDT(tokenIdentifier, valueToSendToSc)
	txData.Str("forward_async_call_half_payment").Bytes(vault).Str("accept_funds")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIssuerBalance := initialSupply - valueToSendToSc
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, tokenIssuerBalance)
	dcdtCommon.CheckAddressHasTokens(t, forwarder, nodes, []byte(tokenIdentifier), 0, valueToSendToSc/2)
	dcdtCommon.CheckAddressHasTokens(t, vault, nodes, []byte(tokenIdentifier), 0, valueToSendToSc/2)

	dcdtCommon.CheckNumCallBacks(t, forwarder, nodes, 1)
	dcdtCommon.CheckForwarderRawSavedCallbackArgs(t, forwarder, nodes, 1, vmcommon.Ok, [][]byte{})
	dcdtCommon.CheckForwarderRawSavedCallbackPayments(t, forwarder, nodes, []*dcdtCommon.ForwarderRawSavedPaymentInfo{})

	// call forwarder to ask the second one to send it back some dcdt
	valueToRequest := valueToSendToSc / 4
	txData.Clear().Func("forward_async_call").Bytes(vault).Str("retrieve_funds").Str(tokenIdentifier).Int64(0).Int64(valueToRequest)

	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, tokenIssuerBalance)
	dcdtCommon.CheckAddressHasTokens(t, forwarder, nodes, []byte(tokenIdentifier), 0, valueToSendToSc*3/4)
	dcdtCommon.CheckAddressHasTokens(t, vault, nodes, []byte(tokenIdentifier), 0, valueToSendToSc/4)

	dcdtCommon.CheckNumCallBacks(t, forwarder, nodes, 2)
	dcdtCommon.CheckForwarderRawSavedCallbackArgs(t, forwarder, nodes, 2, vmcommon.Ok, [][]byte{})
	dcdtCommon.CheckForwarderRawSavedCallbackPayments(t, forwarder, nodes, []*dcdtCommon.ForwarderRawSavedPaymentInfo{
		{
			TokenId: tokenIdentifier,
			Nonce:   0,
			Payment: big.NewInt(valueToRequest),
		},
	})

	// call forwarder to ask the second one to execute a method
	valueToTransferWithExecSc := valueToSendToSc / 4
	txData.Clear().TransferDCDT(tokenIdentifier, valueToTransferWithExecSc)
	txData.Str("forward_transf_exec").Bytes(vault).Str("accept_funds")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(5 * time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	time.Sleep(5 * time.Second)

	tokenIssuerBalance -= valueToTransferWithExecSc
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, tokenIssuerBalance)
	dcdtCommon.CheckAddressHasTokens(t, forwarder, nodes, []byte(tokenIdentifier), 0, valueToSendToSc*3/4)
	dcdtCommon.CheckAddressHasTokens(t, vault, nodes, []byte(tokenIdentifier), 0, valueToSendToSc/2)

	// call forwarder to ask the second one to execute a method that transfers DCDT twice, with execution
	valueToTransferWithExecSc = valueToSendToSc / 10
	txData.Clear().TransferDCDT(tokenIdentifier, valueToTransferWithExecSc)
	txData.Str("forward_transf_exec_twice").Bytes(vault).Str("accept_funds")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(5 * time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	time.Sleep(5 * time.Second)

	tokenIssuerBalance -= valueToTransferWithExecSc
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, tokenIssuerBalance)
	dcdtCommon.CheckAddressHasTokens(t, forwarder, nodes, []byte(tokenIdentifier), 0, valueToSendToSc*3/4)
	dcdtCommon.CheckAddressHasTokens(t, vault, nodes, []byte(tokenIdentifier), 0, valueToSendToSc/2+valueToTransferWithExecSc)
}

func TestCallbackPaymentRewa(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	// send token issue
	initialSupply := int64(10000000000)
	ticker := "TCK"
	dcdtCommon.IssueTestToken(nodes, initialSupply, ticker)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	// deploy the smart contracts

	vaultCode := wasm.GetSCCode("../testdata/vault.wasm")
	secondScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxData(vaultCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(secondScAddress)
	require.Nil(t, err)

	forwarderCode := wasm.GetSCCode("../testdata/forwarder-raw.wasm")
	forwarder, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxData(forwarderCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(forwarder)
	require.Nil(t, err)

	txData := txDataBuilder.NewBuilder()
	// call first sc with dcdt, and first sc automatically calls second sc
	valueToSendToSc := int64(1000)
	txData.Clear().Func("forward_async_call_half_payment").Bytes(secondScAddress).Str("accept_funds")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(valueToSendToSc), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 1, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckNumCallBacks(t, forwarder, nodes, 1)
	dcdtCommon.CheckForwarderRawSavedCallbackArgs(t, forwarder, nodes, 1, vmcommon.Ok, [][]byte{})
	dcdtCommon.CheckForwarderRawSavedCallbackPayments(t, forwarder, nodes, []*dcdtCommon.ForwarderRawSavedPaymentInfo{})

	// call first sc to ask the second one to send it back some dcdt
	valueToRequest := valueToSendToSc / 4
	txData.Clear().Func("forward_async_call").Bytes(secondScAddress).Str("retrieve_funds").Str("REWA").Int64(0).Int64(valueToRequest)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 1, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckNumCallBacks(t, forwarder, nodes, 2)
	dcdtCommon.CheckForwarderRawSavedCallbackArgs(t, forwarder, nodes, 2, vmcommon.Ok, [][]byte{})
	dcdtCommon.CheckForwarderRawSavedCallbackPayments(t, forwarder, nodes, []*dcdtCommon.ForwarderRawSavedPaymentInfo{
		{
			TokenId: "REWA",
			Nonce:   0,
			Payment: big.NewInt(valueToRequest),
		},
	})
}

func TestScCallsScWithDcdtIntraShard_SecondScRefusesPayment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	// send token issue
	initialSupply := int64(10000000000)
	ticker := "TCK"
	dcdtCommon.IssueTestToken(nodes, initialSupply, ticker)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	// deploy the smart contracts

	secondScCode := wasm.GetSCCode("../testdata/second-contract.wasm")
	secondScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(secondScCode)+"@"+
			hex.EncodeToString([]byte(tokenIdentifier)),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(secondScAddress)
	require.Nil(t, err)

	firstScCode := wasm.GetSCCode("../testdata/first-contract.wasm")
	firstScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(firstScCode)+"@"+
			hex.EncodeToString([]byte(tokenIdentifier))+"@"+
			hex.EncodeToString(secondScAddress),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(firstScAddress)
	require.Nil(t, err)

	nonce, round = transferRejectedBySecondContract(t, nonce, round, nodes, tokenIssuer, idxProposers, initialSupply, tokenIdentifier, firstScAddress, secondScAddress, "transferToSecondContractRejected", 2)
	_, _ = transferRejectedBySecondContract(t, nonce, round, nodes, tokenIssuer, idxProposers, initialSupply, tokenIdentifier, firstScAddress, secondScAddress, "transferToSecondContractRejectedWithTransferAndExecute", 2)
}

func TestScACallsScBWithExecOnDestDCDT_TxPending(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	// send token issue
	initialSupply := int64(10000000000)
	ticker := "TCK"
	dcdtCommon.IssueTestToken(nodes, initialSupply, ticker)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 15
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	// deploy smart contracts

	callerScCode := wasm.GetSCCode("../testdata/exec-on-dest-caller.wasm")
	callerScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(callerScCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(callerScAddress)
	require.Nil(t, err)

	receiverScCode := wasm.GetSCCode("../testdata/exec-on-dest-receiver.wasm")
	receiverScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(receiverScCode)+"@"+
			hex.EncodeToString([]byte(tokenIdentifier)),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(receiverScAddress)
	require.Nil(t, err)

	// set receiver address in caller contract | map[ticker] -> receiverAddress

	txData := txDataBuilder.NewBuilder()
	txData.Clear().
		Func("setPoolAddress").
		Str(tokenIdentifier).
		Str(string(receiverScAddress))

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		callerScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(callerScAddress)
	require.Nil(t, err)

	// issue 1:1 dcdt:interestDcdt in receiver contract

	txData = txDataBuilder.NewBuilder()
	issueTokenSupply := big.NewInt(100000000) // 100 tokens
	issueTokenDecimals := 6
	issuePrice := big.NewInt(1000)
	txData.Clear().
		Func("issue").
		Str(tokenIdentifier).
		Str("token-name").
		Str("L").
		BigInt(issueTokenSupply).
		Int(issueTokenDecimals)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		issuePrice,
		receiverScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// call caller sc with DCDTTransfer which will call the second sc with execute_on_dest_context
	txData = txDataBuilder.NewBuilder()
	valueToTransfer := int64(1000)
	txData.Clear().
		TransferDCDT(tokenIdentifier, valueToTransfer).
		Str("deposit").
		Str(string(callerScAddress))

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		callerScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply-valueToTransfer)

	// should be int64(1000)
	dcdtData := dcdtCommon.GetDCDTTokenData(t, receiverScAddress, nodes, []byte(tokenIdentifier), 0)
	require.EqualValues(t, &dcdt.DCDigitalToken{Value: big.NewInt(valueToTransfer)}, dcdtData)

	// no tokens in caller contract
	dcdtData = dcdtCommon.GetDCDTTokenData(t, callerScAddress, nodes, []byte(tokenIdentifier), 0)
	require.EqualValues(t, &dcdt.DCDigitalToken{Value: big.NewInt(0)}, dcdtData)
}

func TestScACallsScBWithExecOnDestScAPerformsAsyncCall_NoCallbackInScB(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	for _, n := range nodes {
		n.EconomicsData.SetMaxGasLimitPerBlock(1500000000, 0)
	}

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	tokenIssuer := nodes[0]

	// deploy parent contract
	callerScCode := wasm.GetSCCode("../../wasm/testdata/community/parent.wasm")
	callerScAddress, err := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)
	require.Nil(t, err)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(callerScCode),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 10, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(callerScAddress)
	require.Nil(t, err)

	// deploy child contract by calling deployChildContract endpoint
	receiverScCode := wasm.GetSCCode("../../wasm/testdata/community/child.wasm")
	txDeployData := txDataBuilder.NewBuilder()
	txDeployData.Func("deployChildContract").Str(receiverScCode)

	indirectDeploy := "deployChildContract@" + receiverScCode
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		callerScAddress,
		indirectDeploy,
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// issue DCDT by calling exec on dest context on child contract
	ticker := "DSN"
	name := "DisplayName"
	issueCost := big.NewInt(1000)
	txIssueData := txDataBuilder.NewBuilder()
	txIssueData.Func("executeOnDestIssueToken").
		Str(name).
		Str(ticker).
		BigInt(big.NewInt(500000))

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		issueCost,
		callerScAddress,
		txIssueData.ToString(),
		1000000000,
	)

	nrRoundsToPropagateMultiShard := 12
	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenID := integrationTests.GetTokenIdentifier(nodes, []byte(ticker))

	scQuery := nodes[0].SCQueryService
	childScAddressQuery := &process.SCQuery{
		ScAddress:  callerScAddress,
		FuncName:   "getChildContractAddress",
		CallerAddr: nil,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{},
	}

	res, _, err := scQuery.ExecuteQuery(childScAddressQuery)
	require.Nil(t, err)

	receiverScAddress := res.ReturnData[0]

	tokenIdQuery := &process.SCQuery{
		ScAddress:  receiverScAddress,
		FuncName:   "getWrappedRewaTokenIdentifier",
		CallerAddr: nil,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{},
	}

	res, _, err = scQuery.ExecuteQuery(tokenIdQuery)
	require.Nil(t, err)
	require.True(t, strings.Contains(string(res.ReturnData[0]), ticker))

	dcdtCommon.CheckAddressHasTokens(t, receiverScAddress, nodes, tokenID, 0, 500000)
}

func TestExecOnDestWithTokenTransferFromScAtoScBWithIntermediaryExecOnDest_NotEnoughGasInTx(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	enableEpochs := config.EnableEpochs{
		GlobalMintBurnDisableEpoch:              integrationTests.UnreachableEpoch,
		SCProcessorV2EnableEpoch:                integrationTests.UnreachableEpoch,
		FailExecutionOnEveryAPIErrorEnableEpoch: integrationTests.UnreachableEpoch,
	}
	andesVersion := config.WasmVMVersionByEpoch{Version: "v1.4"}
	vmConfig := &config.VirtualMachineConfig{WasmVMVersions: []config.WasmVMVersionByEpoch{andesVersion}}
	nodes := integrationTests.CreateNodesWithEnableEpochsAndVmConfig(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		enableEpochs,
		vmConfig,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	initialSupply := int64(10000000000)
	ticker := "TCK"
	dcdtCommon.IssueTestToken(nodes, initialSupply, ticker)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 15
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	// deploy smart contracts
	mapperScCode := wasm.GetSCCode("../testdata/mapper.wasm")
	mapperScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(mapperScCode),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(mapperScAddress)
	require.Nil(t, err)

	senderScCode := wasm.GetSCCode("../testdata/sender.wasm")
	senderScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(senderScCode),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(senderScAddress)
	require.Nil(t, err)

	txData := txDataBuilder.NewBuilder()
	txData.Func("setRouterAddress").Str(string(mapperScAddress))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		senderScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(senderScAddress)
	require.Nil(t, err)

	receiverScCode := wasm.GetSCCode("../testdata/receiver.wasm")
	receiverScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(receiverScCode)+"@"+
			hex.EncodeToString([]byte(tokenIdentifier))+"@"+
			hex.EncodeToString(senderScAddress),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 12, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(receiverScAddress)
	require.Nil(t, err)

	txData.Clear().Func("setAddress").Str(tokenIdentifier).Str(string(receiverScAddress))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		mapperScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	time.Sleep(time.Second)

	issueCost := big.NewInt(1000)
	txData.Clear().Func("issue").Str(ticker).Str(tokenIdentifier).Str("L")
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		issueCost,
		receiverScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	nrRoundsToPropagateMultiShard = 25
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	scQuery := nodes[0].SCQueryService
	tokenIdQuery := &process.SCQuery{
		ScAddress:  receiverScAddress,
		FuncName:   "lendToken",
		CallerAddr: nil,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{},
	}

	res, _, err := scQuery.ExecuteQuery(tokenIdQuery)
	require.Nil(t, err)
	tokenIdStr := string(res.ReturnData[0])
	require.True(t, strings.Contains(tokenIdStr, ticker))

	txData.Clear().Func("setLendTokenRoles").Int(3).Int(4).Int(5)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		receiverScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	valueToTransfer := int64(1000)
	txData.Clear().
		TransferDCDT(tokenIdentifier, valueToTransfer).
		Str("deposit").
		Str(string(tokenIssuer.OwnAccount.Address))

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		senderScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply-valueToTransfer)
	dcdtData := dcdtCommon.GetDCDTTokenData(t, receiverScAddress, nodes, []byte(tokenIdentifier), 0)
	require.EqualValues(t, &dcdt.DCDigitalToken{Value: big.NewInt(valueToTransfer)}, dcdtData)
}

func TestExecOnDestWithTokenTransferFromScAtoScBWithScCall_GasUsedMismatch(t *testing.T) {
	// TODO add missing required WASM binaries
	t.Skip("accidentally missing required WASM binaries")

	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	initialSupply := int64(10000000000)
	ticker := "BUSD"

	tickerWREWA := "WREWA"
	initialSupplyWREWA := int64(21000000)

	dcdtCommon.IssueTestToken(nodes, initialSupply, ticker)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 15
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	dcdtCommon.IssueTestToken(nodes, initialSupplyWREWA, tickerWREWA)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifierWREWA := string(integrationTests.GetTokenIdentifier(nodes, []byte(tickerWREWA)))
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupplyWREWA)

	// deploy smart contracts
	mapperScCode := wasm.GetSCCode("../testdata/mapperA.wasm")
	mapperScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(mapperScCode),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(mapperScAddress)
	require.Nil(t, err)

	senderScCode := wasm.GetSCCode("../testdata/senderA.wasm")
	senderScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(senderScCode),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(senderScAddress)
	require.Nil(t, err)

	txData := txDataBuilder.NewBuilder()
	txData.Func("setRouterAddress").Str(string(mapperScAddress))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		senderScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(senderScAddress)
	require.Nil(t, err)

	receiverScCode := wasm.GetSCCode("../testdata/receiverA.wasm")
	receiverScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(receiverScCode)+"@"+
			hex.EncodeToString([]byte(tokenIdentifier))+"@"+
			hex.EncodeToString(senderScAddress)+"@01@01@01@01@01",
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 12, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(receiverScAddress)
	require.Nil(t, err)

	receiverScCodeWREWA := wasm.GetSCCode("../testdata/receiverA.wasm")
	receiverScAddressWREWA, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(receiverScCodeWREWA)+"@"+
			hex.EncodeToString([]byte(tokenIdentifierWREWA))+"@"+
			hex.EncodeToString(senderScAddress)+"@01@01@01@01@01",
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 12, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(receiverScAddressWREWA)
	require.Nil(t, err)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	time.Sleep(time.Second)

	issueCost := big.NewInt(1000)
	txData.Clear().Func("issue").Str(ticker).Str(tokenIdentifier).Str("L")
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		issueCost,
		receiverScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	nrRoundsToPropagateMultiShard = 100
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	txData.Clear().Func("issue").Str(ticker).Str(tokenIdentifier).Str("B")
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		issueCost,
		receiverScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	nrRoundsToPropagateMultiShard = 100
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	txData.Clear().Func("issue").Str(tickerWREWA).Str(tokenIdentifierWREWA).Str("L")
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		issueCost,
		receiverScAddressWREWA,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	nrRoundsToPropagateMultiShard = 25
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	txData.Clear().Func("issue").Str(tickerWREWA).Str(tokenIdentifierWREWA).Str("B")
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		issueCost,
		receiverScAddressWREWA,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	nrRoundsToPropagateMultiShard = 25
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	txData.Clear().Func("setTicker").Str(tokenIdentifier).Str(string(receiverScAddress))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		senderScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 400, nonce, round, idxProposers)
	time.Sleep(time.Second)

	txData.Clear().Func("setTicker").Str(tokenIdentifierWREWA).Str(string(receiverScAddressWREWA))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		senderScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)

	scQuery := nodes[0].SCQueryService
	tokenIdQuery := &process.SCQuery{
		ScAddress:  receiverScAddress,
		FuncName:   "lendToken",
		CallerAddr: nil,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{},
	}

	res, _, err := scQuery.ExecuteQuery(tokenIdQuery)
	require.Nil(t, err)
	tokenIdStrLendBusd := string(res.ReturnData[0])
	require.True(t, strings.Contains(tokenIdStrLendBusd, ticker))

	scQuery = nodes[0].SCQueryService
	tokenIdQuery = &process.SCQuery{
		ScAddress:  receiverScAddress,
		FuncName:   "borrowToken",
		CallerAddr: nil,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{},
	}

	res, _, err = scQuery.ExecuteQuery(tokenIdQuery)
	require.Nil(t, err)
	tokenIdStrBorrow := string(res.ReturnData[0])
	require.True(t, strings.Contains(tokenIdStrBorrow, ticker))

	txData.Clear().Func("setLendTokenRoles").Int(3).Int(4).Int(5)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		receiverScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	txData.Clear().Func("setBorrowTokenRoles").Int(3).Int(4).Int(5)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		receiverScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)

	//

	scQuery = nodes[0].SCQueryService
	lendWREWAtokenIdQuery := &process.SCQuery{
		ScAddress:  receiverScAddressWREWA,
		FuncName:   "lendToken",
		CallerAddr: nil,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{},
	}

	borrowWREWAtokenIdQuery := &process.SCQuery{
		ScAddress:  receiverScAddressWREWA,
		FuncName:   "borrowToken",
		CallerAddr: nil,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{},
	}

	res, _, err = scQuery.ExecuteQuery(borrowWREWAtokenIdQuery)
	require.Nil(t, err)
	tokenIdStr := string(res.ReturnData[0])
	require.True(t, strings.Contains(tokenIdStr, tickerWREWA))

	res, _, err = scQuery.ExecuteQuery(lendWREWAtokenIdQuery)
	require.Nil(t, err)
	tokenIdStr = string(res.ReturnData[0])
	require.True(t, strings.Contains(tokenIdStr, tickerWREWA))

	txData.Clear().Func("setLendTokenRoles").Int(3).Int(4).Int(5)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		receiverScAddressWREWA,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	txData.Clear().Func("setBorrowTokenRoles").Int(3).Int(4).Int(5)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		receiverScAddressWREWA,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)

	//
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	valueToTransfer := int64(1000)
	txData.Clear().
		TransferDCDT(tokenIdentifier, valueToTransfer).
		Str("deposit").
		Str(string(tokenIssuer.OwnAccount.Address))

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		senderScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 40, nonce, round, idxProposers)
	time.Sleep(time.Second)

	valueToTransferWREWA := int64(1000)

	txData.Clear().
		TransferDCDT(tokenIdentifierWREWA, valueToTransferWREWA).
		Str("deposit").
		Str(string(tokenIssuer.OwnAccount.Address))

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		senderScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 40, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply-valueToTransfer)
	dcdtData := dcdtCommon.GetDCDTTokenData(t, receiverScAddress, nodes, []byte(tokenIdentifier), 0)
	require.EqualValues(t, &dcdt.DCDigitalToken{Value: big.NewInt(valueToTransfer)}, dcdtData)

	txData.Clear().TransferDCDTNFT(tokenIdStrLendBusd, 1, 100).Str(string(senderScAddress)).Str("borrow").Str(tokenIdentifier).Str(tokenIdentifier)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		nodes[0].OwnAccount.Address,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 25, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtBorrowBUSDData := dcdtCommon.GetDCDTTokenData(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdStrBorrow), 0)
	require.EqualValues(t, &dcdt.DCDigitalToken{Value: big.NewInt(100)}, dcdtBorrowBUSDData)

}

func TestIssueDCDT_FromSCWithNotEnoughGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV3.toml")
	for _, n := range nodes {
		n.EconomicsData.SetMaxGasLimitPerBlock(1500000000, 0)
		if check.IfNil(n.SystemSCFactory) {
			continue
		}
		gasScheduleHandler := n.SystemSCFactory.(core.GasScheduleSubscribeHandler)
		gasScheduleHandler.GasScheduleChange(gasSchedule)
	}

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	scAddress := dcdtCommon.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../testdata/local-dcdt-and-nft.wasm")

	alice := nodes[0]
	issuePrice := big.NewInt(1000)
	txData := []byte("issueFungibleToken" + "@" + hex.EncodeToString([]byte("TOKEN")) +
		"@" + hex.EncodeToString([]byte("TKR")) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()))
	integrationTests.CreateAndSendTransaction(
		alice,
		nodes,
		issuePrice,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)
	time.Sleep(time.Second)

	userAccount := dcdtCommon.GetUserAccountWithAddress(t, alice.OwnAccount.Address, nodes)
	balanceAfterTransfer := userAccount.GetBalance()

	nrRoundsToPropagateMultiShard := 15
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)
	userAccount = dcdtCommon.GetUserAccountWithAddress(t, alice.OwnAccount.Address, nodes)
	require.Equal(t, userAccount.GetBalance(), big.NewInt(0).Add(balanceAfterTransfer, issuePrice))
}

func TestIssueAndBurnDCDT_MaxGasPerBlockExceeded(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numIssues := 22
	numBurns := 50

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	enableEpochs := config.EnableEpochs{
		GlobalMintBurnDisableEpoch:           integrationTests.UnreachableEpoch,
		MaxBlockchainHookCountersEnableEpoch: integrationTests.UnreachableEpoch,
	}
	nodes := integrationTests.CreateNodesWithEnableEpochs(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		enableEpochs,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV3.toml")
	for _, n := range nodes {
		n.EconomicsData.SetMaxGasLimitPerBlock(1500000000, 0)
		if check.IfNil(n.SystemSCFactory) {
			continue
		}
		n.EconomicsData.SetMaxGasLimitPerBlock(15000000000, 0)
		gasScheduleHandler := n.SystemSCFactory.(core.GasScheduleSubscribeHandler)
		gasScheduleHandler.GasScheduleChange(gasSchedule)
	}

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, big.NewInt(0).Mul(initialVal, initialVal))

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// send token issue

	initialSupply := int64(10000000000)
	ticker := "TCK"
	dcdtCommon.IssueTestTokenWithCustomGas(nodes, initialSupply, ticker, 60000000)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	tokenName := "token"
	issuePrice := big.NewInt(1000)

	txData := txDataBuilder.NewBuilder()
	txData.Clear().IssueDCDT(tokenName, ticker, initialSupply, 6)
	txData.CanFreeze(true).CanWipe(true).CanPause(true).CanMint(true).CanBurn(true)
	for i := 0; i < numIssues; i++ {
		integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, issuePrice, vm.DCDTSCAddress, txData.ToString(), 60000000)
	}

	txDataBuilderObj := txDataBuilder.NewBuilder()
	txDataBuilderObj.Clear().Func("DCDTBurn").Str(tokenIdentifier).Int(1)

	burnTxData := txDataBuilderObj.ToString()
	for i := 0; i < numBurns; i++ {
		integrationTests.CreateAndSendTransaction(
			nodes[0],
			nodes,
			big.NewInt(0),
			vm.DCDTSCAddress,
			burnTxData,
			60000000,
		)
	}

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 25, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply-int64(numBurns))

	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() != core.MetachainShardId {
			continue
		}

		scQuery := &process.SCQuery{
			ScAddress:  vm.DCDTSCAddress,
			FuncName:   "getTokenProperties",
			CallerAddr: vm.DCDTSCAddress,
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{[]byte(tokenIdentifier)},
		}
		vmOutput, _, err := n.SCQueryService.ExecuteQuery(scQuery)
		require.Nil(t, err)
		require.Equal(t, vmOutput.ReturnCode, vmcommon.Ok)

		burntValue := big.NewInt(int64(numBurns)).String()
		require.Equal(t, string(vmOutput.ReturnData[4]), burntValue)
	}
}

func TestScCallsScWithDcdtCrossShard_SecondScRefusesPayment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	// send token issue

	initialSupply := int64(10000000000)
	ticker := "TCK"
	dcdtCommon.IssueTestToken(nodes, initialSupply, ticker)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	// deploy the smart contracts

	secondScCode := wasm.GetSCCode("../testdata/second-contract.wasm")
	secondScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(secondScCode)+"@"+
			hex.EncodeToString([]byte(tokenIdentifier)),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(secondScAddress)
	require.Nil(t, err)

	firstScCode := wasm.GetSCCode("../testdata/first-contract.wasm")
	firstScAddress, _ := nodes[2].BlockchainHook.NewAddress(nodes[2].OwnAccount.Address, nodes[2].OwnAccount.Nonce, vmFactory.WasmVirtualMachine)
	integrationTests.CreateAndSendTransaction(
		nodes[2],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(firstScCode)+"@"+
			hex.EncodeToString([]byte(tokenIdentifier))+"@"+
			hex.EncodeToString(secondScAddress),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err = nodes[2].AccntState.GetExistingAccount(firstScAddress)
	require.Nil(t, err)

	nonce, round = transferRejectedBySecondContract(t, nonce, round, nodes, tokenIssuer, idxProposers, initialSupply, tokenIdentifier, firstScAddress, secondScAddress, "transferToSecondContractRejected", 20)
	_, _ = transferRejectedBySecondContract(t, nonce, round, nodes, tokenIssuer, idxProposers, initialSupply, tokenIdentifier, firstScAddress, secondScAddress, "transferToSecondContractRejectedWithTransferAndExecute", 20)
}

func transferRejectedBySecondContract(
	t *testing.T,
	nonce, round uint64,
	nodes []*integrationTests.TestProcessorNode,
	tokenIssuer *integrationTests.TestProcessorNode,
	idxProposers []int,
	initialSupply int64,
	tokenIdentifier string,
	firstScAddress []byte,
	secondScAddress []byte,
	functionToCall string,
	nrRoundToPropagate int,
) (uint64, uint64) {
	// call first sc with dcdt, and first sc automatically calls second sc which returns error
	valueToSendToSc := int64(1000)
	txData := txDataBuilder.NewBuilder()
	txData.Clear().TransferDCDT(tokenIdentifier, valueToSendToSc)
	txData.Str(functionToCall)
	integrationTests.CreateAndSendTransaction(
		tokenIssuer,
		nodes,
		big.NewInt(0),
		firstScAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundToPropagate, nonce, round, idxProposers)
	time.Sleep(time.Second)

	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply-valueToSendToSc)

	dcdtData := dcdtCommon.GetDCDTTokenData(t, firstScAddress, nodes, []byte(tokenIdentifier), 0)
	require.Equal(t, &dcdt.DCDigitalToken{Value: big.NewInt(valueToSendToSc)}, dcdtData)

	dcdtData = dcdtCommon.GetDCDTTokenData(t, secondScAddress, nodes, []byte(tokenIdentifier), 0)
	require.Equal(t, &dcdt.DCDigitalToken{Value: big.NewInt(0)}, dcdtData)

	return nonce, round
}

func TestDCDTMultiTransferFromSC_IntraShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	multiTransferFromSC(t, 1)
}

func TestDCDTMultiTransferFromSC_CrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	multiTransferFromSC(t, 2)
}

func multiTransferFromSC(t *testing.T, numOfShards int) {
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	ownerNode := nodes[0]
	destinationNode := nodes[1]

	ownerShardID := ownerNode.ShardCoordinator.ComputeId(ownerNode.OwnAccount.Address)
	destShardID := ownerNode.ShardCoordinator.ComputeId(destinationNode.OwnAccount.Address)

	if numOfShards > 1 && ownerShardID == destShardID {
		for _, node := range nodes {
			nodeShardID := ownerNode.ShardCoordinator.ComputeId(node.OwnAccount.Address)
			if nodeShardID != ownerShardID {
				destinationNode = node
				break
			}
		}
	}

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

	// send token issue

	initialSupply := int64(10000000000)
	ticker := "TCK"
	dcdtCommon.IssueTestTokenWithSpecialRoles(nodes, initialSupply, ticker)
	tokenIssuer := ownerNode

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := integrationTests.GetTokenIdentifier(nodes, []byte(ticker))
	dcdtCommon.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, 0, initialSupply)

	// deploy the smart contract
	scCode := wasm.GetSCCode("../testdata/multi-transfer-dcdt.wasm")
	scAddress, _ := tokenIssuer.BlockchainHook.NewAddress(
		tokenIssuer.OwnAccount.Address,
		tokenIssuer.OwnAccount.Nonce,
		vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		ownerNode,
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxData(scCode)+"@"+hex.EncodeToString(tokenIdentifier),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := ownerNode.AccntState.GetExistingAccount(scAddress)
	require.Nil(t, err)

	roles := [][]byte{
		[]byte(core.DCDTRoleLocalMint),
	}
	dcdtCommon.SetRoles(nodes, scAddress, tokenIdentifier, roles)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 12, nonce, round, idxProposers)

	txData := txDataBuilder.NewBuilder()
	txData.Func("batchTransferDcdtToken")

	txData.Bytes(destinationNode.OwnAccount.Address)
	txData.Bytes(tokenIdentifier)
	txData.Int64(10)

	txData.Bytes(destinationNode.OwnAccount.Address)
	txData.Bytes(tokenIdentifier)
	txData.Int64(10)

	integrationTests.CreateAndSendTransaction(
		ownerNode,
		nodes,
		big.NewInt(0),
		scAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit,
	)

	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 12, nonce, round, idxProposers)
	dcdtCommon.CheckAddressHasTokens(t, destinationNode.OwnAccount.Address, nodes, tokenIdentifier, 0, 20)
}

func TestDCDTIssueUnderProtectedKeyWillReturnTokensBack(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 2
	numMetachainNodes := 2

	enableEpochs := config.EnableEpochs{
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch: integrationTests.UnreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:              integrationTests.UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch:        integrationTests.UnreachableEpoch,
	}

	nodes := integrationTests.CreateNodesWithEnableEpochs(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		enableEpochs,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := int64(10000000000)
	integrationTests.MintAllNodes(nodes, big.NewInt(initialVal))

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// send token issue

	initialSupply := int64(10000000000)
	ticker := "COIN12345678"
	dcdtCommon.IssueTestToken(nodes, initialSupply, ticker)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 1, nonce, round, idxProposers)
	time.Sleep(time.Second)

	userAcc := dcdtCommon.GetUserAccountWithAddress(t, tokenIssuer.OwnAccount.Address, nodes)
	balanceBefore := userAcc.GetBalance()

	nrRoundsToPropagateMultiShard := 12
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	tokenIdentifier := integrationTests.GetTokenIdentifier(nodes, []byte(ticker))
	require.Equal(t, 0, len(tokenIdentifier))

	tokenPrice := big.NewInt(1000)
	userAcc = dcdtCommon.GetUserAccountWithAddress(t, tokenIssuer.OwnAccount.Address, nodes)
	balanceAfter := userAcc.GetBalance()
	require.Equal(t, balanceAfter, big.NewInt(0).Add(balanceBefore, tokenPrice))
}
