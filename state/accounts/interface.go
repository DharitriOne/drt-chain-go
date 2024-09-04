package accounts

import (
	"github.com/DharitriOne/drt-chain-go/state"
)

type dataTrieInteractor interface {
	state.DataTrieTracker
}
