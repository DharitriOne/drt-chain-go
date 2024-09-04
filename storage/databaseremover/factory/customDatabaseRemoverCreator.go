package factory

import (
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/DharitriOne/drt-chain-go/storage"
	"github.com/DharitriOne/drt-chain-go/storage/databaseremover"
	"github.com/DharitriOne/drt-chain-go/storage/databaseremover/disabled"
)

// CreateCustomDatabaseRemover will handle the creation of a custom database remover based on the configuration
func CreateCustomDatabaseRemover(storagePruningConfig config.StoragePruningConfig) (storage.CustomDatabaseRemoverHandler, error) {
	if storagePruningConfig.AccountsTrieCleanOldEpochsData {
		return databaseremover.NewCustomDatabaseRemover(storagePruningConfig)
	}

	return disabled.NewDisabledCustomDatabaseRemover(), nil
}
