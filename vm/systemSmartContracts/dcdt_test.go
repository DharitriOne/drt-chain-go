package systemSmartContracts

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/core"
	vmData "github.com/DharitriOne/drt-chain-core-go/data/vm"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/testscommon"
	"github.com/DharitriOne/drt-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/DharitriOne/drt-chain-go/testscommon/hashingMocks"
	"github.com/DharitriOne/drt-chain-go/vm"
	"github.com/DharitriOne/drt-chain-go/vm/mock"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgumentsForDCDT() ArgsNewDCDTSmartContract {
	return ArgsNewDCDTSmartContract{
		Eei:     &mock.SystemEIStub{},
		GasCost: vm.GasCost{MetaChainSystemSCsCost: vm.MetaChainSystemSCsCost{DCDTIssue: 10}},
		DCDTSCConfig: config.DCDTSystemSCConfig{
			BaseIssuingCost: "1000",
		},
		DCDTSCAddress:          []byte("address"),
		Marshalizer:            &mock.MarshalizerMock{},
		Hasher:                 &hashingMocks.HasherMock{},
		AddressPubKeyConverter: testscommon.NewPubkeyConverterMock(32),
		EndOfEpochSCAddress:    vm.EndOfEpochAddress,
		EnableEpochsHandler: enableEpochsHandlerMock.NewEnableEpochsHandlerStub(
			common.DCDTFlag,
			common.GlobalMintBurnFlag,
			common.MetaDCDTSetFlag,
			common.DCDTRegisterAndSetAllRolesFlag,
			common.DCDTNFTCreateOnMultiShardFlag,
			common.DCDTTransferRoleFlag,
			common.DCDTMetadataContinuousCleanupFlag,
		),
	}
}

func TestNewDCDTSmartContract(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	e, err := NewDCDTSmartContract(args)
	ky := hex.EncodeToString([]byte("NUMBATdcdttxgenDCDTtkn"))
	fmt.Println(ky)

	assert.Nil(t, err)
	assert.NotNil(t, e)
}

func TestNewDCDTSmartContract_NilEEIShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	args.Eei = nil

	e, err := NewDCDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewDCDTSmartContract_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	args.Marshalizer = nil

	e, err := NewDCDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrNilMarshalizer, err)
}

func TestNewDCDTSmartContract_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	args.Hasher = nil

	e, err := NewDCDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrNilHasher, err)
}

func TestNewDCDTSmartContract_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	args.EnableEpochsHandler = nil

	e, err := NewDCDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrNilEnableEpochsHandler, err)
}

func TestNewDCDTSmartContract_InvalidEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()

	e, err := NewDCDTSmartContract(args)
	assert.Nil(t, e)
	assert.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
}

func TestNewDCDTSmartContract_NilPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	args.AddressPubKeyConverter = nil

	e, err := NewDCDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrNilAddressPubKeyConverter, err)
}

func TestNewDCDTSmartContract_BaseIssuingCostLessThanZeroShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	args.DCDTSCConfig.BaseIssuingCost = "-1"

	e, err := NewDCDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrInvalidBaseIssuingCost, err)
}

func TestNewDCDTSmartContract_InvalidBaseIssuingCostShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	args.DCDTSCConfig.BaseIssuingCost = "invalid cost"

	e, err := NewDCDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrInvalidBaseIssuingCost, err)
}

func TestDcdt_ExecuteIssueAlways6charactersForRandom(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewDCDTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("addr"),
			CallValue:   big.NewInt(0),
			GasProvided: 100000,
		},
		RecipientAddr: []byte("addr"),
		Function:      "issueNonFungible",
	}
	eei.gasRemaining = vmInput.GasProvided
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.DCDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.DCDTIssue
	ticker := []byte("TICKER")
	vmInput.Arguments = [][]byte{[]byte("name"), ticker}

	randomWithPreprendedZeros := make([]byte, 32)
	randomWithPreprendedZeros[2] = 1
	e.hasher = &mock.HasherStub{
		ComputeCalled: func(s string) []byte {
			return randomWithPreprendedZeros
		},
	}

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	lastOutput := eei.output[len(eei.output)-1]
	assert.Equal(t, len(lastOutput), len(ticker)+1+6)

	vmInput.Function = "issueSemiFungible"
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	lastOutput = eei.output[len(eei.output)-1]
	assert.Equal(t, len(lastOutput), len(ticker)+1+6)

	vmInput.Arguments = nil
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestDcdt_ExecuteIssueWithMultiNFTCreate(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	e, _ := NewDCDTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("addr"),
			CallValue:   big.NewInt(0),
			GasProvided: 100000,
		},
		RecipientAddr: []byte("addr"),
		Function:      "issue",
	}
	eei.gasRemaining = vmInput.GasProvided
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.DCDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.DCDTIssue
	ticker := []byte("TICKER")
	vmInput.Arguments = [][]byte{[]byte("name"), ticker, []byte(canCreateMultiShard), []byte("true")}

	enableEpochsHandler.RemoveActiveFlags(common.DCDTNFTCreateOnMultiShardFlag)
	returnCode := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)

	enableEpochsHandler.AddActiveFlags(common.DCDTNFTCreateOnMultiShardFlag)
	returnCode = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)

	vmInput.Function = "issueSemiFungible"
	returnCode = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, returnCode)

	upgradePropertiesLog := eei.logs[0]
	expectedTopics := [][]byte{[]byte("TICKER-75fd57"), big.NewInt(0).Bytes(), []byte(canCreateMultiShard), boolToSlice(true), []byte(upgradable), boolToSlice(true), []byte(canAddSpecialRoles), boolToSlice(true)}
	assert.Equal(t, &vmcommon.LogEntry{
		Identifier: []byte(upgradeProperties),
		Address:    []byte("addr"),
		Topics:     expectedTopics,
	}, upgradePropertiesLog)

	lastOutput := eei.output[len(eei.output)-1]
	token, _ := e.getExistingToken(lastOutput)
	assert.True(t, token.CanCreateMultiShard)
}

func TestDcdt_ExecuteIssue(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewDCDTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("addr"),
			CallValue:   big.NewInt(0),
			GasProvided: 100000,
		},
		RecipientAddr: []byte("addr"),
		Function:      "issue",
	}
	eei.gasRemaining = vmInput.GasProvided
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments = [][]byte{[]byte("name"), []byte("TICKER")}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(100).Bytes())
	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(10).Bytes())
	vmInput.Arguments = append(vmInput.Arguments, []byte(upgradable), boolToSlice(false))
	vmInput.Arguments = append(vmInput.Arguments, []byte(canAddSpecialRoles), boolToSlice(false))
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.DCDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.DCDTIssue
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	upgradePropertiesLog := eei.logs[0]
	expectedTopics := [][]byte{[]byte("TICKER-75fd57"), big.NewInt(0).Bytes(), []byte(upgradable), boolToSlice(false), []byte(canAddSpecialRoles), boolToSlice(false)}
	assert.Equal(t, &vmcommon.LogEntry{
		Identifier: []byte(upgradeProperties),
		Address:    []byte("addr"),
		Topics:     expectedTopics,
	}, upgradePropertiesLog)

	vmInput.Arguments[0] = []byte("01234567891&*@")
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestDcdt_ExecuteIssueWithZero(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	e, _ := NewDCDTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("addr"),
			CallValue:   big.NewInt(0),
			GasProvided: 100000,
		},
		RecipientAddr: []byte("addr"),
		Function:      "issue",
	}
	eei.gasRemaining = vmInput.GasProvided
	vmInput.Arguments = [][]byte{[]byte("name"), []byte("TICKER")}
	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(0).Bytes())
	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(10).Bytes())
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.DCDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.DCDTIssue

	enableEpochsHandler.RemoveActiveFlags(common.GlobalMintBurnFlag, common.DCDTNFTCreateOnMultiShardFlag)
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDcdt_ExecuteIssueTooMuchSupply(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewDCDTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("addr"),
			CallValue:   big.NewInt(0),
			GasProvided: 100000,
		},
		RecipientAddr: []byte("addr"),
		Function:      "issue",
	}
	eei.gasRemaining = vmInput.GasProvided

	vmInput.Arguments = [][]byte{[]byte("name"), []byte("TICKER")}
	tooMuchToIssue := make([]byte, 101)
	tooMuchToIssue[0] = 1
	vmInput.Arguments = append(vmInput.Arguments, tooMuchToIssue)
	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(10).Bytes())
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.DCDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.DCDTIssue
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestDcdt_IssueInvalidNumberOfDecimals(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewDCDTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("addr"),
			CallValue:   big.NewInt(0),
			GasProvided: 100000,
		},
		RecipientAddr: []byte("addr"),
		Function:      "issue",
	}
	eei.gasRemaining = vmInput.GasProvided
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments = [][]byte{[]byte("name"), []byte("TICKER")}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(100).Bytes())
	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(25).Bytes())
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.DCDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.DCDTIssue
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestDcdt_ExecuteNilArgsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	e, _ := NewDCDTSmartContract(args)

	output := e.Execute(nil)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestDcdt_ExecuteInit(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	e, _ := NewDCDTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     []byte("addr"),
			Arguments:      nil,
			CallValue:      big.NewInt(0),
			CallType:       0,
			GasPrice:       0,
			GasProvided:    0,
			OriginalTxHash: nil,
			CurrentTxHash:  nil,
		},
		RecipientAddr: []byte("addr"),
		Function:      "_init",
	}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDcdt_ExecuteWrongFunctionCall(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	args.DCDTSCConfig.OwnerAddress = "owner"
	e, _ := NewDCDTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     []byte("addr"),
			Arguments:      nil,
			CallValue:      big.NewInt(0),
			CallType:       0,
			GasPrice:       0,
			GasProvided:    0,
			OriginalTxHash: nil,
			CurrentTxHash:  nil,
		},
		RecipientAddr: []byte("addr"),
		Function:      "wrong function",
	}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionNotFound, output)
}

