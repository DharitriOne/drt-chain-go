package groups_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DharitriOne/drt-chain-core-go/data/api"
	"github.com/DharitriOne/drt-chain-core-go/data/dcdt"
	apiErrors "github.com/DharitriOne/drt-chain-go/api/errors"
	"github.com/DharitriOne/drt-chain-go/api/groups"
	"github.com/DharitriOne/drt-chain-go/api/mock"
	"github.com/DharitriOne/drt-chain-go/api/shared"
	"github.com/DharitriOne/drt-chain-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type accountResponse struct {
	Account struct {
		Address         string `json:"address"`
		Nonce           uint64 `json:"nonce"`
		Balance         string `json:"balance"`
		Code            string `json:"code"`
		CodeHash        []byte `json:"codeHash"`
		RootHash        []byte `json:"rootHash"`
		DeveloperReward string `json:"developerReward"`
	} `json:"account"`
}

type valueForKeyResponseData struct {
	Value string `json:"value"`
}

type valueForKeyResponse struct {
	Data  valueForKeyResponseData `json:"data"`
	Error string                  `json:"error"`
	Code  string                  `json:"code"`
}

type dcdtTokenData struct {
	TokenIdentifier string `json:"tokenIdentifier"`
	Balance         string `json:"balance"`
	Properties      string `json:"properties"`
}

type dcdtNFTTokenData struct {
	TokenIdentifier string   `json:"tokenIdentifier"`
	Balance         string   `json:"balance"`
	Properties      string   `json:"properties"`
	Name            string   `json:"name"`
	Nonce           uint64   `json:"nonce"`
	Creator         string   `json:"creator"`
	Royalties       string   `json:"royalties"`
	Hash            []byte   `json:"hash"`
	URIs            [][]byte `json:"uris"`
	Attributes      []byte   `json:"attributes"`
}

type dcdtNFTResponseData struct {
	dcdtNFTTokenData `json:"tokenData"`
}

type dcdtTokenResponseData struct {
	dcdtTokenData `json:"tokenData"`
}

type dcdtsWithRoleResponseData struct {
	Tokens []string `json:"tokens"`
}

type dcdtsWithRoleResponse struct {
	Data  dcdtsWithRoleResponseData `json:"data"`
	Error string                    `json:"error"`
	Code  string                    `json:"code"`
}

type dcdtTokenResponse struct {
	Data  dcdtTokenResponseData `json:"data"`
	Error string                `json:"error"`
	Code  string                `json:"code"`
}

type guardianDataResponseData struct {
	GuardianData api.GuardianData `json:"guardianData"`
}

type guardianDataResponse struct {
	Data  guardianDataResponseData `json:"data"`
	Error string                   `json:"error"`
	Code  string                   `json:"code"`
}

type dcdtNFTResponse struct {
	Data  dcdtNFTResponseData `json:"data"`
	Error string              `json:"error"`
	Code  string              `json:"code"`
}

type dcdtTokensCompleteResponseData struct {
	Tokens map[string]dcdtNFTTokenData `json:"dcdts"`
}

type dcdtTokensCompleteResponse struct {
	Data  dcdtTokensCompleteResponseData `json:"data"`
	Error string                         `json:"error"`
	Code  string
}

type keyValuePairsResponseData struct {
	Pairs map[string]string `json:"pairs"`
}

type keyValuePairsResponse struct {
	Data  keyValuePairsResponseData `json:"data"`
	Error string                    `json:"error"`
	Code  string
}

type dcdtRolesResponseData struct {
	Roles map[string][]string `json:"roles"`
}

type dcdtRolesResponse struct {
	Data  dcdtRolesResponseData `json:"data"`
	Error string                `json:"error"`
	Code  string
}

type usernameResponseData struct {
	Username string `json:"username"`
}

type usernameResponse struct {
	Data  usernameResponseData `json:"data"`
	Error string               `json:"error"`
	Code  string               `json:"code"`
}

type codeHashResponseData struct {
	CodeHash string `json:"codeHash"`
}

type codeHashResponse struct {
	Data  codeHashResponseData `json:"data"`
	Error string               `json:"error"`
	Code  string               `json:"code"`
}

func TestNewAddressGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewAddressGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewAddressGroup(&mock.FacadeStub{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

func TestAddressRoute_EmptyTrailReturns404(t *testing.T) {
	t.Parallel()
	facade := mock.FacadeStub{}

	addrGroup, err := groups.NewAddressGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest("GET", "/address", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestAddressGroup_getAccount(t *testing.T) {
	t.Parallel()

	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrCouldNotGetAccount, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAccountCalled: func(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				return api.AccountResponse{}, api.BlockInfo{}, expectedErr
			},
		}

		testAddressGroup(
			t,
			facade,
			"/address/addr",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrCouldNotGetAccount, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAccountCalled: func(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				return api.AccountResponse{
					Address:         "addr",
					Balance:         big.NewInt(100).String(),
					Nonce:           1,
					DeveloperReward: big.NewInt(120).String(),
				}, api.BlockInfo{}, nil
			},
		}

		response := &shared.GenericAPIResponse{}
		loadAddressGroupResponse(t, facade, "/address/addr", "GET", nil, response)

		mapResponse := response.Data.(map[string]interface{})
		accResp := accountResponse{}

		mapResponseBytes, _ := json.Marshal(&mapResponse)
		_ = json.Unmarshal(mapResponseBytes, &accResp)

		assert.Equal(t, "addr", accResp.Account.Address)
		assert.Equal(t, uint64(1), accResp.Account.Nonce)
		assert.Equal(t, "100", accResp.Account.Balance)
		assert.Equal(t, "120", accResp.Account.DeveloperReward)
		assert.Empty(t, response.Error)
	})
}

func TestAddressGroup_getBalance(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//balance", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetBalance, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice/balance?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetBalance, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetBalanceCalled: func(s string, _ api.AccountQueryOptions) (i *big.Int, info api.BlockInfo, e error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}

		testAddressGroup(
			t,
			facade,
			"/address/moa1alice/balance",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBalance, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		amount := big.NewInt(10)
		addr := "testAddress"
		facade := &mock.FacadeStub{
			GetBalanceCalled: func(s string, _ api.AccountQueryOptions) (i *big.Int, info api.BlockInfo, e error) {
				return amount, api.BlockInfo{}, nil
			},
		}

		response := &shared.GenericAPIResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			fmt.Sprintf("/address/%s/balance", addr),
			"GET",
			nil,
			response,
		)

		balanceStr := getValueForKey(response.Data, "balance")
		balanceResponse, ok := big.NewInt(0).SetString(balanceStr, 10)
		assert.True(t, ok)
		assert.Equal(t, amount, balanceResponse)
		assert.Equal(t, "", response.Error)
	})
}

func getValueForKey(dataFromResponse interface{}, key string) string {
	dataMap, ok := dataFromResponse.(map[string]interface{})
	if !ok {
		return ""
	}

	valueI, okCast := dataMap[key]
	if okCast {
		return fmt.Sprintf("%v", valueI)
	}
	return ""
}

func TestAddressGroup_getAccounts(t *testing.T) {
	t.Parallel()

	t.Run("wrong request, should err", func(t *testing.T) {
		t.Parallel()

		addrGroup, _ := groups.NewAddressGroup(&mock.FacadeStub{})

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		invalidRequest := []byte("{invalid json}")
		req, _ := http.NewRequest("POST", "/address/bulk", bytes.NewBuffer(invalidRequest))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		require.NotEmpty(t, response.Error)
		require.Equal(t, shared.ReturnCodeRequestError, response.Code)
	})
	t.Run("invalid query options should error",
		testErrorScenario("/address/bulk?blockNonce=not-uint64", "POST", bytes.NewBuffer([]byte(`["moa1", "moa1"]`)),
			formatExpectedErr(apiErrors.ErrCouldNotGetAccount, apiErrors.ErrBadUrlParams)))
	t.Run("facade error, should err", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetAccountsCalled: func(_ []string, _ api.AccountQueryOptions) (map[string]*api.AccountResponse, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		addrGroup, _ := groups.NewAddressGroup(&facade)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("POST", "/address/bulk", bytes.NewBuffer([]byte(`["moa1", "moa1"]`)))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		require.NotEmpty(t, response.Error)
		require.Equal(t, shared.ReturnCodeInternalError, response.Code)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedAccounts := map[string]*api.AccountResponse{
			"moa1alice": {
				Address: "moa1alice",
				Balance: "100000000000000",
				Nonce:   37,
			},
		}
		facade := &mock.FacadeStub{
			GetAccountsCalled: func(_ []string, _ api.AccountQueryOptions) (map[string]*api.AccountResponse, api.BlockInfo, error) {
				return expectedAccounts, api.BlockInfo{}, nil
			},
		}

		type responseType struct {
			Data struct {
				Accounts map[string]*api.AccountResponse `json:"accounts"`
			} `json:"data"`
			Error string            `json:"error"`
			Code  shared.ReturnCode `json:"code"`
		}
		response := &responseType{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/bulk",
			"POST",
			bytes.NewBuffer([]byte(`["moa1", "moa1"]`)),
			response,
		)

		require.Empty(t, response.Error)
		require.Equal(t, shared.ReturnCodeSuccess, response.Code)
		require.Equal(t, expectedAccounts, response.Data.Accounts)
	})
}

