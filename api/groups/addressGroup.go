package groups

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-core-go/core/check"
	"github.com/DharitriOne/drt-chain-core-go/data/api"
	"github.com/DharitriOne/drt-chain-core-go/data/dcdt"
	"github.com/DharitriOne/drt-chain-go/api/errors"
	"github.com/DharitriOne/drt-chain-go/api/shared"
)

const (
	getAccountPath                 = "/:address"
	getAccountsPath                = "/bulk"
	getBalancePath                 = "/:address/balance"
	getUsernamePath                = "/:address/username"
	getCodeHashPath                = "/:address/code-hash"
	getKeysPath                    = "/:address/keys"
	getKeyPath                     = "/:address/key/:key"
	getDataTrieMigrationStatusPath = "/:address/is-data-trie-migrated"
	getDCDTTokensPath              = "/:address/dcdt"
	getDCDTBalancePath             = "/:address/dcdt/:tokenIdentifier"
	getDCDTTokensWithRolePath      = "/:address/dcdts-with-role/:role"
	getDCDTsRolesPath              = "/:address/dcdts/roles"
	getRegisteredNFTsPath          = "/:address/registered-nfts"
	getDCDTNFTDataPath             = "/:address/nft/:tokenIdentifier/nonce/:nonce"
	getGuardianData                = "/:address/guardian-data"
	urlParamOnFinalBlock           = "onFinalBlock"
	urlParamOnStartOfEpoch         = "onStartOfEpoch"
	urlParamBlockNonce             = "blockNonce"
	urlParamBlockHash              = "blockHash"
	urlParamBlockRootHash          = "blockRootHash"
	urlParamHintEpoch              = "hintEpoch"
	urlParamWithKeys               = "withKeys"
)

// addressFacadeHandler defines the methods to be implemented by a facade for handling address requests
type addressFacadeHandler interface {
	GetBalance(address string, options api.AccountQueryOptions) (*big.Int, api.BlockInfo, error)
	GetUsername(address string, options api.AccountQueryOptions) (string, api.BlockInfo, error)
	GetCodeHash(address string, options api.AccountQueryOptions) ([]byte, api.BlockInfo, error)
	GetValueForKey(address string, key string, options api.AccountQueryOptions) (string, api.BlockInfo, error)
	GetAccount(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error)
	GetAccounts(addresses []string, options api.AccountQueryOptions) (map[string]*api.AccountResponse, api.BlockInfo, error)
	GetDCDTData(address string, key string, nonce uint64, options api.AccountQueryOptions) (*dcdt.DCDigitalToken, api.BlockInfo, error)
	GetDCDTsRoles(address string, options api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error)
	GetNFTTokenIDsRegisteredByAddress(address string, options api.AccountQueryOptions) ([]string, api.BlockInfo, error)
	GetDCDTsWithRole(address string, role string, options api.AccountQueryOptions) ([]string, api.BlockInfo, error)
	GetAllDCDTTokens(address string, options api.AccountQueryOptions) (map[string]*dcdt.DCDigitalToken, api.BlockInfo, error)
	GetKeyValuePairs(address string, options api.AccountQueryOptions) (map[string]string, api.BlockInfo, error)
	GetGuardianData(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error)
	IsDataTrieMigrated(address string, options api.AccountQueryOptions) (bool, error)
	IsInterfaceNil() bool
}

type addressGroup struct {
	*baseGroup
	facade    addressFacadeHandler
	mutFacade sync.RWMutex
}

type dcdtTokenData struct {
	TokenIdentifier string `json:"tokenIdentifier"`
	Balance         string `json:"balance"`
	Properties      string `json:"properties"`
}

type dcdtNFTTokenData struct {
	TokenIdentifier string   `json:"tokenIdentifier"`
	Balance         string   `json:"balance"`
	Properties      string   `json:"properties,omitempty"`
	Name            string   `json:"name,omitempty"`
	Nonce           uint64   `json:"nonce,omitempty"`
	Creator         string   `json:"creator,omitempty"`
	Royalties       string   `json:"royalties,omitempty"`
	Hash            []byte   `json:"hash,omitempty"`
	URIs            [][]byte `json:"uris,omitempty"`
	Attributes      []byte   `json:"attributes,omitempty"`
}

