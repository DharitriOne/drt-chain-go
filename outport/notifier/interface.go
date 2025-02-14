package notifier

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data/block"
)

type httpClientHandler interface {
	Post(route string, payload interface{}) error
	IsInterfaceNil() bool
}

// BlockContainerHandler defines what a block container should be able to do
type BlockContainerHandler interface {
	Get(headerType core.HeaderType) (block.EmptyBlockCreator, error)
}