func TestAddressGroup_getUsername(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//username", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetUsername, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice/username?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetUsername, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetUsernameCalled: func(_ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
				return "", api.BlockInfo{}, expectedErr
			},
		}

		testAddressGroup(
			t,
			facade,
			"/address/moa1alice/username",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetUsername, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testUsername := "provided username"
		facade := &mock.FacadeStub{
			GetUsernameCalled: func(_ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
				return testUsername, api.BlockInfo{}, nil
			},
		}

		usernameResponseObj := &usernameResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/moa1alice/username",
			"GET",
			nil,
			usernameResponseObj,
		)
		assert.Equal(t, testUsername, usernameResponseObj.Data.Username)
	})
}

func TestAddressGroup_getCodeHash(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//code-hash", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetCodeHash, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice/code-hash?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetCodeHash, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetCodeHashCalled: func(_ string, _ api.AccountQueryOptions) ([]byte, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}

		testAddressGroup(
			t,
			facade,
			"/address/moa1alice/code-hash",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetCodeHash, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testCodeHash := []byte("value")
		expectedResponseCodeHash := base64.StdEncoding.EncodeToString(testCodeHash)
		facade := &mock.FacadeStub{
			GetCodeHashCalled: func(_ string, _ api.AccountQueryOptions) ([]byte, api.BlockInfo, error) {
				return testCodeHash, api.BlockInfo{}, nil
			},
		}

		codeHashResponseObj := &codeHashResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/moa1alice/code-hash",
			"GET",
			nil,
			codeHashResponseObj,
		)
		assert.Equal(t, expectedResponseCodeHash, codeHashResponseObj.Data.CodeHash)
	})
}

func TestAddressGroup_getValueForKey(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//key/test", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetValueForKey, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice/key/test?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetValueForKey, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetValueForKeyCalled: func(_ string, _ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
				return "", api.BlockInfo{}, expectedErr
			},
		}

		testAddressGroup(
			t,
			facade,
			"/address/moa1alice/key/test",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetValueForKey, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testValue := "value"
		facade := &mock.FacadeStub{
			GetValueForKeyCalled: func(_ string, _ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
				return testValue, api.BlockInfo{}, nil
			},
		}

		valueForKeyResponseObj := &valueForKeyResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/moa1alice/key/test",
			"GET",
			nil,
			valueForKeyResponseObj,
		)
		assert.Equal(t, testValue, valueForKeyResponseObj.Data.Value)
	})
}

func TestAddressGroup_getGuardianData(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//guardian-data", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetGuardianData, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice/guardian-data?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetGuardianData, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetGuardianDataCalled: func(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error) {
				return api.GuardianData{}, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/moa1alice/guardian-data",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetGuardianData, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedGuardianData := api.GuardianData{
			ActiveGuardian: &api.Guardian{
				Address:         "guardian1",
				ActivationEpoch: 0,
			},
			PendingGuardian: &api.Guardian{
				Address:         "guardian2",
				ActivationEpoch: 10,
			},
			Guarded: true,
		}
		facade := &mock.FacadeStub{
			GetGuardianDataCalled: func(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error) {
				return expectedGuardianData, api.BlockInfo{}, nil
			},
		}

		response := &guardianDataResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/moa1alice/guardian-data",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedGuardianData, response.Data.GuardianData)
	})
}

