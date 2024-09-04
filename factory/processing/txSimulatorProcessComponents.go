package processing

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	dataBlock "github.com/DharitriOne/drt-chain-core-go/data/block"
	"github.com/DharitriOne/drt-chain-go/common/disabled"
	bootstrapDisabled "github.com/DharitriOne/drt-chain-go/epochStart/bootstrap/disabled"
	"github.com/DharitriOne/drt-chain-go/factory"
	"github.com/DharitriOne/drt-chain-go/genesis"
	processDisabled "github.com/DharitriOne/drt-chain-go/genesis/process/disabled"
	"github.com/DharitriOne/drt-chain-go/process"
	"github.com/DharitriOne/drt-chain-go/process/block/preprocess"
	"github.com/DharitriOne/drt-chain-go/process/coordinator"
	"github.com/DharitriOne/drt-chain-go/process/factory/shard"
	"github.com/DharitriOne/drt-chain-go/process/smartContract"
	"github.com/DharitriOne/drt-chain-go/process/smartContract/scrCommon"
	"github.com/DharitriOne/drt-chain-go/process/transaction"
	"github.com/DharitriOne/drt-chain-go/process/transactionEvaluator"
	"github.com/DharitriOne/drt-chain-go/process/transactionLog"
	"github.com/DharitriOne/drt-chain-go/state"
	"github.com/DharitriOne/drt-chain-go/state/syncer"
	"github.com/DharitriOne/drt-chain-go/storage"
	storageFactory "github.com/DharitriOne/drt-chain-go/storage/factory"
	"github.com/DharitriOne/drt-chain-go/storage/storageunit"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
	"github.com/DharitriOne/drt-chain-vm-common-go/parsers"
	datafield "github.com/DharitriOne/drt-chain-vm-common-go/parsers/dataField"
)

func (pcf *processComponentsFactory) createAPITransactionEvaluator() (factory.TransactionEvaluator, process.VirtualMachinesContainerFactory, error) {
	simulationAccountsDB, err := transactionEvaluator.NewSimulationAccountsDB(pcf.state.AccountsAdapterAPI())
	if err != nil {
		return nil, nil, err
	}

	vmOutputCacherConfig := storageFactory.GetCacherFromConfig(pcf.config.VMOutputCacher)
	vmOutputCacher, err := storageunit.NewCache(vmOutputCacherConfig)
	if err != nil {
		return nil, nil, err
	}

	txLogsProcessor, err := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Marshalizer:          pcf.coreData.InternalMarshalizer(),
		SaveInStorageEnabled: false, // no storer needed for tx simulator
	})
	if err != nil {
		return nil, nil, err
	}

	txSimulatorProcessorArgs, vmContainerFactory, txTypeHandler, err := pcf.createArgsTxSimulatorProcessor(simulationAccountsDB, vmOutputCacher, txLogsProcessor)
	if err != nil {
		return nil, nil, err
	}

	dataFieldParser, err := datafield.NewOperationDataFieldParser(&datafield.ArgsOperationDataFieldParser{
		AddressLength: pcf.coreData.AddressPubKeyConverter().Len(),
		Marshalizer:   pcf.coreData.InternalMarshalizer(),
	})
	if err != nil {
		return nil, nil, err
	}

	txSimulatorProcessorArgs.VMOutputCacher = vmOutputCacher
	txSimulatorProcessorArgs.AddressPubKeyConverter = pcf.coreData.AddressPubKeyConverter()
	txSimulatorProcessorArgs.ShardCoordinator = pcf.bootstrapComponents.ShardCoordinator()
	txSimulatorProcessorArgs.Hasher = pcf.coreData.Hasher()
	txSimulatorProcessorArgs.Marshalizer = pcf.coreData.InternalMarshalizer()
	txSimulatorProcessorArgs.DataFieldParser = dataFieldParser

	txSimulator, err := transactionEvaluator.NewTransactionSimulator(txSimulatorProcessorArgs)
	if err != nil {
		return nil, nil, err
	}

	apiTransactionEvaluator, err := transactionEvaluator.NewAPITransactionEvaluator(transactionEvaluator.ArgsApiTransactionEvaluator{
		TxTypeHandler:       txTypeHandler,
		FeeHandler:          pcf.coreData.EconomicsData(),
		TxSimulator:         txSimulator,
		Accounts:            simulationAccountsDB,
		ShardCoordinator:    pcf.bootstrapComponents.ShardCoordinator(),
		EnableEpochsHandler: pcf.coreData.EnableEpochsHandler(),
		BlockChain:          pcf.data.Blockchain(),
	})

	return apiTransactionEvaluator, vmContainerFactory, err
}

