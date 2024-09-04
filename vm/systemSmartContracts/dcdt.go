//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/DharitriOne/protobuf/protobuf  --gogoslick_out=. dcdt.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/core/check"
	"github.com/DharitriOne/drt-chain-core-go/hashing"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/vm"
	logger "github.com/DharitriOne/drt-chain-logger-go"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
)

const numOfRetriesForIdentifier = 50
const tickerSeparator = "-"
const tickerRandomSequenceLength = 3
const minLengthForTickerName = 3
const maxLengthForTickerName = 10
const minLengthForInitTokenName = 10
const minLengthForTokenName = 3
const maxLengthForTokenName = 20
const minNumberOfDecimals = 0
const maxNumberOfDecimals = 18
const configKeyPrefix = "dcdtConfig"
const burnable = "canBurn"
const mintable = "canMint"
const canPause = "canPause"
const canFreeze = "canFreeze"
const canWipe = "canWipe"
const canChangeOwner = "canChangeOwner"
const canAddSpecialRoles = "canAddSpecialRoles"
const canTransferNFTCreateRole = "canTransferNFTCreateRole"
const upgradable = "canUpgrade"
const canCreateMultiShard = "canCreateMultiShard"
const upgradeProperties = "upgradeProperties"

const conversionBase = 10
const metaDCDT = "MetaDCDT"

type dcdt struct {
	eei                    vm.SystemEI
	gasCost                vm.GasCost
	baseIssuingCost        *big.Int
	ownerAddress           []byte // do not use this in functions. Should use e.getDcdtOwner()
	dcdtSCAddress          []byte
	endOfEpochSCAddress    []byte
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	mutExecution           sync.RWMutex
	addressPubKeyConverter core.PubkeyConverter
	enableEpochsHandler    common.EnableEpochsHandler
}

// ArgsNewDCDTSmartContract defines the arguments needed for the dcdt contract
type ArgsNewDCDTSmartContract struct {
	Eei                    vm.SystemEI
	GasCost                vm.GasCost
	DCDTSCConfig           config.DCDTSystemSCConfig
	DCDTSCAddress          []byte
	Marshalizer            marshal.Marshalizer
	Hasher                 hashing.Hasher
	EndOfEpochSCAddress    []byte
	AddressPubKeyConverter core.PubkeyConverter
	EnableEpochsHandler    common.EnableEpochsHandler
}

// NewDCDTSmartContract creates the dcdt smart contract, which controls the issuing of tokens
func NewDCDTSmartContract(args ArgsNewDCDTSmartContract) (*dcdt, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, vm.ErrNilHasher
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, vm.ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.DCDTMetadataContinuousCleanupFlag,
		common.GlobalMintBurnFlag,
		common.MultiClaimOnDelegationFlag,
		common.MetaDCDTSetFlag,
		common.DCDTMetadataContinuousCleanupFlag,
		common.ManagedCryptoAPIsFlag,
		common.DCDTFlag,
		common.DCDTTransferRoleFlag,
		common.GlobalMintBurnFlag,
		common.DCDTRegisterAndSetAllRolesFlag,
		common.MetaDCDTSetFlag,
		common.DCDTNFTCreateOnMultiShardFlag,
		common.NFTStopCreateFlag,
	})
	if err != nil {
		return nil, err
	}
	if check.IfNil(args.AddressPubKeyConverter) {
		return nil, vm.ErrNilAddressPubKeyConverter
	}
	if len(args.EndOfEpochSCAddress) == 0 {
		return nil, vm.ErrNilEndOfEpochSmartContractAddress
	}
	baseIssuingCost, okConvert := big.NewInt(0).SetString(args.DCDTSCConfig.BaseIssuingCost, conversionBase)
	if !okConvert || baseIssuingCost.Cmp(big.NewInt(0)) < 0 {
		return nil, vm.ErrInvalidBaseIssuingCost
	}

	return &dcdt{
		eei:             args.Eei,
		gasCost:         args.GasCost,
		baseIssuingCost: baseIssuingCost,
		// we should have called pubkeyConverter.Decode here instead of a byte slice cast. Since that change would break
		// backwards compatibility, the fix was carried in the epochStart/metachain/systemSCs.go
		ownerAddress:           []byte(args.DCDTSCConfig.OwnerAddress),
		dcdtSCAddress:          args.DCDTSCAddress,
		hasher:                 args.Hasher,
		marshalizer:            args.Marshalizer,
		endOfEpochSCAddress:    args.EndOfEpochSCAddress,
		addressPubKeyConverter: args.AddressPubKeyConverter,
		enableEpochsHandler:    args.EnableEpochsHandler,
	}, nil
}

// Execute calls one of the functions from the dcdt smart contract and runs the code according to the input
func (e *dcdt) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	e.mutExecution.RLock()
	defer e.mutExecution.RUnlock()

	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	if args.Function == core.SCDeployInitFunctionName {
		return e.init(args)
	}

	if !e.enableEpochsHandler.IsFlagEnabled(common.DCDTFlag) {
		e.eei.AddReturnMessage("DCDT SC disabled")
		return vmcommon.UserError
	}

	switch args.Function {
	case "issue":
		return e.issue(args)
	case "issueSemiFungible":
		return e.registerSemiFungible(args)
	case "issueNonFungible":
		return e.registerNonFungible(args)
	case "registerMetaDCDT":
		return e.registerMetaDCDT(args)
	case "changeSFTToMetaDCDT":
		return e.changeSFTToMetaDCDT(args)
	case "registerAndSetAllRoles":
		return e.registerAndSetRoles(args)
	case core.BuiltInFunctionDCDTBurn:
		return e.burn(args)
	case "mint":
		return e.mint(args)
	case "freeze":
		return e.toggleFreeze(args, core.BuiltInFunctionDCDTFreeze)
	case "unFreeze":
		return e.toggleFreeze(args, core.BuiltInFunctionDCDTUnFreeze)
	case "wipe":
		return e.wipe(args)
	case "pause":
		return e.togglePause(args, core.BuiltInFunctionDCDTPause)
	case "unPause":
		return e.togglePause(args, core.BuiltInFunctionDCDTUnPause)
	case "freezeSingleNFT":
		return e.toggleFreezeSingleNFT(args, core.BuiltInFunctionDCDTFreeze)
	case "unFreezeSingleNFT":
		return e.toggleFreezeSingleNFT(args, core.BuiltInFunctionDCDTUnFreeze)
	case "wipeSingleNFT":
		return e.wipeSingleNFT(args)
	case "claim":
		return e.claim(args)
	case "configChange":
		return e.configChange(args)
	case "controlChanges":
		return e.controlChanges(args)
	case "transferOwnership":
		return e.transferOwnership(args)
	case "getTokenProperties":
		return e.getTokenProperties(args)
	case "getSpecialRoles":
		return e.getSpecialRoles(args)
	case "setSpecialRole":
		return e.setSpecialRole(args)
	case "unSetSpecialRole":
		return e.unSetSpecialRole(args)
	case "transferNFTCreateRole":
		return e.transferNFTCreateRole(args)
	case "stopNFTCreate":
		return e.stopNFTCreateForever(args)
	case "getAllAddressesAndRoles":
		return e.getAllAddressesAndRoles(args)
	case "getContractConfig":
		return e.getContractConfig(args)
	case "changeToMultiShardCreate":
		return e.changeToMultiShardCreate(args)
	case "setBurnRoleGlobally":
		return e.setBurnRoleGlobally(args)
	case "unsetBurnRoleGlobally":
		return e.unsetBurnRoleGlobally(args)
	case "sendAllTransferRoleAddresses":
		return e.sendAllTransferRoleAddresses(args)
	}

	e.eei.AddReturnMessage("invalid method to call")
	return vmcommon.FunctionNotFound
}

func (e *dcdt) init(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	scConfig := &DCDTConfig{
		OwnerAddress:       e.ownerAddress,
		BaseIssuingCost:    e.baseIssuingCost,
		MinTokenNameLength: minLengthForInitTokenName,
		MaxTokenNameLength: maxLengthForTokenName,
	}
	err := e.saveDCDTConfig(scConfig)
	if err != nil {
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *dcdt) checkBasicCreateArguments(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.DCDTIssue)
	if err != nil {
		e.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}
	dcdtConfig, err := e.getDCDTConfig()
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if len(args.Arguments) < 1 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.UserError
	}
	if len(args.Arguments[0]) < minLengthForTokenName ||
		len(args.Arguments[0]) > int(dcdtConfig.MaxTokenNameLength) {
		e.eei.AddReturnMessage("token name length not in parameters")
		return vmcommon.FunctionWrongSignature
	}
	if args.CallValue.Cmp(dcdtConfig.BaseIssuingCost) != 0 {
		e.eei.AddReturnMessage("callValue not equals with baseIssuingCost")
		return vmcommon.OutOfFunds
	}

	return vmcommon.Ok
}