func TestAddressGroup_getKeyValuePairs(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//keys", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetKeyValuePairs, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice/keys?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetKeyValuePairs, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetKeyValuePairsCalled: func(_ string, _ api.AccountQueryOptions) (map[string]string, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/moa1alice/keys",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetKeyValuePairs, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pairs := map[string]string{
			"k1": "v1",
			"k2": "v2",
		}
		facade := &mock.FacadeStub{
			GetKeyValuePairsCalled: func(_ string, _ api.AccountQueryOptions) (map[string]string, api.BlockInfo, error) {
				return pairs, api.BlockInfo{}, nil
			},
		}

		response := &keyValuePairsResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/moa1alice/keys",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, pairs, response.Data.Pairs)
	})
}

func TestAddressGroup_getDCDTBalance(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//dcdt/newToken", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetDCDTBalance, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice/dcdt/newToken?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetDCDTBalance, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetDCDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*dcdt.DCDigitalToken, api.BlockInfo, error) {
				return &dcdt.DCDigitalToken{}, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/moa1alice/dcdt/newToken",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetDCDTBalance, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testValue := big.NewInt(100).String()
		testProperties := []byte{byte(0), byte(1), byte(0)}
		facade := &mock.FacadeStub{
			GetDCDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*dcdt.DCDigitalToken, api.BlockInfo, error) {
				return &dcdt.DCDigitalToken{Value: big.NewInt(100), Properties: testProperties}, api.BlockInfo{}, nil
			},
		}

		dcdtBalanceResponseObj := &dcdtTokenResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/moa1alice/dcdt/newToken",
			"GET",
			nil,
			dcdtBalanceResponseObj,
		)
		assert.Equal(t, testValue, dcdtBalanceResponseObj.Data.Balance)
		assert.Equal(t, "000100", dcdtBalanceResponseObj.Data.Properties)
	})
}

func TestAddressGroup_getDCDTsRoles(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//dcdts/roles", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetRolesForAccount, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice/dcdts/roles?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetRolesForAccount, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetDCDTsRolesCalled: func(_ string, _ api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/moa1alice/dcdts/roles",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetRolesForAccount, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		roles := map[string][]string{
			"token0": {"role0", "role1"},
			"token1": {"role3", "role1"},
		}
		facade := &mock.FacadeStub{
			GetDCDTsRolesCalled: func(_ string, _ api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error) {
				return roles, api.BlockInfo{}, nil
			},
		}

		response := &dcdtRolesResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/moa1alice/dcdts/roles",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, roles, response.Data.Roles)
	})
}

func TestAddressGroup_getDCDTTokensWithRole(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//dcdts-with-role/DCDTRoleNFTCreate", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetDCDTTokensWithRole, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice/dcdts-with-role/DCDTRoleNFTCreate?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetDCDTTokensWithRole, apiErrors.ErrBadUrlParams)))
	t.Run("invalid role should error",
		testErrorScenario("/address/moa1alice/dcdts-with-role/invalid", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetDCDTTokensWithRole, fmt.Errorf("invalid role: %s", "invalid"))))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetDCDTsWithRoleCalled: func(_ string, _ string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/moa1alice/dcdts-with-role/DCDTRoleNFTCreate",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetDCDTTokensWithRole, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedTokens := []string{"ABC-0o9i8u", "XYZ-r5y7i9"}
		facade := &mock.FacadeStub{
			GetDCDTsWithRoleCalled: func(address string, role string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
				return expectedTokens, api.BlockInfo{}, nil
			},
		}

		dcdtResponseObj := &dcdtsWithRoleResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/moa1alice/dcdts-with-role/DCDTRoleNFTCreate",
			"GET",
			nil,
			dcdtResponseObj,
		)
		assert.Equal(t, expectedTokens, dcdtResponseObj.Data.Tokens)
	})
}

func TestAddressGroup_getNFTTokenIDsRegisteredByAddress(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//registered-nfts", "GET", nil,
			formatExpectedErr(apiErrors.ErrRegisteredNFTTokenIDs, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice/registered-nfts?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrRegisteredNFTTokenIDs, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetNFTTokenIDsRegisteredByAddressCalled: func(_ string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/moa1alice/registered-nfts",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrRegisteredNFTTokenIDs, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedTokens := []string{"ABC-0o9i8u", "XYZ-r5y7i9"}
		facade := &mock.FacadeStub{
			GetNFTTokenIDsRegisteredByAddressCalled: func(address string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
				return expectedTokens, api.BlockInfo{}, nil
			},
		}

		dcdtResponseObj := &dcdtsWithRoleResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/moa1alice/registered-nfts",
			"GET",
			nil,
			dcdtResponseObj,
		)
		assert.Equal(t, expectedTokens, dcdtResponseObj.Data.Tokens)
	})
}

