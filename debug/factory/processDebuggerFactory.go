package factory

import (
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/debug/process"
)

// CreateProcessDebugger creates a new instance of type ProcessDebugger
func CreateProcessDebugger(configs config.ProcessDebugConfig) (ProcessDebugger, error) {
	if !configs.Enabled {
		return process.NewDisabledDebugger(), nil
	}

	return process.NewProcessDebugger(configs)
}
