package filters

import (
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/data/api"
	"github.com/DharitriOne/drt-chain-core-go/data/block"
	"github.com/DharitriOne/drt-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestStatusFilters_ApplyStatusFilters(t *testing.T) {
	t.Parallel()

	sf := NewStatusFilters(0)

	dcdtTransferTx := &transaction.ApiTransactionResult{
		Hash:             "myHash",
		Nonce:            1,
		SourceShard:      1,
		DestinationShard: 0,
		Data:             []byte("DCDTTransfer@42524f2d343663663439@a688906bd8b00000"),
	}
	mbs := []*api.MiniBlock{
		{
			SourceShard:      1,
			DestinationShard: 0,
			Transactions: []*transaction.ApiTransactionResult{
				dcdtTransferTx,
				{},
			},
			Type: block.TxBlock.String(),
		},
		{
			Type: block.TxBlock.String(),
		},
		{
			DestinationShard: 1,
			SourceShard:      0,
			Type:             block.SmartContractResultBlock.String(),
			Transactions: []*transaction.ApiTransactionResult{
				{},
				{
					OriginalTransactionHash: "myHash",
					Nonce:                   1,
					SourceShard:             1,
					DestinationShard:        0,
					Data:                    []byte("DCDTTransfer@42524f2d343663663439@a688906bd8b00000@75736572206572726f72"),
				},
			},
		},
		{
			Type: block.RewardsBlock.String(),
		},
	}
	sf.ApplyStatusFilters(mbs)
	require.Equal(t, transaction.TxStatusFail, dcdtTransferTx.Status)
}

func TestStatusFilters_SetStatusIfIsFailedDCDTTransfer(t *testing.T) {
	t.Parallel()

	sf := NewStatusFilters(0)
	// DCDT transfer fail
	tx1 := &transaction.ApiTransactionResult{
		Nonce:            1,
		Hash:             "myHash",
		SourceShard:      1,
		DestinationShard: 0,
		Data:             []byte("DCDTTransfer@42524f2d343663663439@a688906bd8b00000"),
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{
				OriginalTxHash: "myHash",
				Nonce:          1,
				Data:           "DCDTTransfer@42524f2d343663663439@a688906bd8b00000@75736572206572726f72",
			},
		},
	}

	sf.SetStatusIfIsFailedDCDTTransfer(tx1)
	require.Equal(t, transaction.TxStatusFail, tx1.Status)

	// transaction with no SCR should be ignored
	tx2 := &transaction.ApiTransactionResult{
		Status: transaction.TxStatusSuccess,
	}
	sf.SetStatusIfIsFailedDCDTTransfer(tx2)
	require.Equal(t, transaction.TxStatusSuccess, tx2.Status)

	// intra shard transaction should be ignored
	tx3 := &transaction.ApiTransactionResult{
		Status: transaction.TxStatusSuccess,
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{},
			{},
		},
	}
	sf.SetStatusIfIsFailedDCDTTransfer(tx3)
	require.Equal(t, transaction.TxStatusSuccess, tx3.Status)

	// no DCDT transfer should be ignored
	tx4 := &transaction.ApiTransactionResult{
		Status:           transaction.TxStatusSuccess,
		SourceShard:      1,
		DestinationShard: 0,
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{},
			{},
		},
	}
	sf.SetStatusIfIsFailedDCDTTransfer(tx4)
	require.Equal(t, transaction.TxStatusSuccess, tx4.Status)
}