func TestAddressGroup_getDCDTNFTData(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//nft/newToken/nonce/10", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetDCDTNFTData, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice/nft/newToken/nonce/10?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetDCDTNFTData, apiErrors.ErrBadUrlParams)))
	t.Run("invalid nonce should error",
		testErrorScenario("/address/moa1alice/nft/newToken/nonce/not-int", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetDCDTNFTData, apiErrors.ErrNonceInvalid)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetDCDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*dcdt.DCDigitalToken, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/moa1alice/nft/newToken/nonce/10",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetDCDTNFTData, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testAddress := "address"
		testValue := big.NewInt(100).String()
		testNonce := uint64(37)
		testProperties := []byte{byte(1), byte(0), byte(0)}
		facade := &mock.FacadeStub{
			GetDCDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*dcdt.DCDigitalToken, api.BlockInfo, error) {
				return &dcdt.DCDigitalToken{
					Value:         big.NewInt(100),
					Properties:    testProperties,
					TokenMetaData: &dcdt.MetaData{Nonce: testNonce, Creator: []byte(testAddress)}}, api.BlockInfo{}, nil
			},
		}

		dcdtResponseObj := &dcdtNFTResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/moa1alice/nft/newToken/nonce/10",
			"GET",
			nil,
			dcdtResponseObj,
		)
		assert.Equal(t, testValue, dcdtResponseObj.Data.Balance)
		assert.Equal(t, "010000", dcdtResponseObj.Data.Properties)
		assert.Equal(t, testAddress, dcdtResponseObj.Data.Creator)
		assert.Equal(t, testNonce, dcdtResponseObj.Data.Nonce)
	})
}

func TestAddressGroup_getAllDCDTData(t *testing.T) {
	t.Parallel()

	t.Run("empty address should error",
		testErrorScenario("/address//dcdt", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetDCDTNFTData, apiErrors.ErrEmptyAddress)))
	t.Run("invalid query options should error",
		testErrorScenario("/address/moa1alice/dcdt?blockNonce=not-uint64", "GET", nil,
			formatExpectedErr(apiErrors.ErrGetDCDTNFTData, apiErrors.ErrBadUrlParams)))
	t.Run("with node fail should err", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAllDCDTTokensCalled: func(address string, options api.AccountQueryOptions) (map[string]*dcdt.DCDigitalToken, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, expectedErr
			},
		}
		testAddressGroup(
			t,
			facade,
			"/address/moa1alice/dcdt",
			"GET",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetDCDTNFTData, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testValue1 := "token1"
		testValue2 := "token2"
		facade := &mock.FacadeStub{
			GetAllDCDTTokensCalled: func(address string, _ api.AccountQueryOptions) (map[string]*dcdt.DCDigitalToken, api.BlockInfo, error) {
				tokens := make(map[string]*dcdt.DCDigitalToken)
				tokens[testValue1] = &dcdt.DCDigitalToken{Value: big.NewInt(10)}
				tokens[testValue2] = &dcdt.DCDigitalToken{Value: big.NewInt(100)}
				return tokens, api.BlockInfo{}, nil
			},
		}

		dcdtTokenResponseObj := &dcdtTokensCompleteResponse{}
		loadAddressGroupResponse(
			t,
			facade,
			"/address/moa1alice/dcdt",
			"GET",
			nil,
			dcdtTokenResponseObj,
		)
		assert.Equal(t, 2, len(dcdtTokenResponseObj.Data.Tokens))
	})
}

func TestAddressGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		addrGroup, err := groups.NewAddressGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = addrGroup.UpdateFacade(nil)
		require.Equal(t, apiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("cast failure should error", func(t *testing.T) {
		t.Parallel()

		addrGroup, err := groups.NewAddressGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = addrGroup.UpdateFacade("this is not a facade handler")
		require.True(t, errors.Is(err, apiErrors.ErrFacadeWrongTypeAssertion))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		roles := map[string][]string{
			"token0": {"role0", "role1"},
			"token1": {"role3", "role1"},
		}
		testAddress := "address"
		facade := mock.FacadeStub{
			GetDCDTsRolesCalled: func(_ string, _ api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error) {
				return roles, api.BlockInfo{}, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/dcdts/roles", testAddress), nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := dcdtRolesResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, roles, response.Data.Roles)

		newErr := errors.New("new error")
		newFacade := mock.FacadeStub{
			GetDCDTsRolesCalled: func(_ string, _ api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error) {
				return nil, api.BlockInfo{}, newErr
			},
		}
		err = addrGroup.UpdateFacade(&newFacade)
		require.NoError(t, err)

		req, _ = http.NewRequest("GET", fmt.Sprintf("/address/%s/dcdts/roles", testAddress), nil)
		resp = httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response = dcdtRolesResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, newErr.Error()))
	})
}

func TestAddressGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	addrGroup, _ := groups.NewAddressGroup(nil)
	require.True(t, addrGroup.IsInterfaceNil())

	addrGroup, _ = groups.NewAddressGroup(&mock.FacadeStub{})
	require.False(t, addrGroup.IsInterfaceNil())
}

func testErrorScenario(url string, method string, body io.Reader, expectedErr string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		testAddressGroup(
			t,
			&mock.FacadeStub{},
			url,
			method,
			body,
			http.StatusBadRequest,
			expectedErr,
		)
	}
}

func loadAddressGroupResponse(
	t *testing.T,
	facade shared.FacadeHandler,
	url string,
	method string,
	body io.Reader,
	destination interface{},
) {
	addrGroup, err := groups.NewAddressGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest(method, url, body)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	loadResponse(resp.Body, destination)
}

func testAddressGroup(
	t *testing.T,
	facade shared.FacadeHandler,
	url string,
	method string,
	body io.Reader,
	expectedRespCode int,
	expectedRespError string,
) {
	addrGroup, err := groups.NewAddressGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

	req, _ := http.NewRequest(method, url, body)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, expectedRespCode, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedRespError))
}

func formatExpectedErr(err, innerErr error) string {
	return fmt.Sprintf("%s: %s", err.Error(), innerErr.Error())
}

func getAddressRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"address": {
				Routes: []config.RouteConfig{
					{Name: "/:address", Open: true},
					{Name: "/bulk", Open: true},
					{Name: "/:address/guardian-data", Open: true},
					{Name: "/:address/balance", Open: true},
					{Name: "/:address/username", Open: true},
					{Name: "/:address/code-hash", Open: true},
					{Name: "/:address/keys", Open: true},
					{Name: "/:address/key/:key", Open: true},
					{Name: "/:address/dcdt", Open: true},
					{Name: "/:address/dcdts/roles", Open: true},
					{Name: "/:address/dcdt/:tokenIdentifier", Open: true},
					{Name: "/:address/nft/:tokenIdentifier/nonce/:nonce", Open: true},
					{Name: "/:address/dcdts-with-role/:role", Open: true},
					{Name: "/:address/registered-nfts", Open: true},
					{Name: "/:address/is-data-trie-migrated", Open: true},
				},
			},
		},
	}
}

func TestIsDataTrieMigrated(t *testing.T) {
	t.Parallel()

	testAddress := "address"
	expectedErr := errors.New("expected error")

	t.Run("should return error if IsDataTrieMigrated returns error", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			IsDataTrieMigratedCalled: func(address string, _ api.AccountQueryOptions) (bool, error) {
				return false, expectedErr
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)
		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/is-data-trie-migrated", testAddress), nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})

	t.Run("should return true if IsDataTrieMigrated returns true", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			IsDataTrieMigratedCalled: func(address string, _ api.AccountQueryOptions) (bool, error) {
				return true, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)
		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/is-data-trie-migrated", testAddress), nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.True(t, response.Error == "")

		respData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.True(t, respData["isMigrated"].(bool))
	})

	t.Run("should return false if IsDataTrieMigrated returns false", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			IsDataTrieMigratedCalled: func(address string, _ api.AccountQueryOptions) (bool, error) {
				return false, nil
			},
		}

		addrGroup, err := groups.NewAddressGroup(&facade)
		require.NoError(t, err)
		ws := startWebServer(addrGroup, "address", getAddressRoutesConfig())

		req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/is-data-trie-migrated", testAddress), nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.True(t, response.Error == "")

		respData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.False(t, respData["isMigrated"].(bool))
	})
}
