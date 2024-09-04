package disabled

import (
	"github.com/DharitriOne/drt-chain-core-go/data/rewardTx"
)

// RewardTxProcessor implements the RewardTransactionProcessor interface but does nothing as it is disabled
type RewardTxProcessor struct {
}

// ProcessRewardTransaction does nothing as it is disabled
func (rtp *RewardTxProcessor) ProcessRewardTransaction(_ *rewardTx.RewardTx) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtp *RewardTxProcessor) IsInterfaceNil() bool {
	return rtp == nil
}
