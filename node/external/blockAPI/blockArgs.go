package blockAPI

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/data/transaction"
	"github.com/DharitriOne/drt-chain-core-go/data/typeConverters"
	"github.com/DharitriOne/drt-chain-core-go/hashing"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/DharitriOne/drt-chain-go/common"
	"github.com/DharitriOne/drt-chain-go/dataRetriever"
	"github.com/DharitriOne/drt-chain-go/dblookupext"
	outportProcess "github.com/DharitriOne/drt-chain-go/outport/process"
	"github.com/DharitriOne/drt-chain-go/process"
	"github.com/DharitriOne/drt-chain-go/state"
)

// ArgAPIBlockProcessor is structure that store components that are needed to create an api block processor
type ArgAPIBlockProcessor struct {
	SelfShardID                  uint32
	Store                        dataRetriever.StorageService
	Marshalizer                  marshal.Marshalizer
	Uint64ByteSliceConverter     typeConverters.Uint64ByteSliceConverter
	HistoryRepo                  dblookupext.HistoryRepository
	APITransactionHandler        APITransactionHandler
	StatusComputer               transaction.StatusComputerHandler
	Hasher                       hashing.Hasher
	AddressPubkeyConverter       core.PubkeyConverter
	LogsFacade                   logsFacade
	ReceiptsRepository           receiptsRepository
	AlteredAccountsProvider      outportProcess.AlteredAccountsProviderHandler
	AccountsRepository           state.AccountsRepository
	ScheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	EnableEpochsHandler          common.EnableEpochsHandler
}
