package dcdtMultiTransferThroughForwarder

import (
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-go/integrationTests"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm/dcdt"
	multitransfer "github.com/DharitriOne/drt-chain-go/integrationTests/vm/dcdt/multi-transfer"
	"github.com/DharitriOne/drt-chain-go/testscommon/txDataBuilder"
)

func TestDCDTMultiTransferThroughForwarder(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net := integrationTests.NewTestNetworkSized(t, 2, 1, 1)
	net.Start()
	defer net.Close()

	initialVal := uint64(1000000000)
	net.MintNodeAccountsUint64(initialVal)
	net.Step()

	senderNode := net.NodesSharded[0][0]
	destAccount := net.NodesSharded[1][0].OwnAccount

	owner := senderNode.OwnAccount
	forwarder := net.DeployPayableSC(owner, "../../testdata/forwarder.wasm")
	vault := net.DeployNonpayableSC(owner, "../../testdata/vaultV2.wasm")
	vaultOtherShard := net.DeployNonpayableSC(net.NodesSharded[1][0].OwnAccount, "../../testdata/vaultV2.wasm")

	DCDTMultiTransferThroughForwarder_RunStepsAndAsserts(
		t,
		net,
		senderNode,
		senderNode.OwnAccount,
		destAccount,
		forwarder,
		vault,
		vaultOtherShard,
	)
}

func DCDTMultiTransferThroughForwarder_RunStepsAndAsserts(
	t *testing.T,
	net *integrationTests.TestNetwork,
	ownerShard1Node *integrationTests.TestProcessorNode,
	ownerWallet *integrationTests.TestWalletAccount,
	ownerShard2Wallet *integrationTests.TestWalletAccount,
	forwarder []byte,
	vaultShard1 []byte,
	vaultShard2 []byte,
) {
	// Create the fungible token
	supply := int64(1000)
	tokenID := multitransfer.IssueFungibleTokenWithIssuerAddress(t, net, ownerShard1Node, ownerWallet, "FUNG1", supply)

	// Issue and create an SFT
	sftID := multitransfer.IssueNftWithIssuerAddress(net, ownerShard1Node, ownerWallet, "SFT1", true)
	multitransfer.CreateSFT(t, net, ownerShard1Node, ownerWallet, sftID, 1, supply)

	// Send the tokens to the forwarder SC
	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionMultiDCDTNFTTransfer)
	txData.Bytes(forwarder).Int(2)
	txData.Str(tokenID).Int(0).Int64(supply)
	txData.Str(sftID).Int(1).Int64(supply)

	tx := net.CreateTxUint64(ownerWallet, ownerWallet.Address, 0, txData.ToBytes())
	tx.GasLimit = net.MaxGasLimit / 2
	_ = net.SignAndSendTx(ownerWallet, tx)
	net.Steps(4)

	dcdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, supply)
	dcdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, supply)

	// transfer to a user from another shard
	transfers := []*multitransfer.DcdtTransfer{
		{
			TokenIdentifier: tokenID,
			Nonce:           0,
			Amount:          100,
		}}

	multiTransferThroughForwarder(
		net,
		ownerWallet,
		forwarder,
		"multi_transfer_via_async",
		transfers,
		ownerShard2Wallet.Address)

	dcdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 900)
	dcdt.CheckAddressHasTokens(t, ownerShard2Wallet.Address, net.Nodes, []byte(tokenID), 0, 100)

	// transfer to vault, same shard
	multiTransferThroughForwarder(
		net,
		ownerWallet,
		forwarder,
		"forward_sync_accept_funds_multi_transfer",
		transfers,
		vaultShard1)

	dcdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 800)
	dcdt.CheckAddressHasTokens(t, vaultShard1, net.Nodes, []byte(tokenID), 0, 100)

	// transfer fungible and non-fungible
	// transfer to vault, same shard
	transfers = []*multitransfer.DcdtTransfer{
		{
			TokenIdentifier: tokenID,
			Nonce:           0,
			Amount:          100,
		},
		{
			TokenIdentifier: sftID,
			Nonce:           1,
			Amount:          100,
		},
	}
	multiTransferThroughForwarder(
		net,
		ownerWallet,
		forwarder,
		"forward_sync_accept_funds_multi_transfer",
		transfers,
		vaultShard1)

	dcdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 700)
	dcdt.CheckAddressHasTokens(t, vaultShard1, net.Nodes, []byte(tokenID), 0, 200)

	dcdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 900)
	dcdt.CheckAddressHasTokens(t, vaultShard1, net.Nodes, []byte(sftID), 1, 100)

	// transfer fungible and non-fungible
	// transfer to vault, cross shard via transfer and execute
	transfers = []*multitransfer.DcdtTransfer{
		{
			TokenIdentifier: tokenID,
			Nonce:           0,
			Amount:          100,
		},
		{
			TokenIdentifier: sftID,
			Nonce:           1,
			Amount:          100,
		},
	}
	multiTransferThroughForwarder(
		net,
		ownerWallet,
		forwarder,
		"forward_transf_exec_accept_funds_multi_transfer",
		transfers,
		vaultShard2)

	dcdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 600)
	dcdt.CheckAddressHasTokens(t, vaultShard2, net.Nodes, []byte(tokenID), 0, 100)

	dcdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 800)
	dcdt.CheckAddressHasTokens(t, vaultShard2, net.Nodes, []byte(sftID), 1, 100)

	// transfer to vault, cross shard, via async call
	transfers = []*multitransfer.DcdtTransfer{
		{
			TokenIdentifier: tokenID,
			Nonce:           0,
			Amount:          100,
		},
		{
			TokenIdentifier: sftID,
			Nonce:           1,
			Amount:          100,
		},
	}
	multiTransferThroughForwarder(
		net,
		ownerWallet,
		forwarder,
		"multi_transfer_via_async",
		transfers,
		vaultShard2)

	dcdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 500)
	dcdt.CheckAddressHasTokens(t, vaultShard2, net.Nodes, []byte(tokenID), 0, 200)

	dcdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 700)
	dcdt.CheckAddressHasTokens(t, vaultShard2, net.Nodes, []byte(sftID), 1, 200)
}

