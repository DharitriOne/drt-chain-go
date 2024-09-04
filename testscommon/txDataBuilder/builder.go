package txDataBuilder

import (
	"encoding/hex"
	"math/big"

	"github.com/DharitriOne/drt-chain-core-go/core"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
)

// TxDataBuilder constructs a string to be used for transaction arguments
type TxDataBuilder struct {
	function        string
	elements        []string
	elementsAsBytes [][]byte
	separator       string
}

// NewBuilder creates a new txDataBuilder instance.
func NewBuilder() *TxDataBuilder {
	return &TxDataBuilder{
		function:        "",
		elements:        make([]string, 0),
		elementsAsBytes: make([][]byte, 0),
		separator:       "@",
	}
}

// Clear resets the internal state of the txDataBuilder, allowing a new data
// string to be built.
func (builder *TxDataBuilder) Clear() *TxDataBuilder {
	builder.function = ""
	builder.elements = make([]string, 0)
	builder.elementsAsBytes = make([][]byte, 0)
	return builder
}

// Elements returns the individual elements added to the builder
func (builder *TxDataBuilder) Elements() []string {
	return builder.elements
}

// ElementsAsBytes returns the individual elements added to the builder
func (builder *TxDataBuilder) ElementsAsBytes() [][]byte {
	return builder.elementsAsBytes
}

// Function returns the individual elements added to the builder
func (builder *TxDataBuilder) Function() string {
	return builder.function
}

// ToString returns the data as a string.
func (builder *TxDataBuilder) ToString() string {
	if len(builder.function) > 0 {
		return builder.toStringWithFunction()
	}

	return builder.toStringWithoutFunction()
}

// ToBytes returns the data as a slice of bytes.
func (builder *TxDataBuilder) ToBytes() []byte {
	return []byte(builder.ToString())
}

// GetLast returns the currently last element.
func (builder *TxDataBuilder) GetLast() string {
	if len(builder.elements) == 0 {
		return ""
	}

	return builder.elements[len(builder.elements)-1]
}

// SetLast replaces the last element with the provided one.
func (builder *TxDataBuilder) SetLast(element string) {
	if len(builder.elements) == 0 {
		builder.elements = []string{element}
	}

	builder.elements[len(builder.elements)-1] = element
}

// Func sets the function to be invoked by the data string.
func (builder *TxDataBuilder) Func(function string) *TxDataBuilder {
	builder.function = function

	return builder
}

// Byte appends a single byte to the data string.
func (builder *TxDataBuilder) Byte(value byte) *TxDataBuilder {
	elementAsBytes := []byte{value}
	element := hex.EncodeToString(elementAsBytes)
	builder.elements = append(builder.elements, element)
	builder.elementsAsBytes = append(builder.elementsAsBytes, elementAsBytes)
	return builder
}

// Bytes appends a slice of bytes to the data string.
func (builder *TxDataBuilder) Bytes(bytes []byte) *TxDataBuilder {
	element := hex.EncodeToString(bytes)
	builder.elements = append(builder.elements, element)
	builder.elementsAsBytes = append(builder.elementsAsBytes, bytes)
	return builder
}

// Str appends a string to the data string.
func (builder *TxDataBuilder) Str(str string) *TxDataBuilder {
	elementAsBytes := []byte(str)
	element := hex.EncodeToString(elementAsBytes)
	builder.elements = append(builder.elements, element)
	builder.elementsAsBytes = append(builder.elementsAsBytes, elementAsBytes)
	return builder
}

// Int appends an integer to the data string.
func (builder *TxDataBuilder) Int(value int) *TxDataBuilder {
	elementAsBytes := big.NewInt(int64(value)).Bytes()
	element := hex.EncodeToString(elementAsBytes)
	builder.elements = append(builder.elements, element)
	builder.elementsAsBytes = append(builder.elementsAsBytes, elementAsBytes)
	return builder
}

// Int64 appends an int64 to the data string.
func (builder *TxDataBuilder) Int64(value int64) *TxDataBuilder {
	elementAsBytes := big.NewInt(value).Bytes()
	element := hex.EncodeToString(elementAsBytes)
	builder.elements = append(builder.elements, element)
	builder.elementsAsBytes = append(builder.elementsAsBytes, elementAsBytes)
	return builder
}

// True appends the string "true" to the data string.
func (builder *TxDataBuilder) True() *TxDataBuilder {
	return builder.Str("true")
}

// False appends the string "false" to the data string.
func (builder *TxDataBuilder) False() *TxDataBuilder {
	return builder.Str("false")
}

// Bool appends either "true" or "false" to the data string, depending on the
// `value` argument.
func (builder *TxDataBuilder) Bool(value bool) *TxDataBuilder {
	if value {
		return builder.True()
	}

	return builder.False()
}

// BigInt appends the bytes of a big.Int to the data string.
func (builder *TxDataBuilder) BigInt(value *big.Int) *TxDataBuilder {
	return builder.Bytes(value.Bytes())
}

// IssueDCDT appends to the data string all the elements required to request an DCDT issuing.
func (builder *TxDataBuilder) IssueDCDT(token string, ticker string, supply int64, numDecimals byte) *TxDataBuilder {
	return builder.Func("issue").Str(token).Str(ticker).Int64(supply).Byte(numDecimals)
}