func TestDcdt_ExecuteBurnWrongNumOfArgsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionDCDTBurn, [][]byte{[]byte("dcdtToken"), {100}})
	vmInput.Arguments = [][]byte{[]byte("wrong_token_name")}

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "number of arguments must be equal with 2"))
}

func TestDcdt_ExecuteBurnWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionDCDTBurn, [][]byte{[]byte("dcdtToken"), {100}})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestDcdt_ExecuteBurnWrongValueToBurnShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionDCDTBurn, [][]byte{[]byte("dcdtToken"), {100}})
	vmInput.Arguments[1] = []byte{0}

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "negative or 0 value to burn"))
}

func TestDcdt_ExecuteBurnOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionDCDTBurn, [][]byte{[]byte("dcdtToken"), {100}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestDcdt_ExecuteBurnAndMintDisabled(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.RemoveActiveFlags(common.GlobalMintBurnFlag)
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionDCDTBurn, [][]byte{[]byte("dcdtToken"), {100}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "global burn is no more enabled, use local burn"))

	vmInput = getDefaultVmInputForFunc("mint", [][]byte{[]byte("dcdtToken"), {100}})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "global mint is no more enabled, use local mint"))
}

func TestDcdt_ExecuteBurnOnNonBurnableTokenShouldWorkAndReturnBurntTokens(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		Burnable: false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	burnValue := []byte{100}
	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionDCDTBurn, [][]byte{tokenName, burnValue})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.True(t, strings.Contains(eei.returnMessage, "token is not burnable"))

	outputTransfer := eei.outputAccounts["owner"].OutputTransfers[0]
	expectedReturnData := []byte(core.BuiltInFunctionDCDTTransfer + "@" + hex.EncodeToString(tokenName) + "@" + hex.EncodeToString(burnValue))
	assert.Equal(t, expectedReturnData, outputTransfer.Data)
}

func TestDcdt_ExecuteBurn(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:  tokenName,
		Burnable:   true,
		BurntValue: big.NewInt(100),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionDCDTBurn, [][]byte{[]byte("dcdtToken"), {100}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dcdtData := &DCDTDataV2{}
	_ = args.Marshalizer.Unmarshal(dcdtData, eei.GetStorage(tokenName))
	assert.Equal(t, big.NewInt(200), dcdtData.BurntValue)
}

func TestDcdt_ExecuteMintTooFewArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{[]byte("dcdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "accepted arguments number 2/3"))
}

func TestDcdt_ExecuteMintTooManyArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{[]byte("dcdtToken"), {200}, []byte("dest"), []byte("arg")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "accepted arguments number 2/3"))
}

func TestDcdt_ExecuteMintWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{[]byte("dcdtToken"), {200}})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestDcdt_ExecuteMintNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.DCDTOperations = 10

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{[]byte("dcdtToken"), {200}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestDcdt_ExecuteMintOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{[]byte("dcdtToken"), {200}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestDcdt_ExecuteMintNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: []byte("random address"),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, {200}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestDcdt_ExecuteMintWrongMintValueShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: []byte("owner"),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, {0}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "negative or zero mint value"))
}

func TestDcdt_ExecuteMintNonMintableTokenShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: []byte("owner"),
		Mintable:     false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, {200}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "token is not mintable"))
}

func TestDcdt_ExecuteMintSavesTokenWithMintedTokensAdded(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    []byte("dcdtToken"),
		OwnerAddress: []byte("owner"),
		Mintable:     true,
		MintedValue:  big.NewInt(100),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, {200}})

	_ = e.Execute(vmInput)

	dcdtData := &DCDTDataV2{}
	_ = args.Marshalizer.Unmarshal(dcdtData, eei.GetStorage(tokenName))
	assert.Equal(t, big.NewInt(300), dcdtData.MintedValue)

	vmInput.Arguments[1] = make([]byte, 101)
	returnCode := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestDcdt_ExecuteMintInvalidDestinationAddressShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: []byte("owner"),
		Mintable:     true,
		MintedValue:  big.NewInt(100),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, {200}, []byte("dest")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "destination address of invalid length"))
}

func TestDcdt_ExecuteMintTransferNoErr(t *testing.T) {
	t.Parallel()

	err := errors.New("transfer error")
	args := createMockArgumentsForDCDT()
	args.Eei.(*mock.SystemEIStub).GetStorageCalled = func(key []byte) []byte {
		marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
			OwnerAddress: []byte("owner"),
			Mintable:     true,
			MintedValue:  big.NewInt(100),
		})
		return marshalizedData
	}
	args.Eei.(*mock.SystemEIStub).AddReturnMessageCalled = func(msg string) {
		assert.Equal(t, err.Error(), msg)
	}

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{[]byte("dcdtToken"), {200}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDcdt_ExecuteMintWithTwoArgsShouldSetOwnerAsDestination(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	mintValue := []byte{200}
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		Mintable:     true,
		MintedValue:  big.NewInt(100),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, mintValue})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.DCDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(owner)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionDCDTTransfer + "@" + hex.EncodeToString(tokenName) + "@" + hex.EncodeToString(mintValue)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestDcdt_ExecuteMintWithThreeArgsShouldSetThirdArgAsDestination(t *testing.T) {
	t.Parallel()

	dest := []byte("_dest")
	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	mintValue := []byte{200}
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		Mintable:     true,
		MintedValue:  big.NewInt(100),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, mintValue, dest})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.DCDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(dest)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionDCDTTransfer + "@" + hex.EncodeToString(tokenName) + "@" + hex.EncodeToString(mintValue)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestDcdt_ExecuteIssueDisabled(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.RemoveActiveFlags(common.DCDTFlag)
	e, _ := NewDCDTSmartContract(args)

	callValue, _ := big.NewInt(0).SetString(args.DCDTSCConfig.BaseIssuingCost, 10)
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     []byte("addr"),
			Arguments:      [][]byte{[]byte("01234567891")},
			CallValue:      callValue,
			CallType:       0,
			GasPrice:       0,
			GasProvided:    args.GasCost.MetaChainSystemSCsCost.DCDTIssue,
			OriginalTxHash: nil,
			CurrentTxHash:  nil,
		},
		RecipientAddr: []byte("addr"),
		Function:      "issue",
	}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestDcdt_ExecuteToggleFreezeTooFewArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("dcdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 2"))

	vmInput.Function = "freezeSingleNFT"
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 3"))
}

func TestDcdt_ExecuteToggleFreezeWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("dcdtToken"), []byte("owner")})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))

	vmInput.Function = "freezeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("owner"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestDcdt_ExecuteToggleFreezeNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.DCDTOperations = 10

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("dcdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))

	vmInput.Function = "freezeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("owner"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestDcdt_ExecuteToggleFreezeOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("dcdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))

	vmInput.Function = "freezeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("owner"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestDcdt_ExecuteToggleFreezeNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := "dcdtToken"
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: []byte("random address"),
	})
	tokensMap[tokenName] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte(tokenName), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))

	vmInput.Function = "freezeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("owner"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestDcdt_ExecuteToggleFreezeNonFreezableTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: owner,
		CanFreeze:    false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{tokenName, owner})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot freeze"))

	vmInput.Function = "freezeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, owner)
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot freeze"))
}

func TestDcdt_ExecuteToggleFreezeTransferNoErr(t *testing.T) {
	t.Parallel()

	err := errors.New("transfer error")
	args := createMockArgumentsForDCDT()
	args.Eei.(*mock.SystemEIStub).GetStorageCalled = func(key []byte) []byte {
		marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
			OwnerAddress: []byte("owner"),
			CanFreeze:    true,
		})
		return marshalizedData
	}
	args.Eei.(*mock.SystemEIStub).AddReturnMessageCalled = func(msg string) {
		assert.Equal(t, err.Error(), msg)
	}

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("dcdtToken"), getAddress()})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDcdt_ExecuteToggleFreezeSingleNFTTransferNoErr(t *testing.T) {
	t.Parallel()

	err := errors.New("transfer error")
	args := createMockArgumentsForDCDT()
	args.Eei.(*mock.SystemEIStub).GetStorageCalled = func(key []byte) []byte {
		marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
			OwnerAddress: []byte("owner"),
			CanFreeze:    true,
			TokenType:    []byte(core.NonFungibleDCDT),
		})
		return marshalizedData
	}
	args.Eei.(*mock.SystemEIStub).AddReturnMessageCalled = func(msg string) {
		assert.Equal(t, err.Error(), msg)
	}

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freezeSingleNFT", [][]byte{[]byte("dcdtToken"), big.NewInt(10).Bytes(), getAddress()})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDcdt_ExecuteToggleFreezeShouldWorkWithRealBech32Address(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	args.AddressPubKeyConverter = testscommon.RealWorldBech32PubkeyConverter

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	addressToFreezeBech32 := "moa158tgst07d6rt93td6nh5cd2mmpfhtp7hr24l4wfgtlggqpnp6kjs7e2zuz"
	addressToFreeze, err := args.AddressPubKeyConverter.Decode(addressToFreezeBech32)
	assert.NoError(t, err)

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{tokenName, addressToFreeze})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.DCDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToFreeze)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionDCDTFreeze + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestDcdt_ExecuteToggleFreezeShouldFailWithBech32Converter(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	args.AddressPubKeyConverter = testscommon.RealWorldBech32PubkeyConverter

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	addressToFreeze := []byte("not a bech32 address")

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{tokenName, addressToFreeze})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid address to freeze/unfreeze"))

	vmInput.Function = "freezeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, addressToFreeze)
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid address to freeze/unfreeze"))
}

func TestDcdt_ExecuteToggleFreezeShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	addressToFreeze := getAddress()

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{tokenName, addressToFreeze})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.DCDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToFreeze)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionDCDTFreeze + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestDcdt_ExecuteToggleUnFreezeShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	addressToUnfreeze := getAddress()

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unFreeze", [][]byte{tokenName, addressToUnfreeze})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.DCDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToUnfreeze)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionDCDTUnFreeze + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestDcdt_ExecuteToggleFreezeSingleNFTShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
		TokenType:    []byte(core.NonFungibleDCDT),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	addressToFreeze := getAddress()

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freezeSingleNFT", [][]byte{tokenName, big.NewInt(10).Bytes(), addressToFreeze})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.DCDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToFreeze)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionDCDTFreeze + "@" + hex.EncodeToString(append(tokenName, big.NewInt(10).Bytes()...))
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestDcdt_ExecuteToggleUnFreezeSingleNFTShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
		TokenType:    []byte(core.NonFungibleDCDT),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	addressToUnfreeze := getAddress()

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unFreezeSingleNFT", [][]byte{tokenName, big.NewInt(10).Bytes(), addressToUnfreeze})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.DCDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToUnfreeze)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionDCDTUnFreeze + "@" + hex.EncodeToString(append(tokenName, big.NewInt(10).Bytes()...))
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestDcdt_ExecuteWipeTooFewArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("dcdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 2"))

	vmInput.Function = "wipeSingleNFT"
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 3"))
}

func TestDcdt_ExecuteWipeWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("dcdtToken"), []byte("owner")})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))

	vmInput.Function = "wipeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("one"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestDcdt_ExecuteWipeNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.DCDTOperations = 10

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("dcdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))

	vmInput.Function = "wipeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("one"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestDcdt_ExecuteWipeOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("dcdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))

	vmInput.Function = "wipeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("one"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestDcdt_ExecuteWipeNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := "dcdtToken"
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: []byte("random address"),
	})
	tokensMap[tokenName] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte(tokenName), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))

	vmInput.Function = "wipeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("one"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestDcdt_ExecuteWipeNonWipeableTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: owner,
		CanWipe:      false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{tokenName, owner})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot wipe"))

	vmInput.Function = "wipeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("one"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot wipe"))
}

func TestDcdt_ExecuteWipeInvalidDestShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: owner,
		CanWipe:      true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{tokenName, []byte("dest")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid"))

	vmInput.Function = "wipeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("one"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid"))
}

func TestDcdt_ExecuteWipeTransferFailsNoErr(t *testing.T) {
	t.Parallel()

	err := errors.New("transfer error")
	args := createMockArgumentsForDCDT()
	args.Eei.(*mock.SystemEIStub).GetStorageCalled = func(key []byte) []byte {
		marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
			OwnerAddress: []byte("owner"),
			CanWipe:      true,
			TokenType:    []byte(core.FungibleDCDT),
		})
		return marshalizedData
	}
	args.Eei.(*mock.SystemEIStub).AddReturnMessageCalled = func(msg string) {
		assert.Equal(t, err.Error(), msg)
	}

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("dcdtToken"), getAddress()})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDcdt_ExecuteWipeSingleNFTTransferNoErr(t *testing.T) {
	t.Parallel()

	err := errors.New("transfer error")
	args := createMockArgumentsForDCDT()
	args.Eei.(*mock.SystemEIStub).GetStorageCalled = func(key []byte) []byte {
		marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
			OwnerAddress: []byte("owner"),
			CanWipe:      true,
			TokenType:    []byte(core.NonFungibleDCDT),
		})
		return marshalizedData
	}
	args.Eei.(*mock.SystemEIStub).AddReturnMessageCalled = func(msg string) {
		assert.Equal(t, err.Error(), msg)
	}

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipeSingleNFT", [][]byte{[]byte("dcdtToken"), big.NewInt(10).Bytes(), getAddress()})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDcdt_ExecuteWipeShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	addressToWipe := getAddress()
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		TokenType:    []byte(core.FungibleDCDT),
		OwnerAddress: owner,
		CanWipe:      true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{tokenName, addressToWipe})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.DCDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToWipe)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionDCDTWipe + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestDcdt_ExecuteWipeSingleNFTShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	addressToWipe := getAddress()
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		TokenType:    []byte(core.NonFungibleDCDT),
		OwnerAddress: owner,
		CanWipe:      true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipeSingleNFT", [][]byte{tokenName, big.NewInt(10).Bytes(), addressToWipe})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.DCDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToWipe)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionDCDTWipe + "@" + hex.EncodeToString(append(tokenName, big.NewInt(10).Bytes()...))
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestDcdt_ExecutePauseTooFewArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 1"))
}

func TestDcdt_ExecutePauseWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{[]byte("dcdtToken")})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestDcdt_ExecutePauseNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.DCDTOperations = 10

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{[]byte("dcdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestDcdt_ExecutePauseOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{[]byte("dcdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestDcdt_ExecutePauseNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := "dcdtToken"
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: []byte("random address"),
	})
	tokensMap[tokenName] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{[]byte(tokenName)})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestDcdt_ExecutePauseNonPauseableTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: owner,
		CanPause:     false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{tokenName})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot pause/un-pause"))
}

func TestDcdt_ExecutePauseOnAPausedTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: owner,
		CanPause:     true,
		IsPaused:     true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{tokenName})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot pause an already paused contract"))
}

func TestDcdt_ExecuteTogglePauseSavesTokenWithPausedFlagSet(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: []byte("owner"),
		CanPause:     true,
		IsPaused:     false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{tokenName})

	_ = e.Execute(vmInput)

	dcdtData := &DCDTDataV2{}
	_ = args.Marshalizer.Unmarshal(dcdtData, eei.GetStorage(tokenName))
	assert.Equal(t, true, dcdtData.IsPaused)

	require.Equal(t, &vmcommon.LogEntry{
		Identifier: []byte(core.BuiltInFunctionDCDTPause),
		Topics:     [][]byte{[]byte("dcdtToken")},
		Address:    []byte("owner"),
	}, eei.logs[0])
}

func TestDcdt_ExecuteTogglePauseShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanPause:     true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{tokenName})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()

	systemAddress := make([]byte, len(core.SystemAccountAddress))
	copy(systemAddress, core.SystemAccountAddress)
	systemAddress[len(core.SystemAccountAddress)-1] = 0

	createdAcc, accCreated := vmOutput.OutputAccounts[string(systemAddress)]
	assert.True(t, accCreated)

	assert.True(t, len(createdAcc.OutputTransfers) == 1)
	outputTransfer := createdAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	expectedInput := core.BuiltInFunctionDCDTPause + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestDcdt_ExecuteUnPauseOnAnUnPausedTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: owner,
		CanPause:     true,
		IsPaused:     false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unPause", [][]byte{tokenName})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot unPause an already un-paused contract"))
}

func TestDcdt_ExecuteUnPauseSavesTokenWithPausedFlagSetToFalse(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: []byte("owner"),
		CanPause:     true,
		IsPaused:     true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unPause", [][]byte{tokenName})

	_ = e.Execute(vmInput)

	dcdtData := &DCDTDataV2{}
	_ = args.Marshalizer.Unmarshal(dcdtData, eei.GetStorage(tokenName))
	assert.Equal(t, false, dcdtData.IsPaused)

	require.Equal(t, &vmcommon.LogEntry{
		Identifier: []byte(core.BuiltInFunctionDCDTUnPause),
		Topics:     [][]byte{[]byte("dcdtToken")},
		Address:    []byte("owner"),
	}, eei.logs[0])
}

func TestDcdt_ExecuteUnPauseShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanPause:     true,
		IsPaused:     true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unPause", [][]byte{tokenName})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()

	systemAddress := make([]byte, len(core.SystemAccountAddress))
	copy(systemAddress, core.SystemAccountAddress)
	systemAddress[len(core.SystemAccountAddress)-1] = 0

	createdAcc, accCreated := vmOutput.OutputAccounts[string(systemAddress)]
	assert.True(t, accCreated)

	assert.True(t, len(createdAcc.OutputTransfers) == 1)
	outputTransfer := createdAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	expectedInput := core.BuiltInFunctionDCDTUnPause + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestDcdt_ExecuteTransferOwnershipWrongNumOfArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("dcdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "expected num of arguments 2"))
}

func TestDcdt_ExecuteTransferOwnershipWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("dcdtToken"), []byte("newOwner")})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestDcdt_ExecuteTransferOwnershipNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.DCDTOperations = 10

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("dcdtToken"), []byte("newOwner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestDcdt_ExecuteTransferOwnershipOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("dcdtToken"), []byte("newOwner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestDcdt_ExecuteTransferOwnershipNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: []byte("random address"),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("dcdtToken"), []byte("newOwner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestDcdt_ExecuteTransferOwnershipNonTransferableTokenShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress:   []byte("owner"),
		CanChangeOwner: false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("dcdtToken"), []byte("newOwner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot change owner of the token"))
}