// NewAddressGroup returns a new instance of addressGroup
func NewAddressGroup(facade addressFacadeHandler) (*addressGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for address group", errors.ErrNilFacadeHandler)
	}

	ag := &addressGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    getAccountPath,
			Method:  http.MethodGet,
			Handler: ag.getAccount,
		},
		{
			Path:    getAccountsPath,
			Method:  http.MethodPost,
			Handler: ag.getAccounts,
		},
		{
			Path:    getBalancePath,
			Method:  http.MethodGet,
			Handler: ag.getBalance,
		},
		{
			Path:    getUsernamePath,
			Method:  http.MethodGet,
			Handler: ag.getUsername,
		},
		{
			Path:    getCodeHashPath,
			Method:  http.MethodGet,
			Handler: ag.getCodeHash,
		},
		{
			Path:    getKeyPath,
			Method:  http.MethodGet,
			Handler: ag.getValueForKey,
		},
		{
			Path:    getKeysPath,
			Method:  http.MethodGet,
			Handler: ag.getKeyValuePairs,
		},
		{
			Path:    getDCDTBalancePath,
			Method:  http.MethodGet,
			Handler: ag.getDCDTBalance,
		},
		{
			Path:    getDCDTNFTDataPath,
			Method:  http.MethodGet,
			Handler: ag.getDCDTNFTData,
		},
		{
			Path:    getDCDTTokensPath,
			Method:  http.MethodGet,
			Handler: ag.getAllDCDTData,
		},
		{
			Path:    getRegisteredNFTsPath,
			Method:  http.MethodGet,
			Handler: ag.getNFTTokenIDsRegisteredByAddress,
		},
		{
			Path:    getDCDTTokensWithRolePath,
			Method:  http.MethodGet,
			Handler: ag.getDCDTTokensWithRole,
		},
		{
			Path:    getDCDTsRolesPath,
			Method:  http.MethodGet,
			Handler: ag.getDCDTsRoles,
		},
		{
			Path:    getGuardianData,
			Method:  http.MethodGet,
			Handler: ag.getGuardianData,
		},
		{
			Path:    getDataTrieMigrationStatusPath,
			Method:  http.MethodGet,
			Handler: ag.isDataTrieMigrated,
		},
	}
	ag.endpoints = endpoints

	return ag, nil
}

// getAccount returns a response containing information about the account correlated with provided address
func (ag *addressGroup) getAccount(c *gin.Context) {
	addr, options, err := extractBaseParams(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrCouldNotGetAccount, err)
		return
	}

	withKeys, err := parseBoolUrlParam(c, urlParamWithKeys)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrCouldNotGetAccount, err)
		return
	}

	options.WithKeys = withKeys

	accountResponse, blockInfo, err := ag.getFacade().GetAccount(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrCouldNotGetAccount, err)
		return
	}

	accountResponse.Address = addr
	shared.RespondWithSuccess(c, gin.H{"account": accountResponse, "blockInfo": blockInfo})
}

// getAccounts returns the state of the provided addresses on the specified block
func (ag *addressGroup) getAccounts(c *gin.Context) {
	var addresses []string
	err := c.ShouldBindJSON(&addresses)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrValidation, err)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrCouldNotGetAccount, err)
		return
	}

	accountsResponse, blockInfo, err := ag.getFacade().GetAccounts(addresses, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrCouldNotGetAccount, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"accounts": accountsResponse, "blockInfo": blockInfo})
}

// getBalance returns the balance for the address parameter
func (ag *addressGroup) getBalance(c *gin.Context) {
	addr, options, err := extractBaseParams(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBalance, err)
		return
	}

	balance, blockInfo, err := ag.getFacade().GetBalance(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBalance, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"balance": balance.String(), "blockInfo": blockInfo})
}

// getUsername returns the username for the address parameter
func (ag *addressGroup) getUsername(c *gin.Context) {
	addr, options, err := extractBaseParams(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetUsername, err)
		return
	}

	userName, blockInfo, err := ag.getFacade().GetUsername(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetUsername, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"username": userName, "blockInfo": blockInfo})
}

// getCodeHash returns the code hash for the address parameter
func (ag *addressGroup) getCodeHash(c *gin.Context) {
	addr, options, err := extractBaseParams(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetCodeHash, err)
		return
	}

	codeHash, blockInfo, err := ag.getFacade().GetCodeHash(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetCodeHash, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"codeHash": codeHash, "blockInfo": blockInfo})
}

// getValueForKey returns the value for the given address and key
func (ag *addressGroup) getValueForKey(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrGetValueForKey, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetValueForKey, err)
		return
	}

	key := c.Param("key")
	if key == "" {
		shared.RespondWithValidationError(c, errors.ErrGetValueForKey, errors.ErrEmptyKey)
		return
	}

	value, blockInfo, err := ag.getFacade().GetValueForKey(addr, key, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetValueForKey, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"value": value, "blockInfo": blockInfo})
}

