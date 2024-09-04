package mock

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-crypto-go"
)

// PeerSignatureHandlerStub -
type PeerSignatureHandlerStub struct {
	VerifyPeerSignatureCalled func(pk []byte, pid core.PeerID, signature []byte) error
	GetPeerSignatureCalled    func(key crypto.PrivateKey, pid []byte) ([]byte, error)
}

// VerifyPeerSignature -
func (stub *PeerSignatureHandlerStub) VerifyPeerSignature(pk []byte, pid core.PeerID, signature []byte) error {
	if stub.VerifyPeerSignatureCalled != nil {
		return stub.VerifyPeerSignatureCalled(pk, pid, signature)
	}

	return nil
}

// GetPeerSignature -
func (stub *PeerSignatureHandlerStub) GetPeerSignature(key crypto.PrivateKey, pid []byte) ([]byte, error) {
	if stub.GetPeerSignatureCalled != nil {
		return stub.GetPeerSignatureCalled(key, pid)
	}

	return make([]byte, 0), nil
}

// IsInterfaceNil -
func (stub *PeerSignatureHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
