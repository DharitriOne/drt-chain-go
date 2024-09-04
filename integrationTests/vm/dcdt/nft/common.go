package nft

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-go/integrationTests"
	"github.com/DharitriOne/drt-chain-go/integrationTests/vm/dcdt"
	"github.com/stretchr/testify/require"
)

// NftArguments -
type NftArguments struct {
	Name       []byte
	Quantity   int64
	Royalties  int64
	Hash       []byte
	Attributes []byte
	URI        [][]byte
}

// CreateNFT -
func CreateNFT(tokenIdentifier []byte, issuer *integrationTests.TestProcessorNode, nodes []*integrationTests.TestProcessorNode, args *NftArguments) {
	txData := fmt.Sprintf("%s@%s@%s@%s@%s@%s@%s@%s@",
		core.BuiltInFunctionDCDTNFTCreate,
		hex.EncodeToString(tokenIdentifier),
		hex.EncodeToString(big.NewInt(args.Quantity).Bytes()),
		hex.EncodeToString(args.Name),
		hex.EncodeToString(big.NewInt(args.Royalties).Bytes()),
		hex.EncodeToString(args.Hash),
		hex.EncodeToString(args.Attributes),
		hex.EncodeToString(args.URI[0]),
	)

	integrationTests.CreateAndSendTransaction(issuer, nodes, big.NewInt(0), issuer.OwnAccount.Address, txData, integrationTests.AdditionalGasLimit)
}

// CheckNftData -
func CheckNftData(
	t *testing.T,
	creator []byte,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tickerID []byte,
	args *NftArguments,
	nonce uint64,
) {
	dcdtData := dcdt.GetDCDTTokenData(t, address, nodes, tickerID, nonce)

	if args.Quantity == 0 {
		require.Nil(t, dcdtData.TokenMetaData)
		return
	}

	require.NotNil(t, dcdtData.TokenMetaData)
	require.Equal(t, creator, dcdtData.TokenMetaData.Creator)
	require.Equal(t, args.URI[0], dcdtData.TokenMetaData.URIs[0])
	require.Equal(t, args.Attributes, dcdtData.TokenMetaData.Attributes)
	require.Equal(t, args.Name, dcdtData.TokenMetaData.Name)
	require.Equal(t, args.Hash, dcdtData.TokenMetaData.Hash)
	require.Equal(t, uint32(args.Royalties), dcdtData.TokenMetaData.Royalties)
	require.Equal(t, big.NewInt(args.Quantity).Bytes(), dcdtData.Value.Bytes())
}

// PrepareNFTWithRoles -
func PrepareNFTWithRoles(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxProposers []int,
	nftCreator *integrationTests.TestProcessorNode,
	round *uint64,
	nonce *uint64,
	dcdtType string,
	quantity int64,
	roles [][]byte,
) (string, *NftArguments) {
	dcdt.IssueNFT(nodes, dcdtType, "SFT")

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte("SFT")))

	// ----- set special roles
	dcdt.SetRoles(nodes, nftCreator.OwnAccount.Address, []byte(tokenIdentifier), roles)

	time.Sleep(time.Second)
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	nftMetaData := NftArguments{
		Name:       []byte("nft name"),
		Quantity:   quantity,
		Royalties:  9000,
		Hash:       []byte("hash"),
		Attributes: []byte("attr"),
		URI:        [][]byte{[]byte("uri")},
	}
	CreateNFT([]byte(tokenIdentifier), nftCreator, nodes, &nftMetaData)

	time.Sleep(time.Second)
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, 3, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	CheckNftData(
		t,
		nftCreator.OwnAccount.Address,
		nftCreator.OwnAccount.Address,
		nodes,
		[]byte(tokenIdentifier),
		&nftMetaData,
		1,
	)

	return tokenIdentifier, &nftMetaData
}
