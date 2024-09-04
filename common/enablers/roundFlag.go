package enablers

import (
	"github.com/DharitriOne/drt-chain-core-go/core/atomic"
)

type roundFlag struct {
	*atomic.Flag
	round   uint64
	options []string
}
