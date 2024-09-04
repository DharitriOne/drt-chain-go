package transactionAPI

import (
	"time"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data/typeConverters"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/DharitriOne/drt-chain-go/dataRetriever"
	"github.com/DharitriOne/drt-chain-go/dblookupext"
	"github.com/DharitriOne/drt-chain-go/process"
	"github.com/DharitriOne/drt-chain-go/sharding"
)

// ArgAPITransactionProcessor is structure that store components that are needed to create an api transaction processor
type ArgAPITransactionProcessor struct {
	RoundDuration            uint64
	GenesisTime              time.Time
	Marshalizer              marshal.Marshalizer
	AddressPubKeyConverter   core.PubkeyConverter
	ShardCoordinator         sharding.Coordinator
	HistoryRepository        dblookupext.HistoryRepository
	StorageService           dataRetriever.StorageService
	DataPool                 dataRetriever.PoolsHolder
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	FeeComputer              feeComputer
	TxTypeHandler            process.TxTypeHandler
	LogsFacade               LogsFacade
	DataFieldParser          DataFieldParser
}
