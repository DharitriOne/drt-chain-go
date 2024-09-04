package antiflood

import "github.com/DharitriOne/drt-chain-go/process"

func (af *p2pAntiflood) Debugger() process.AntifloodDebugger {
	return af.debugger
}
