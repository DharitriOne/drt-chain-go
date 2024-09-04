package gin

import (
	"errors"
	"testing"

	apiErrors "github.com/DharitriOne/drt-chain-go/api/errors"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/facade/initial"
	"github.com/DharitriOne/drt-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestCommon_checkArgs(t *testing.T) {
	t.Parallel()

	args := ArgsNewWebServer{
		Facade:          nil,
		ApiConfig:       config.ApiRoutesConfig{},
		AntiFloodConfig: config.WebServerAntifloodConfig{},
	}
	err := checkArgs(args)
	require.True(t, errors.Is(err, apiErrors.ErrCannotCreateGinWebServer))

	args.Facade, err = initial.NewInitialNodeFacade(initial.ArgInitialNodeFacade{
		ApiInterface:                "api interface",
		PprofEnabled:                false,
		P2PPrometheusMetricsEnabled: false,
		StatusMetricsHandler:        &testscommon.StatusMetricsStub{},
	})
	require.NoError(t, err)
	err = checkArgs(args)
	require.NoError(t, err)
}

func TestCommon_isLogRouteEnabled(t *testing.T) {
	t.Parallel()

	routesConfigWithMissingLog := config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{},
	}
	require.False(t, isLogRouteEnabled(routesConfigWithMissingLog))

	routesConfig := config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"log": {
				Routes: []config.RouteConfig{
					{Name: "/log", Open: true},
				},
			},
		},
	}
	require.True(t, isLogRouteEnabled(routesConfig))
	require.False(t, isLogRouteEnabled(config.ApiRoutesConfig{}))
}