// format: issue@tokenName@ticker@initialSupply@numOfDecimals@optional-list-of-properties
func (e *dcdt) issue(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 4 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}

	returnCode := e.checkBasicCreateArguments(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if len(args.Arguments[2]) > core.MaxLenForDCDTIssueMint {
		returnMessage := fmt.Sprintf("max length for dcdt issue is %d", core.MaxLenForDCDTIssueMint)
		e.eei.AddReturnMessage(returnMessage)
		return vmcommon.UserError
	}

	initialSupply := big.NewInt(0).SetBytes(args.Arguments[2])
	isGlobalMintBurnFlagEnabled := e.enableEpochsHandler.IsFlagEnabled(common.GlobalMintBurnFlag)
	isSupplyZeroAfterFlag := isGlobalMintBurnFlagEnabled && initialSupply.Cmp(zero) == 0
	isInvalidSupply := initialSupply.Cmp(zero) < 0 || isSupplyZeroAfterFlag
	if isInvalidSupply {
		e.eei.AddReturnMessage(vm.ErrNegativeOrZeroInitialSupply.Error())
		return vmcommon.UserError
	}

	numOfDecimals := uint32(big.NewInt(0).SetBytes(args.Arguments[3]).Uint64())
	if numOfDecimals < minNumberOfDecimals || numOfDecimals > maxNumberOfDecimals {
		e.eei.AddReturnMessage(fmt.Errorf("%w, minimum: %d, maximum: %d, provided: %d",
			vm.ErrInvalidNumberOfDecimals,
			minNumberOfDecimals,
			maxNumberOfDecimals,
			numOfDecimals,
		).Error())
		return vmcommon.UserError
	}

	tokenIdentifier, _, err := e.createNewToken(
		args.CallerAddr,
		args.Arguments[0],
		args.Arguments[1],
		initialSupply,
		numOfDecimals,
		args.Arguments[4:],
		[]byte(core.FungibleDCDT))
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if initialSupply.Cmp(zero) > 0 {
		dcdtTransferData := core.BuiltInFunctionDCDTTransfer + "@" + hex.EncodeToString(tokenIdentifier) + "@" + hex.EncodeToString(initialSupply.Bytes())
		e.eei.Transfer(args.CallerAddr, e.dcdtSCAddress, big.NewInt(0), []byte(dcdtTransferData), 0)
	} else {
		e.eei.Finish(tokenIdentifier)
	}

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(args.Function),
		Address:    args.CallerAddr,
		Topics:     [][]byte{tokenIdentifier, args.Arguments[0], args.Arguments[1], []byte(core.FungibleDCDT), big.NewInt(int64(numOfDecimals)).Bytes()},
	}
	e.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

func (e *dcdt) registerNonFungible(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := e.checkBasicCreateArguments(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if len(args.Arguments) < 2 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.UserError
	}

	tokenIdentifier, _, err := e.createNewToken(
		args.CallerAddr,
		args.Arguments[0],
		args.Arguments[1],
		big.NewInt(0),
		0,
		args.Arguments[2:],
		[]byte(core.NonFungibleDCDT))
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	e.eei.Finish(tokenIdentifier)

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(args.Function),
		Address:    args.CallerAddr,
		Topics:     [][]byte{tokenIdentifier, args.Arguments[0], args.Arguments[1], []byte(core.NonFungibleDCDT), big.NewInt(0).Bytes()},
	}
	e.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

func (e *dcdt) registerSemiFungible(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := e.checkBasicCreateArguments(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if len(args.Arguments) < 2 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.UserError
	}

	tokenIdentifier, _, err := e.createNewToken(
		args.CallerAddr,
		args.Arguments[0],
		args.Arguments[1],
		big.NewInt(0),
		0,
		args.Arguments[2:],
		[]byte(core.SemiFungibleDCDT))
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	e.eei.Finish(tokenIdentifier)

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(args.Function),
		Address:    args.CallerAddr,
		Topics:     [][]byte{tokenIdentifier, args.Arguments[0], args.Arguments[1], []byte(core.SemiFungibleDCDT), big.NewInt(0).Bytes()},
	}
	e.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