func multiTransferThroughForwarder(
	net *integrationTests.TestNetwork,
	ownerWallet *integrationTests.TestWalletAccount,
	forwarderAddress []byte,
	function string,
	transfers []*multitransfer.DcdtTransfer,
	destAddress []byte) {

	txData := txDataBuilder.NewBuilder()
	txData.Func(function).Bytes(destAddress)

	for _, transfer := range transfers {
		txData.Str(transfer.TokenIdentifier).Int64(transfer.Nonce).Int64(transfer.Amount)
	}

	tx := net.CreateTxUint64(ownerWallet, forwarderAddress, 0, txData.ToBytes())
	tx.GasLimit = net.MaxGasLimit / 2
	_ = net.SignAndSendTx(ownerWallet, tx)
	net.Steps(10)
}

func TestDCDTMultiTransferWithWrongArgumentsSFT(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net := integrationTests.NewTestNetworkSized(t, 2, 1, 1)
	net.Start()
	defer net.Close()

	initialVal := uint64(1000000000)
	net.MintNodeAccountsUint64(initialVal)
	net.Step()

	senderNode := net.NodesSharded[0][0]
	owner := senderNode.OwnAccount
	forwarder := net.DeployNonpayableSC(owner, "../../testdata/execute/output/execute.wasm")
	vaultOtherShard := net.DeployNonpayableSC(net.NodesSharded[1][0].OwnAccount, "../../testdata/vault.wasm")

	DCDTMultiTransferWithWrongArgumentsSFT_RunStepsAndAsserts(t, net, senderNode, senderNode.OwnAccount, forwarder, vaultOtherShard)
}

