package resolvers_test

import (
	"github.com/DharitriOne/drt-chain-go/dataRetriever"
	"github.com/DharitriOne/drt-chain-go/dataRetriever/mock"
	"github.com/DharitriOne/drt-chain-go/p2p"
	"github.com/DharitriOne/drt-chain-go/testscommon/p2pmocks"
)

func createRequestMsg(dataType dataRetriever.RequestDataType, val []byte) p2p.MessageP2P {
	marshalizer := &mock.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataType, Value: val})
	return &p2pmocks.P2PMessageMock{DataField: buff}
}

func createRequestMsgWithChunkIndex(dataType dataRetriever.RequestDataType, val []byte, epoch uint32, chunkIndex uint32) p2p.MessageP2P {
	marshalizer := &mock.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(&dataRetriever.RequestData{
		Type:       dataType,
		Value:      val,
		Epoch:      epoch,
		ChunkIndex: chunkIndex,
	})
	return &p2pmocks.P2PMessageMock{DataField: buff}
}
