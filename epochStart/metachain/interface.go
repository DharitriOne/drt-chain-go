package metachain

import (
	"github.com/DharitriOne/drt-chain-core-go/display"
	"github.com/DharitriOne/drt-chain-go/state"
)

// AuctionListDisplayHandler should be able to display auction list data during selection process
type AuctionListDisplayHandler interface {
	DisplayOwnersData(ownersData map[string]*OwnerAuctionData)
	DisplayOwnersSelectedNodes(ownersData map[string]*OwnerAuctionData)
	DisplayAuctionList(
		auctionList []state.ValidatorInfoHandler,
		ownersData map[string]*OwnerAuctionData,
		numOfSelectedNodes uint32,
	)
	IsInterfaceNil() bool
}

// TableDisplayHandler should be able to display tables in log
type TableDisplayHandler interface {
	DisplayTable(tableHeader []string, lines []*display.LineData, message string)
	IsInterfaceNil() bool
}
