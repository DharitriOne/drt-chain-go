package factory

import "github.com/DharitriOne/drt-chain-go/cmd/assessment/benchmarks"

type benchmarkCoordinator interface {
	RunAllTests() *benchmarks.TestResults
	IsInterfaceNil() bool
}