// getGuardianData returns the guardian data and guarded state for a given account
func (ag *addressGroup) getGuardianData(c *gin.Context) {
	addr, options, err := extractBaseParams(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetGuardianData, err)
		return
	}

	guardianData, blockInfo, err := ag.getFacade().GetGuardianData(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetGuardianData, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"guardianData": guardianData, "blockInfo": blockInfo})
}

// addressGroup returns all the key-value pairs for the given address
func (ag *addressGroup) getKeyValuePairs(c *gin.Context) {
	addr, options, err := extractBaseParams(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetKeyValuePairs, err)
		return
	}

	value, blockInfo, err := ag.getFacade().GetKeyValuePairs(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetKeyValuePairs, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"pairs": value, "blockInfo": blockInfo})
}

// getDCDTBalance returns the balance for the given address and dcdt token
func (ag *addressGroup) getDCDTBalance(c *gin.Context) {
	addr, tokenIdentifier, options, err := extractGetDCDTBalanceParams(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetDCDTBalance, err)
		return
	}

	dcdtData, blockInfo, err := ag.getFacade().GetDCDTData(addr, tokenIdentifier, 0, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetDCDTBalance, err)
		return
	}

	tokenData := dcdtTokenData{
		TokenIdentifier: tokenIdentifier,
		Balance:         dcdtData.Value.String(),
		Properties:      hex.EncodeToString(dcdtData.Properties),
	}

	shared.RespondWithSuccess(c, gin.H{"tokenData": tokenData, "blockInfo": blockInfo})
}

// getDCDTsRoles returns the token identifiers and roles for a given address
func (ag *addressGroup) getDCDTsRoles(c *gin.Context) {
	addr, options, err := extractBaseParams(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetRolesForAccount, err)
		return
	}

	tokensRoles, blockInfo, err := ag.getFacade().GetDCDTsRoles(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetRolesForAccount, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"roles": tokensRoles, "blockInfo": blockInfo})
}

// getDCDTTokensWithRole returns the token identifiers where a given address has the given role
func (ag *addressGroup) getDCDTTokensWithRole(c *gin.Context) {
	addr, role, options, err := extractGetDCDTTokensWithRoleParams(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetDCDTTokensWithRole, err)
		return
	}

	tokens, blockInfo, err := ag.getFacade().GetDCDTsWithRole(addr, role, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetDCDTTokensWithRole, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"tokens": tokens, "blockInfo": blockInfo})
}

// getNFTTokenIDsRegisteredByAddress returns the token identifiers of the tokens where a given address is the owner
func (ag *addressGroup) getNFTTokenIDsRegisteredByAddress(c *gin.Context) {
	addr, options, err := extractBaseParams(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrRegisteredNFTTokenIDs, err)
		return
	}

	tokens, blockInfo, err := ag.getFacade().GetNFTTokenIDsRegisteredByAddress(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrRegisteredNFTTokenIDs, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"tokens": tokens, "blockInfo": blockInfo})
}

// getDCDTNFTData returns the nft data for the given token
func (ag *addressGroup) getDCDTNFTData(c *gin.Context) {
	addr, tokenIdentifier, nonce, options, err := extractGetDCDTNFTDataParams(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetDCDTNFTData, err)
		return
	}

	dcdtData, blockInfo, err := ag.getFacade().GetDCDTData(addr, tokenIdentifier, nonce.Uint64(), options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetDCDTNFTData, err)
		return
	}

	tokenData := buildTokenDataApiResponse(tokenIdentifier, dcdtData)
	shared.RespondWithSuccess(c, gin.H{"tokenData": tokenData, "blockInfo": blockInfo})
}

// getAllDCDTData returns the tokens list from this account
func (ag *addressGroup) getAllDCDTData(c *gin.Context) {
	addr, options, err := extractBaseParams(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetDCDTNFTData, err)
		return
	}

	tokens, blockInfo, err := ag.getFacade().GetAllDCDTTokens(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetDCDTNFTData, err)
		return
	}

	formattedTokens := make(map[string]*dcdtNFTTokenData)
	for tokenID, dcdtData := range tokens {
		tokenData := buildTokenDataApiResponse(tokenID, dcdtData)

		formattedTokens[tokenID] = tokenData
	}

	shared.RespondWithSuccess(c, gin.H{"dcdts": formattedTokens, "blockInfo": blockInfo})
}