func TestDcdt_ExecuteTransferOwnershipInvalidDestinationAddressShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:      tokenName,
		OwnerAddress:   []byte("owner"),
		CanChangeOwner: true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("dcdtToken"), []byte("invalid address")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid"))
}

func TestDcdt_ExecuteTransferOwnershipSavesTokenWithNewOwnerAddressSet(t *testing.T) {
	t.Parallel()

	newOwner := getAddress()
	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:      []byte("dcdtToken"),
		OwnerAddress:   []byte("owner"),
		CanChangeOwner: true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("dcdtToken"), newOwner})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dcdtData := &DCDTDataV2{}
	_ = args.Marshalizer.Unmarshal(dcdtData, eei.GetStorage(tokenName))
	assert.Equal(t, newOwner, dcdtData.OwnerAddress)
}

func TestDcdt_ExecuteDcdtControlChangesWrongNumOfArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("dcdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))
}

func TestDcdt_ExecuteDcdtControlChangesWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("dcdtToken"), []byte("burnable")})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestDcdt_ExecuteDcdtControlChangesNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.DCDTOperations = 10

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("dcdtToken"), []byte("burnable")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestDcdt_ExecuteDcdtControlChangesOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("dcdtToken"), []byte("burnable")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestDcdt_ExecuteDcdtControlChangesNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: []byte("random address"),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("dcdtToken"), []byte("burnable")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestDcdt_ExecuteDcdtControlChangesNonUpgradableTokenShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress: []byte("owner"),
		Upgradable:   false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("dcdtToken"), []byte("burnable")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "token is not upgradable"))
}

func TestDcdt_ExecuteDcdtControlChangesSavesTokenWithUpgradedProperties(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:        []byte("dcdtToken"),
		TokenType:        []byte(core.FungibleDCDT),
		OwnerAddress:     []byte("owner"),
		Upgradable:       true,
		BurntValue:       big.NewInt(100),
		MintedValue:      big.NewInt(1000),
		NumWiped:         37,
		NFTCreateStopped: true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("dcdtToken"), []byte(burnable)})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput = getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("dcdtToken"),
		[]byte(burnable), []byte("true"),
		[]byte(mintable), []byte("true"),
		[]byte(canPause), []byte("true"),
		[]byte(canFreeze), []byte("true"),
		[]byte(canWipe), []byte("true"),
		[]byte(upgradable), []byte("false"),
		[]byte(canChangeOwner), []byte("true"),
		[]byte(canTransferNFTCreateRole), []byte("true"),
	})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dcdtData := &DCDTDataV2{}
	_ = args.Marshalizer.Unmarshal(dcdtData, eei.GetStorage(tokenName))
	assert.True(t, dcdtData.Burnable)
	assert.True(t, dcdtData.Mintable)
	assert.True(t, dcdtData.CanPause)
	assert.True(t, dcdtData.CanFreeze)
	assert.True(t, dcdtData.CanWipe)
	assert.False(t, dcdtData.Upgradable)
	assert.True(t, dcdtData.CanChangeOwner)

	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFunc("getTokenProperties", [][]byte{[]byte("dcdtToken")})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	assert.Equal(t, 18, len(eei.output))
	assert.Equal(t, []byte("dcdtToken"), eei.output[0])
	assert.Equal(t, []byte(core.FungibleDCDT), eei.output[1])
	assert.Equal(t, vmInput.CallerAddr, eei.output[2])
	assert.Equal(t, "1000", string(eei.output[3]))
	assert.Equal(t, "100", string(eei.output[4]))
	assert.Equal(t, []byte("NumDecimals-0"), eei.output[5])
	assert.Equal(t, []byte("IsPaused-false"), eei.output[6])
	assert.Equal(t, []byte("CanUpgrade-false"), eei.output[7])
	assert.Equal(t, []byte("CanMint-true"), eei.output[8])
	assert.Equal(t, []byte("CanBurn-true"), eei.output[9])
	assert.Equal(t, []byte("CanChangeOwner-true"), eei.output[10])
	assert.Equal(t, []byte("CanPause-true"), eei.output[11])
	assert.Equal(t, []byte("CanFreeze-true"), eei.output[12])
	assert.Equal(t, []byte("CanWipe-true"), eei.output[13])
	assert.Equal(t, []byte("CanAddSpecialRoles-false"), eei.output[14])
	assert.Equal(t, []byte("CanTransferNFTCreateRole-true"), eei.output[15])
	assert.Equal(t, []byte("NFTCreateStopped-true"), eei.output[16])
	assert.Equal(t, []byte("NumWiped-37"), eei.output[17])
}

func TestDcdt_ExecuteDcdtControlChangesForMultiNFTTransferShouldFaild(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:        []byte("dcdtToken"),
		TokenType:        []byte(core.NonFungibleDCDT),
		OwnerAddress:     []byte("owner"),
		Upgradable:       true,
		BurntValue:       big.NewInt(0),
		MintedValue:      big.NewInt(0),
		NumWiped:         37,
		NFTCreateStopped: true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("dcdtToken"),
		[]byte(canCreateMultiShard), []byte("true"),
	})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestDcdt_GetSpecialRolesValueNotZeroShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("dcdtToken")})
	vmInput.CallValue = big.NewInt(37)
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestDcdt_GetSpecialRolesInvalidNumOfArgsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("dcdtToken"), []byte("additional arg")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))
}

func TestDcdt_GetSpecialRolesNotEnoughGasShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.DCDTOperations = 10

	e, _ := NewDCDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("dcdtToken")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)

	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestDcdt_GetSpecialRolesInvalidTokenShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("invalid dcdtToken")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	assert.True(t, strings.Contains(eei.returnMessage, "no ticker with given name"))
}

func TestDcdt_GetSpecialRolesNoSpecialRoles(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("dcdtToken")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	assert.Equal(t, 0, len(eei.output))
}

func TestDcdt_GetSpecialRolesShouldWork(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	addr1 := "moa1kzzv2uw97q5k9mt458qk3q9u3cwhwqykvyk598q2f6wwx7gvrd9s2wkd6x"
	addr1Bytes, _ := testscommon.RealWorldBech32PubkeyConverter.Decode(addr1)

	addr2 := "moa1e7n8rzxdtl2n2fl6mrsg4l7stp2elxhfy6l9p7eeafspjhhrjq7qmhjnv7"
	addr2Bytes, _ := testscommon.RealWorldBech32PubkeyConverter.Decode(addr2)

	specialRoles := []*DCDTRoles{
		{
			Address: addr1Bytes,
			Roles: [][]byte{
				[]byte(core.DCDTRoleLocalMint),
				[]byte(core.DCDTRoleLocalBurn),
			},
		},
		{
			Address: addr2Bytes,
			Roles: [][]byte{
				[]byte(core.DCDTRoleNFTAddQuantity),
				[]byte(core.DCDTRoleNFTCreate),
				[]byte(core.DCDTRoleNFTBurn),
			},
		},
	}
	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		SpecialRoles: specialRoles,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	args.AddressPubKeyConverter = testscommon.RealWorldBech32PubkeyConverter

	e, _ := NewDCDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("dcdtToken")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	assert.Equal(t, 2, len(eei.output))
	assert.Equal(t, []byte("moa1kzzv2uw97q5k9mt458qk3q9u3cwhwqykvyk598q2f6wwx7gvrd9s2wkd6x:DCDTRoleLocalMint,DCDTRoleLocalBurn"), eei.output[0])
	assert.Equal(t, []byte("moa1e7n8rzxdtl2n2fl6mrsg4l7stp2elxhfy6l9p7eeafspjhhrjq7qmhjnv7:DCDTRoleNFTAddQuantity,DCDTRoleNFTCreate,DCDTRoleNFTBurn"), eei.output[1])
}

func TestDcdt_GetSpecialRolesWithEmptyAddressShouldWork(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	addr := ""
	addrBytes, _ := testscommon.RealWorldBech32PubkeyConverter.Decode(addr)

	specialRoles := []*DCDTRoles{
		{
			Address: addrBytes,
			Roles: [][]byte{
				[]byte(core.DCDTRoleLocalMint),
				[]byte(core.DCDTRoleLocalBurn),
			},
		},
		{
			Address: addrBytes,
			Roles: [][]byte{
				[]byte(core.DCDTRoleNFTAddQuantity),
				[]byte(core.DCDTRoleNFTCreate),
				[]byte(core.DCDTRoleNFTBurn),
			},
		},
		{
			Address: addrBytes,
			Roles: [][]byte{
				[]byte(vmcommon.DCDTRoleBurnForAll),
			},
		},
	}
	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		SpecialRoles: specialRoles,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	args.AddressPubKeyConverter = testscommon.RealWorldBech32PubkeyConverter

	e, _ := NewDCDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("dcdtToken")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	assert.Equal(t, 3, len(eei.output))
	assert.Equal(t, []byte(":DCDTRoleLocalMint,DCDTRoleLocalBurn"), eei.output[0])
	assert.Equal(t, []byte(":DCDTRoleNFTAddQuantity,DCDTRoleNFTCreate,DCDTRoleNFTBurn"), eei.output[1])
	assert.Equal(t, []byte(":DCDTRoleBurnForAll"), eei.output[2])
}

