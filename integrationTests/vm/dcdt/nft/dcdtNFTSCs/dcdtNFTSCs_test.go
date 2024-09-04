package dcdtNFTSCs

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-go/integrationTests"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm/dcdt"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm/dcdt/nft"
	"github.com/stretchr/testify/require"
)

func TestDCDTNFTIssueCreateBurnSendViaAsyncViaExecuteOnSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := dcdt.CreateNodesAndPrepareBalances(1)

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

	scAddress, tokenIdentifier := deployAndIssueNFTSFTThroughSC(t, nodes, idxProposers, &nonce, &round, "nftIssue", "@03@05")

	txData := []byte("nftCreate" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("name")) +
		"@" + hex.EncodeToString(big.NewInt(10).Bytes()) + "@" + hex.EncodeToString(scAddress) +
		"@" + hex.EncodeToString([]byte("abc")) + "@" + hex.EncodeToString([]byte("NFT")))
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
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 3, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 3, big.NewInt(1))
	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 2, big.NewInt(1))
	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(1))

	txData = []byte("nftBurn" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()))
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
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 3, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 2, big.NewInt(1))
	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(0))

	destinationAddress := nodes[0].OwnAccount.Address
	txData = []byte("transferNftViaAsyncCall" + "@" + hex.EncodeToString(destinationAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + hex.EncodeToString(big.NewInt(2).Bytes()) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("NFT")))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	txData = []byte("transfer_nft_and_execute" + "@" + hex.EncodeToString(destinationAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + hex.EncodeToString(big.NewInt(3).Bytes()) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("NFT")))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 3, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasNft(t, scAddress, destinationAddress, nodes, []byte(tokenIdentifier), 2, big.NewInt(1))
	checkAddressHasNft(t, scAddress, destinationAddress, nodes, []byte(tokenIdentifier), 3, big.NewInt(1))
	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 2, big.NewInt(0))
	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 3, big.NewInt(0))
}

func TestDCDTSemiFTIssueCreateBurnSendViaAsyncViaExecuteOnSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := dcdt.CreateNodesAndPrepareBalances(1)

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

	scAddress, tokenIdentifier := deployAndIssueNFTSFTThroughSC(t, nodes, idxProposers, &nonce, &round, "sftIssue", "@03@04@05")

	txData := []byte("nftCreate" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("name")) +
		"@" + hex.EncodeToString(big.NewInt(10).Bytes()) + "@" + hex.EncodeToString(scAddress) +
		"@" + hex.EncodeToString([]byte("abc")) + "@" + hex.EncodeToString([]byte("NFT")))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	txData = []byte("nftAddQuantity" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString(big.NewInt(10).Bytes()))
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

	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(11))

	txData = []byte("nftBurn" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()))
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

	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(9))

	destinationAddress := nodes[0].OwnAccount.Address
	txData = []byte("transferNftViaAsyncCall" + "@" + hex.EncodeToString(destinationAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()) +
		"@" + hex.EncodeToString(big.NewInt(5).Bytes()) + "@" + hex.EncodeToString([]byte("NFT")))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	txData = []byte("transfer_nft_and_execute" + "@" + hex.EncodeToString(destinationAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()) +
		"@" + hex.EncodeToString(big.NewInt(4).Bytes()) + "@" + hex.EncodeToString([]byte("NFT")))
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

	checkAddressHasNft(t, scAddress, destinationAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(9))
	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(0))
}

func TestDCDTTransferNFTBetweenContractsAcceptAndNotAcceptWithRevert(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := dcdt.CreateNodesAndPrepareBalances(1)

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

	scAddress, tokenIdentifier := deployAndIssueNFTSFTThroughSC(t, nodes, idxProposers, &nonce, &round, "nftIssue", "@03@05")

	txData := []byte("nftCreate" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("name")) +
		"@" + hex.EncodeToString(big.NewInt(10).Bytes()) + "@" + hex.EncodeToString(scAddress) +
		"@" + hex.EncodeToString([]byte("abc")) + "@" + hex.EncodeToString([]byte("NFT")))
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

	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 2, big.NewInt(1))
	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(1))

	destinationSCAddress := dcdt.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../../testdata/nft-receiver.wasm")
	txData = []byte("transferNftViaAsyncCall" + "@" + hex.EncodeToString(destinationSCAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("wrongFunctionToCall")))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	txData = []byte("transfer_nft_and_execute" + "@" + hex.EncodeToString(destinationSCAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + hex.EncodeToString(big.NewInt(2).Bytes()) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("wrongFunctionToCall")))
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

	checkAddressHasNft(t, scAddress, destinationSCAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(0))
	checkAddressHasNft(t, scAddress, destinationSCAddress, nodes, []byte(tokenIdentifier), 2, big.NewInt(0))
	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(1))
	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 2, big.NewInt(1))

	txData = []byte("transferNftViaAsyncCall" + "@" + hex.EncodeToString(destinationSCAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("acceptAndReturnCallData")))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	txData = []byte("transfer_nft_and_execute" + "@" + hex.EncodeToString(destinationSCAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + hex.EncodeToString(big.NewInt(2).Bytes()) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("acceptAndReturnCallData")))
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

	checkAddressHasNft(t, scAddress, destinationSCAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(1))
	checkAddressHasNft(t, scAddress, destinationSCAddress, nodes, []byte(tokenIdentifier), 2, big.NewInt(1))
	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(0))
	checkAddressHasNft(t, scAddress, scAddress, nodes, []byte(tokenIdentifier), 2, big.NewInt(0))
}

func TestDCDTTransferNFTToSCIntraShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := dcdt.CreateNodesAndPrepareBalances(1)

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

	roles := [][]byte{
		[]byte(core.DCDTRoleNFTCreate),
		[]byte(core.DCDTRoleNFTBurn),
	}
	tokenIdentifier, _ := nft.PrepareNFTWithRoles(
		t,
		nodes,
		idxProposers,
		nodes[0],
		&round,
		&nonce,
		core.NonFungibleDCDT,
		1,
		roles,
	)

	nonceArg := hex.EncodeToString(big.NewInt(0).SetUint64(1).Bytes())
	quantityToTransfer := hex.EncodeToString(big.NewInt(1).Bytes())
	destinationSCAddress := dcdt.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../../testdata/nft-receiver.wasm")
	txData := core.BuiltInFunctionDCDTNFTTransfer + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + nonceArg + "@" + quantityToTransfer + "@" + hex.EncodeToString(destinationSCAddress) + "@" + hex.EncodeToString([]byte("acceptAndReturnCallData"))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		nodes[0].OwnAccount.Address,
		txData,
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 3, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasNft(t, nodes[0].OwnAccount.Address, destinationSCAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(1))
}

func TestDCDTTransferNFTToSCCrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := dcdt.CreateNodesAndPrepareBalances(2)

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

	destinationSCAddress := dcdt.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../../testdata/nft-receiver.wasm")

	destinationSCShardID := nodes[0].ShardCoordinator.ComputeId(destinationSCAddress)

	nodeFromOtherShard := nodes[1]
	for _, node := range nodes {
		shID := node.ShardCoordinator.ComputeId(node.OwnAccount.Address)
		if shID != destinationSCShardID {
			nodeFromOtherShard = node
			break
		}
	}

	roles := [][]byte{
		[]byte(core.DCDTRoleNFTCreate),
		[]byte(core.DCDTRoleNFTBurn),
	}
	tokenIdentifier, _ := nft.PrepareNFTWithRoles(
		t,
		nodes,
		idxProposers,
		nodeFromOtherShard,
		&round,
		&nonce,
		core.NonFungibleDCDT,
		1,
		roles,
	)

	nonceArg := hex.EncodeToString(big.NewInt(0).SetUint64(1).Bytes())
	quantityToTransfer := hex.EncodeToString(big.NewInt(1).Bytes())

	txData := core.BuiltInFunctionDCDTNFTTransfer + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + nonceArg + "@" + quantityToTransfer + "@" + hex.EncodeToString(destinationSCAddress) + "@" + hex.EncodeToString([]byte("acceptAndReturnCallData"))
	integrationTests.CreateAndSendTransaction(
		nodeFromOtherShard,
		nodes,
		big.NewInt(0),
		nodeFromOtherShard.OwnAccount.Address,
		txData,
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 10, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasNft(t, nodeFromOtherShard.OwnAccount.Address, destinationSCAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(1))

	txData = core.BuiltInFunctionDCDTNFTTransfer + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + nonceArg + "@" + quantityToTransfer + "@" + hex.EncodeToString(destinationSCAddress) + "@" + hex.EncodeToString([]byte("wrongFunction"))
	integrationTests.CreateAndSendTransaction(
		nodeFromOtherShard,
		nodes,
		big.NewInt(0),
		nodeFromOtherShard.OwnAccount.Address,
		txData,
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 10, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasNft(t, nodeFromOtherShard.OwnAccount.Address, destinationSCAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(1))

	txData = core.BuiltInFunctionDCDTNFTTransfer + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + nonceArg + "@" + quantityToTransfer + "@" + hex.EncodeToString(destinationSCAddress)
	integrationTests.CreateAndSendTransaction(
		nodeFromOtherShard,
		nodes,
		big.NewInt(0),
		nodeFromOtherShard.OwnAccount.Address,
		txData,
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 10, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasNft(t, nodeFromOtherShard.OwnAccount.Address, destinationSCAddress, nodes, []byte(tokenIdentifier), 1, big.NewInt(1))
}

func deployAndIssueNFTSFTThroughSC(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxProposers []int,
	nonce *uint64,
	round *uint64,
	issueFunc string,
	rolesEncoded string,
) ([]byte, string) {
	scAddress := dcdt.DeployNonPayableSmartContract(t, nodes, idxProposers, nonce, round, "../../testdata/local-dcdt-and-nft.wasm")

	issuePrice := big.NewInt(1000)
	txData := []byte(issueFunc + "@" + hex.EncodeToString([]byte("TOKEN")) +
		"@" + hex.EncodeToString([]byte("TKR")))
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
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte("TKR")))
	txData = []byte("setLocalRoles" + "@" + hex.EncodeToString(scAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + rolesEncoded)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)

	time.Sleep(time.Second)
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	return scAddress, tokenIdentifier
}

func checkAddressHasNft(
	t *testing.T,
	creator []byte,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tickerID []byte,
	nonce uint64,
	quantity *big.Int,
) {
	dcdtData := dcdt.GetDCDTTokenData(t, address, nodes, tickerID, nonce)

	if quantity.Cmp(big.NewInt(0)) == 0 {
		require.Nil(t, dcdtData.TokenMetaData)
		return
	}

	require.NotNil(t, dcdtData.TokenMetaData)
	require.Equal(t, creator, dcdtData.TokenMetaData.Creator)
	require.Equal(t, quantity.Bytes(), dcdtData.Value.Bytes())
}
