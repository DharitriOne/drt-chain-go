package dblookupext

import (
	"github.com/DharitriOne/drt-chain-core-go/data"
	"github.com/DharitriOne/drt-chain-core-go/data/block"
	"github.com/DharitriOne/drt-chain-go/dblookupext/dcdtSupply"
)

// HistoryRepositoryFactory can create new instances of HistoryRepository
type HistoryRepositoryFactory interface {
	Create() (HistoryRepository, error)
	IsInterfaceNil() bool
}

// HistoryRepository provides methods needed for the history data processing
type HistoryRepository interface {
	RecordBlock(blockHeaderHash []byte,
		blockHeader data.HeaderHandler,
		blockBody data.BodyHandler,
		scrResultsFromPool map[string]data.TransactionHandler,
		receiptsFromPool map[string]data.TransactionHandler,
		createdIntraShardMiniBlocks []*block.MiniBlock,
		logs []*data.LogData) error
	OnNotarizedBlocks(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
	GetMiniblockMetadataByTxHash(hash []byte) (*MiniblockMetadata, error)
	GetEpochByHash(hash []byte) (uint32, error)
	GetResultsHashesByTxHash(txHash []byte, epoch uint32) (*ResultsHashesByTxHash, error)
	RevertBlock(blockHeader data.HeaderHandler, blockBody data.BodyHandler) error
	GetDCDTSupply(token string) (*dcdtSupply.SupplyDCDT, error)
	IsEnabled() bool
	IsInterfaceNil() bool
}

// BlockTracker defines the interface of the block tracker
type BlockTracker interface {
	RegisterCrossNotarizedHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterSelfNotarizedFromCrossHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterSelfNotarizedHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterFinalMetachainHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	IsInterfaceNil() bool
}

// SuppliesHandler defines the interface of a supplies processor
type SuppliesHandler interface {
	ProcessLogs(blockNonce uint64, logs []*data.LogData) error
	RevertChanges(header data.HeaderHandler, body data.BodyHandler) error
	GetDCDTSupply(token string) (*dcdtSupply.SupplyDCDT, error)
	IsInterfaceNil() bool
}
