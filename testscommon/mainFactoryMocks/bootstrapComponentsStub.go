package mainFactoryMocks

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	nodeFactory "github.com/DharitriOne/drt-chain-go/cmd/node/factory"
	"github.com/DharitriOne/drt-chain-go/factory"
	"github.com/DharitriOne/drt-chain-go/process"
	"github.com/DharitriOne/drt-chain-go/sharding"
	"github.com/DharitriOne/drt-chain-go/sharding/nodesCoordinator"
)

// BootstrapComponentsStub -
type BootstrapComponentsStub struct {
	Bootstrapper                         factory.EpochStartBootstrapper
	BootstrapParams                      factory.BootstrapParamsHolder
	NodeRole                             core.NodeType
	ShCoordinator                        sharding.Coordinator
	ShardCoordinatorCalled               func() sharding.Coordinator
	HdrVersionHandler                    nodeFactory.HeaderVersionHandler
	VersionedHdrFactory                  nodeFactory.VersionedHeaderFactory
	HdrIntegrityVerifier                 nodeFactory.HeaderIntegrityVerifierHandler
	GuardedAccountHandlerField           process.GuardedAccountHandler
	NodesCoordinatorRegistryFactoryField nodesCoordinator.NodesCoordinatorRegistryFactory
}

// Create -
func (bcs *BootstrapComponentsStub) Create() error {
	return nil
}

// Close -
func (bcs *BootstrapComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (bcs *BootstrapComponentsStub) CheckSubcomponents() error {
	return nil
}

// EpochStartBootstrapper -
func (bcs *BootstrapComponentsStub) EpochStartBootstrapper() factory.EpochStartBootstrapper {
	return bcs.Bootstrapper
}

// EpochBootstrapParams -
func (bcs *BootstrapComponentsStub) EpochBootstrapParams() factory.BootstrapParamsHolder {
	return bcs.BootstrapParams
}

// NodeType -
func (bcs *BootstrapComponentsStub) NodeType() core.NodeType {
	return bcs.NodeRole
}

// ShardCoordinator -
func (bcs *BootstrapComponentsStub) ShardCoordinator() sharding.Coordinator {
	if bcs.ShardCoordinatorCalled != nil {
		return bcs.ShardCoordinatorCalled()
	}
	return bcs.ShCoordinator
}

// HeaderVersionHandler -
func (bcs *BootstrapComponentsStub) HeaderVersionHandler() nodeFactory.HeaderVersionHandler {
	return bcs.HdrVersionHandler
}

// VersionedHeaderFactory -
func (bcs *BootstrapComponentsStub) VersionedHeaderFactory() nodeFactory.VersionedHeaderFactory {
	return bcs.VersionedHdrFactory
}

// HeaderIntegrityVerifier -
func (bcs *BootstrapComponentsStub) HeaderIntegrityVerifier() nodeFactory.HeaderIntegrityVerifierHandler {
	return bcs.HdrIntegrityVerifier
}

// SetShardCoordinator -
func (bcs *BootstrapComponentsStub) SetShardCoordinator(shardCoordinator sharding.Coordinator) error {
	bcs.ShCoordinator = shardCoordinator
	return nil
}

// GuardedAccountHandler -
func (bcs *BootstrapComponentsStub) GuardedAccountHandler() process.GuardedAccountHandler {
	return bcs.GuardedAccountHandlerField
}

// NodesCoordinatorRegistryFactory -
func (bcs *BootstrapComponentsStub) NodesCoordinatorRegistryFactory() nodesCoordinator.NodesCoordinatorRegistryFactory {
	return bcs.NodesCoordinatorRegistryFactoryField
}

// String -
func (bcs *BootstrapComponentsStub) String() string {
	return "BootstrapComponentsStub"
}

// IsInterfaceNil -
func (bcs *BootstrapComponentsStub) IsInterfaceNil() bool {
	return bcs == nil
}
