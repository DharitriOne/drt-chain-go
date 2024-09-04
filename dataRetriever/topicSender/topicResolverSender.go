package topicsender

import (
	"github.com/DharitriOne/drt-chain-core-go/core"
	"github.com/DharitriOne/drt-chain-go/dataRetriever"
	"github.com/DharitriOne/drt-chain-go/p2p"
)

var _ dataRetriever.TopicResolverSender = (*topicResolverSender)(nil)

// ArgTopicResolverSender is the argument structure used to create new TopicResolverSender instance
type ArgTopicResolverSender struct {
	ArgBaseTopicSender
}

type topicResolverSender struct {
	*baseTopicSender
}

// NewTopicResolverSender returns a new topic resolver instance
func NewTopicResolverSender(arg ArgTopicResolverSender) (*topicResolverSender, error) {
	err := checkBaseTopicSenderArgs(arg.ArgBaseTopicSender)
	if err != nil {
		return nil, err
	}

	return &topicResolverSender{
		baseTopicSender: createBaseTopicSender(arg.ArgBaseTopicSender),
	}, nil
}

// Send is used to send an array buffer to a connected peer
// It is used when replying to a request
func (trs *topicResolverSender) Send(buff []byte, peer core.PeerID, destination p2p.MessageHandler) error {
	return trs.sendToConnectedPeer(trs.topicName, buff, peer, destination)
}

// IsInterfaceNil returns true if there is no value under the interface
func (trs *topicResolverSender) IsInterfaceNil() bool {
	return trs == nil
}