func (e *dcdt) registerMetaDCDT(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !e.enableEpochsHandler.IsFlagEnabled(common.MetaDCDTSetFlag) {
		e.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	returnCode := e.checkBasicCreateArguments(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if len(args.Arguments) < 3 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.UserError
	}
	numOfDecimals := uint32(big.NewInt(0).SetBytes(args.Arguments[2]).Uint64())
	if numOfDecimals < minNumberOfDecimals || numOfDecimals > maxNumberOfDecimals {
		e.eei.AddReturnMessage(fmt.Errorf("%w, minimum: %d, maximum: %d, provided: %d",
			vm.ErrInvalidNumberOfDecimals,
			minNumberOfDecimals,
			maxNumberOfDecimals,
			numOfDecimals,
		).Error())
		return vmcommon.UserError
	}

	tokenIdentifier, _, err := e.createNewToken(
		args.CallerAddr,
		args.Arguments[0],
		args.Arguments[1],
		big.NewInt(0),
		numOfDecimals,
		args.Arguments[3:],
		[]byte(metaDCDT))
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	e.eei.Finish(tokenIdentifier)

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(args.Function),
		Address:    args.CallerAddr,
		Topics:     [][]byte{tokenIdentifier, args.Arguments[0], args.Arguments[1], []byte(metaDCDT), big.NewInt(int64(numOfDecimals)).Bytes()},
	}
	e.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

// arguments list: tokenName, tickerID prefix, type of token, numDecimals, numGlobalSettings, listGlobalSettings, list(address, special roles)
func (e *dcdt) registerAndSetRoles(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !e.enableEpochsHandler.IsFlagEnabled(common.DCDTRegisterAndSetAllRolesFlag) {
		e.eei.AddReturnMessage("invalid method to call")
		return vmcommon.FunctionNotFound
	}

	returnCode := e.checkBasicCreateArguments(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if len(args.Arguments) != 4 {
		e.eei.AddReturnMessage("arguments length mismatch")
		return vmcommon.UserError
	}
	isWithDecimals, tokenType, err := getTokenType(args.Arguments[2])
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	numOfDecimals := uint32(0)
	if isWithDecimals {
		numOfDecimals = uint32(big.NewInt(0).SetBytes(args.Arguments[3]).Uint64())
		if numOfDecimals < minNumberOfDecimals || numOfDecimals > maxNumberOfDecimals {
			e.eei.AddReturnMessage(fmt.Errorf("%w, minimum: %d, maximum: %d, provided: %d",
				vm.ErrInvalidNumberOfDecimals,
				minNumberOfDecimals,
				maxNumberOfDecimals,
				numOfDecimals,
			).Error())
			return vmcommon.UserError
		}
	}

	tokenIdentifier, token, err := e.createNewToken(
		args.CallerAddr,
		args.Arguments[0],
		args.Arguments[1],
		big.NewInt(0),
		numOfDecimals,
		[][]byte{},
		tokenType)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	allRoles, err := getAllRolesForTokenType(string(tokenType))
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	properties, returnCode := e.setRolesForTokenAndAddress(token, args.CallerAddr, allRoles)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	returnCode = e.prepareAndSendRoleChangeData(tokenIdentifier, args.CallerAddr, allRoles, properties)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	err = e.saveToken(tokenIdentifier, token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(args.Function),
		Address:    args.CallerAddr,
		Topics:     [][]byte{tokenIdentifier, args.Arguments[0], args.Arguments[1], tokenType, big.NewInt(int64(numOfDecimals)).Bytes()},
	}
	e.eei.Finish(tokenIdentifier)
	e.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

func getAllRolesForTokenType(tokenType string) ([][]byte, error) {
	switch tokenType {
	case core.NonFungibleDCDT:
		return [][]byte{[]byte(core.DCDTRoleNFTCreate), []byte(core.DCDTRoleNFTBurn), []byte(core.DCDTRoleNFTUpdateAttributes), []byte(core.DCDTRoleNFTAddURI)}, nil
	case core.SemiFungibleDCDT, metaDCDT:
		return [][]byte{[]byte(core.DCDTRoleNFTCreate), []byte(core.DCDTRoleNFTBurn), []byte(core.DCDTRoleNFTAddQuantity)}, nil
	case core.FungibleDCDT:
		return [][]byte{[]byte(core.DCDTRoleLocalMint), []byte(core.DCDTRoleLocalBurn)}, nil
	}

	return nil, vm.ErrInvalidArgument
}

func getTokenType(compressed []byte) (bool, []byte, error) {
	// TODO: might extract the compressed constants to core, alongside metaDCDT
	switch string(compressed) {
	case "NFT":
		return false, []byte(core.NonFungibleDCDT), nil
	case "SFT":
		return false, []byte(core.SemiFungibleDCDT), nil
	case "META":
		return true, []byte(metaDCDT), nil
	case "FNG":
		return true, []byte(core.FungibleDCDT), nil
	}
	return false, nil, vm.ErrInvalidArgument
}

func (e *dcdt) changeSFTToMetaDCDT(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !e.enableEpochsHandler.IsFlagEnabled(common.MetaDCDTSetFlag) {
		e.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.UserError
	}
	numOfDecimals := uint32(big.NewInt(0).SetBytes(args.Arguments[1]).Uint64())
	if numOfDecimals < minNumberOfDecimals || numOfDecimals > maxNumberOfDecimals {
		e.eei.AddReturnMessage(fmt.Errorf("%w, minimum: %d, maximum: %d, provided: %d",
			vm.ErrInvalidNumberOfDecimals,
			minNumberOfDecimals,
			maxNumberOfDecimals,
			numOfDecimals,
		).Error())
		return vmcommon.UserError
	}
	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if !bytes.Equal(token.TokenType, []byte(core.SemiFungibleDCDT)) {
		e.eei.AddReturnMessage("change can happen to semi fungible tokens only")
		return vmcommon.UserError
	}

	token.TokenType = []byte(metaDCDT)
	token.NumDecimals = numOfDecimals
	err := e.saveToken(args.Arguments[0], token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(args.Function),
		Address:    args.CallerAddr,
		Topics:     [][]byte{args.Arguments[0], token.TokenName, token.TickerName, []byte(metaDCDT), args.Arguments[1]},
	}
	e.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

func (e *dcdt) createNewToken(
	owner []byte,
	tokenName []byte,
	tickerName []byte,
	initialSupply *big.Int,
	numDecimals uint32,
	properties [][]byte,
	tokenType []byte,
) ([]byte, *DCDTDataV2, error) {
	if !isTokenNameHumanReadable(tokenName) {
		return nil, nil, vm.ErrTokenNameNotHumanReadable
	}
	if !isTickerValid(tickerName) {
		return nil, nil, vm.ErrTickerNameNotValid
	}

	tokenIdentifier, err := e.createNewTokenIdentifier(owner, tickerName)
	if err != nil {
		return nil, nil, err
	}

	newDCDTToken := &DCDTDataV2{
		OwnerAddress:       owner,
		TokenName:          tokenName,
		TickerName:         tickerName,
		TokenType:          tokenType,
		NumDecimals:        numDecimals,
		MintedValue:        big.NewInt(0).Set(initialSupply),
		BurntValue:         big.NewInt(0),
		Upgradable:         true,
		CanAddSpecialRoles: true,
	}
	err = e.upgradeProperties(tokenIdentifier, newDCDTToken, properties, true, owner)
	if err != nil {
		return nil, nil, err
	}
	e.addBurnRoleAndSendToAllShards(newDCDTToken, tokenIdentifier)

	err = e.saveToken(tokenIdentifier, newDCDTToken)
	if err != nil {
		return nil, nil, err
	}

	return tokenIdentifier, newDCDTToken, nil
}

func isTickerValid(tickerName []byte) bool {
	if len(tickerName) < minLengthForTickerName || len(tickerName) > maxLengthForTickerName {
		return false
	}

	for _, ch := range tickerName {
		isBigCharacter := ch >= 'A' && ch <= 'Z'
		isNumber := ch >= '0' && ch <= '9'
		isReadable := isBigCharacter || isNumber
		if !isReadable {
			return false
		}
	}

	return true
}

func isTokenNameHumanReadable(tokenName []byte) bool {
	for _, ch := range tokenName {
		isSmallCharacter := ch >= 'a' && ch <= 'z'
		isBigCharacter := ch >= 'A' && ch <= 'Z'
		isNumber := ch >= '0' && ch <= '9'
		isReadable := isSmallCharacter || isBigCharacter || isNumber
		if !isReadable {
			return false
		}
	}
	return true
}

func (e *dcdt) createNewTokenIdentifier(caller []byte, ticker []byte) ([]byte, error) {
	newRandomBase := append(caller, e.eei.BlockChainHook().CurrentRandomSeed()...)
	newRandom := e.hasher.Compute(string(newRandomBase))
	newRandomForTicker := newRandom[:tickerRandomSequenceLength]

	tickerPrefix := append(ticker, []byte(tickerSeparator)...)
	newRandomAsBigInt := big.NewInt(0).SetBytes(newRandomForTicker)

	one := big.NewInt(1)
	for i := 0; i < numOfRetriesForIdentifier; i++ {
		encoded := fmt.Sprintf("%06x", newRandomAsBigInt)
		newIdentifier := append(tickerPrefix, encoded...)
		buff := e.eei.GetStorage(newIdentifier)
		if len(buff) == 0 {
			return newIdentifier, nil
		}
		newRandomAsBigInt.Add(newRandomAsBigInt, one)
	}

	return nil, vm.ErrCouldNotCreateNewTokenIdentifier
}

func (e *dcdt) upgradeProperties(tokenIdentifier []byte, token *DCDTDataV2, args [][]byte, isCreate bool, callerAddr []byte) error {
	mintBurnable := true
	if string(token.TokenType) != core.FungibleDCDT {
		mintBurnable = false
	}

	topics := make([][]byte, 0)
	nonce := big.NewInt(0)
	topics = append(topics, tokenIdentifier, nonce.Bytes())
	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(upgradeProperties),
		Address:    callerAddr,
	}

	if len(args) == 0 {
		topics = append(topics, []byte(upgradable), boolToSlice(token.Upgradable), []byte(canAddSpecialRoles), boolToSlice(token.CanAddSpecialRoles))
		logEntry.Topics = topics
		e.eei.AddLogEntry(logEntry)

		return nil
	}
	if len(args)%2 != 0 {
		return vm.ErrInvalidNumOfArguments
	}

	isUpgradablePropertyInArgs := false
	isCanAddSpecialRolePropertyInArgs := false
	for i := 0; i < len(args); i += 2 {
		optionalArg := string(args[i])
		val, err := checkAndGetSetting(string(args[i+1]))
		if err != nil {
			return err
		}
		switch optionalArg {
		case burnable:
			if !mintBurnable {
				return fmt.Errorf("%w only fungible tokens are burnable at system SC", vm.ErrInvalidArgument)
			}
			token.Burnable = val
		case mintable:
			if !mintBurnable {
				return fmt.Errorf("%w only fungible tokens are mintable at system SC", vm.ErrInvalidArgument)
			}
			token.Mintable = val
		case canPause:
			token.CanPause = val
		case canFreeze:
			token.CanFreeze = val
		case canWipe:
			token.CanWipe = val
		case upgradable:
			token.Upgradable = val
			isUpgradablePropertyInArgs = true
		case canChangeOwner:
			token.CanChangeOwner = val
		case canAddSpecialRoles:
			token.CanAddSpecialRoles = val
			isCanAddSpecialRolePropertyInArgs = true
		case canTransferNFTCreateRole:
			token.CanTransferNFTCreateRole = val
		case canCreateMultiShard:
			if !e.enableEpochsHandler.IsFlagEnabled(common.DCDTNFTCreateOnMultiShardFlag) {
				return vm.ErrInvalidArgument
			}
			if mintBurnable {
				return vm.ErrInvalidArgument
			}
			if !isCreate {
				return vm.ErrInvalidArgument
			}
			token.CanCreateMultiShard = val
		default:
			return vm.ErrInvalidArgument
		}
	}

	topics = append(topics, args...)
	if !isUpgradablePropertyInArgs {
		topics = append(topics, []byte(upgradable), boolToSlice(token.Upgradable))
	}
	if !isCanAddSpecialRolePropertyInArgs {
		topics = append(topics, []byte(canAddSpecialRoles), boolToSlice(token.CanAddSpecialRoles))
	}

	logEntry.Topics = topics
	e.eei.AddLogEntry(logEntry)

	return nil
}

func checkAndGetSetting(arg string) (bool, error) {
	if arg == "true" {
		return true, nil
	}
	if arg == "false" {
		return false, nil
	}
	return false, vm.ErrInvalidArgument
}

func getStringFromBool(val bool) string {
	if val {
		return "true"
	}
	return "false"
}

func (e *dcdt) burn(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !e.enableEpochsHandler.IsFlagEnabled(common.GlobalMintBurnFlag) {
		e.eei.AddReturnMessage("global burn is no more enabled, use local burn")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 2 {
		e.eei.AddReturnMessage("number of arguments must be equal with 2")
		return vmcommon.FunctionWrongSignature
	}
	if args.CallValue.Cmp(zero) != 0 {
		e.eei.AddReturnMessage("callValue must be 0")
		return vmcommon.OutOfFunds
	}
	burntValue := big.NewInt(0).SetBytes(args.Arguments[1])
	if burntValue.Cmp(big.NewInt(0)) <= 0 {
		e.eei.AddReturnMessage("negative or 0 value to burn")
		return vmcommon.UserError
	}
	token, err := e.getExistingToken(args.Arguments[0])
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if !token.Burnable {
		dcdtTransferData := core.BuiltInFunctionDCDTTransfer + "@" + hex.EncodeToString(args.Arguments[0]) + "@" + hex.EncodeToString(args.Arguments[1])
		e.eei.Transfer(args.CallerAddr, e.dcdtSCAddress, big.NewInt(0), []byte(dcdtTransferData), 0)
		e.eei.AddReturnMessage("token is not burnable")
		if e.enableEpochsHandler.IsFlagEnabled(common.MultiClaimOnDelegationFlag) {
			return vmcommon.UserError
		}
		return vmcommon.Ok
	}

	token.BurntValue.Add(token.BurntValue, burntValue)

	err = e.saveToken(args.Arguments[0], token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = e.eei.UseGas(args.GasProvided)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	return vmcommon.Ok
}

func (e *dcdt) mint(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !e.enableEpochsHandler.IsFlagEnabled(common.GlobalMintBurnFlag) {
		e.eei.AddReturnMessage("global mint is no more enabled, use local mint")
		return vmcommon.UserError
	}

	if len(args.Arguments) < 2 || len(args.Arguments) > 3 {
		e.eei.AddReturnMessage("accepted arguments number 2/3")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if len(args.Arguments[1]) > core.MaxLenForDCDTIssueMint {
		returnMessage := fmt.Sprintf("max length for dcdt mint is %d", core.MaxLenForDCDTIssueMint)
		e.eei.AddReturnMessage(returnMessage)
		return vmcommon.UserError
	}

	mintValue := big.NewInt(0).SetBytes(args.Arguments[1])
	if mintValue.Cmp(big.NewInt(0)) <= 0 {
		e.eei.AddReturnMessage("negative or zero mint value")
		return vmcommon.UserError
	}
	if !token.Mintable {
		e.eei.AddReturnMessage("token is not mintable")
		return vmcommon.UserError
	}

	token.MintedValue.Add(token.MintedValue, mintValue)
	err := e.saveToken(args.Arguments[0], token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	destination := token.OwnerAddress
	if len(args.Arguments) == 3 {
		if len(args.Arguments[2]) != len(args.CallerAddr) {
			e.eei.AddReturnMessage("destination address of invalid length")
			return vmcommon.UserError
		}
		destination = args.Arguments[2]
	}

	dcdtTransferData := core.BuiltInFunctionDCDTTransfer + "@" + hex.EncodeToString(args.Arguments[0]) + "@" + hex.EncodeToString(mintValue.Bytes())
	e.eei.Transfer(destination, e.dcdtSCAddress, big.NewInt(0), []byte(dcdtTransferData), 0)

	return vmcommon.Ok
}

func (e *dcdt) toggleFreeze(args *vmcommon.ContractCallInput, builtInFunc string) vmcommon.ReturnCode {
	if len(args.Arguments) != 2 {
		e.eei.AddReturnMessage("invalid number of arguments, wanted 2")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if !token.CanFreeze {
		e.eei.AddReturnMessage("cannot freeze")
		return vmcommon.UserError
	}

	if !e.isAddressValid(args.Arguments[1]) {
		e.eei.AddReturnMessage("invalid address to freeze/unfreeze")
		return vmcommon.UserError
	}

	dcdtTransferData := builtInFunc + "@" + hex.EncodeToString(args.Arguments[0])
	e.eei.Transfer(args.Arguments[1], e.dcdtSCAddress, big.NewInt(0), []byte(dcdtTransferData), 0)

	return vmcommon.Ok
}

func (e *dcdt) wipe(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 2 {
		e.eei.AddReturnMessage("invalid number of arguments, wanted 2")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	returnCode = e.wipeTokenFromAddress(args.Arguments[1], token, args.Arguments[0], args.Arguments[0])
	return returnCode
}

func (e *dcdt) toggleFreezeSingleNFT(args *vmcommon.ContractCallInput, builtInFunc string) vmcommon.ReturnCode {
	if len(args.Arguments) != 3 {
		e.eei.AddReturnMessage("invalid number of arguments, wanted 3")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if !token.CanFreeze {
		e.eei.AddReturnMessage("cannot freeze")
		return vmcommon.UserError
	}
	if string(token.TokenType) == core.FungibleDCDT {
		e.eei.AddReturnMessage("only non fungible tokens can be freezed per nonce")
		return vmcommon.UserError
	}
	if !e.isAddressValid(args.Arguments[2]) {
		e.eei.AddReturnMessage("invalid address to freeze/unfreeze")
		return vmcommon.UserError
	}
	if !isArgumentsUint64(args.Arguments[1]) {
		e.eei.AddReturnMessage("invalid second argument, wanted nonce as bigInt")
		return vmcommon.UserError
	}

	composedArg := append(args.Arguments[0], args.Arguments[1]...)
	dcdtTransferData := builtInFunc + "@" + hex.EncodeToString(composedArg)
	e.eei.Transfer(args.Arguments[2], e.dcdtSCAddress, big.NewInt(0), []byte(dcdtTransferData), 0)

	return vmcommon.Ok
}

func isArgumentsUint64(arg []byte) bool {
	argAsBigInt := big.NewInt(0).SetBytes(arg)
	return argAsBigInt.IsUint64()
}

func (e *dcdt) wipeTokenFromAddress(
	address []byte,
	token *DCDTDataV2,
	tokenID []byte,
	wipeArgument []byte,
) vmcommon.ReturnCode {
	if !token.CanWipe {
		e.eei.AddReturnMessage("cannot wipe")
		return vmcommon.UserError
	}
	if !e.isAddressValid(address) {
		e.eei.AddReturnMessage("invalid address to wipe")
		return vmcommon.UserError
	}

	dcdtTransferData := core.BuiltInFunctionDCDTWipe + "@" + hex.EncodeToString(wipeArgument)
	e.eei.Transfer(address, e.dcdtSCAddress, big.NewInt(0), []byte(dcdtTransferData), 0)

	token.NumWiped++
	err := e.saveToken(tokenID, token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *dcdt) wipeSingleNFT(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 3 {
		e.eei.AddReturnMessage("invalid number of arguments, wanted 3")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if string(token.TokenType) == core.FungibleDCDT {
		e.eei.AddReturnMessage("only non fungible tokens can be wiped per nonce")
		return vmcommon.UserError
	}
	if !isArgumentsUint64(args.Arguments[1]) {
		e.eei.AddReturnMessage("invalid second argument, wanted nonce as bigInt")
		return vmcommon.UserError
	}

	composedArg := append(args.Arguments[0], args.Arguments[1]...)
	returnCode = e.wipeTokenFromAddress(args.Arguments[2], token, args.Arguments[0], composedArg)
	return returnCode
}

func (e *dcdt) togglePause(args *vmcommon.ContractCallInput, builtInFunc string) vmcommon.ReturnCode {
	if len(args.Arguments) != 1 {
		e.eei.AddReturnMessage("invalid number of arguments, wanted 1")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if !token.CanPause {
		e.eei.AddReturnMessage("cannot pause/un-pause")
		return vmcommon.UserError
	}
	if token.IsPaused && builtInFunc == core.BuiltInFunctionDCDTPause {
		e.eei.AddReturnMessage("cannot pause an already paused contract")
		return vmcommon.UserError
	}
	if !token.IsPaused && builtInFunc == core.BuiltInFunctionDCDTUnPause {
		e.eei.AddReturnMessage("cannot unPause an already un-paused contract")
		return vmcommon.UserError
	}

	token.IsPaused = !token.IsPaused
	err := e.saveToken(args.Arguments[0], token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	dcdtTransferData := builtInFunc + "@" + hex.EncodeToString(args.Arguments[0])
	e.eei.SendGlobalSettingToAll(e.dcdtSCAddress, []byte(dcdtTransferData))

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(builtInFunc),
		Address:    args.CallerAddr,
		Topics:     [][]byte{args.Arguments[0]},
		Data:       nil,
	}
	e.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

func (e *dcdt) checkInputReturnDataBurnForAll(args *vmcommon.ContractCallInput) (*DCDTDataV2, vmcommon.ReturnCode) {
	isBurnForAllFlagEnabled := e.enableEpochsHandler.IsFlagEnabled(common.DCDTMetadataContinuousCleanupFlag)
	if !isBurnForAllFlagEnabled {
		e.eei.AddReturnMessage("invalid method to call")
		return nil, vmcommon.FunctionNotFound
	}
	if len(args.Arguments) != 1 {
		e.eei.AddReturnMessage("invalid number of arguments, wanted 1")
		return nil, vmcommon.FunctionWrongSignature
	}
	return e.basicOwnershipChecks(args)
}

func (e *dcdt) saveTokenAndSendForAll(token *DCDTDataV2, tokenID []byte, builtInCall string) vmcommon.ReturnCode {
	err := e.saveToken(tokenID, token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	dcdtTransferData := builtInCall + "@" + hex.EncodeToString(tokenID)
	e.eei.SendGlobalSettingToAll(e.dcdtSCAddress, []byte(dcdtTransferData))

	return vmcommon.Ok
}

func (e *dcdt) setBurnRoleGlobally(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	token, returnCode := e.checkInputReturnDataBurnForAll(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	burnForAllExists := checkIfDefinedRoleExistsInToken(token, []byte(vmcommon.DCDTRoleBurnForAll))
	if burnForAllExists {
		e.eei.AddReturnMessage("cannot set burn role globally as it was already set")
		return vmcommon.UserError
	}

	e.addBurnRoleAndSendToAllShards(token, args.Arguments[0])
	err := e.saveToken(args.Arguments[0], token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *dcdt) unsetBurnRoleGlobally(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	token, returnCode := e.checkInputReturnDataBurnForAll(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	burnForAllExists := checkIfDefinedRoleExistsInToken(token, []byte(vmcommon.DCDTRoleBurnForAll))
	if !burnForAllExists {
		e.eei.AddReturnMessage("cannot unset burn role globally as it was not set")
		return vmcommon.UserError
	}

	deleteRoleFromToken(token, []byte(vmcommon.DCDTRoleBurnForAll))

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(vmcommon.BuiltInFunctionDCDTUnSetBurnRoleForAll),
		Address:    args.CallerAddr,
		Topics:     [][]byte{args.Arguments[0], zero.Bytes(), zero.Bytes(), []byte(vmcommon.DCDTRoleBurnForAll)},
	}
	e.eei.AddLogEntry(logEntry)

	returnCode = e.saveTokenAndSendForAll(token, args.Arguments[0], vmcommon.BuiltInFunctionDCDTUnSetBurnRoleForAll)
	return returnCode
}

func (e *dcdt) addBurnRoleAndSendToAllShards(token *DCDTDataV2, tokenID []byte) {
	isBurnForAllFlagEnabled := e.enableEpochsHandler.IsFlagEnabled(common.DCDTMetadataContinuousCleanupFlag)
	if !isBurnForAllFlagEnabled {
		return
	}

	burnForAllRole := &DCDTRoles{Roles: [][]byte{[]byte(vmcommon.DCDTRoleBurnForAll)}, Address: []byte{}}
	token.SpecialRoles = append(token.SpecialRoles, burnForAllRole)

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(vmcommon.BuiltInFunctionDCDTSetBurnRoleForAll),
		Address:    token.OwnerAddress,
		Topics:     [][]byte{tokenID, zero.Bytes(), zero.Bytes(), []byte(vmcommon.DCDTRoleBurnForAll)},
	}
	e.eei.AddLogEntry(logEntry)

	dcdtTransferData := vmcommon.BuiltInFunctionDCDTSetBurnRoleForAll + "@" + hex.EncodeToString(tokenID)
	e.eei.SendGlobalSettingToAll(e.dcdtSCAddress, []byte(dcdtTransferData))
}

func (e *dcdt) configChange(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	owner, err := e.getDcdtOwner()
	if err != nil {
		e.eei.AddReturnMessage(fmt.Sprintf("could not get stored owner, error: %s", err.Error()))
		return vmcommon.UserError
	}

	isCorrectCaller := bytes.Equal(args.CallerAddr, owner) || bytes.Equal(args.CallerAddr, e.endOfEpochSCAddress)
	if !isCorrectCaller {
		e.eei.AddReturnMessage("configChange can be called by whitelisted address only")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		e.eei.AddReturnMessage("callValue must be 0")
		return vmcommon.UserError
	}
	err = e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.DCDTOperations)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 4 {
		e.eei.AddReturnMessage(vm.ErrInvalidNumOfArguments.Error())
		return vmcommon.UserError
	}

	newConfig := &DCDTConfig{
		OwnerAddress:       args.Arguments[0],
		BaseIssuingCost:    big.NewInt(0).SetBytes(args.Arguments[1]),
		MinTokenNameLength: uint32(big.NewInt(0).SetBytes(args.Arguments[2]).Uint64()),
		MaxTokenNameLength: uint32(big.NewInt(0).SetBytes(args.Arguments[3]).Uint64()),
	}

	if len(newConfig.OwnerAddress) != len(args.RecipientAddr) {
		e.eei.AddReturnMessage("invalid arguments, first argument must be a valid address")
		return vmcommon.UserError
	}
	if newConfig.BaseIssuingCost.Cmp(zero) < 0 {
		e.eei.AddReturnMessage("invalid new base issueing cost")
		return vmcommon.UserError
	}
	if newConfig.MinTokenNameLength > newConfig.MaxTokenNameLength {
		e.eei.AddReturnMessage("invalid min and max token name lengths")
		return vmcommon.UserError
	}

	err = e.saveDCDTConfig(newConfig)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *dcdt) claim(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	owner, err := e.getDcdtOwner()
	if err != nil {
		e.eei.AddReturnMessage(fmt.Sprintf("could not get stored owner, error: %s", err.Error()))
		return vmcommon.UserError
	}

	if !bytes.Equal(args.CallerAddr, owner) {
		e.eei.AddReturnMessage("claim can be called by whitelisted address only")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		e.eei.AddReturnMessage("callValue must be 0")
		return vmcommon.UserError
	}
	err = e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.DCDTOperations)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 0 {
		e.eei.AddReturnMessage(vm.ErrInvalidNumOfArguments.Error())
		return vmcommon.UserError
	}

	scBalance := e.eei.GetBalance(args.RecipientAddr)
	e.eei.Transfer(args.CallerAddr, args.RecipientAddr, scBalance, nil, 0)

	return vmcommon.Ok
}

func (e *dcdt) getTokenProperties(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		e.eei.AddReturnMessage("callValue must be 0")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		e.eei.AddReturnMessage(vm.ErrInvalidNumOfArguments.Error())
		return vmcommon.UserError
	}
	err := e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.DCDTOperations)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	dcdtToken, err := e.getExistingToken(args.Arguments[0])
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	e.eei.Finish(dcdtToken.TokenName)
	e.eei.Finish(dcdtToken.TokenType)
	e.eei.Finish(dcdtToken.OwnerAddress)
	e.eei.Finish([]byte(dcdtToken.MintedValue.String()))
	e.eei.Finish([]byte(dcdtToken.BurntValue.String()))
	e.eei.Finish([]byte(fmt.Sprintf("NumDecimals-%d", dcdtToken.NumDecimals)))
	e.eei.Finish([]byte("IsPaused-" + getStringFromBool(dcdtToken.IsPaused)))
	e.eei.Finish([]byte("CanUpgrade-" + getStringFromBool(dcdtToken.Upgradable)))
	e.eei.Finish([]byte("CanMint-" + getStringFromBool(dcdtToken.Mintable)))
	e.eei.Finish([]byte("CanBurn-" + getStringFromBool(dcdtToken.Burnable)))
	e.eei.Finish([]byte("CanChangeOwner-" + getStringFromBool(dcdtToken.CanChangeOwner)))
	e.eei.Finish([]byte("CanPause-" + getStringFromBool(dcdtToken.CanPause)))
	e.eei.Finish([]byte("CanFreeze-" + getStringFromBool(dcdtToken.CanFreeze)))
	e.eei.Finish([]byte("CanWipe-" + getStringFromBool(dcdtToken.CanWipe)))
	e.eei.Finish([]byte("CanAddSpecialRoles-" + getStringFromBool(dcdtToken.CanAddSpecialRoles)))
	e.eei.Finish([]byte("CanTransferNFTCreateRole-" + getStringFromBool(dcdtToken.CanTransferNFTCreateRole)))
	e.eei.Finish([]byte("NFTCreateStopped-" + getStringFromBool(dcdtToken.NFTCreateStopped)))
	e.eei.Finish([]byte(fmt.Sprintf("NumWiped-%d", dcdtToken.NumWiped)))

	return vmcommon.Ok
}

func (e *dcdt) getSpecialRoles(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		e.eei.AddReturnMessage("callValue must be 0")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		e.eei.AddReturnMessage(vm.ErrInvalidNumOfArguments.Error())
		return vmcommon.UserError
	}
	err := e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.DCDTOperations)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	dcdtToken, err := e.getExistingToken(args.Arguments[0])
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	specialRoles := dcdtToken.SpecialRoles
	for _, specialRole := range specialRoles {
		rolesAsString := make([]string, 0)
		for _, role := range specialRole.Roles {
			rolesAsString = append(rolesAsString, string(role))
		}

		roles := strings.Join(rolesAsString, ",")

		specialRoleAddress, errEncode := e.addressPubKeyConverter.Encode(specialRole.Address)
		e.treatEncodeErrorForGetSpecialRoles(errEncode, rolesAsString, specialRole.Address)

		message := fmt.Sprintf("%s:%s", specialRoleAddress, roles)
		e.eei.Finish([]byte(message))
	}

	return vmcommon.Ok
}

func (e *dcdt) treatEncodeErrorForGetSpecialRoles(err error, roles []string, address []byte) {
	if err == nil {
		return
	}

	logLevel := logger.LogTrace
	for _, role := range roles {
		if role != vmcommon.DCDTRoleBurnForAll {
			logLevel = logger.LogWarning
			break
		}
	}

	log.Log(logLevel, "dcdt.treatEncodeErrorForGetSpecialRoles",
		"hex specialRole.Address", hex.EncodeToString(address),
		"roles", strings.Join(roles, ", "),
		"error", err)
}

func (e *dcdt) basicOwnershipChecks(args *vmcommon.ContractCallInput) (*DCDTDataV2, vmcommon.ReturnCode) {
	if args.CallValue.Cmp(zero) != 0 {
		e.eei.AddReturnMessage("callValue must be 0")
		return nil, vmcommon.OutOfFunds
	}
	err := e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.DCDTOperations)
	if err != nil {
		e.eei.AddReturnMessage("not enough gas")
		return nil, vmcommon.OutOfGas
	}
	if len(args.Arguments) < 1 {
		e.eei.AddReturnMessage("not enough arguments")
		return nil, vmcommon.UserError
	}
	token, err := e.getExistingToken(args.Arguments[0])
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}
	if !bytes.Equal(token.OwnerAddress, args.CallerAddr) {
		e.eei.AddReturnMessage("can be called by owner only")
		return nil, vmcommon.UserError
	}

	return token, vmcommon.Ok
}

func (e *dcdt) transferOwnership(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 2 {
		e.eei.AddReturnMessage("expected num of arguments 2")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if !token.CanChangeOwner {
		e.eei.AddReturnMessage("cannot change owner of the token")
		return vmcommon.UserError
	}
	if !e.isAddressValid(args.Arguments[1]) {
		e.eei.AddReturnMessage("destination address is invalid")
		return vmcommon.UserError
	}

	token.OwnerAddress = args.Arguments[1]
	err := e.saveToken(args.Arguments[0], token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(args.Function),
		Address:    args.CallerAddr,
		Topics:     [][]byte{args.Arguments[0], token.TokenName, token.TickerName, token.TokenType, args.Arguments[1]},
	}
	e.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

func (e *dcdt) controlChanges(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 2 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if !token.Upgradable {
		e.eei.AddReturnMessage("token is not upgradable")
		return vmcommon.UserError
	}

	err := e.upgradeProperties(args.Arguments[0], token, args.Arguments[1:], false, args.CallerAddr)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	err = e.saveToken(args.Arguments[0], token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *dcdt) checkArgumentsForSpecialRoleChanges(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 3 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	if len(args.Arguments[1]) != len(args.CallerAddr) {
		e.eei.AddReturnMessage("second argument must be an address")
		return vmcommon.FunctionWrongSignature
	}
	err := checkDuplicates(args.Arguments[2:])
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func isDefinedRoleInArgs(args [][]byte, definedRole []byte) bool {
	for _, arg := range args {
		if bytes.Equal(arg, definedRole) {
			return true
		}
	}
	return false
}

func checkIfDefinedRoleExistsInArgsAndToken(args [][]byte, token *DCDTDataV2, definedRole []byte) bool {
	if !isDefinedRoleInArgs(args, definedRole) {
		return false
	}

	return checkIfDefinedRoleExistsInToken(token, definedRole)
}

func checkIfDefinedRoleExistsInToken(token *DCDTDataV2, definedRole []byte) bool {
	for _, dcdtRole := range token.SpecialRoles {
		for _, role := range dcdtRole.Roles {
			if bytes.Equal(role, definedRole) {
				return true
			}
		}
	}
	return false
}

func deleteRoleFromToken(token *DCDTDataV2, definedRole []byte) {
	var newRoles []*DCDTRoles
	for _, dcdtRole := range token.SpecialRoles {
		index := getRoleIndex(dcdtRole, definedRole)
		if index < 0 {
			newRoles = append(newRoles, dcdtRole)
			continue
		}

		copy(dcdtRole.Roles[index:], dcdtRole.Roles[index+1:])
		dcdtRole.Roles[len(dcdtRole.Roles)-1] = nil
		dcdtRole.Roles = dcdtRole.Roles[:len(dcdtRole.Roles)-1]

		if len(dcdtRole.Roles) > 0 {
			newRoles = append(newRoles, dcdtRole)
		}
	}
	token.SpecialRoles = newRoles
}

func (e *dcdt) isSpecialRoleValidForFungible(argument string) error {
	switch argument {
	case core.DCDTRoleLocalMint:
		return nil
	case core.DCDTRoleLocalBurn:
		return nil
	case core.DCDTRoleTransfer:
		if e.enableEpochsHandler.IsFlagEnabled(common.DCDTTransferRoleFlag) {
			return nil
		}
		return vm.ErrInvalidArgument
	default:
		return vm.ErrInvalidArgument
	}
}

func (e *dcdt) isSpecialRoleValidForSemiFungible(argument string) error {
	switch argument {
	case core.DCDTRoleNFTBurn:
		return nil
	case core.DCDTRoleNFTAddQuantity:
		return nil
	case core.DCDTRoleNFTCreate:
		return nil
	case core.DCDTRoleTransfer:
		if e.enableEpochsHandler.IsFlagEnabled(common.DCDTTransferRoleFlag) {
			return nil
		}
		return vm.ErrInvalidArgument
	default:
		return vm.ErrInvalidArgument
	}
}

func (e *dcdt) isSpecialRoleValidForNonFungible(argument string) error {
	switch argument {
	case core.DCDTRoleNFTBurn:
		return nil
	case core.DCDTRoleNFTCreate:
		return nil
	case core.DCDTRoleTransfer, core.DCDTRoleNFTUpdateAttributes, core.DCDTRoleNFTAddURI:
		if e.enableEpochsHandler.IsFlagEnabled(common.DCDTTransferRoleFlag) {
			return nil
		}
		return vm.ErrInvalidArgument
	default:
		return vm.ErrInvalidArgument
	}
}

func (e *dcdt) checkSpecialRolesAccordingToTokenType(args [][]byte, token *DCDTDataV2) error {
	switch string(token.TokenType) {
	case core.FungibleDCDT:
		return validateRoles(args, e.isSpecialRoleValidForFungible)
	case core.NonFungibleDCDT:
		return validateRoles(args, e.isSpecialRoleValidForNonFungible)
	case core.SemiFungibleDCDT:
		return validateRoles(args, e.isSpecialRoleValidForSemiFungible)
	case metaDCDT:
		isCheckMetaDCDTOnRolesFlagEnabled := e.enableEpochsHandler.IsFlagEnabled(common.ManagedCryptoAPIsFlag)
		if isCheckMetaDCDTOnRolesFlagEnabled {
			return validateRoles(args, e.isSpecialRoleValidForSemiFungible)
		}
	}
	return nil
}

type rolesProperties struct {
	transferRoleExists       bool
	isMultiShardNFTCreateSet bool
}

func getFirstAddressWithGivenRole(token *DCDTDataV2, definedRole []byte) ([]byte, error) {
	for _, dcdtRole := range token.SpecialRoles {
		for _, role := range dcdtRole.Roles {
			if bytes.Equal(role, definedRole) {
				return dcdtRole.Address, nil
			}
		}
	}

	return nil, vm.ErrElementNotFound
}

func (e *dcdt) changeToMultiShardCreate(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 1 {
		e.eei.AddReturnMessage("invalid number of arguments")
		return vmcommon.UserError
	}
	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if !token.CanAddSpecialRoles {
		e.eei.AddReturnMessage("cannot add special roles")
		return vmcommon.UserError
	}
	if token.CanCreateMultiShard {
		e.eei.AddReturnMessage("it is already multi shard create")
		return vmcommon.UserError
	}

	addressWithCreateRole, err := getFirstAddressWithGivenRole(token, []byte(core.DCDTRoleNFTCreate))
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	token.CanCreateMultiShard = true
	isAddressLastByteZero := addressWithCreateRole[len(addressWithCreateRole)-1] == 0
	if !isAddressLastByteZero {
		multiCreateRoleOnly := [][]byte{[]byte(core.DCDTRoleNFTCreateMultiShard)}
		e.sendRoleChangeData(args.Arguments[0], addressWithCreateRole, multiCreateRoleOnly, core.BuiltInFunctionSetDCDTRole)
	}

	err = e.saveToken(args.Arguments[0], token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *dcdt) setRolesForTokenAndAddress(
	token *DCDTDataV2,
	address []byte,
	roles [][]byte,
) (*rolesProperties, vmcommon.ReturnCode) {

	if !token.CanAddSpecialRoles {
		e.eei.AddReturnMessage("cannot add special roles")
		return nil, vmcommon.UserError
	}

	if e.enableEpochsHandler.IsFlagEnabled(common.NFTStopCreateFlag) && token.NFTCreateStopped && isDefinedRoleInArgs(roles, []byte(core.DCDTRoleNFTCreate)) {
		e.eei.AddReturnMessage("cannot add NFT create role as NFT creation was stopped")
		return nil, vmcommon.UserError
	}

	if !token.CanCreateMultiShard && checkIfDefinedRoleExistsInArgsAndToken(roles, token, []byte(core.DCDTRoleNFTCreate)) {
		e.eei.AddReturnMessage(vm.ErrNFTCreateRoleAlreadyExists.Error())
		return nil, vmcommon.UserError
	}

	err := e.checkSpecialRolesAccordingToTokenType(roles, token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}

	transferRoleExists := checkIfDefinedRoleExistsInArgsAndToken(roles, token, []byte(core.DCDTRoleTransfer))

	dcdtRole, isNew := getRolesForAddress(token, address)
	if isNew {
		dcdtRole.Roles = append(dcdtRole.Roles, roles...)
		token.SpecialRoles = append(token.SpecialRoles, dcdtRole)
	} else {
		for _, arg := range roles {
			index := getRoleIndex(dcdtRole, arg)
			if index >= 0 {
				e.eei.AddReturnMessage("special role already exists for given address")
				return nil, vmcommon.UserError
			}

			dcdtRole.Roles = append(dcdtRole.Roles, arg)
		}
	}

	isMultiShardNFTCreateSet := token.CanCreateMultiShard && isDefinedRoleInArgs(roles, []byte(core.DCDTRoleNFTCreate))
	if isMultiShardNFTCreateSet {
		err = checkCorrectAddressForNFTCreateMultiShard(token)
		if err != nil {
			e.eei.AddReturnMessage(err.Error())
			return nil, vmcommon.UserError
		}
	}

	properties := &rolesProperties{
		transferRoleExists:       transferRoleExists,
		isMultiShardNFTCreateSet: isMultiShardNFTCreateSet,
	}
	return properties, vmcommon.Ok
}

func (e *dcdt) prepareAndSendRoleChangeData(
	tokenID []byte,
	address []byte,
	roles [][]byte,
	properties *rolesProperties,
) vmcommon.ReturnCode {
	allRoles := roles
	if properties.isMultiShardNFTCreateSet {
		allRoles = append(allRoles, []byte(core.DCDTRoleNFTCreateMultiShard))
	}
	e.sendRoleChangeData(tokenID, address, allRoles, core.BuiltInFunctionSetDCDTRole)

	isTransferRoleDefinedInArgs := isDefinedRoleInArgs(roles, []byte(core.DCDTRoleTransfer))
	firstTransferRoleSet := !properties.transferRoleExists && isDefinedRoleInArgs(roles, []byte(core.DCDTRoleTransfer))
	if firstTransferRoleSet {
		dcdtTransferData := core.BuiltInFunctionDCDTSetLimitedTransfer + "@" + hex.EncodeToString(tokenID)
		e.eei.SendGlobalSettingToAll(e.dcdtSCAddress, []byte(dcdtTransferData))
	}

	if isTransferRoleDefinedInArgs {
		e.sendNewTransferRoleAddressToSystemAccount(tokenID, address)
	}

	return vmcommon.Ok
}

func (e *dcdt) setSpecialRole(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := e.checkArgumentsForSpecialRoleChanges(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	properties, returnCode := e.setRolesForTokenAndAddress(token, args.Arguments[1], args.Arguments[2:])
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	returnCode = e.prepareAndSendRoleChangeData(args.Arguments[0], args.Arguments[1], args.Arguments[2:], properties)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	err := e.saveToken(args.Arguments[0], token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func checkCorrectAddressForNFTCreateMultiShard(token *DCDTDataV2) error {
	nftCreateRole := []byte(core.DCDTRoleNFTCreate)
	mapLastBytes := make(map[uint8]struct{})
	for _, dcdtRole := range token.SpecialRoles {
		hasCreateRole := false
		for _, role := range dcdtRole.Roles {
			if bytes.Equal(nftCreateRole, role) {
				hasCreateRole = true
				break
			}
		}

		if !hasCreateRole {
			continue
		}

		lastBytesOfAddress := dcdtRole.Address[len(dcdtRole.Address)-1]
		_, exists := mapLastBytes[lastBytesOfAddress]
		if exists {
			return vm.ErrInvalidAddress
		}

		mapLastBytes[lastBytesOfAddress] = struct{}{}
	}

	return nil
}

func (e *dcdt) unSetSpecialRole(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := e.checkArgumentsForSpecialRoleChanges(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if isDefinedRoleInArgs(args.Arguments[2:], []byte(core.DCDTRoleNFTCreate)) {
		e.eei.AddReturnMessage("cannot un set NFT create role")
		return vmcommon.UserError
	}

	address := args.Arguments[1]
	dcdtRole, isNew := getRolesForAddress(token, address)
	if isNew {
		e.eei.AddReturnMessage("address does not have special role")
		return vmcommon.UserError
	}

	for _, arg := range args.Arguments[2:] {
		index := getRoleIndex(dcdtRole, arg)
		if index < 0 {
			e.eei.AddReturnMessage("special role does not exist for address")
			return vmcommon.UserError
		}

		copy(dcdtRole.Roles[index:], dcdtRole.Roles[index+1:])
		dcdtRole.Roles[len(dcdtRole.Roles)-1] = nil
		dcdtRole.Roles = dcdtRole.Roles[:len(dcdtRole.Roles)-1]
	}

	e.sendRoleChangeData(args.Arguments[0], args.Arguments[1], args.Arguments[2:], core.BuiltInFunctionUnSetDCDTRole)
	if len(dcdtRole.Roles) == 0 {
		for i, roles := range token.SpecialRoles {
			if bytes.Equal(roles.Address, address) {
				copy(token.SpecialRoles[i:], token.SpecialRoles[i+1:])
				token.SpecialRoles[len(token.SpecialRoles)-1] = nil
				token.SpecialRoles = token.SpecialRoles[:len(token.SpecialRoles)-1]

				break
			}
		}
	}

	isTransferRoleInArgs := isDefinedRoleInArgs(args.Arguments[2:], []byte(core.DCDTRoleTransfer))
	transferRoleExists := checkIfDefinedRoleExistsInArgsAndToken(args.Arguments[2:], token, []byte(core.DCDTRoleTransfer))
	lastTransferRoleWasDeleted := isTransferRoleInArgs && !transferRoleExists
	if lastTransferRoleWasDeleted {
		dcdtTransferData := core.BuiltInFunctionDCDTUnSetLimitedTransfer + "@" + hex.EncodeToString(args.Arguments[0])
		e.eei.SendGlobalSettingToAll(e.dcdtSCAddress, []byte(dcdtTransferData))
	}

	if isTransferRoleInArgs {
		e.deleteTransferRoleAddressFromSystemAccount(args.Arguments[0], address)
	}

	err := e.saveToken(args.Arguments[0], token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *dcdt) sendNewTransferRoleAddressToSystemAccount(token []byte, address []byte) {
	isSendTransferRoleAddressFlagEnabled := e.enableEpochsHandler.IsFlagEnabled(common.DCDTMetadataContinuousCleanupFlag)
	if !isSendTransferRoleAddressFlagEnabled {
		return
	}

	dcdtTransferData := vmcommon.BuiltInFunctionDCDTTransferRoleAddAddress + "@" + hex.EncodeToString(token) + "@" + hex.EncodeToString(address)
	e.eei.SendGlobalSettingToAll(e.dcdtSCAddress, []byte(dcdtTransferData))
}

func (e *dcdt) deleteTransferRoleAddressFromSystemAccount(token []byte, address []byte) {
	isSendTransferRoleAddressFlagEnabled := e.enableEpochsHandler.IsFlagEnabled(common.DCDTMetadataContinuousCleanupFlag)
	if !isSendTransferRoleAddressFlagEnabled {
		return
	}

	dcdtTransferData := vmcommon.BuiltInFunctionDCDTTransferRoleDeleteAddress + "@" + hex.EncodeToString(token) + "@" + hex.EncodeToString(address)
	e.eei.SendGlobalSettingToAll(e.dcdtSCAddress, []byte(dcdtTransferData))
}

func (e *dcdt) sendAllTransferRoleAddresses(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	isSendTransferRoleAddressFlagEnabled := e.enableEpochsHandler.IsFlagEnabled(common.DCDTMetadataContinuousCleanupFlag)
	if !isSendTransferRoleAddressFlagEnabled {
		e.eei.AddReturnMessage("invalid method to call")
		return vmcommon.FunctionNotFound
	}
	if len(args.Arguments) != 1 {
		e.eei.AddReturnMessage("wrong number of arguments, expected 1")
		return vmcommon.UserError
	}

	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	transferRoleFound := false
	dcdtTransferData := vmcommon.BuiltInFunctionDCDTTransferRoleAddAddress + "@" + hex.EncodeToString(args.Arguments[0])
	for _, role := range token.SpecialRoles {
		for _, actualRole := range role.Roles {
			if bytes.Equal(actualRole, []byte(core.DCDTRoleTransfer)) {
				dcdtTransferData += "@" + hex.EncodeToString(role.Address)
				transferRoleFound = true
				break
			}
		}
	}

	if !transferRoleFound {
		e.eei.AddReturnMessage("no address with transfer role")
		return vmcommon.UserError
	}

	e.eei.SendGlobalSettingToAll(e.dcdtSCAddress, []byte(dcdtTransferData))

	return vmcommon.Ok
}

func (e *dcdt) deleteNFTCreateRole(token *DCDTDataV2, currentNFTCreateOwner []byte) ([][]byte, vmcommon.ReturnCode) {
	creatorAddresses := make([][]byte, 0)
	for _, dcdtRole := range token.SpecialRoles {
		index := getRoleIndex(dcdtRole, []byte(core.DCDTRoleNFTCreate))
		if index < 0 {
			continue
		}

		if len(currentNFTCreateOwner) > 0 && !bytes.Equal(dcdtRole.Address, currentNFTCreateOwner) {
			e.eei.AddReturnMessage("address mismatch, second argument must be the one holding the NFT create role")
			return nil, vmcommon.UserError
		}

		copy(dcdtRole.Roles[index:], dcdtRole.Roles[index+1:])
		dcdtRole.Roles[len(dcdtRole.Roles)-1] = nil
		dcdtRole.Roles = dcdtRole.Roles[:len(dcdtRole.Roles)-1]

		creatorAddresses = append(creatorAddresses, dcdtRole.Address)
	}

	if len(creatorAddresses) > 0 {
		return creatorAddresses, vmcommon.Ok
	}
	e.eei.AddReturnMessage("no address is holding the NFT create role")
	return nil, vmcommon.UserError
}

func (e *dcdt) transferNFTCreateRole(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := e.checkArgumentsForSpecialRoleChanges(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if !token.CanTransferNFTCreateRole {
		e.eei.AddReturnMessage("NFT create role transfer is not allowed")
		return vmcommon.UserError
	}

	if bytes.Equal(token.TokenType, []byte(core.FungibleDCDT)) {
		e.eei.AddReturnMessage("invalid function call for fungible tokens")
		return vmcommon.UserError
	}
	if len(args.Arguments[2]) != len(args.CallerAddr) {
		e.eei.AddReturnMessage("third argument must be an address")
		return vmcommon.FunctionWrongSignature
	}
	if bytes.Equal(args.Arguments[1], args.Arguments[2]) {
		e.eei.AddReturnMessage("second and third arguments must not be equal")
		return vmcommon.FunctionWrongSignature
	}

	if token.CanCreateMultiShard {
		lastIndex := len(args.CallerAddr) - 1
		isLastByteTheSame := args.Arguments[1][lastIndex] == args.Arguments[2][lastIndex]
		if !isLastByteTheSame {
			e.eei.AddReturnMessage("transfer NFT create cannot happen for addresses that don't end in the same byte")
			return vmcommon.UserError
		}
	}

	_, returnCode = e.deleteNFTCreateRole(token, args.Arguments[1])
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	nextNFTCreateOwner := args.Arguments[2]
	dcdtRole, isNew := getRolesForAddress(token, nextNFTCreateOwner)
	dcdtRole.Roles = append(dcdtRole.Roles, []byte(core.DCDTRoleNFTCreate))
	if isNew {
		token.SpecialRoles = append(token.SpecialRoles, dcdtRole)
	}

	err := e.saveToken(args.Arguments[0], token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	dcdtTransferNFTCreateData := core.BuiltInFunctionDCDTNFTCreateRoleTransfer + "@" +
		hex.EncodeToString(args.Arguments[0]) + "@" + hex.EncodeToString(args.Arguments[2])
	e.eei.Transfer(args.Arguments[1], e.dcdtSCAddress, big.NewInt(0), []byte(dcdtTransferNFTCreateData), 0)

	return vmcommon.Ok
}

func (e *dcdt) stopNFTCreateForever(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 1 {
		e.eei.AddReturnMessage("number of arguments must be equal to 1")
		return vmcommon.FunctionWrongSignature
	}

	token, returnCode := e.basicOwnershipChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if token.NFTCreateStopped {
		e.eei.AddReturnMessage("NFT create was already stopped")
		return vmcommon.UserError
	}
	if bytes.Equal(token.TokenType, []byte(core.FungibleDCDT)) {
		e.eei.AddReturnMessage("invalid function call for fungible tokens")
		return vmcommon.UserError
	}

	currentOwners, returnCode := e.deleteNFTCreateRole(token, nil)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	token.NFTCreateStopped = true
	err := e.saveToken(args.Arguments[0], token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	for _, currentOwner := range currentOwners {
		e.sendRoleChangeData(args.Arguments[0], currentOwner, [][]byte{[]byte(core.DCDTRoleNFTCreate)}, core.BuiltInFunctionUnSetDCDTRole)
	}

	return vmcommon.Ok
}

func (e *dcdt) sendRoleChangeData(tokenID []byte, destination []byte, roles [][]byte, builtInFunc string) {
	dcdtSetRoleData := builtInFunc + "@" + hex.EncodeToString(tokenID)
	for _, arg := range roles {
		dcdtSetRoleData += "@" + hex.EncodeToString(arg)
	}

	e.eei.Transfer(destination, e.dcdtSCAddress, big.NewInt(0), []byte(dcdtSetRoleData), 0)
}

func (e *dcdt) getAllAddressesAndRoles(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 1 {
		e.eei.AddReturnMessage("needs 1 argument")
		return vmcommon.UserError
	}

	if args.CallValue.Cmp(zero) != 0 {
		e.eei.AddReturnMessage("callValue must be 0")
		return vmcommon.UserError
	}

	token, err := e.getExistingToken(args.Arguments[0])
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	for _, dcdtRole := range token.SpecialRoles {
		e.eei.Finish(dcdtRole.Address)
		for _, role := range dcdtRole.Roles {
			e.eei.Finish(role)
		}
	}

	return vmcommon.Ok
}

func checkDuplicates(arguments [][]byte) error {
	mapArgs := make(map[string]struct{})
	for _, arg := range arguments {
		_, found := mapArgs[string(arg)]
		if !found {
			mapArgs[string(arg)] = struct{}{}
			continue
		}
		return vm.ErrDuplicatesFoundInArguments
	}

	return nil
}

func validateRoles(arguments [][]byte, isSpecialRoleValid func(role string) error) error {
	for _, arg := range arguments {
		err := isSpecialRoleValid(string(arg))
		if err != nil {
			return err
		}
	}

	return nil
}

func getRoleIndex(dcdtRoles *DCDTRoles, role []byte) int {
	for i, currentRole := range dcdtRoles.Roles {
		if bytes.Equal(currentRole, role) {
			return i
		}
	}
	return -1
}

func getRolesForAddress(token *DCDTDataV2, address []byte) (*DCDTRoles, bool) {
	for _, dcdtRole := range token.SpecialRoles {
		if bytes.Equal(address, dcdtRole.Address) {
			return dcdtRole, false
		}
	}

	dcdtRole := &DCDTRoles{
		Address: address,
	}
	return dcdtRole, true
}

func (e *dcdt) saveTokenV1(identifier []byte, token *DCDTDataV2) error {
	tokenV1 := &DCDTDataV1{
		OwnerAddress:             token.OwnerAddress,
		TokenName:                token.TokenName,
		TickerName:               token.TickerName,
		TokenType:                token.TokenType,
		Mintable:                 token.Mintable,
		Burnable:                 token.Burnable,
		CanPause:                 token.CanPause,
		CanFreeze:                token.CanFreeze,
		CanWipe:                  token.CanWipe,
		Upgradable:               token.Upgradable,
		CanChangeOwner:           token.CanChangeOwner,
		IsPaused:                 token.IsPaused,
		MintedValue:              token.MintedValue,
		BurntValue:               token.BurntValue,
		NumDecimals:              token.NumDecimals,
		CanAddSpecialRoles:       token.CanAddSpecialRoles,
		NFTCreateStopped:         token.NFTCreateStopped,
		CanTransferNFTCreateRole: token.CanTransferNFTCreateRole,
		SpecialRoles:             token.SpecialRoles,
		NumWiped:                 token.NumWiped,
	}
	marshaledData, err := e.marshalizer.Marshal(tokenV1)
	if err != nil {
		return err
	}

	e.eei.SetStorage(identifier, marshaledData)
	return nil
}

func (e *dcdt) saveToken(identifier []byte, token *DCDTDataV2) error {
	if !e.enableEpochsHandler.IsFlagEnabled(common.DCDTNFTCreateOnMultiShardFlag) {
		return e.saveTokenV1(identifier, token)
	}

	marshaledData, err := e.marshalizer.Marshal(token)
	if err != nil {
		return err
	}

	e.eei.SetStorage(identifier, marshaledData)
	return nil
}

func (e *dcdt) getExistingToken(tokenIdentifier []byte) (*DCDTDataV2, error) {
	marshaledData := e.eei.GetStorage(tokenIdentifier)
	if len(marshaledData) == 0 {
		return nil, vm.ErrNoTickerWithGivenName
	}

	token := &DCDTDataV2{}
	err := e.marshalizer.Unmarshal(token, marshaledData)
	return token, err
}

func (e *dcdt) getDCDTConfig() (*DCDTConfig, error) {
	dcdtConfig := &DCDTConfig{
		OwnerAddress:       e.ownerAddress,
		BaseIssuingCost:    e.baseIssuingCost,
		MinTokenNameLength: minLengthForInitTokenName,
		MaxTokenNameLength: maxLengthForTokenName,
	}
	marshalledData := e.eei.GetStorage([]byte(configKeyPrefix))
	if len(marshalledData) == 0 {
		return dcdtConfig, nil
	}

	err := e.marshalizer.Unmarshal(dcdtConfig, marshalledData)
	return dcdtConfig, err
}

func (e *dcdt) getDcdtOwner() ([]byte, error) {
	dcdtConfig, err := e.getDCDTConfig()
	if err != nil {
		return nil, err
	}

	return dcdtConfig.OwnerAddress, nil
}

func (e *dcdt) saveDCDTConfig(dcdtConfig *DCDTConfig) error {
	marshaledData, err := e.marshalizer.Marshal(dcdtConfig)
	if err != nil {
		return err
	}

	e.eei.SetStorage([]byte(configKeyPrefix), marshaledData)
	return nil
}

// SetNewGasCost is called whenever a gas cost was changed
func (e *dcdt) SetNewGasCost(gasCost vm.GasCost) {
	e.mutExecution.Lock()
	e.gasCost = gasCost
	e.mutExecution.Unlock()
}

func (e *dcdt) isAddressValid(addressBytes []byte) bool {
	isLengthOk := len(addressBytes) == e.addressPubKeyConverter.Len()
	if !isLengthOk {
		return false
	}

	encodedAddress := e.addressPubKeyConverter.SilentEncode(addressBytes, log)

	return encodedAddress != ""
}

// CanUseContract returns true if contract can be used
func (e *dcdt) CanUseContract() bool {
	return true
}

func (e *dcdt) getContractConfig(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := e.checkArgumentsForGeneralViewFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	dcdtConfig, err := e.getDCDTConfig()
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	e.eei.Finish(dcdtConfig.OwnerAddress)
	e.eei.Finish(dcdtConfig.BaseIssuingCost.Bytes())
	e.eei.Finish(big.NewInt(int64(dcdtConfig.MinTokenNameLength)).Bytes())
	e.eei.Finish(big.NewInt(int64(dcdtConfig.MaxTokenNameLength)).Bytes())

	return vmcommon.Ok
}

func (e *dcdt) checkArgumentsForGeneralViewFunc(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		e.eei.AddReturnMessage(vm.ErrCallValueMustBeZero.Error())
		return vmcommon.UserError
	}
	err := e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.DCDTOperations)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 0 {
		e.eei.AddReturnMessage(vm.ErrInvalidNumOfArguments.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

// IsInterfaceNil returns true if underlying object is nil
func (e *dcdt) IsInterfaceNil() bool {
	return e == nil
}