func DCDTMultiTransferWithWrongArgumentsSFT_RunStepsAndAsserts(
	t *testing.T,
	net *integrationTests.TestNetwork,
	ownerShard1Node *integrationTests.TestProcessorNode,
	ownerWallet *integrationTests.TestWalletAccount,
	forwarder []byte,
	vaultShard2 []byte) {
	// Issue and create SFT
	supply := int64(1000)
	sftID := multitransfer.IssueNftWithIssuerAddress(net, ownerShard1Node, ownerWallet, "SFT1", true)
	multitransfer.CreateSFT(t, net, ownerShard1Node, ownerWallet, sftID, 1, supply)

	// Send the tokens to the forwarder SC
	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionMultiDCDTNFTTransfer)
	txData.Bytes(forwarder).Int(1)
	txData.Str(sftID).Int(1).Int64(10).Str("doAsyncCall").Bytes(forwarder)
	txData.Bytes([]byte{}).Str(core.BuiltInFunctionMultiDCDTNFTTransfer).Int(6).Bytes(vaultShard2).Int(1).Str(sftID).Int(1).Int(1).Bytes([]byte{})
	tx := net.CreateTxUint64(ownerWallet, ownerWallet.Address, 0, txData.ToBytes())
	tx.GasLimit = net.MaxGasLimit / 2
	_ = net.SignAndSendTx(ownerWallet, tx)
	net.Steps(12)

	dcdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 10)
	dcdt.CheckAddressHasTokens(t, vaultShard2, net.Nodes, []byte(sftID), 1, 0)
	dcdt.CheckAddressHasTokens(t, ownerWallet.Address, net.Nodes, []byte(sftID), 1, supply-10)
}

func TestDCDTMultiTransferWithWrongArgumentsFungible(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net := integrationTests.NewTestNetworkSized(t, 2, 1, 1)
	net.Start()
	defer net.Close()

	initialVal := uint64(1000000000)
	net.MintNodeAccountsUint64(initialVal)
	net.Step()

	senderNode := net.NodesSharded[0][0]
	owner := senderNode.OwnAccount
	forwarder := net.DeployNonpayableSC(owner, "../../testdata/execute/output/execute.wasm")
	vaultOtherShard := net.DeploySCWithInitArgs(net.NodesSharded[1][0].OwnAccount, "../../testdata/contract.wasm", false, []byte{10})

	DCDTMultiTransferWithWrongArgumentsFungible_RunStepsAndAsserts(t, net, senderNode, senderNode.OwnAccount, forwarder, vaultOtherShard)
}

func DCDTMultiTransferWithWrongArgumentsFungible_RunStepsAndAsserts(
	t *testing.T,
	net *integrationTests.TestNetwork,
	senderNode *integrationTests.TestProcessorNode,
	ownerWallet *integrationTests.TestWalletAccount,
	forwarder []byte, vaultOtherShard []byte) {
	// Create the fungible token
	supply := int64(1000)
	tokenID := multitransfer.IssueFungibleTokenWithIssuerAddress(t, net, senderNode, ownerWallet, "FUNG1", supply)

	// Send the tokens to the forwarder SC
	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionMultiDCDTNFTTransfer)
	txData.Bytes(forwarder).Int(1)
	txData.Str(tokenID).Int(0).Int64(80).Str("doAsyncCall").Bytes(forwarder)
	txData.Bytes([]byte{}).Str(core.BuiltInFunctionMultiDCDTNFTTransfer).Int(6).Bytes(vaultOtherShard).Int(1).Str(tokenID).Int(0).Int(42).Bytes([]byte{})
	tx := net.CreateTxUint64(ownerWallet, ownerWallet.Address, 0, txData.ToBytes())
	tx.GasLimit = 104000
	_ = net.SignAndSendTx(ownerWallet, tx)
	net.Steps(12)

	dcdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 80)
	dcdt.CheckAddressHasTokens(t, vaultOtherShard, net.Nodes, []byte(tokenID), 0, 0)
	dcdt.CheckAddressHasTokens(t, ownerWallet.Address, net.Nodes, []byte(tokenID), 0, supply-80)
}
