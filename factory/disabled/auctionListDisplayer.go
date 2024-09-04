package disabled

import (
	"github.com/DharitriOne/drt-chain-go/epochStart/metachain"
	"github.com/DharitriOne/drt-chain-go/state"
)

type auctionListDisplayer struct {
}

// NewDisabledAuctionListDisplayer creates a disabled auction list displayer
func NewDisabledAuctionListDisplayer() *auctionListDisplayer {
	return &auctionListDisplayer{}
}

// DisplayOwnersData does nothing
func (ald *auctionListDisplayer) DisplayOwnersData(_ map[string]*metachain.OwnerAuctionData) {
}

// DisplayOwnersSelectedNodes does nothing
func (ald *auctionListDisplayer) DisplayOwnersSelectedNodes(_ map[string]*metachain.OwnerAuctionData) {
}

// DisplayAuctionList does nothing
func (ald *auctionListDisplayer) DisplayAuctionList(
	_ []state.ValidatorInfoHandler,
	_ map[string]*metachain.OwnerAuctionData,
	_ uint32,
) {
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ald *auctionListDisplayer) IsInterfaceNil() bool {
	return ald == nil
}
