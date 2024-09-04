package chronology

import (
	"time"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-go/consensus"
	"github.com/DharitriOne/drt-chain-go/ntp"
)

// ArgChronology holds all dependencies required by the chronology component
type ArgChronology struct {
	GenesisTime      time.Time
	RoundHandler     consensus.RoundHandler
	SyncTimer        ntp.SyncTimer
	Watchdog         core.WatchdogTimer
	AppStatusHandler core.AppStatusHandler
}
