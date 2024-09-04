package factory

import (
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/debug/handler"
)

// NewInterceptorDebuggerFactory will instantiate an InterceptorDebugHandler based on the provided config
func NewInterceptorDebuggerFactory(config config.InterceptorResolverDebugConfig) (InterceptorDebugHandler, error) {
	if !config.Enabled {
		return handler.NewDisabledInterceptorDebugHandler(), nil
	}

	return handler.NewInterceptorDebugHandler(config)
}
