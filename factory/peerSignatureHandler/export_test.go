package peerSignatureHandler

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-go/errors"
)

func (psh *peerSignatureHandler) GetPIDAndSig(entry interface{}) (core.PeerID, []byte, error) {
	pidSig, ok := entry.(*pidSignature)
	if !ok {
		return "", nil, errors.ErrWrongTypeAssertion
	}

	return pidSig.pid, pidSig.signature, nil
}

func (psh *peerSignatureHandler) GetCacheEntry(pid core.PeerID, sig []byte) *pidSignature {
	return &pidSignature{
		pid:       pid,
		signature: sig,
	}
}
