package jobs

//
//import (
//	"bytes"
//	"context"
//	"fmt"
//	"github.com/application-research/edge-ur/core"
//	"github.com/filecoin-project/go-data-segment/datasegment"
//	"github.com/filecoin-project/go-data-segment/util"
//	"github.com/filecoin-project/go-state-types/abi"
//	"github.com/google/uuid"
//	"github.com/ipfs/go-cid"
//	"io"
//	"time"
//)
//
//type BucketCarBundler struct {
//	Miner string
//	Processor
//}
//
//func (b BucketCarBundler) Info() error {
//	panic("implement me")
//}
//
//func (b BucketCarBundler) Run() error {
//
//	// get all buckets cid
//	var buckets []core.Bucket
//	b.LightNode.DB.Model(&core.Bucket{}).Where("status = ? and miner = ?", "filled", b.Miner).Find(&buckets)
//
//	if len(buckets) < 5 {
//		fmt.Println("Not enough buckets to aggregate")
//		return nil
//	}
//
//	// create one bundle
//	bundleUuid, err := uuid.NewUUID()
//	if err != nil {
//		panic(err)
//	}
//
//	bundle := core.Bundle{
//		Uuid:      bundleUuid.String(),
//		Name:      bundleUuid.String(),
//		Status:    "open",
//		CreatedAt: time.Now(),
//		UpdatedAt: time.Now(),
//	}
//	b.LightNode.DB.Create(&bundle)
//
//	var intTotalSize int64
//	// create piece info of each bucket
//	var subPieceInfos []abi.PieceInfo
//	for _, bucket := range buckets {
//		bucketPieceCid, err := cid.Decode(bucket.PieceCid)
//
//		subPieceInfos = append(subPieceInfos, abi.PieceInfo{
//			Size:     abi.PaddedPieceSize(bucket.PieceSize),
//			PieceCID: bucketPieceCid,
//		})
//		intTotalSize += bucket.PieceSize
//
//		bucket.BundleUuid = bundle.Uuid
//		err = b.LightNode.DB.Save(&bucket).Error
//		if err != nil {
//			panic(err)
//		}
//	}
//
//	totalSizePow2, err := util.CeilPow2(uint64(intTotalSize * 2))
//	if err != nil {
//		panic(err)
//	}
//
//	bundle.Status = "processing"
//	b.LightNode.DB.Save(&bundle)
//
//	//create the aggregate object
//	a, err := datasegment.NewAggregate(abi.PaddedPieceSize(totalSizePow2), subPieceInfos)
//	if err != nil {
//		panic(err)
//	}
//
//	// create the verifiable data aggregation car
//	var readers []io.Reader
//	for _, bucketV := range buckets {
//		cCid, err := cid.Decode(bucketV.Cid)
//		if err != nil {
//			panic(err)
//		}
//		cData, errCData := b.LightNode.Node.GetFile(context.Background(), cCid) // get the node
//		if errCData != nil {
//			panic(errCData)
//		}
//		readers = append(readers, cData)
//	}
//	rootReader, err := a.AggregateObjectReader(readers)
//	if err != nil {
//		panic(err)
//	}
//
//	cidPC, errP := a.PieceCID()
//	if errP != nil {
//		fmt.Printf("%+v\n", errP)
//	}
//
//	// add this to the node
//	rootBundle, err := b.LightNode.Node.AddPinFile(context.Background(), rootReader, nil)
//	if err != nil {
//		panic(err)
//	}
//
//	bundle.FileCid = rootBundle.Cid().String()
//	bundle.AggregatePieceCid = cidPC.String()
//	bundle.Status = "filled"
//	bundle.Size = int64(a.DealSize)
//	bundle.DeltaNodeUrl = b.LightNode.Config.ExternalApi.DeltaNodeApiUrl
//	bundle.CreatedAt = time.Now()
//	bundle.UpdatedAt = time.Now()
//	b.LightNode.DB.Save(&bundle)
//
//	// update the bucket with proof piece info
//	for _, bucketX := range buckets {
//		bucketPieceCid, err := cid.Decode(bucketX.PieceCid)
//		if err != nil {
//			panic(err)
//		}
//
//		pieceInfo := abi.PieceInfo{
//			Size:     abi.PaddedPieceSize(bucketX.PieceSize),
//			PieceCID: bucketPieceCid,
//		}
//		proofForEach, err := a.ProofForPieceInfo(pieceInfo)
//		//aux, err := proofForEach.ComputeExpectedAuxData(datasegment.VerifierDataForPieceInfo(pieceInfo))
//
//		//bucketX.CommPa = aux.CommPa.String()
//		//bucketX.SizePa = int64(aux.SizePa)
//
//		incPW := &bytes.Buffer{}
//		proofForEach.MarshalCBOR(incPW)
//		bucketX.InclusionProof = incPW.Bytes()
//		bucketX.BundleUuid = bundle.Uuid
//
//		if err != nil {
//			panic(err)
//		}
//
//		b.LightNode.DB.Save(bucketX)
//
//	}
//
//	job := CreateNewDispatcher()
//	job.AddJob(NewUploadBundleToDeltaProcessor(b.LightNode, rootReader, bundle))
//	job.Start(1)
//
//	return nil
//}
//
//func NewBucketCarBundler(ln *core.LightNode, miner string) IProcessor {
//	return &BucketCarBundler{
//		miner,
//		Processor{
//			LightNode: ln,
//		},
//	}
//}
