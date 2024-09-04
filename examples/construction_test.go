package examples

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/core/pubkeyConverter"
	"github.com/DharitriOne/drt-chain-core-go/data/block"
	"github.com/DharitriOne/drt-chain-core-go/data/transaction"
	"github.com/DharitriOne/drt-chain-core-go/hashing/blake2b"
	"github.com/DharitriOne/drt-chain-core-go/hashing/keccak"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/DharitriOne/drt-chain-crypto-go/signing"
	"github.com/DharitriOne/drt-chain-crypto-go/signing/ed25519"
	"github.com/DharitriOne/drt-chain-crypto-go/signing/ed25519/singlesig"
	"github.com/stretchr/testify/require"
)

var (
	addressEncoder, _  = pubkeyConverter.NewBech32PubkeyConverter(32, "moa")
	signingMarshalizer = &marshal.JsonMarshalizer{}
	txSignHasher       = keccak.NewKeccak()
	signer             = &singlesig.Ed25519Signer{}
	signingCryptoSuite = ed25519.NewEd25519()
	contentMarshalizer = &marshal.GogoProtoMarshalizer{}
	contentHasher      = blake2b.NewBlake2b()
)

const alicePrivateKeyHex = "413f42575f7f26fad3317a778771212fdb80245850981e48b58a4f25e344e8f9"

func TestConstructTransaction_NoDataNoValue(t *testing.T) {
	tx := &transaction.Transaction{
		Nonce:    89,
		Value:    big.NewInt(0),
		RcvAddr:  getPubkeyOfAddress(t, "moa1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruq0yu4wk"),
		SndAddr:  getPubkeyOfAddress(t, "moa1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssfq94h8"),
		GasPrice: 1000000000,
		GasLimit: 50000,
		ChainID:  []byte("local-testnet"),
		Version:  1,
	}

	tx.Signature = computeTransactionSignature(t, alicePrivateKeyHex, tx)
	require.Equal(t, "686bcb4057948d1a6109ea27c350807855b7ef874cd6ca0ab353acaec5bfa2c8b42aa4755cd2316c54cbc82d440ca16c4d87b878f5283371be93a5adf838630b", hex.EncodeToString(tx.Signature))

	data, _ := contentMarshalizer.Marshal(tx)
	require.Equal(t, "0859120200001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc0340d08603520d6c6f63616c2d746573746e657458016240686bcb4057948d1a6109ea27c350807855b7ef874cd6ca0ab353acaec5bfa2c8b42aa4755cd2316c54cbc82d440ca16c4d87b878f5283371be93a5adf838630b", hex.EncodeToString(data))

	txHash := contentHasher.Compute(string(data))
	require.Equal(t, "79e4747017865198d73e6281c6cc4a55acaf52d00d06084c7ff3b46798423807", hex.EncodeToString(txHash))
}

func TestConstructTransaction_Usernames(t *testing.T) {
	tx := &transaction.Transaction{
		Nonce:       89,
		Value:       big.NewInt(0),
		RcvAddr:     getPubkeyOfAddress(t, "moa1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruq0yu4wk"),
		SndAddr:     getPubkeyOfAddress(t, "moa1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssfq94h8"),
		GasPrice:    1000000000,
		GasLimit:    50000,
		ChainID:     []byte("local-testnet"),
		Version:     1,
		SndUserName: []byte("alice"),
		RcvUserName: []byte("bob"),
	}

	tx.Signature = computeTransactionSignature(t, alicePrivateKeyHex, tx)
	require.Equal(t, "06af5d563cf22fb5db9086ab8c9708070bddd50adb0f668c60af5e2a7d075e6249a90095938cf8451ad8d673d4019ea97c1de5a67126d6b0c1a46bddffd0b300", hex.EncodeToString(tx.Signature))

	data, _ := contentMarshalizer.Marshal(tx)
	require.Equal(t, "0859120200001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82203626f622a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e13205616c696365388094ebdc0340d08603520d6c6f63616c2d746573746e65745801624006af5d563cf22fb5db9086ab8c9708070bddd50adb0f668c60af5e2a7d075e6249a90095938cf8451ad8d673d4019ea97c1de5a67126d6b0c1a46bddffd0b300", hex.EncodeToString(data))

	txHash := contentHasher.Compute(string(data))
	require.Equal(t, "ee5ea0ca436737165333b045575513e71cb5069103c04bd00020d30af5346b27", hex.EncodeToString(txHash))
}

