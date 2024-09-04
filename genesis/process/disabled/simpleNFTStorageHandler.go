package disabled

import (
	"math/big"

	"github.com/DharitriOne/drt-chain-core-go/data"
	"github.com/DharitriOne/drt-chain-core-go/data/dcdt"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
)

// SimpleNFTStorage implements the SimpleNFTStorage interface but does nothing as it is disabled
type SimpleNFTStorage struct {
}

// GetDCDTNFTTokenOnDestination is disabled
func (s *SimpleNFTStorage) GetDCDTNFTTokenOnDestination(_ vmcommon.UserAccountHandler, _ []byte, _ uint64) (*dcdt.DCDigitalToken, bool, error) {
	return &dcdt.DCDigitalToken{Value: big.NewInt(0)}, true, nil
}

// SaveNFTMetaDataToSystemAccount is disabled
func (s *SimpleNFTStorage) SaveNFTMetaDataToSystemAccount(_ data.TransactionHandler) error {
	return nil
}

// IsInterfaceNil return true if underlying object is nil
func (s *SimpleNFTStorage) IsInterfaceNil() bool {
	return s == nil
}
