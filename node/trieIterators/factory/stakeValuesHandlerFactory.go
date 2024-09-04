package factory

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-go/node/external"
	"github.com/DharitriOne/drt-chain-go/node/trieIterators"
	"github.com/DharitriOne/drt-chain-go/node/trieIterators/disabled"
)

// CreateTotalStakedValueHandler will create a new instance of TotalStakedValueHandler
func CreateTotalStakedValueHandler(args trieIterators.ArgTrieIteratorProcessor) (external.TotalStakedValueHandler, error) {
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledStakeValuesProcessor(), nil
	}

	return trieIterators.NewTotalStakedValueProcessor(args)
}