// isDataTrieMigrated returns true if the data trie is migrated for the given address
func (ag *addressGroup) isDataTrieMigrated(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		shared.RespondWithValidationError(c, errors.ErrIsDataTrieMigrated, errors.ErrEmptyAddress)
		return
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrIsDataTrieMigrated, err)
		return
	}

	isMigrated, err := ag.getFacade().IsDataTrieMigrated(addr, options)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrIsDataTrieMigrated, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"isMigrated": isMigrated})
}

func buildTokenDataApiResponse(tokenIdentifier string, dcdtData *dcdt.DCDigitalToken) *dcdtNFTTokenData {
	tokenData := &dcdtNFTTokenData{
		TokenIdentifier: tokenIdentifier,
		Balance:         dcdtData.Value.String(),
		Properties:      hex.EncodeToString(dcdtData.Properties),
	}
	if dcdtData.TokenMetaData != nil {
		tokenData.Name = string(dcdtData.TokenMetaData.Name)
		tokenData.Nonce = dcdtData.TokenMetaData.Nonce
		tokenData.Creator = string(dcdtData.TokenMetaData.Creator)
		tokenData.Royalties = big.NewInt(int64(dcdtData.TokenMetaData.Royalties)).String()
		tokenData.Hash = dcdtData.TokenMetaData.Hash
		tokenData.URIs = dcdtData.TokenMetaData.URIs
		tokenData.Attributes = dcdtData.TokenMetaData.Attributes
	}

	return tokenData
}

func (ag *addressGroup) getFacade() addressFacadeHandler {
	ag.mutFacade.RLock()
	defer ag.mutFacade.RUnlock()

	return ag.facade
}

func extractBaseParams(c *gin.Context) (string, api.AccountQueryOptions, error) {
	addr := c.Param("address")
	if addr == "" {
		return "", api.AccountQueryOptions{}, errors.ErrEmptyAddress
	}

	options, err := extractAccountQueryOptions(c)
	if err != nil {
		return "", api.AccountQueryOptions{}, err
	}

	return addr, options, nil
}

func extractGetDCDTBalanceParams(c *gin.Context) (string, string, api.AccountQueryOptions, error) {
	addr, options, err := extractBaseParams(c)
	if err != nil {
		return "", "", api.AccountQueryOptions{}, err
	}

	tokenIdentifier := c.Param("tokenIdentifier")
	if tokenIdentifier == "" {
		return "", "", api.AccountQueryOptions{}, errors.ErrEmptyTokenIdentifier
	}

	return addr, tokenIdentifier, options, nil
}

func extractGetDCDTTokensWithRoleParams(c *gin.Context) (string, string, api.AccountQueryOptions, error) {
	addr, options, err := extractBaseParams(c)
	if err != nil {
		return "", "", api.AccountQueryOptions{}, err
	}

	role := c.Param("role")
	if role == "" {
		return "", "", api.AccountQueryOptions{}, errors.ErrEmptyRole
	}

	if !core.IsValidDCDTRole(role) {
		return "", "", api.AccountQueryOptions{}, fmt.Errorf("%w: %s", errors.ErrInvalidRole, role)
	}

	return addr, role, options, nil
}

func extractGetDCDTNFTDataParams(c *gin.Context) (string, string, *big.Int, api.AccountQueryOptions, error) {
	addr, options, err := extractBaseParams(c)
	if err != nil {
		return "", "", nil, api.AccountQueryOptions{}, err
	}

	tokenIdentifier := c.Param("tokenIdentifier")
	if tokenIdentifier == "" {
		return "", "", nil, api.AccountQueryOptions{}, errors.ErrEmptyTokenIdentifier
	}

	nonceAsStr := c.Param("nonce")
	if nonceAsStr == "" {
		return "", "", nil, api.AccountQueryOptions{}, errors.ErrNonceInvalid
	}

	nonceAsBigInt, okConvert := big.NewInt(0).SetString(nonceAsStr, 10)
	if !okConvert {
		return "", "", nil, api.AccountQueryOptions{}, errors.ErrNonceInvalid
	}

	return addr, tokenIdentifier, nonceAsBigInt, options, nil
}

// UpdateFacade will update the facade
func (ag *addressGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castFacade, ok := newFacade.(addressFacadeHandler)
	if !ok {
		return fmt.Errorf("%w for address group", errors.ErrFacadeWrongTypeAssertion)
	}

	ag.mutFacade.Lock()
	ag.facade = castFacade
	ag.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ag *addressGroup) IsInterfaceNil() bool {
	return ag == nil
}
