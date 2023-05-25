package jobs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/application-research/edge-ur/core"
	"github.com/filecoin-project/go-data-segment/datasegment"
	"github.com/filecoin-project/go-data-segment/util"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"io"
	"os"
)

type BucketsAggregator struct {
	Processor
}

func (b BucketsAggregator) Info() error {
	//TODO implement me
	panic("implement me")
}

func (b BucketsAggregator) Run() error {

	// get all buckets cid
	var buckets []core.Bucket
	b.LightNode.DB.Model(&core.Bucket{}).Where("status = ?", "filled").Find(&buckets)

	if len(buckets) < 2 {
		fmt.Println("Not enough buckets to aggregate")
		return nil
	}

	var intTotalSize int64
	// create piece info of each bucket
	var subPieceInfos []abi.PieceInfo
	for _, bucket := range buckets {
		bucketPieceCid, err := cid.Decode(bucket.PieceCid)

		subPieceInfos = append(subPieceInfos, abi.PieceInfo{
			Size:     abi.PaddedPieceSize(bucket.PieceSize),
			PieceCID: bucketPieceCid,
		})
		intTotalSize += bucket.PieceSize

		if err != nil {
			panic(err)
		}
	}

	totalSizePow2, err := util.CeilPow2(uint64(intTotalSize * 2))
	if err != nil {
		panic(err)
	}

	//create the aggregate object
	fmt.Println("intTotalSize", intTotalSize)
	a, err := datasegment.NewAggregate(abi.PaddedPieceSize(totalSizePow2), subPieceInfos)
	if err != nil {
		panic(err)
	}

	// create the verifiable data aggregation car
	indexStart, err := a.IndexStartPosition()
	buffer := bytes.NewBuffer(nil)
	reader := bytes.NewReader(buffer.Bytes())
	_, err = reader.Seek(int64(indexStart), io.SeekStart)
	_, err = io.Copy(buffer, reader)

	for _, bucketV := range buckets {
		cCid, err := cid.Decode(bucketV.Cid)
		if err != nil {
			panic(err)
		}
		cData, errCData := b.LightNode.Node.GetFile(context.Background(), cCid) // get the node
		if errCData != nil {
			panic(errCData)
		}

		reader.Seek(0, os.SEEK_SET)
		io.Copy(buffer, cData)
		cData.Close()
	}

	cidIPC, _ := a.IndexPieceCID()
	fmt.Println("a.IndexPieceCid()", cidIPC)

	cidPC, _ := a.PieceCID()
	fmt.Println("a.PieceCID()", cidPC)
	if err != nil {
		panic(err)
	}

	return nil
}

func NewBucketsAggregator(ln *core.LightNode) IProcessor {
	return &BucketsAggregator{
		Processor{
			LightNode: ln,
		},
	}
}