func TestDcdt_UnsetSpecialRoleWithRemoveEntryFromSpecialRoles(t *testing.T) {
	t.Parallel()

	tokenName := []byte("dcdtToken")
	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	owner := "moa1e7n8rzxdtl2n2fl6mrsg4l7stp2elxhfy6l9p7eeafspjhhrjq7qmhjnv7"
	ownerBytes, _ := testscommon.RealWorldBech32PubkeyConverter.Decode(owner)

	addr1 := "moa1kzzv2uw97q5k9mt458qk3q9u3cwhwqykvyk598q2f6wwx7gvrd9s2wkd6x"
	addr1Bytes, _ := testscommon.RealWorldBech32PubkeyConverter.Decode(addr1)

	addr2 := "moa1rsq30t33aqeg8cuc3q4kfnx0jukzsx52yfua92r233zhhmndl3us0qkmuz"
	addr2Bytes, _ := testscommon.RealWorldBech32PubkeyConverter.Decode(addr2)

	specialRoles := []*DCDTRoles{
		{
			Address: addr1Bytes,
			Roles: [][]byte{
				[]byte(core.DCDTRoleLocalMint),
			},
		},
		{
			Address: addr2Bytes,
			Roles: [][]byte{
				[]byte(core.DCDTRoleNFTAddQuantity),
				[]byte(core.DCDTRoleNFTCreate),
				[]byte(core.DCDTRoleNFTBurn),
			},
		},
	}
	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		OwnerAddress:       ownerBytes,
		SpecialRoles:       specialRoles,
		CanAddSpecialRoles: true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	args.AddressPubKeyConverter = testscommon.RealWorldBech32PubkeyConverter

	e, _ := NewDCDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("dcdtToken")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 2, len(eei.output))
	assert.Equal(t, []byte("moa1kzzv2uw97q5k9mt458qk3q9u3cwhwqykvyk598q2f6wwx7gvrd9s2wkd6x:DCDTRoleLocalMint"), eei.output[0])
	assert.Equal(t, []byte("moa1rsq30t33aqeg8cuc3q4kfnx0jukzsx52yfua92r233zhhmndl3us0qkmuz:DCDTRoleNFTAddQuantity,DCDTRoleNFTCreate,DCDTRoleNFTBurn"), eei.output[1])

	// unset the role for the address
	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.CallerAddr = ownerBytes
	vmInput.Arguments = [][]byte{[]byte("dcdtToken"), addr1Bytes, []byte(core.DCDTRoleLocalMint)}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	// get roles again
	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("dcdtToken")})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))

	// set the role for the address
	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.CallerAddr = ownerBytes
	vmInput.Arguments = [][]byte{[]byte("dcdtToken"), addr1Bytes, []byte(core.DCDTRoleLocalMint)}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	// get roles again
	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("dcdtToken")})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 2, len(eei.output))
	assert.Equal(t, []byte("moa1kzzv2uw97q5k9mt458qk3q9u3cwhwqykvyk598q2f6wwx7gvrd9s2wkd6x:DCDTRoleLocalMint"), eei.output[1])
}

func TestDcdt_ExecuteConfigChangeGetContractConfig(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("configChange", [][]byte{[]byte("dcdtToken"), []byte(burnable)})
	vmInput.CallerAddr = e.ownerAddress
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	newBaseIssingCost := big.NewInt(100)
	newMinTokenNameLength := int64(5)
	newMaxTokenNameLength := int64(20)
	newOwner := vmInput.RecipientAddr
	vmInput = getDefaultVmInputForFunc("configChange",
		[][]byte{newOwner, newBaseIssingCost.Bytes(), big.NewInt(newMinTokenNameLength).Bytes(),
			big.NewInt(newMaxTokenNameLength).Bytes()})
	vmInput.CallerAddr = e.ownerAddress
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dcdtData := &DCDTConfig{}
	_ = args.Marshalizer.Unmarshal(dcdtData, eei.GetStorage([]byte(configKeyPrefix)))
	assert.True(t, dcdtData.BaseIssuingCost.Cmp(newBaseIssingCost) == 0)
	assert.Equal(t, uint32(newMaxTokenNameLength), dcdtData.MaxTokenNameLength)
	assert.Equal(t, uint32(newMinTokenNameLength), dcdtData.MinTokenNameLength)
	assert.Equal(t, newOwner, dcdtData.OwnerAddress)

	vmInput = getDefaultVmInputForFunc("getContractConfig", make([][]byte, 0))
	vmInput.CallerAddr = []byte("any address")
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	require.Equal(t, 4, len(eei.output))
	assert.Equal(t, newOwner, eei.output[0])
	assert.Equal(t, newBaseIssingCost.Bytes(), eei.output[1])
	assert.Equal(t, big.NewInt(newMinTokenNameLength).Bytes(), eei.output[2])
	assert.Equal(t, big.NewInt(newMaxTokenNameLength).Bytes(), eei.output[3])

}

func TestDcdt_ExecuteClaim(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("claim", [][]byte{})
	vmInput.CallerAddr = e.ownerAddress

	eei.outputAccounts[string(vmInput.RecipientAddr)] = &vmcommon.OutputAccount{
		Address:      vmInput.RecipientAddr,
		Nonce:        0,
		BalanceDelta: big.NewInt(0),
		Balance:      big.NewInt(100),
	}

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	scOutAcc := eei.outputAccounts[string(vmInput.RecipientAddr)]
	assert.True(t, scOutAcc.BalanceDelta.Cmp(big.NewInt(-100)) == 0)

	receiver := eei.outputAccounts[string(vmInput.CallerAddr)]
	assert.True(t, receiver.BalanceDelta.Cmp(big.NewInt(100)) == 0)
}

func getAddress() []byte {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	return key
}

func TestDcdt_SetSpecialRoleCheckArgumentsErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
}

func TestDcdt_SetSpecialRoleCheckBasicOwnershipErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("1"), []byte("caller"), []byte(core.DCDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.OutOfFunds, retCode)
}