func TestConstructTransaction_WithDataNoValue(t *testing.T) {
	tx := &transaction.Transaction{
		Nonce:    90,
		Value:    big.NewInt(0),
		RcvAddr:  getPubkeyOfAddress(t, "moa1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruq0yu4wk"),
		SndAddr:  getPubkeyOfAddress(t, "moa1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssfq94h8"),
		GasPrice: 1000000000,
		GasLimit: 80000,
		Data:     []byte("hello"),
		ChainID:  []byte("local-testnet"),
		Version:  1,
	}

	tx.Signature = computeTransactionSignature(t, alicePrivateKeyHex, tx)
	require.Equal(t, "096eb86b07034d19dba855e571d34860684c6e46a8dd551ce16ea7257b08f1952af1df3f58085077e65d1ef1a4abb3cc55c44d761a4547341f1628c705f7b001", hex.EncodeToString(tx.Signature))

	data, _ := contentMarshalizer.Marshal(tx)
	require.Equal(t, "085a120200001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc034080f1044a0568656c6c6f520d6c6f63616c2d746573746e657458016240096eb86b07034d19dba855e571d34860684c6e46a8dd551ce16ea7257b08f1952af1df3f58085077e65d1ef1a4abb3cc55c44d761a4547341f1628c705f7b001", hex.EncodeToString(data))

	txHash := contentHasher.Compute(string(data))
	require.Equal(t, "9ebce74d595b54011e7584b702b1ab8002f210567e7f183fc4014d43d019feb6", hex.EncodeToString(txHash))
}

func TestConstructTransaction_WithDataWithValue(t *testing.T) {
	tx := &transaction.Transaction{
		Nonce:    91,
		Value:    stringToBigInt("10000000000000000000"),
		RcvAddr:  getPubkeyOfAddress(t, "moa1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruq0yu4wk"),
		SndAddr:  getPubkeyOfAddress(t, "moa1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssfq94h8"),
		GasPrice: 1000000000,
		GasLimit: 100000,
		Data:     []byte("for the book"),
		ChainID:  []byte("local-testnet"),
		Version:  1,
	}

	tx.Signature = computeTransactionSignature(t, alicePrivateKeyHex, tx)
	require.Equal(t, "498e84fdf4c9b7c1bf2dc3949a10d6e8a3f06cebe73ef8759dc211e8a5c22fa88496257c4ba8d752abf025a44f1cb959a3817c6a2e098f98f470c87b111c500a", hex.EncodeToString(tx.Signature))

	data, _ := contentMarshalizer.Marshal(tx)
	require.Equal(t, "085b1209008ac7230489e800001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc0340a08d064a0c666f722074686520626f6f6b520d6c6f63616c2d746573746e657458016240498e84fdf4c9b7c1bf2dc3949a10d6e8a3f06cebe73ef8759dc211e8a5c22fa88496257c4ba8d752abf025a44f1cb959a3817c6a2e098f98f470c87b111c500a", hex.EncodeToString(data))

	txHash := contentHasher.Compute(string(data))
	require.Equal(t, "db291d86db9e79725feeade579540e79f46bbacbfe4a621499693e4474920e67", hex.EncodeToString(txHash))
}

func TestConstructTransaction_WithDataWithLargeValue(t *testing.T) {
	tx := &transaction.Transaction{
		Nonce:    92,
		Value:    stringToBigInt("123456789000000000000000000000"),
		RcvAddr:  getPubkeyOfAddress(t, "moa1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruq0yu4wk"),
		SndAddr:  getPubkeyOfAddress(t, "moa1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssfq94h8"),
		GasPrice: 1000000000,
		GasLimit: 100000,
		Data:     []byte("for the spaceship"),
		ChainID:  []byte("local-testnet"),
		Version:  1,
	}

	tx.Signature = computeTransactionSignature(t, alicePrivateKeyHex, tx)
	require.Equal(t, "7daf71024964d2c569b59f03689f5afe9a0012edddc290a653f0d76bb981a25538c2d44f467cb70e9a58745f099311090977d18e165af6a4b113527f46321006", hex.EncodeToString(tx.Signature))

	data, _ := contentMarshalizer.Marshal(tx)
	require.Equal(t, "085c120e00018ee90ff6181f3761632000001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc0340a08d064a11666f722074686520737061636573686970520d6c6f63616c2d746573746e6574580162407daf71024964d2c569b59f03689f5afe9a0012edddc290a653f0d76bb981a25538c2d44f467cb70e9a58745f099311090977d18e165af6a4b113527f46321006", hex.EncodeToString(data))

	txHash := contentHasher.Compute(string(data))
	require.Equal(t, "d65273c27347f393904db17725f8db945ffe513e535b49eb7d369c5b70cf2e41", hex.EncodeToString(txHash))
}

