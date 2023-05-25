package jobs

import (
	"bytes"
	"context"
	"github.com/application-research/edge-ur/core"
	"github.com/filecoin-project/go-data-segment/datasegment"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"io"
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
	b.LightNode.DB.Model(&core.Bucket{}).Where("status = ?", "uploaded-to-delta").Find(&buckets)
	var intTotalSize int64
	// create piece info of each bucket
	var subPieceInfos []abi.PieceInfo
	for _, bucket := range buckets {
		bucketPieceCid, err := cid.Decode(bucket.Cid)

		// subPieceInfo
		for _, subPieceInfo := range subPieceInfos {
			if subPieceInfo.PieceCID == bucketPieceCid {
				continue
			}

			subPieceInfos = append(subPieceInfos, abi.PieceInfo{
				Size:     abi.PaddedPieceSize(bucket.PieceSize),
				PieceCID: bucketPieceCid,
			})
			intTotalSize += bucket.PieceSize
		}

		if err != nil {
			panic(err)
		}
	}

	//create the aggregate object
	a, err := datasegment.NewAggregate(abi.PaddedPieceSize(intTotalSize), subPieceInfos)
	if err != nil {
		panic(err)
	}
	// create the verifiable data aggregation car
	buff := &bytes.Buffer{}
	for _, bucketV := range buckets {
		cCid, err := cid.Decode(bucketV.Cid)
		if err != nil {
			panic(err)
		}
		cData, errCData := b.LightNode.Node.GetFile(context.Background(), cCid) // get the node
		if errCData != nil {
			panic(errCData)
		}
		cData.WriteTo(buff)
	}

	//index_start, err := a.IndexStartPosition()
	r, err := a.IndexReader()
	_, err = io.Copy(buff, r)

	if err != nil {
		panic(err)
	}

	//index_size, err := a.IndexSize()

	return nil
}

func NewBucketsAggregator(ln *core.LightNode) IProcessor {
	return &BucketsAggregator{
		Processor{
			LightNode: ln,
		},
	}
}
