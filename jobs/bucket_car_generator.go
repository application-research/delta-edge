package jobs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/application-research/edge-ur/core"
	"github.com/application-research/filclient"
	"github.com/filecoin-project/go-data-segment/datasegment"
	"github.com/filecoin-project/go-data-segment/util"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	uio "github.com/ipfs/go-unixfs/io"
	"io"
)

// The BucketCarGenerator type has a Bucket field and implements the Processor interface.
// @property Bucket - The `Bucket` property is a field of type `core.Bucket`. It is likely used to store or retrieve data
// related to cars, such as their make, model, year, and other attributes. The `BucketCarGenerator` struct likely
// represents a component or module that is responsible for generating new
// @property {Processor}  - The `BucketCarGenerator` struct has two properties:
type BucketCarGenerator struct {
	Bucket core.Bucket
	Processor
}

func (g BucketCarGenerator) Info() error {
	panic("implement me")
}

func (g BucketCarGenerator) Run() error {
	g.GenerateCarForBucket(g.Bucket.Uuid)
	return nil
}

func NewBucketCarGenerator(ln *core.LightNode, bucketToProcess core.Bucket) IProcessor {
	return &BucketCarGenerator{
		bucketToProcess,
		Processor{
			LightNode: ln,
		},
	}
}

func (r *BucketCarGenerator) GenerateCarForBucket(bucketUuid string) {

	// create node and raw per file (layer them)
	var content []core.Content
	r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid = ?", bucketUuid).Find(&content)

	// for each content, generate a node and a raw
	dir := uio.NewDirectory(r.LightNode.Node.DAGService)
	dir.SetCidBuilder(GetCidBuilderDefault())

	//buf := &bytes.Buffer{}
	var subPieceInfos []abi.PieceInfo
	var intTotalSize int64

	totalSizePow2, err := util.CeilPow2(uint64(intTotalSize * 2))

	for _, c := range content {
		cCid, err := cid.Decode(c.Cid)
		if err != nil {
			panic(err)
		}
		//cData, errCData := r.LightNode.Node.GetFile(context.Background(), cCid) // get the node
		cDataNode, errCData := r.LightNode.Node.Get(context.Background(), cCid) // get the file
		if errCData != nil {
			panic(errCData)
		}
		dir.AddChild(context.Background(), c.Name, cDataNode)

		pieceCid, pieceSize, _, err := filclient.GeneratePieceCommitment(context.Background(), cCid, r.LightNode.Node.Blockstore)
		c.PieceCid = pieceCid.String()
		c.PieceSize = int64(pieceSize)

		subPieceInfos = append(subPieceInfos, abi.PieceInfo{
			Size:     abi.PaddedPieceSize(pieceSize),
			PieceCID: pieceCid,
		})
		intTotalSize += c.Size
		r.LightNode.DB.Save(&c)
	}

	agg, err := datasegment.NewAggregate(abi.PaddedPieceSize(totalSizePow2), subPieceInfos)
	var aggReaders []io.Reader
	for _, cAgg := range content {
		cCidAgg, err := cid.Decode(cAgg.Cid)
		if err != nil {
			panic(err)
		}
		cDataAgg, errCData := r.LightNode.Node.GetFile(context.Background(), cCidAgg) // get the node
		if errCData != nil {
			panic(errCData)
		}
		aggReaders = append(aggReaders, cDataAgg)
	}

	rootReader, err := agg.AggregateObjectReader(aggReaders)
	if err != nil {
		panic(err)
	}

	aggNd, err := r.LightNode.Node.AddPinFile(context.Background(), rootReader, nil)
	if err != nil {
		panic(err)
	}

	var bucket core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("uuid = ?", bucketUuid).First(&bucket)
	bucket.Cid = aggNd.Cid().String()
	bucket.RequestingApiKey = r.Bucket.RequestingApiKey
	aggCid, err := agg.PieceCID()
	if err != nil {
		panic(err)
	}

	bucket.PieceCid = aggCid.String()
	bucket.PieceSize = int64(agg.DealSize)
	bucket.Status = "filled"
	bucket.Size = intTotalSize
	r.LightNode.DB.Save(&bucket)

	for _, c := range content {
		fmt.Println("PieceCid: ", c.PieceCid)
		pieceCidStr, err := cid.Decode(c.PieceCid)
		if err != nil {
			panic(err)
		}
		pieceInfo := abi.PieceInfo{
			Size:     abi.PaddedPieceSize(c.PieceSize),
			PieceCID: pieceCidStr,
		}
		proofForEach, err := agg.ProofForPieceInfo(pieceInfo)
		aux, err := proofForEach.ComputeExpectedAuxData(datasegment.VerifierDataForPieceInfo(pieceInfo))
		if err != nil {
			panic(err)
		}

		bucketPieceCid, _ := cid.Decode(bucket.PieceCid)
		if aux.CommPa != bucketPieceCid {
			panic("commPa does not match")
		}

		incW := &bytes.Buffer{}
		proofForEach.MarshalCBOR(incW)
		c.InclusionProof = incW.Bytes()

		r.LightNode.DB.Save(&c)
	}

	fmt.Println("Bucket CID: ", bucket.Cid)
	fmt.Println("Bucket Size: ", bucket.Size)
	fmt.Println("Bucket Piece CID: ", bucket.PieceCid)
	fmt.Println("Bucket Piece Size: ", bucket.PieceSize)

	//job := CreateNewDispatcher()
	//job.AddJob(NewBucketCarBundler(r.LightNode, bucket.Miner))
	//job.Start(1)

}