func TestConstructTransaction_WithGuardianFields(t *testing.T) {
	tx := &transaction.Transaction{
		Nonce:    92,
		Value:    stringToBigInt("123456789000000000000000000000"),
		RcvAddr:  getPubkeyOfAddress(t, "moa1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruq0yu4wk"),
		SndAddr:  getPubkeyOfAddress(t, "moa1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssfq94h8"),
		GasPrice: 1000000000,
		GasLimit: 150000,
		Data:     []byte("test data field"),
		ChainID:  []byte("local-testnet"),
		Version:  2,
		Options:  2,
	}

	tx.GuardianAddr = getPubkeyOfAddress(t, "moa1x23lzn8483xs2su4fak0r0dqx6w38enpmmqf2yrkylwq7mfnvyhstcgmz5")
	tx.GuardianSignature = bytes.Repeat([]byte{0}, 64)

	tx.Signature = computeTransactionSignature(t, alicePrivateKeyHex, tx)
	require.Equal(t, "c74c3e5b276c32ae72ab1a0bac17939d7577e55e1467cd3b4d4a45b04ad3957e94a855657b42601384819502882d559dd5f8e31a93097e113fe9fa8261515104", hex.EncodeToString(tx.Signature))

	data, _ := contentMarshalizer.Marshal(tx)
	require.Equal(t, "085c120e00018ee90ff6181f3761632000001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc0340f093094a0f746573742064617461206669656c64520d6c6f63616c2d746573746e657458026240c74c3e5b276c32ae72ab1a0bac17939d7577e55e1467cd3b4d4a45b04ad3957e94a855657b42601384819502882d559dd5f8e31a93097e113fe9fa82615151046802722032a3f14cf53c4d0543954f6cf1bda0369d13e661dec095107627dc0f6d33612f7a4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(data))

	txHash := contentHasher.Compute(string(data))
	require.Equal(t, "0896751a0c0eb3316041295ae1d71a4200aef5b359609b0cc181cae4e22b531e", hex.EncodeToString(txHash))
}

func TestConstructTransaction_WithNonceZero(t *testing.T) {
	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  getPubkeyOfAddress(t, "moa1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruq0yu4wk"),
		SndAddr:  getPubkeyOfAddress(t, "moa1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssfq94h8"),
		GasPrice: 1000000000,
		GasLimit: 80000,
		Data:     []byte("hello"),
		ChainID:  []byte("local-testnet"),
		Version:  1,
	}

	tx.Signature = computeTransactionSignature(t, alicePrivateKeyHex, tx)
	require.Equal(t, "d73e9c2f978d248eaba41c2453088f0e3488c02eb88ede7b1f40a22527a0f90c747056163a1625c127b3b6bd602a3a0cb607478a45d0a847e471911b3b87e805", hex.EncodeToString(tx.Signature))

	data, _ := contentMarshalizer.Marshal(tx)
	require.Equal(t, "120200001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc034080f1044a0568656c6c6f520d6c6f63616c2d746573746e657458016240d73e9c2f978d248eaba41c2453088f0e3488c02eb88ede7b1f40a22527a0f90c747056163a1625c127b3b6bd602a3a0cb607478a45d0a847e471911b3b87e805", hex.EncodeToString(data))

	txHash := contentHasher.Compute(string(data))
	require.Equal(t, "18361f53a01d70faa43b563c0da4ea95444abe85ffdce980131cee26cd7ba06d", hex.EncodeToString(txHash))
}

func stringToBigInt(input string) *big.Int {
	result := big.NewInt(0)
	_, _ = result.SetString(input, 10)
	return result
}

func getPubkeyOfAddress(t *testing.T, address string) []byte {
	pubkey, err := addressEncoder.Decode(address)
	require.Nil(t, err)
	return pubkey
}

func computeTransactionSignature(t *testing.T, senderSeedHex string, tx *transaction.Transaction) []byte {
	keyGenerator := signing.NewKeyGenerator(signingCryptoSuite)

	senderSeed, err := hex.DecodeString(senderSeedHex)
	require.Nil(t, err)

	privateKey, err := keyGenerator.PrivateKeyFromByteArray(senderSeed)
	require.Nil(t, err)

	dataToSign, err := tx.GetDataForSigning(addressEncoder, signingMarshalizer, txSignHasher)
	require.Nil(t, err)

	signature, err := signer.Sign(privateKey, dataToSign)
	require.Nil(t, err)
	require.Len(t, signature, 64)

	return signature
}

func TestConstructMiniBlockHeaderReserved_WithMaxValues(t *testing.T) {
	mbhr := &block.MiniBlockHeaderReserved{
		ExecutionType: block.ProcessingType(math.MaxInt32),
		State:         block.MiniBlockState(math.MaxInt32),
	}

	data, _ := contentMarshalizer.Marshal(mbhr)
	fmt.Printf("size: %d\n", len(data))
	require.True(t, len(data) < 16)
}
