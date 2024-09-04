package localFuncs

import (
	"math/big"

	"github.com/DharitriOne/drt-chain-go/testscommon/txDataBuilder"
	mock "github.com/DharitriOne/drt-chain-vm-go/mock/context"
	vmhooks "github.com/DharitriOne/drt-chain-vm-go/vmhost/vmhooks"
)

// LocalMintMock is an exposed mock contract method
func LocalMintMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("localMint", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		scAddress := host.Runtime().GetContextAddress()
		args := host.Runtime().Arguments()

		callData := txDataBuilder.NewBuilder()
		callData.LocalMintDCDT(
			string(args[0]),
			big.NewInt(0).SetBytes(args[1]).Int64())

		vmhooks.ExecuteOnDestContextWithTypedArgs(
			host,
			1_000_000,
			big.NewInt(0),
			[]byte(callData.Function()),
			scAddress,
			callData.ElementsAsBytes())

		return instance
	})
}

// LocalBurnMock is an exposed mock contract method
func LocalBurnMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("localBurn", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		scAddress := host.Runtime().GetContextAddress()
		args := host.Runtime().Arguments()

		callData := txDataBuilder.NewBuilder()
		callData.LocalBurnDCDT(
			string(args[0]),
			big.NewInt(0).SetBytes(args[1]).Int64())

		vmhooks.ExecuteOnDestContextWithTypedArgs(
			host,
			1_000_000,
			big.NewInt(0),
			[]byte(callData.Function()),
			scAddress,
			callData.ElementsAsBytes())

		return instance
	})
}