func TestDcdt_SetSpecialRoleNewSendRoleChangeDataErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller"),
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte, _ uint64) {
			require.Equal(t, []byte("DCDTSetRole@6d79546f6b656e@44434454526f6c654c6f63616c4275726e"), input)
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("caller"), []byte(core.DCDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_SetSpecialRoleAlreadyExists(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.DCDTRoleLocalBurn)},
					},
				},
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte, _ uint64) {
			require.Equal(t, []byte("DCDTSetRole@6d79546f6b656e@44434454526f6c654c6f63616c4275726e"), input)
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.DCDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_SetSpecialRoleCannotSaveToken(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
					},
				},
				TokenType:          []byte(core.FungibleDCDT),
				CanAddSpecialRoles: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte, _ uint64) {
			require.Equal(t, []byte("DCDTSetRole@6d79546f6b656e@44434454526f6c654c6f63616c4275726e"), input)
			castedMarshalizer := args.Marshalizer.(*mock.MarshalizerMock)
			castedMarshalizer.Fail = true
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.DCDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_SetSpecialRoleShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
					},
				},
				TokenType:          []byte(core.FungibleDCDT),
				CanAddSpecialRoles: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte, _ uint64) {
			require.Equal(t, []byte("DCDTSetRole@6d79546f6b656e@44434454526f6c654c6f63616c4275726e"), input)
		},
		SetStorageCalled: func(key []byte, value []byte) {
			token := &DCDTDataV2{}
			_ = args.Marshalizer.Unmarshal(token, value)
			require.Equal(t, [][]byte{[]byte(core.DCDTRoleLocalMint), []byte(core.DCDTRoleLocalBurn)}, token.SpecialRoles[0].Roles)
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.DCDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestDcdt_SetSpecialRoleNFTShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
					},
				},
				TokenType:          []byte(core.NonFungibleDCDT),
				CanAddSpecialRoles: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte, _ uint64) {
			require.Equal(t, []byte("DCDTSetRole@6d79546f6b656e@44434454526f6c654e4654437265617465"), input)
		},
		SetStorageCalled: func(key []byte, value []byte) {
			token := &DCDTDataV2{}
			_ = args.Marshalizer.Unmarshal(token, value)
			require.Equal(t, [][]byte{[]byte(core.DCDTRoleLocalMint), []byte(core.DCDTRoleNFTCreate)}, token.SpecialRoles[0].Roles)
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.DCDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	vmInput.Arguments[2] = []byte(core.DCDTRoleNFTAddQuantity)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	vmInput.Arguments[2] = []byte(core.DCDTRoleNFTCreate)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestDcdt_SetSpecialRoleTransferNotEnabledShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.RemoveActiveFlags(common.DCDTTransferRoleFlag)

	token := &DCDTDataV2{
		OwnerAddress: []byte("caller123"),
		SpecialRoles: []*DCDTRoles{
			{
				Address: []byte("myAddress"),
				Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
			},
		},
		TokenType:          []byte(core.NonFungibleDCDT),
		CanAddSpecialRoles: true,
	}
	dcdtTransferData := core.BuiltInFunctionDCDTSetLimitedTransfer + "@" + hex.EncodeToString([]byte("myToken"))
	called := false
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		SendGlobalSettingToAllCalled: func(sender []byte, input []byte) {
			assert.Equal(t, input, []byte(dcdtTransferData))
			called = true
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	enableEpochsHandler.RemoveActiveFlags(common.DCDTMetadataContinuousCleanupFlag)
	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.DCDTRoleTransfer)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	token.TokenType = []byte(core.NonFungibleDCDT)
	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	token.TokenType = []byte(core.FungibleDCDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	token.TokenType = []byte(core.SemiFungibleDCDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	enableEpochsHandler.AddActiveFlags(common.DCDTTransferRoleFlag)
	called = false
	token.TokenType = []byte(core.NonFungibleDCDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.True(t, called)

	token.TokenType = []byte(core.FungibleDCDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)

	called = false
	newAddressRole := &DCDTRoles{
		Address: []byte("address"),
		Roles:   [][]byte{[]byte(core.DCDTRoleTransfer)},
	}
	token.SpecialRoles = append(token.SpecialRoles, newAddressRole)
	token.TokenType = []byte(core.SemiFungibleDCDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.False(t, called)

	token.SpecialRoles[0].Roles = append(token.SpecialRoles[0].Roles, []byte(core.DCDTRoleTransfer))
	token.TokenType = []byte(core.SemiFungibleDCDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.False(t, called)

	vmInput.Function = "unSetSpecialRole"
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.False(t, called)

	dcdtTransferData = core.BuiltInFunctionDCDTUnSetLimitedTransfer + "@" + hex.EncodeToString([]byte("myToken"))
	token.SpecialRoles = token.SpecialRoles[:1]
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.True(t, called)
}

func TestDcdt_SetSpecialRoleTransferWithTransferRoleEnhancement(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.RemoveActiveFlags(common.DCDTTransferRoleFlag)

	token := &DCDTDataV2{
		OwnerAddress: []byte("caller123"),
		SpecialRoles: []*DCDTRoles{
			{
				Address: []byte("myAddress"),
				Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
			},
		},
		TokenType:          []byte(core.NonFungibleDCDT),
		CanAddSpecialRoles: true,
	}
	called := 0
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.DCDTRoleTransfer)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	enableEpochsHandler.AddActiveFlags(common.DCDTTransferRoleFlag)
	called = 0
	token.TokenType = []byte(core.NonFungibleDCDT)
	eei.SendGlobalSettingToAllCalled = func(sender []byte, input []byte) {
		if called == 0 {
			assert.Equal(t, core.BuiltInFunctionDCDTSetLimitedTransfer+"@"+hex.EncodeToString([]byte("myToken")), string(input))
		} else {
			assert.Equal(t, vmcommon.BuiltInFunctionDCDTTransferRoleAddAddress+"@"+hex.EncodeToString([]byte("myToken"))+"@"+hex.EncodeToString([]byte("myAddress")), string(input))
		}
		called++
	}

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, called, 2)

	called = 0
	newAddressRole := &DCDTRoles{
		Address: []byte("address"),
		Roles:   [][]byte{[]byte(core.DCDTRoleTransfer)},
	}
	token.SpecialRoles = append(token.SpecialRoles, newAddressRole)
	token.TokenType = []byte(core.SemiFungibleDCDT)
	eei.SendGlobalSettingToAllCalled = func(sender []byte, input []byte) {
		assert.Equal(t, vmcommon.BuiltInFunctionDCDTTransferRoleAddAddress+"@"+hex.EncodeToString([]byte("myToken"))+"@"+hex.EncodeToString([]byte("myAddress")), string(input))
		called++
	}
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, called, 1)

	token.SpecialRoles[0].Roles = append(token.SpecialRoles[0].Roles, []byte(core.DCDTRoleTransfer))
	vmInput.Function = "unSetSpecialRole"
	called = 0
	eei.SendGlobalSettingToAllCalled = func(sender []byte, input []byte) {
		assert.Equal(t, vmcommon.BuiltInFunctionDCDTTransferRoleDeleteAddress+"@"+hex.EncodeToString([]byte("myToken"))+"@"+hex.EncodeToString([]byte("myAddress")), string(input))
		called++
	}
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, called, 1)

	called = 0
	eei.SendGlobalSettingToAllCalled = func(sender []byte, input []byte) {
		if called == 0 {
			assert.Equal(t, core.BuiltInFunctionDCDTUnSetLimitedTransfer+"@"+hex.EncodeToString([]byte("myToken")), string(input))
		} else {
			assert.Equal(t, vmcommon.BuiltInFunctionDCDTTransferRoleDeleteAddress+"@"+hex.EncodeToString([]byte("myToken"))+"@"+hex.EncodeToString([]byte("myAddress")), string(input))
		}

		called++
	}
	token.SpecialRoles = token.SpecialRoles[:1]
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, called, 2)
}

func TestDcdt_SendAllTransferRoleAddresses(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.RemoveActiveFlags(common.DCDTMetadataContinuousCleanupFlag)

	token := &DCDTDataV2{
		OwnerAddress: []byte("caller1234"),
		SpecialRoles: []*DCDTRoles{
			{
				Address: []byte("myAddress1"),
				Roles:   [][]byte{[]byte(core.DCDTRoleTransfer)},
			},
			{
				Address: []byte("myAddress2"),
				Roles:   [][]byte{[]byte(core.DCDTRoleTransfer)},
			},
			{
				Address: []byte("myAddress3"),
				Roles:   [][]byte{[]byte(core.DCDTRoleTransfer)},
			},
		},
		TokenType:          []byte(core.NonFungibleDCDT),
		CanAddSpecialRoles: true,
	}
	called := 0
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("sendAllTransferRoleAddresses", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress")}
	vmInput.CallerAddr = []byte("caller1234")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionNotFound, retCode)

	enableEpochsHandler.AddActiveFlags(common.DCDTMetadataContinuousCleanupFlag)
	eei.ReturnMessage = ""
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, "wrong number of arguments, expected 1", eei.ReturnMessage)

	called = 0
	token.TokenType = []byte(core.NonFungibleDCDT)
	eei.SendGlobalSettingToAllCalled = func(sender []byte, input []byte) {
		assert.Equal(t, vmcommon.BuiltInFunctionDCDTTransferRoleAddAddress+"@"+hex.EncodeToString([]byte("myToken"))+"@"+hex.EncodeToString([]byte("myAddress1"))+"@"+hex.EncodeToString([]byte("myAddress2"))+"@"+hex.EncodeToString([]byte("myAddress3")), string(input))
		called++
	}
	vmInput.Arguments = [][]byte{[]byte("myToken")}
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, called, 1)

	called = 0
	token.SpecialRoles = make([]*DCDTRoles, 0)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, "no address with transfer role", eei.ReturnMessage)
}

func TestDcdt_SetSpecialRoleSFTShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
					},
				},
				TokenType:          []byte(core.SemiFungibleDCDT),
				CanAddSpecialRoles: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte, _ uint64) {
			require.Equal(t, []byte("DCDTSetRole@6d79546f6b656e@44434454526f6c654e46544164645175616e74697479"), input)
		},
		SetStorageCalled: func(key []byte, value []byte) {
			token := &DCDTDataV2{}
			_ = args.Marshalizer.Unmarshal(token, value)
			require.Equal(t, [][]byte{[]byte(core.DCDTRoleLocalMint), []byte(core.DCDTRoleNFTAddQuantity)}, token.SpecialRoles[0].Roles)
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.DCDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	vmInput.Arguments[2] = []byte(core.DCDTRoleNFTAddQuantity)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestDcdt_SetSpecialRoleCreateNFTTwoTimesShouldError(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.DCDTRoleNFTCreate)},
					},
				},
				TokenType:          []byte(core.NonFungibleDCDT),
				CanAddSpecialRoles: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("caller234"), []byte(core.DCDTRoleNFTCreate)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_SetSpecialRoleCreateNFTTwoTimesMultiShardShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddres4"),
						Roles:   [][]byte{[]byte(core.DCDTRoleNFTCreate)},
					},
				},
				TokenType:           []byte(core.NonFungibleDCDT),
				CanAddSpecialRoles:  true,
				CanCreateMultiShard: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("caller234"), []byte(core.DCDTRoleNFTCreate)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.ReturnMessage, vm.ErrInvalidAddress.Error())

	vmInput.Arguments[1] = []byte("caller23X")
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestDcdt_UnSetSpecialRoleCreateNFTShouldError(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.DCDTRoleNFTCreate)},
					},
				},
				TokenType:          []byte(core.NonFungibleDCDT),
				CanAddSpecialRoles: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("caller234"), []byte(core.DCDTRoleNFTCreate)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_UnsetSpecialRoleCheckArgumentsErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("1"), []byte("caller"), []byte(core.DCDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller2")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
}

func TestDcdt_UnsetSpecialRoleCheckArgumentsInvalidRoleErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("1"), []byte("caller"), []byte("mirage")}
	vmInput.CallerAddr = []byte("caller")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.OutOfFunds, retCode)
}

func TestDcdt_UnsetSpecialRoleCheckArgumentsDuplicatedRoleInArgsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("1"), []byte("caller"), []byte(core.DCDTRoleLocalBurn), []byte(core.DCDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_UnsetSpecialRoleCheckBasicOwnershipErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("1"), []byte("caller"), []byte(core.DCDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.OutOfFunds, retCode)
}

func TestDcdt_UnsetSpecialRoleNewShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller"),
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("caller"), []byte(core.DCDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_UnsetSpecialRoleCannotRemoveRoleNotExistsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
					},
				},
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.DCDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_UnsetSpecialRoleRemoveRoleTransfer(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
					},
				},
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte, _ uint64) {
			require.Equal(t, []byte("DCDTUnSetRole@6d79546f6b656e@44434454526f6c654c6f63616c4d696e74"), input)
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.DCDTRoleLocalMint)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestDcdt_UnsetSpecialRoleRemoveRoleSaveTokenErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
					},
				},
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte, _ uint64) {
			require.Equal(t, []byte("DCDTUnSetRole@6d79546f6b656e@44434454526f6c654c6f63616c4d696e74"), input)
			castedMarshalizer := args.Marshalizer.(*mock.MarshalizerMock)
			castedMarshalizer.Fail = true
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.DCDTRoleLocalMint)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_UnsetSpecialRoleRemoveRoleShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
					},
				},
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte, _ uint64) {
			require.Equal(t, []byte("DCDTUnSetRole@6d79546f6b656e@44434454526f6c654c6f63616c4d696e74"), input)
		},
		SetStorageCalled: func(key []byte, value []byte) {
			token := &DCDTDataV2{}
			_ = args.Marshalizer.Unmarshal(token, value)
			require.Len(t, token.SpecialRoles, 0)
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.DCDTRoleLocalMint)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestDcdt_StopNFTCreateForeverCheckArgumentsErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("stopNFTCreate", [][]byte{})
	vmInput.Arguments = [][]byte{{1}, {2}}
	vmInput.CallerAddr = []byte("caller2")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)

	vmInput.CallValue = big.NewInt(0)
	vmInput.Arguments = [][]byte{{1}}
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_StopNFTCreateForeverCallErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	token := &DCDTDataV2{
		OwnerAddress: []byte("caller1"),
		SpecialRoles: []*DCDTRoles{
			{
				Address: []byte("myAddress"),
				Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
			},
		},
	}
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("stopNFTCreate", [][]byte{[]byte("tokenID")})
	vmInput.CallerAddr = []byte("caller2")
	vmInput.CallValue = big.NewInt(0)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	vmInput.CallerAddr = token.OwnerAddress
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	token.TokenType = []byte(core.NonFungibleDCDT)
	token.NFTCreateStopped = true
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	token.NFTCreateStopped = false
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_StopNFTCreateForeverCallShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	token := &DCDTDataV2{
		OwnerAddress: []byte("caller1"),
		SpecialRoles: []*DCDTRoles{
			{
				Address: []byte("myAddress"),
				Roles:   [][]byte{[]byte(core.DCDTRoleNFTCreate)},
			},
		},
	}
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte, _ uint64) {
			require.Equal(t, []byte("DCDTUnSetRole@746f6b656e4944@44434454526f6c654e4654437265617465"), input)
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("stopNFTCreate", [][]byte{[]byte("tokenID")})
	vmInput.CallerAddr = token.OwnerAddress
	vmInput.CallValue = big.NewInt(0)

	token.TokenType = []byte(core.NonFungibleDCDT)
	token.NFTCreateStopped = false
	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestDcdt_TransferNFTCreateCheckArgumentsErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("transferNFTCreateRole", [][]byte{})
	vmInput.Arguments = [][]byte{{1}, {2}}
	vmInput.CallerAddr = []byte("caller2")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)

	vmInput.CallValue = big.NewInt(0)
	vmInput.Arguments = [][]byte{{1}, []byte("caller3"), {3}}
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_TransferNFTCreateCallErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	token := &DCDTDataV2{
		OwnerAddress: []byte("caller1"),
		SpecialRoles: []*DCDTRoles{
			{
				Address: []byte("caller1"),
				Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
			},
		},
	}
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("transferNFTCreateRole", [][]byte{[]byte("tokenID"), []byte("caller3"), []byte("caller22")})
	vmInput.CallerAddr = []byte("caller2")
	vmInput.CallValue = big.NewInt(0)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	vmInput.CallerAddr = token.OwnerAddress
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	token.TokenType = []byte(core.FungibleDCDT)
	token.CanTransferNFTCreateRole = true
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	token.TokenType = []byte(core.NonFungibleDCDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)

	vmInput.Arguments[2] = vmInput.Arguments[1]
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)

	vmInput.Arguments[2] = []byte("caller2")
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_TransferNFTCreateCallShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	token := &DCDTDataV2{
		OwnerAddress: []byte("caller1"),
		SpecialRoles: []*DCDTRoles{
			{
				Address: []byte("caller3"),
				Roles:   [][]byte{[]byte(core.DCDTRoleNFTCreate)},
			},
		},
	}
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte, _ uint64) {
			require.Equal(t, []byte("DCDTNFTCreateRoleTransfer@746f6b656e4944@63616c6c657232"), input)
			require.Equal(t, destination, []byte("caller3"))
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("transferNFTCreateRole", [][]byte{[]byte("tokenID"), []byte("caller3"), []byte("caller2")})
	vmInput.CallerAddr = token.OwnerAddress
	vmInput.CallValue = big.NewInt(0)

	token.TokenType = []byte(core.NonFungibleDCDT)
	token.CanTransferNFTCreateRole = true
	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestDcdt_TransferNFTCreateCallMultiShardShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	token := &DCDTDataV2{
		OwnerAddress: []byte("caller1"),
		SpecialRoles: []*DCDTRoles{
			{
				Address: []byte("3caller"),
				Roles:   [][]byte{[]byte(core.DCDTRoleNFTCreate)},
			},
		},
	}
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("transferNFTCreateRole", [][]byte{[]byte("tokenID"), []byte("3caller"), []byte("caller2")})
	vmInput.CallerAddr = token.OwnerAddress
	vmInput.CallValue = big.NewInt(0)

	token.TokenType = []byte(core.NonFungibleDCDT)
	token.CanTransferNFTCreateRole = true
	token.CanCreateMultiShard = true
	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	vmInput.Arguments = [][]byte{[]byte("tokenID"), []byte("3caller"), []byte("2caller")}
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestDcdt_SetNewGasCost(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	e.SetNewGasCost(vm.GasCost{BuiltInCost: vm.BuiltInCost{
		ChangeOwnerAddress: 10000,
	}})

	require.Equal(t, uint64(10000), e.gasCost.BuiltInCost.ChangeOwnerAddress)
}

func TestDcdt_GetAllAddressesAndRolesNoArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("getAllAddressesAndRoles", [][]byte{})
	vmInput.Arguments = nil

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_GetAllAddressesAndRolesCallWithValueShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("getAllAddressesAndRoles", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("arg")}
	vmInput.CallValue = big.NewInt(0)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestDcdt_GetAllAddressesAndRolesCallGetExistingTokenErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &DCDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*DCDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.DCDTRoleLocalMint)},
					},
				},
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("getAllAddressesAndRoles", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("arg")}
	vmInput.CallValue = big.NewInt(0)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestDcdt_CanUseContract(t *testing.T) {
	args := createMockArgumentsForDCDT()
	eei := &mock.SystemEIStub{}
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	require.True(t, e.CanUseContract())
}

func TestDcdt_ExecuteIssueMetaDCDT(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	e, _ := NewDCDTSmartContract(args)

	enableEpochsHandler.RemoveActiveFlags(common.MetaDCDTSetFlag)
	vmInput := getDefaultVmInputForFunc("registerMetaDCDT", nil)
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "invalid method to call")

	eei.returnMessage = ""
	eei.gasRemaining = 9999
	enableEpochsHandler.AddActiveFlags(common.MetaDCDTSetFlag)
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "not enough arguments")

	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)
	vmInput.Arguments = [][]byte{[]byte("tokenName")}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "not enough arguments")

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("ticker"), big.NewInt(20).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of decimals"))

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("ticker"), big.NewInt(10).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "ticker name is not valid"))

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("TICKER"), big.NewInt(10).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, len(eei.output), 1)
	assert.True(t, strings.Contains(string(eei.output[0]), "TICKER-"))
}

func TestDcdt_ExecuteChangeSFTToMetaDCDT(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	e, _ := NewDCDTSmartContract(args)

	enableEpochsHandler.RemoveActiveFlags(common.MetaDCDTSetFlag)
	vmInput := getDefaultVmInputForFunc("changeSFTToMetaDCDT", nil)
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "invalid method to call")

	eei.returnMessage = ""
	eei.gasRemaining = 9999
	enableEpochsHandler.AddActiveFlags(common.MetaDCDTSetFlag)
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "not enough arguments")

	vmInput.Arguments = [][]byte{[]byte("tokenName"), big.NewInt(20).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of decimals"))

	vmInput.Arguments = [][]byte{[]byte("tokenName"), big.NewInt(10).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "no ticker with given name"))

	_ = e.saveToken(vmInput.Arguments[0], &DCDTDataV2{TokenType: []byte(core.NonFungibleDCDT), OwnerAddress: vmInput.CallerAddr})
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "change can happen to semi fungible tokens only"))

	_ = e.saveToken(vmInput.Arguments[0], &DCDTDataV2{TokenType: []byte(core.SemiFungibleDCDT), OwnerAddress: vmInput.CallerAddr})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	token, _ := e.getExistingToken(vmInput.Arguments[0])
	assert.Equal(t, token.NumDecimals, uint32(10))
	assert.Equal(t, token.TokenType, []byte(metaDCDT))
}

func TestDcdt_ExecuteIssueSFTAndChangeSFTToMetaDCDT(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewDCDTSmartContract(args)

	eei.returnMessage = ""
	eei.gasRemaining = 9999

	vmInput := getDefaultVmInputForFunc("issueSemiFungible", nil)
	vmInput.CallValue = e.baseIssuingCost
	vmInput.Arguments = [][]byte{[]byte("name"), []byte("TOKEN")}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	fullTicker := eei.output[0]

	token, _ := e.getExistingToken(fullTicker)
	assert.Equal(t, token.NumDecimals, uint32(0))
	assert.Equal(t, token.TokenType, []byte(core.SemiFungibleDCDT))

	vmInput.CallValue = big.NewInt(0)
	vmInput.Function = "changeSFTToMetaDCDT"
	vmInput.Arguments = [][]byte{fullTicker, big.NewInt(10).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	token, _ = e.getExistingToken(fullTicker)
	assert.Equal(t, token.NumDecimals, uint32(10))
	assert.Equal(t, token.TokenType, []byte(metaDCDT))

	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "change can happen to semi fungible tokens only"))
}

func TestDcdt_ExecuteRegisterAndSetErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	e, _ := NewDCDTSmartContract(args)

	enableEpochsHandler.RemoveActiveFlags(common.DCDTRegisterAndSetAllRolesFlag)
	vmInput := getDefaultVmInputForFunc("registerAndSetAllRoles", nil)
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionNotFound, output)
	assert.Equal(t, eei.returnMessage, "invalid method to call")

	eei.returnMessage = ""
	eei.gasRemaining = 9999
	enableEpochsHandler.AddActiveFlags(common.DCDTRegisterAndSetAllRolesFlag)
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "not enough arguments")

	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)
	vmInput.Arguments = [][]byte{[]byte("tokenName")}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "arguments length mismatch")

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("ticker"), []byte("VAL"), big.NewInt(20).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidArgument.Error()))

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("ticker"), []byte("FNG"), big.NewInt(10).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "ticker name is not valid"))

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("ticker"), []byte("FNG"), big.NewInt(20).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of decimals"))
}

func TestDcdt_ExecuteRegisterAndSetFungible(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("registerAndSetAllRoles", nil)
	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("TICKER"), []byte("FNG"), big.NewInt(10).Bytes()}
	eei.gasRemaining = 9999
	eei.returnMessage = ""
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, len(eei.output), 1)
	assert.True(t, strings.Contains(string(eei.output[0]), "TICKER-"))

	token, _ := e.getExistingToken(eei.output[0])
	assert.Equal(t, token.TokenType, []byte(core.FungibleDCDT))
}

func TestDcdt_ExecuteRegisterAndSetNonFungible(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("registerAndSetAllRoles", nil)
	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("TICKER"), []byte("NFT"), big.NewInt(10).Bytes()}
	eei.gasRemaining = 9999
	eei.returnMessage = ""
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, len(eei.output), 1)
	assert.True(t, strings.Contains(string(eei.output[0]), "TICKER-"))

	token, _ := e.getExistingToken(eei.output[0])
	assert.Equal(t, token.TokenType, []byte(core.NonFungibleDCDT))
}

func TestDcdt_ExecuteRegisterAndSetSemiFungible(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("registerAndSetAllRoles", nil)
	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("TICKER"), []byte("SFT"), big.NewInt(10).Bytes()}
	eei.gasRemaining = 9999
	eei.returnMessage = ""
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, len(eei.output), 1)
	assert.True(t, strings.Contains(string(eei.output[0]), "TICKER-"))

	token, _ := e.getExistingToken(eei.output[0])
	assert.Equal(t, token.TokenType, []byte(core.SemiFungibleDCDT))
	lenOutTransfer := 0
	for _, outAcc := range eei.outputAccounts {
		lenOutTransfer += len(outAcc.OutputTransfers)
	}
	assert.Equal(t, uint32(lenOutTransfer), 1+eei.blockChainHook.NumberOfShards())
}

func TestDcdt_ExecuteRegisterAndSetMetaDCDTShouldSetType(t *testing.T) {
	t.Parallel()

	registerAndSetAllRolesWithTypeCheck(t, []byte("NFT"), []byte(core.NonFungibleDCDT))
	registerAndSetAllRolesWithTypeCheck(t, []byte("SFT"), []byte(core.SemiFungibleDCDT))
	registerAndSetAllRolesWithTypeCheck(t, []byte("META"), []byte(metaDCDT))
	registerAndSetAllRolesWithTypeCheck(t, []byte("FNG"), []byte(core.FungibleDCDT))
}

func registerAndSetAllRolesWithTypeCheck(t *testing.T, typeArgument []byte, expectedType []byte) {
	args := createMockArgumentsForDCDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewDCDTSmartContract(args)

	enableEpochsHandler.RemoveActiveFlags(common.DCDTMetadataContinuousCleanupFlag)
	vmInput := getDefaultVmInputForFunc("registerAndSetAllRoles", nil)
	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("TICKER"), typeArgument, big.NewInt(10).Bytes()}
	eei.gasRemaining = 9999
	eei.returnMessage = ""
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, len(eei.output), 1)
	assert.True(t, strings.Contains(string(eei.output[0]), "TICKER-"))

	token, _ := e.getExistingToken(eei.output[0])
	assert.Equal(t, expectedType, token.TokenType)

	lenOutTransfer := 0
	for _, outAcc := range eei.outputAccounts {
		lenOutTransfer += len(outAcc.OutputTransfers)
	}
	assert.Equal(t, lenOutTransfer, 1)
}

func TestDcdt_setBurnRoleGlobally(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("setBurnRoleGlobally", [][]byte{})

	enableEpochsHandler.RemoveActiveFlags(common.DCDTMetadataContinuousCleanupFlag)
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionNotFound, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid method to call"))

	enableEpochsHandler.AddActiveFlags(common.DCDTMetadataContinuousCleanupFlag)
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 1"))

	owner := bytes.Repeat([]byte{1}, 32)
	tokenName := []byte("TOKEN-ABABAB")
	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanPause:     true,
		IsPaused:     true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap

	vmInput.CallerAddr = owner
	vmInput.Arguments = [][]byte{tokenName}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()

	systemAddress := make([]byte, len(core.SystemAccountAddress))
	copy(systemAddress, core.SystemAccountAddress)
	systemAddress[len(core.SystemAccountAddress)-1] = 0

	createdAcc, accCreated := vmOutput.OutputAccounts[string(systemAddress)]
	assert.True(t, accCreated)

	assert.True(t, len(createdAcc.OutputTransfers) == 1)
	outputTransfer := createdAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	expectedInput := vmcommon.BuiltInFunctionDCDTSetBurnRoleForAll + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)

	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot set burn role globally as it was already set"))
}

func TestDcdt_unsetBurnRoleGlobally(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewDCDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unsetBurnRoleGlobally", [][]byte{})

	enableEpochsHandler.RemoveActiveFlags(common.DCDTMetadataContinuousCleanupFlag)
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionNotFound, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid method to call"))

	enableEpochsHandler.AddActiveFlags(common.DCDTMetadataContinuousCleanupFlag)
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 1"))

	owner := bytes.Repeat([]byte{1}, 32)
	tokenName := []byte("TOKEN-ABABAB")
	tokensMap := map[string][]byte{}
	token := &DCDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanPause:     true,
		IsPaused:     true,
	}

	burnForAllRole := &DCDTRoles{Roles: [][]byte{[]byte(vmcommon.DCDTRoleBurnForAll)}, Address: []byte{}}
	token.SpecialRoles = append(token.SpecialRoles, burnForAllRole)

	marshalizedData, _ := args.Marshalizer.Marshal(token)
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap

	vmInput.CallerAddr = owner
	vmInput.Arguments = [][]byte{tokenName}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	vmOutput := eei.CreateVMOutput()

	systemAddress := make([]byte, len(core.SystemAccountAddress))
	copy(systemAddress, core.SystemAccountAddress)
	systemAddress[len(core.SystemAccountAddress)-1] = 0

	createdAcc, accCreated := vmOutput.OutputAccounts[string(systemAddress)]
	assert.True(t, accCreated)

	assert.True(t, len(createdAcc.OutputTransfers) == 1)
	outputTransfer := createdAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	expectedInput := vmcommon.BuiltInFunctionDCDTUnSetBurnRoleForAll + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)

	token, err := e.getExistingToken(tokenName)
	assert.Nil(t, err)
	assert.Equal(t, len(token.SpecialRoles), 0)

	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot unset burn role globally as it was not set"))
}

func TestDcdt_CheckRolesOnMetaDCDT(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewDCDTSmartContract(args)

	err := e.checkSpecialRolesAccordingToTokenType([][]byte{[]byte("random")}, &DCDTDataV2{TokenType: []byte(metaDCDT)})
	assert.Nil(t, err)

	enableEpochsHandler.AddActiveFlags(common.ManagedCryptoAPIsFlag)
	err = e.checkSpecialRolesAccordingToTokenType([][]byte{[]byte("random")}, &DCDTDataV2{TokenType: []byte(metaDCDT)})
	assert.Equal(t, err, vm.ErrInvalidArgument)
}

func TestDcdt_SetNFTCreateRoleAfterStopNFTCreateShouldNotWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	eei := createDefaultEei()
	args.Eei = eei

	owner := bytes.Repeat([]byte{1}, 32)
	tokenName := []byte("TOKEN-ABABAB")
	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(DCDTDataV2{
		TokenName:          tokenName,
		OwnerAddress:       owner,
		CanPause:           true,
		IsPaused:           true,
		TokenType:          []byte(core.NonFungibleDCDT),
		CanAddSpecialRoles: true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap

	e, _ := NewDCDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{tokenName, owner, []byte(core.DCDTRoleNFTCreate)})
	vmInput.CallerAddr = owner
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmInput = getDefaultVmInputForFunc("stopNFTCreate", [][]byte{tokenName})
	vmInput.CallerAddr = owner
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmInput = getDefaultVmInputForFunc("setSpecialRole", [][]byte{tokenName, owner, []byte(core.DCDTRoleNFTCreate)})
	vmInput.CallerAddr = owner
	enableEpochsHandler.AddActiveFlags(common.NFTStopCreateFlag)
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot add NFT create role as NFT creation was stopped"))

	enableEpochsHandler.RemoveActiveFlags(common.NFTStopCreateFlag)
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}
