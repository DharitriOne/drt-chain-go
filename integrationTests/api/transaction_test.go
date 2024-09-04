package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/data/transaction"
	"github.com/DharitriOne/drt-chain-go/integrationTests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	node := integrationTests.NewTestProcessorNodeWithTestWebServer(3, 0, 0)

	testTransactionGasCostWithMissingFields(t, node)
}

func testTransactionGasCostWithMissingFields(tb testing.TB, node *integrationTests.TestProcessorNodeWithTestWebServer) {
	// this is an example found in the wild, should not add more fields in order to pass the tests
	tx := transaction.FrontendTransaction{
		Sender:   "moa1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awq9cstml",
		Receiver: "moa188anxz35atlef7cucszypmvx88lhz4m7a7t7lhcwt6sfphpsqlksr33h66",
		Value:    "100",
		Nonce:    0,
		GasPrice: 100,
		Version:  1,
		ChainID:  "T",
	}

	jsonBytes, _ := json.Marshal(tx)
	req, _ := http.NewRequest("POST", "/transaction/cost", bytes.NewBuffer(jsonBytes))

	resp := node.DoRequest(req)
	require.NotNil(tb, resp)

	type transactionCostResponseData struct {
		Cost uint64 `json:"txGasUnits"`
	}
	type transactionCostResponse struct {
		Data  transactionCostResponseData `json:"data"`
		Error string                      `json:"error"`
		Code  string                      `json:"code"`
	}

	txCost := &transactionCostResponse{}
	loadResponse(tb, resp.Body, txCost)
	assert.Empty(tb, txCost.Error)

	assert.Equal(tb, integrationTests.MinTxGasLimit, txCost.Data.Cost)
}

func loadResponse(tb testing.TB, rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)

	assert.Nil(tb, err)
}
