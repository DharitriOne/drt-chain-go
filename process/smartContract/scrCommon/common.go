package scrCommon

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data"
	"github.com/DharitriOne/drt-chain-core-go/hashing"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/process"
	"github.com/DharitriOne/drt-chain-go/sharding"
	"github.com/DharitriOne/drt-chain-go/state"
	"github.com/DharitriOne/drt-chain-go/storage"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
	"math/big"
)

// TestSmartContractProcessor is a SmartContractProcessor used in integration tests
type TestSmartContractProcessor interface {
	process.SmartContractProcessorFacade
	GetCompositeTestError() error
	GetGasRemaining() uint64
	GetAllSCRs() []data.TransactionHandler
	CleanGasRefunded()
}

// ExecutableChecker is an interface for checking if a builtin function is executable
type ExecutableChecker interface {
	CheckIsExecutable(senderAddr []byte, value *big.Int, receiverAddr []byte, gasProvidedForCall uint64, arguments [][]byte) error
}

// ArgsNewSmartContractProcessor defines the arguments needed for new smart contract processor
type ArgsNewSmartContractProcessor struct {
	VmContainer         process.VirtualMachinesContainer
	ArgsParser          process.ArgumentsParser
	Hasher              hashing.Hasher
	Marshalizer         marshal.Marshalizer
	AccountsDB          state.AccountsAdapter
	BlockChainHook      process.BlockChainHookHandler
	BuiltInFunctions    vmcommon.BuiltInFunctionContainer
	PubkeyConv          core.PubkeyConverter
	ShardCoordinator    sharding.Coordinator
	ScrForwarder        process.IntermediateTransactionHandler
	TxFeeHandler        process.TransactionFeeHandler
	EconomicsFee        process.FeeHandler
	TxTypeHandler       process.TxTypeHandler
	GasHandler          process.GasHandler
	GasSchedule         core.GasScheduleNotifier
	TxLogsProcessor     process.TransactionLogProcessor
	BadTxForwarder      process.IntermediateTransactionHandler
	EnableRoundsHandler process.EnableRoundsHandler
	EnableEpochsHandler common.EnableEpochsHandler
	EnableEpochs        config.EnableEpochs
	VMOutputCacher      storage.Cacher
	WasmVMChangeLocker  common.Locker
	IsGenesisProcessing bool
}

// FindVMByScAddress is exported for use in all version of scr processors
func FindVMByScAddress(container process.VirtualMachinesContainer, scAddress []byte) (vmcommon.VMExecutionHandler, []byte, error) {
	vmType, err := vmcommon.ParseVMTypeFromContractAddress(scAddress)
	if err != nil {
		return nil, nil, err
	}

	vm, err := container.Get(vmType)
	if err != nil {
		return nil, nil, err
	}

	return vm, vmType, nil
}

// CreateExecutableCheckersMap creates a map of executable checker builtin functions
func CreateExecutableCheckersMap(builtinFunctions vmcommon.BuiltInFunctionContainer) map[string]ExecutableChecker {
	executableCheckers := make(map[string]ExecutableChecker)

	for key := range builtinFunctions.Keys() {
		builtinFunc, err := builtinFunctions.Get(key)
		if err != nil {
			continue
		}
		executableCheckerFunc, ok := builtinFunc.(ExecutableChecker)
		if !ok {
			continue
		}
		executableCheckers[key] = executableCheckerFunc
	}

	return executableCheckers
}
