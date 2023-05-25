package jobs

import (
	"context"
	"fmt"
	"github.com/application-research/edge-ur/core"
	"github.com/filecoin-project/go-data-segment/datasegment"
	"github.com/filecoin-project/go-data-segment/util"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
	"io"
	"time"
)

type BucketCarBundler struct {
	Processor
}

func (b BucketCarBundler) Info() error {
	panic("implement me")
}

func (b BucketCarBundler) Run() error {

	// get all buckets cid
	var buckets []core.Bucket
	b.LightNode.DB.Model(&core.Bucket{}).Where("status = ?", "filled").Find(&buckets)

	if len(buckets) < 1 {
		fmt.Println("Not enough buckets to aggregate")
		return nil
	}

	// create one bundle
	bundleUuid, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}

	bundle := core.Bundle{
		Uuid:      bundleUuid.String(),
		Name:      bundleUuid.String(),
		Status:    "open",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	b.LightNode.DB.Create(&bundle)

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

		bucket.BundleUuid = bundle.Uuid
		if err != nil {
			panic(err)
		}
	}

	totalSizePow2, err := util.CeilPow2(uint64(intTotalSize * 2))
	if err != nil {
		panic(err)
	}

	bundle.Status = "processing"
	b.LightNode.DB.Save(&bundle)
	
	//create the aggregate object
	a, err := datasegment.NewAggregate(abi.PaddedPieceSize(totalSizePow2), subPieceInfos)
	if err != nil {
		panic(err)
	}

	// create the verifiable data aggregation car
	var readers []io.Reader
	for _, bucketV := range buckets {
		cCid, err := cid.Decode(bucketV.Cid)
		if err != nil {
			panic(err)
		}
		cData, errCData := b.LightNode.Node.GetFile(context.Background(), cCid) // get the node
		if errCData != nil {
			panic(errCData)
		}
		readers = append(readers, cData)
	}
	rootReader, err := a.AggregateObjectReader(readers)
	if err != nil {
		panic(err)
	}

	cidIPC, errPiece := a.IndexPieceCID()
	if errPiece != nil {
		fmt.Printf("%+v\n", errPiece)
	}
	fmt.Println("a.IndexPieceCid()", cidIPC)

	cidPC, errP := a.PieceCID()
	fmt.Println("a.PieceCID()", cidPC)
	if errP != nil {
		fmt.Printf("%+v\n", errP)
	}

	// process the deal
	fmt.Println("rootReader", rootReader)

	// add this to the node
	rootBundle, err := b.LightNode.Node.AddPinFile(context.Background(), rootReader, nil)
	if err != nil {
		panic(err)
	}

	bundle.Cid = rootBundle.Cid().String()
	bundle.Status = "filled"
	b.LightNode.DB.Save(&bundle)

	job := CreateNewDispatcher()
	job.AddJob(NewUploadBundleToDeltaProcessor(b.LightNode, rootReader, bundle))
	job.Start(1)

	return nil
}

func NewBucketCarBundler(ln *core.LightNode) IProcessor {
	return &BucketCarBundler{
		Processor{
			LightNode: ln,
		},
	}
}
