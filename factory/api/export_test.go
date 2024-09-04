package api

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/factory"
	"github.com/DharitriOne/drt-chain-go/process"
	"github.com/DharitriOne/drt-chain-go/vm"
)

// SCQueryElementArgs -
type SCQueryElementArgs struct {
	GeneralConfig         *config.Config
	EpochConfig           *config.EpochConfig
	CoreComponents        factory.CoreComponentsHolder
	StateComponents       factory.StateComponentsHolder
	StatusCoreComponents  factory.StatusCoreComponentsHolder
	DataComponents        factory.DataComponentsHolder
	ProcessComponents     factory.ProcessComponentsHolder
	GasScheduleNotifier   core.GasScheduleNotifier
	MessageSigVerifier    vm.MessageSignVerifier
	SystemSCConfig        *config.SystemSmartContractsConfig
	Bootstrapper          process.Bootstrapper
	AllowVMQueriesChan    chan struct{}
	WorkingDir            string
	Index                 int
	GuardedAccountHandler process.GuardedAccountHandler
}

// CreateScQueryElement -
func CreateScQueryElement(args SCQueryElementArgs) (process.SCQueryService, common.StorageManager, error) {
	return createScQueryElement(scQueryElementArgs{
		generalConfig:         args.GeneralConfig,
		epochConfig:           args.EpochConfig,
		coreComponents:        args.CoreComponents,
		stateComponents:       args.StateComponents,
		dataComponents:        args.DataComponents,
		processComponents:     args.ProcessComponents,
		statusCoreComponents:  args.StatusCoreComponents,
		gasScheduleNotifier:   args.GasScheduleNotifier,
		messageSigVerifier:    args.MessageSigVerifier,
		systemSCConfig:        args.SystemSCConfig,
		bootstrapper:          args.Bootstrapper,
		allowVMQueriesChan:    args.AllowVMQueriesChan,
		workingDir:            args.WorkingDir,
		index:                 args.Index,
		guardedAccountHandler: args.GuardedAccountHandler,
	})
}

// CreateBlockchainForScQuery -
func CreateBlockchainForScQuery(selfShardID uint32) (data.ChainHandler, error) {
	return createBlockchainForScQuery(selfShardID)
}