func (pcf *processComponentsFactory) createArgsTxSimulatorProcessor(
	accountsAdapter state.AccountsAdapter,
	vmOutputCacher storage.Cacher,
	txLogsProcessor process.TransactionLogProcessor,
) (transactionEvaluator.ArgsTxSimulator, process.VirtualMachinesContainerFactory, process.TxTypeHandler, error) {
	shardID := pcf.bootstrapComponents.ShardCoordinator().SelfId()
	if shardID == core.MetachainShardId {
		return pcf.createArgsTxSimulatorProcessorForMeta(accountsAdapter, vmOutputCacher, txLogsProcessor)
	} else {
		return pcf.createArgsTxSimulatorProcessorShard(accountsAdapter, vmOutputCacher, txLogsProcessor)
	}
}

func (pcf *processComponentsFactory) createArgsTxSimulatorProcessorForMeta(
	accountsAdapter state.AccountsAdapter,
	vmOutputCacher storage.Cacher,
	txLogsProcessor process.TransactionLogProcessor,
) (transactionEvaluator.ArgsTxSimulator, process.VirtualMachinesContainerFactory, process.TxTypeHandler, error) {
	args := transactionEvaluator.ArgsTxSimulator{}

	argsFactory := shard.ArgsNewIntermediateProcessorsContainerFactory{
		ShardCoordinator:        pcf.bootstrapComponents.ShardCoordinator(),
		Marshalizer:             pcf.coreData.InternalMarshalizer(),
		Hasher:                  pcf.coreData.Hasher(),
		PubkeyConverter:         pcf.coreData.AddressPubKeyConverter(),
		Store:                   bootstrapDisabled.NewChainStorer(),
		PoolsHolder:             pcf.data.Datapool(),
		EconomicsFee:            &processDisabled.FeeHandler{},
		EnableEpochsHandler:     pcf.coreData.EnableEpochsHandler(),
		TxExecutionOrderHandler: disabled.NewOrderedCollection(),
	}
	intermediateProcessorsFactory, err := shard.NewIntermediateProcessorsContainerFactory(argsFactory)
	if err != nil {
		return args, nil, nil, err
	}

	intermediateProcessorsContainer, err := intermediateProcessorsFactory.Create()
	if err != nil {
		return args, nil, nil, err
	}

	builtInFuncFactory, err := pcf.createBuiltInFunctionContainer(accountsAdapter, make(map[string]struct{}))
	if err != nil {
		return args, nil, nil, err
	}

	vmContainerFactory, err := pcf.createVMFactoryMeta(
		accountsAdapter,
		builtInFuncFactory.BuiltInFunctionContainer(),
		pcf.config.SmartContractsStorageSimulate,
		builtInFuncFactory.NFTStorageHandler(),
		builtInFuncFactory.DCDTGlobalSettingsHandler(),
	)
	if err != nil {
		return args, nil, nil, err
	}

	args.BlockChainHook = vmContainerFactory.BlockChainHookImpl()

	vmContainer, err := vmContainerFactory.Create()
	if err != nil {
		return args, nil, nil, err
	}

	txTypeHandler, err := pcf.createTxTypeHandler(builtInFuncFactory)
	if err != nil {
		return args, nil, nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(
		pcf.coreData.EconomicsData(),
		txTypeHandler,
		pcf.coreData.EnableEpochsHandler(),
	)
	if err != nil {
		return args, nil, nil, err
	}

	scForwarder, err := intermediateProcessorsContainer.Get(dataBlock.SmartContractResultBlock)
	if err != nil {
		return args, nil, nil, err
	}
	badTxInterim, err := intermediateProcessorsContainer.Get(dataBlock.InvalidBlock)
	if err != nil {
		return args, nil, nil, err
	}

	scProcArgs := scrCommon.ArgsNewSmartContractProcessor{
		VmContainer:         vmContainer,
		ArgsParser:          smartContract.NewArgumentParser(),
		Hasher:              pcf.coreData.Hasher(),
		Marshalizer:         pcf.coreData.InternalMarshalizer(),
		AccountsDB:          accountsAdapter,
		BlockChainHook:      vmContainerFactory.BlockChainHookImpl(),
		BuiltInFunctions:    builtInFuncFactory.BuiltInFunctionContainer(),
		PubkeyConv:          pcf.coreData.AddressPubKeyConverter(),
		ShardCoordinator:    pcf.bootstrapComponents.ShardCoordinator(),
		ScrForwarder:        scForwarder,
		TxFeeHandler:        &processDisabled.FeeHandler{},
		EconomicsFee:        pcf.coreData.EconomicsData(),
		TxTypeHandler:       txTypeHandler,
		GasHandler:          gasHandler,
		GasSchedule:         pcf.gasSchedule,
		TxLogsProcessor:     txLogsProcessor,
		EnableEpochsHandler: pcf.coreData.EnableEpochsHandler(),
		EnableRoundsHandler: pcf.coreData.EnableRoundsHandler(),
		BadTxForwarder:      badTxInterim,
		VMOutputCacher:      vmOutputCacher,
		WasmVMChangeLocker:  pcf.coreData.WasmVMChangeLocker(),
		IsGenesisProcessing: false,
	}

	scProcessor, err := smartContract.NewSmartContractProcessor(scProcArgs)
	if err != nil {
		return args, nil, nil, err
	}

	argsTxProcessor := transaction.ArgsNewMetaTxProcessor{
		Hasher:              pcf.coreData.Hasher(),
		Marshalizer:         pcf.coreData.InternalMarshalizer(),
		Accounts:            accountsAdapter,
		PubkeyConv:          pcf.coreData.AddressPubKeyConverter(),
		ShardCoordinator:    pcf.bootstrapComponents.ShardCoordinator(),
		ScProcessor:         scProcessor,
		TxTypeHandler:       txTypeHandler,
		EconomicsFee:        pcf.coreData.EconomicsData(),
		EnableEpochsHandler: pcf.coreData.EnableEpochsHandler(),
		TxVersionChecker:    pcf.coreData.TxVersionChecker(),
		GuardianChecker:     pcf.bootstrapComponents.GuardedAccountHandler(),
	}

	txProcessor, err := transaction.NewMetaTxProcessor(argsTxProcessor)
	if err != nil {
		return args, nil, nil, err
	}

	args.TransactionProcessor = txProcessor
	args.IntermediateProcContainer = intermediateProcessorsContainer

	return args, vmContainerFactory, txTypeHandler, nil
}

func (pcf *processComponentsFactory) createTxTypeHandler(builtInFuncFactory vmcommon.BuiltInFunctionFactory) (process.TxTypeHandler, error) {
	dcdtTransferParser, err := parsers.NewDCDTTransferParser(pcf.coreData.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:     pcf.coreData.AddressPubKeyConverter(),
		ShardCoordinator:    pcf.bootstrapComponents.ShardCoordinator(),
		BuiltInFunctions:    builtInFuncFactory.BuiltInFunctionContainer(),
		ArgumentParser:      parsers.NewCallArgsParser(),
		DCDTTransferParser:  dcdtTransferParser,
		EnableEpochsHandler: pcf.coreData.EnableEpochsHandler(),
	}

	return coordinator.NewTxTypeHandler(argsTxTypeHandler)
}

func (pcf *processComponentsFactory) createArgsTxSimulatorProcessorShard(
	accountsAdapter state.AccountsAdapter,
	vmOutputCacher storage.Cacher,
	txLogsProcessor process.TransactionLogProcessor,
) (transactionEvaluator.ArgsTxSimulator, process.VirtualMachinesContainerFactory, process.TxTypeHandler, error) {
	args := transactionEvaluator.ArgsTxSimulator{}

	argsFactory := shard.ArgsNewIntermediateProcessorsContainerFactory{
		ShardCoordinator:        pcf.bootstrapComponents.ShardCoordinator(),
		Marshalizer:             pcf.coreData.InternalMarshalizer(),
		Hasher:                  pcf.coreData.Hasher(),
		PubkeyConverter:         pcf.coreData.AddressPubKeyConverter(),
		Store:                   bootstrapDisabled.NewChainStorer(),
		PoolsHolder:             pcf.data.Datapool(),
		EconomicsFee:            &processDisabled.FeeHandler{},
		EnableEpochsHandler:     pcf.coreData.EnableEpochsHandler(),
		TxExecutionOrderHandler: disabled.NewOrderedCollection(),
	}

	intermediateProcessorsFactory, err := shard.NewIntermediateProcessorsContainerFactory(argsFactory)
	if err != nil {
		return args, nil, nil, err
	}

	intermediateProcessorsContainer, err := intermediateProcessorsFactory.Create()
	if err != nil {
		return args, nil, nil, err
	}

	mapDNSAddresses, err := pcf.smartContractParser.GetDeployedSCAddresses(genesis.DNSType)
	if err != nil {
		return args, nil, nil, err
	}

	builtInFuncFactory, err := pcf.createBuiltInFunctionContainer(accountsAdapter, mapDNSAddresses)
	if err != nil {
		return args, nil, nil, err
	}

	smartContractStorageSimulate := pcf.config.SmartContractsStorageSimulate
	dcdtTransferParser, err := parsers.NewDCDTTransferParser(pcf.coreData.InternalMarshalizer())
	if err != nil {
		return args, nil, nil, err
	}

	vmContainerFactory, err := pcf.createVMFactoryShard(
		accountsAdapter,
		syncer.NewMissingTrieNodesNotifier(),
		builtInFuncFactory.BuiltInFunctionContainer(),
		dcdtTransferParser,
		pcf.coreData.WasmVMChangeLocker(),
		smartContractStorageSimulate,
		builtInFuncFactory.NFTStorageHandler(),
		builtInFuncFactory.DCDTGlobalSettingsHandler(),
	)
	if err != nil {
		return args, nil, nil, err
	}

	args.BlockChainHook = vmContainerFactory.BlockChainHookImpl()

	err = builtInFuncFactory.SetPayableHandler(vmContainerFactory.BlockChainHookImpl())
	if err != nil {
		return args, nil, nil, err
	}

	vmContainer, err := vmContainerFactory.Create()
	if err != nil {
		return args, nil, nil, err
	}

	txTypeHandler, err := pcf.createTxTypeHandler(builtInFuncFactory)
	if err != nil {
		return args, nil, nil, err
	}
	txFeeHandler := &processDisabled.FeeHandler{}

	gasHandler, err := preprocess.NewGasComputation(
		pcf.coreData.EconomicsData(),
		txTypeHandler,
		pcf.coreData.EnableEpochsHandler(),
	)
	if err != nil {
		return args, nil, nil, err
	}

	scForwarder, err := intermediateProcessorsContainer.Get(dataBlock.SmartContractResultBlock)
	if err != nil {
		return args, nil, nil, err
	}
	badTxInterim, err := intermediateProcessorsContainer.Get(dataBlock.InvalidBlock)
	if err != nil {
		return args, nil, nil, err
	}
	receiptTxInterim, err := intermediateProcessorsContainer.Get(dataBlock.ReceiptBlock)
	if err != nil {
		return args, nil, nil, err
	}

	argsParser := smartContract.NewArgumentParser()

	scProcArgs := scrCommon.ArgsNewSmartContractProcessor{
		VmContainer:         vmContainer,
		ArgsParser:          argsParser,
		Hasher:              pcf.coreData.Hasher(),
		Marshalizer:         pcf.coreData.InternalMarshalizer(),
		AccountsDB:          accountsAdapter,
		BlockChainHook:      vmContainerFactory.BlockChainHookImpl(),
		BuiltInFunctions:    builtInFuncFactory.BuiltInFunctionContainer(),
		PubkeyConv:          pcf.coreData.AddressPubKeyConverter(),
		ShardCoordinator:    pcf.bootstrapComponents.ShardCoordinator(),
		ScrForwarder:        scForwarder,
		TxFeeHandler:        &processDisabled.FeeHandler{},
		EconomicsFee:        pcf.coreData.EconomicsData(),
		TxTypeHandler:       txTypeHandler,
		GasHandler:          gasHandler,
		GasSchedule:         pcf.gasSchedule,
		TxLogsProcessor:     txLogsProcessor,
		EnableEpochsHandler: pcf.coreData.EnableEpochsHandler(),
		EnableRoundsHandler: pcf.coreData.EnableRoundsHandler(),
		BadTxForwarder:      badTxInterim,
		VMOutputCacher:      vmOutputCacher,
		WasmVMChangeLocker:  pcf.coreData.WasmVMChangeLocker(),
		IsGenesisProcessing: false,
	}

	scProcessor, err := smartContract.NewSmartContractProcessor(scProcArgs)
	if err != nil {
		return args, nil, nil, err
	}

	argsTxProcessor := transaction.ArgsNewTxProcessor{
		Accounts:            accountsAdapter,
		Hasher:              pcf.coreData.Hasher(),
		PubkeyConv:          pcf.coreData.AddressPubKeyConverter(),
		Marshalizer:         pcf.coreData.InternalMarshalizer(),
		SignMarshalizer:     pcf.coreData.TxMarshalizer(),
		ShardCoordinator:    pcf.bootstrapComponents.ShardCoordinator(),
		ScProcessor:         scProcessor,
		TxFeeHandler:        txFeeHandler,
		TxTypeHandler:       txTypeHandler,
		EconomicsFee:        pcf.coreData.EconomicsData(),
		ReceiptForwarder:    receiptTxInterim,
		BadTxForwarder:      badTxInterim,
		ArgsParser:          argsParser,
		ScrForwarder:        scForwarder,
		EnableEpochsHandler: pcf.coreData.EnableEpochsHandler(),
		EnableRoundsHandler: pcf.coreData.EnableRoundsHandler(),
		TxVersionChecker:    pcf.coreData.TxVersionChecker(),
		GuardianChecker:     pcf.bootstrapComponents.GuardedAccountHandler(),
		TxLogsProcessor:     txLogsProcessor,
	}

	txProcessor, err := transaction.NewTxProcessor(argsTxProcessor)
	if err != nil {
		return args, nil, nil, err
	}

	args.TransactionProcessor = txProcessor
	args.IntermediateProcContainer = intermediateProcessorsContainer

	return args, vmContainerFactory, txTypeHandler, nil
}
