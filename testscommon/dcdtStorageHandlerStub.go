package testscommon

import (
	"math/big"

	"github.com/DharitriOne/drt-chain-core-go/data"
	"github.com/DharitriOne/drt-chain-core-go/data/dcdt"
	vmcommon "github.com/DharitriOne/drt-chain-vm-common-go"
)

// DcdtStorageHandlerStub -
type DcdtStorageHandlerStub struct {
	SaveDCDTNFTTokenCalled                                    func(senderAddress []byte, acnt vmcommon.UserAccountHandler, dcdtTokenKey []byte, nonce uint64, dcdtData *dcdt.DCDigitalToken, isCreation bool, isReturnWithError bool) ([]byte, error)
	GetDCDTNFTTokenOnSenderCalled                             func(acnt vmcommon.UserAccountHandler, dcdtTokenKey []byte, nonce uint64) (*dcdt.DCDigitalToken, error)
	GetDCDTNFTTokenOnDestinationCalled                        func(acnt vmcommon.UserAccountHandler, dcdtTokenKey []byte, nonce uint64) (*dcdt.DCDigitalToken, bool, error)
	GetDCDTNFTTokenOnDestinationWithCustomSystemAccountCalled func(accnt vmcommon.UserAccountHandler, dcdtTokenKey []byte, nonce uint64, systemAccount vmcommon.UserAccountHandler) (*dcdt.DCDigitalToken, bool, error)
	WasAlreadySentToDestinationShardAndUpdateStateCalled      func(tickerID []byte, nonce uint64, dstAddress []byte) (bool, error)
	SaveNFTMetaDataToSystemAccountCalled                      func(tx data.TransactionHandler) error
	AddToLiquiditySystemAccCalled                             func(dcdtTokenKey []byte, nonce uint64, transferValue *big.Int) error
}

// SaveDCDTNFTToken -
func (e *DcdtStorageHandlerStub) SaveDCDTNFTToken(senderAddress []byte, acnt vmcommon.UserAccountHandler, dcdtTokenKey []byte, nonce uint64, dcdtData *dcdt.DCDigitalToken, isCreation bool, isReturnWithError bool) ([]byte, error) {
	if e.SaveDCDTNFTTokenCalled != nil {
		return e.SaveDCDTNFTTokenCalled(senderAddress, acnt, dcdtTokenKey, nonce, dcdtData, isCreation, isReturnWithError)
	}

	return nil, nil
}

// GetDCDTNFTTokenOnSender -
func (e *DcdtStorageHandlerStub) GetDCDTNFTTokenOnSender(acnt vmcommon.UserAccountHandler, dcdtTokenKey []byte, nonce uint64) (*dcdt.DCDigitalToken, error) {
	if e.GetDCDTNFTTokenOnSenderCalled != nil {
		return e.GetDCDTNFTTokenOnSenderCalled(acnt, dcdtTokenKey, nonce)
	}

	return nil, nil
}

// GetDCDTNFTTokenOnDestination -
func (e *DcdtStorageHandlerStub) GetDCDTNFTTokenOnDestination(acnt vmcommon.UserAccountHandler, dcdtTokenKey []byte, nonce uint64) (*dcdt.DCDigitalToken, bool, error) {
	if e.GetDCDTNFTTokenOnDestinationCalled != nil {
		return e.GetDCDTNFTTokenOnDestinationCalled(acnt, dcdtTokenKey, nonce)
	}

	return nil, false, nil
}

// GetDCDTNFTTokenOnDestinationWithCustomSystemAccount -
func (e *DcdtStorageHandlerStub) GetDCDTNFTTokenOnDestinationWithCustomSystemAccount(accnt vmcommon.UserAccountHandler, dcdtTokenKey []byte, nonce uint64, systemAccount vmcommon.UserAccountHandler) (*dcdt.DCDigitalToken, bool, error) {
	if e.GetDCDTNFTTokenOnDestinationWithCustomSystemAccountCalled != nil {
		return e.GetDCDTNFTTokenOnDestinationWithCustomSystemAccountCalled(accnt, dcdtTokenKey, nonce, systemAccount)
	}

	return nil, false, nil
}

// WasAlreadySentToDestinationShardAndUpdateState -
func (e *DcdtStorageHandlerStub) WasAlreadySentToDestinationShardAndUpdateState(tickerID []byte, nonce uint64, dstAddress []byte) (bool, error) {
	if e.WasAlreadySentToDestinationShardAndUpdateStateCalled != nil {
		return e.WasAlreadySentToDestinationShardAndUpdateStateCalled(tickerID, nonce, dstAddress)
	}

	return false, nil
}

// SaveNFTMetaDataToSystemAccount -
func (e *DcdtStorageHandlerStub) SaveNFTMetaDataToSystemAccount(tx data.TransactionHandler) error {
	if e.SaveNFTMetaDataToSystemAccountCalled != nil {
		return e.SaveNFTMetaDataToSystemAccountCalled(tx)
	}

	return nil
}

// AddToLiquiditySystemAcc -
func (e *DcdtStorageHandlerStub) AddToLiquiditySystemAcc(dcdtTokenKey []byte, nonce uint64, transferValue *big.Int) error {
	if e.AddToLiquiditySystemAccCalled != nil {
		return e.AddToLiquiditySystemAccCalled(dcdtTokenKey, nonce, transferValue)
	}

	return nil
}

// IsInterfaceNil -
func (e *DcdtStorageHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
