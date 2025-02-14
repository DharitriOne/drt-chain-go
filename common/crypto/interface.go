package crypto

import crypto "github.com/DharitriOne/drt-chain-crypto-go"

// MultiSignerContainer defines the container for different versioned multiSigner instances
type MultiSignerContainer interface {
	GetMultiSigner(epoch uint32) (crypto.MultiSigner, error)
	IsInterfaceNil() bool
}
