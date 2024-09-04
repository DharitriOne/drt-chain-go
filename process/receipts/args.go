package receipts

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/core/check"
	"github.com/DharitriOne/drt-chain-core-go/hashing"
	"github.com/DharitriOne/drt-chain-core-go/marshal"
	"github.com/DharitriOne/drt-chain-go/dataRetriever"
)

// ArgsNewReceiptsRepository holds arguments for creating a receiptsRepository
type ArgsNewReceiptsRepository struct {
	Marshaller marshal.Marshalizer
	Hasher     hashing.Hasher
	Store      dataRetriever.StorageService
}

func (args *ArgsNewReceiptsRepository) check() error {
	if check.IfNil(args.Marshaller) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return core.ErrNilHasher
	}
	if check.IfNil(args.Store) {
		return core.ErrNilStore
	}

	return nil
}
