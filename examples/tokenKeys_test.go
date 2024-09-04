package examples

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data/dcdt"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/stretchr/testify/require"
)

func TestComputeTokenStorageKey(t *testing.T) {
	t.Parallel()

	prefix := "NUMBAT"
	require.Equal(t, hex.EncodeToString([]byte(prefix+"dcdtWREWA-bd4d79")), computeStorageKey("WREWA-bd4d79", 0))
	require.Equal(t, hex.EncodeToString([]byte(prefix+"dcdtMYNFT-aaabbbF")), computeStorageKey("MYNFT-aaabbb", 70))
}

func TestDecodeTokenFromProtoBytes(t *testing.T) {
	t.Parallel()

	keyHex := "080112020001222d084612056d794e4654321168747470733a2f2f6d794e46542e636f6d3a0f6d794e465441747472696275746573"
	valueBytes, err := hex.DecodeString(keyHex)
	require.NoError(t, err)

	marshaller := marshal.GogoProtoMarshalizer{}
	recoveredToken := dcdt.DCDigitalToken{}

	err = marshaller.Unmarshal(&recoveredToken, valueBytes)
	require.NoError(t, err)

	expectedToken := dcdt.DCDigitalToken{
		Type:  uint32(core.NonFungible),
		Value: big.NewInt(1),
		TokenMetaData: &dcdt.MetaData{
			Name:       []byte("myNFT"),
			Nonce:      70,
			URIs:       [][]byte{[]byte("https://myNFT.com")},
			Attributes: []byte("myNFTAttributes")},
	}
	require.Equal(t, expectedToken, recoveredToken)
}

func computeStorageKey(tokenIdentifier string, tokenNonce uint64) string {
	key := []byte(core.ProtectedKeyPrefix)
	key = append(key, core.DCDTKeyIdentifier...)
	key = append(key, []byte(tokenIdentifier)...)

	if tokenNonce > 0 {
		nonceBI := big.NewInt(int64(tokenNonce))

		key = append(key, nonceBI.Bytes()...)
	}

	return hex.EncodeToString(key)
}