// IssueDCDTWithAsyncArgs appends to the data string all the elements required to request an DCDT issuing.
func (builder *TxDataBuilder) IssueDCDTWithAsyncArgs(token string, ticker string, supply int64, numDecimals byte) *TxDataBuilder {
	return builder.Func("issue").
		Str(token).
		Str(ticker).
		Int64(supply).
		Byte(numDecimals)
}

// TransferDCDT appends to the data string all the elements required to request an DCDT transfer.
func (builder *TxDataBuilder) TransferDCDT(token string, value int64) *TxDataBuilder {
	return builder.Func(core.BuiltInFunctionDCDTTransfer).Str(token).Int64(value)
}

// TransferDCDTNFT appends to the data string all the elements required to request an DCDT NFT transfer.
func (builder *TxDataBuilder) TransferDCDTNFT(token string, nonce int, value int64) *TxDataBuilder {
	return builder.Func(core.BuiltInFunctionDCDTNFTTransfer).Str(token).Int(nonce).Int64(value)
}

// MultiTransferDCDTNFT appends to the data string all the elements required to request an Multi DCDT NFT transfer.
func (builder *TxDataBuilder) MultiTransferDCDTNFT(destinationAddress []byte, transfers []*vmcommon.DCDTTransfer) *TxDataBuilder {
	txBuilder := builder.Func(core.BuiltInFunctionMultiDCDTNFTTransfer).Bytes(destinationAddress).Int(len(transfers))
	for _, transfer := range transfers {
		txBuilder.Bytes(transfer.DCDTTokenName).Int(int(transfer.DCDTTokenNonce)).BigInt(transfer.DCDTValue)
	}
	return txBuilder
}

// BurnDCDT appends to the data string all the elements required to burn DCDT tokens.
func (builder *TxDataBuilder) BurnDCDT(token string, value int64) *TxDataBuilder {
	return builder.Func(core.BuiltInFunctionDCDTBurn).Str(token).Int64(value)
}

// LocalBurnDCDT appends to the data string all the elements required to local burn DCDT tokens.
func (builder *TxDataBuilder) LocalBurnDCDT(token string, value int64) *TxDataBuilder {
	return builder.Func(core.BuiltInFunctionDCDTLocalBurn).Str(token).Int64(value)
}

// LocalMintDCDT appends to the data string all the elements required to local burn DCDT tokens.
func (builder *TxDataBuilder) LocalMintDCDT(token string, value int64) *TxDataBuilder {
	return builder.Func(core.BuiltInFunctionDCDTLocalMint).Str(token).Int64(value)
}

// CanFreeze appends "canFreeze" followed by the provided boolean value.
func (builder *TxDataBuilder) CanFreeze(prop bool) *TxDataBuilder {
	return builder.Str("canFreeze").Bool(prop)
}

// CanWipe appends "canWipe" followed by the provided boolean value.
func (builder *TxDataBuilder) CanWipe(prop bool) *TxDataBuilder {
	return builder.Str("canWipe").Bool(prop)
}

// CanPause appends "canPause" followed by the provided boolean value.
func (builder *TxDataBuilder) CanPause(prop bool) *TxDataBuilder {
	return builder.Str("canPause").Bool(prop)
}

// CanMint appends "canMint" followed by the provided boolean value.
func (builder *TxDataBuilder) CanMint(prop bool) *TxDataBuilder {
	return builder.Str("canMint").Bool(prop)
}

// CanBurn appends "canBurn" followed by the provided boolean value.
func (builder *TxDataBuilder) CanBurn(prop bool) *TxDataBuilder {
	return builder.Str("canBurn").Bool(prop)
}

// CanTransferNFTCreateRole appends "canTransferNFTCreateRole" followed by the provided boolean value.
func (builder *TxDataBuilder) CanTransferNFTCreateRole(prop bool) *TxDataBuilder {
	return builder.Str("canTransferNFTCreateRole").Bool(prop)
}

// CanAddSpecialRoles appends "canAddSpecialRoles" followed by the provided boolean value.
func (builder *TxDataBuilder) CanAddSpecialRoles(prop bool) *TxDataBuilder {
	return builder.Str("canAddSpecialRoles").Bool(prop)
}

// TransferMultiDCDT appends to the data string all the elements required to request an multi DCDT transfer.
func (builder *TxDataBuilder) TransferMultiDCDT(destAddress []byte, args [][]byte) *TxDataBuilder {
	builder.Func(core.BuiltInFunctionMultiDCDTNFTTransfer)
	builder.Bytes(destAddress)
	builder.Int(len(args) / 3) // no of triplets
	for a := 0; a < len(args); a++ {
		builder.Bytes(args[a])
	}
	return builder
}

// IsInterfaceNil returns true if there is no value under the interface
func (builder *TxDataBuilder) IsInterfaceNil() bool {
	return builder == nil
}

func (builder *TxDataBuilder) toStringWithFunction() string {
	data := builder.function
	for _, element := range builder.elements {
		data = data + builder.separator + element
	}

	return data
}

func (builder *TxDataBuilder) toStringWithoutFunction() string {
	data := ""
	for i, element := range builder.elements {
		if i == 0 {
			data = element
			continue
		}
		data = data + builder.separator + element
	}

	return data
}
