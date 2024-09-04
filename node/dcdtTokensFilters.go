package node

import (
	"bytes"

	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-go/vm/systemSmartContracts"
)

type getRegisteredNftsFilter struct {
	addressBytes []byte
}

func (f *getRegisteredNftsFilter) filter(_ string, dcdtData *systemSmartContracts.DCDTDataV2) bool {
	return !bytes.Equal(dcdtData.TokenType, []byte(core.FungibleDCDT)) && bytes.Equal(dcdtData.OwnerAddress, f.addressBytes)
}

type getTokensWithRoleFilter struct {
	addressBytes []byte
	role         string
}

func (f *getTokensWithRoleFilter) filter(_ string, dcdtData *systemSmartContracts.DCDTDataV2) bool {
	for _, dcdtRoles := range dcdtData.SpecialRoles {
		if !bytes.Equal(dcdtRoles.Address, f.addressBytes) {
			continue
		}

		for _, specialRole := range dcdtRoles.Roles {
			if bytes.Equal(specialRole, []byte(f.role)) {
				return true
			}
		}
	}

	return false
}

type getAllTokensRolesFilter struct {
	addressBytes []byte
	outputRoles  map[string][]string
}

func (f *getAllTokensRolesFilter) filter(tokenIdentifier string, dcdtData *systemSmartContracts.DCDTDataV2) bool {
	for _, dcdtRoles := range dcdtData.SpecialRoles {
		if !bytes.Equal(dcdtRoles.Address, f.addressBytes) {
			continue
		}

		rolesStr := make([]string, 0, len(dcdtRoles.Roles))
		for _, roleBytes := range dcdtRoles.Roles {
			rolesStr = append(rolesStr, string(roleBytes))
		}

		f.outputRoles[tokenIdentifier] = rolesStr
		return true
	}
	return false
}
