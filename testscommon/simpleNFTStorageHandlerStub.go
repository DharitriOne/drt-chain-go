package testscommon

import (
	"math/big"

	"github.com/DharitriOne/drt-chain-core-go/data"
	"github.com/DharitriOne/drt-chain-core-go/data/dcdt"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
)

// SimpleNFTStorageHandlerStub -
type SimpleNFTStorageHandlerStub struct {
	GetDCDTNFTTokenOnDestinationCalled   func(accnt vmcommon.UserAccountHandler, dcdtTokenKey []byte, nonce uint64) (*dcdt.DCDigitalToken, bool, error)
	SaveNFTMetaDataToSystemAccountCalled func(tx data.TransactionHandler) error
}

// GetDCDTNFTTokenOnDestination -
func (s *SimpleNFTStorageHandlerStub) GetDCDTNFTTokenOnDestination(accnt vmcommon.UserAccountHandler, dcdtTokenKey []byte, nonce uint64) (*dcdt.DCDigitalToken, bool, error) {
	if s.GetDCDTNFTTokenOnDestinationCalled != nil {
		return s.GetDCDTNFTTokenOnDestinationCalled(accnt, dcdtTokenKey, nonce)
	}
	return &dcdt.DCDigitalToken{Value: big.NewInt(0)}, true, nil
}

// SaveNFTMetaDataToSystemAccount -
func (s *SimpleNFTStorageHandlerStub) SaveNFTMetaDataToSystemAccount(tx data.TransactionHandler) error {
	if s.SaveNFTMetaDataToSystemAccountCalled != nil {
		return s.SaveNFTMetaDataToSystemAccountCalled(tx)
	}
	return nil
}

// IsInterfaceNil -
func (s *SimpleNFTStorageHandlerStub) IsInterfaceNil() bool {
	return s == nil
}
