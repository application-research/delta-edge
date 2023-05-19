package jobs

import (
	"context"
	"fmt"
	"github.com/application-research/edge-ur/core"
	"github.com/application-research/filclient"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-merkledag"
	uio "github.com/ipfs/go-unixfs/io"
	"github.com/multiformats/go-multihash"
	"io"
)

type GenerateCarProcessor struct {
	Content core.Content `json:"content"`
	File    io.Reader    `json:"file"`
	Processor
}

func NewGenerateCarProcessor(ln *core.LightNode, contentToProcess core.Content, fileNode io.Reader) IProcessor {
	return &GenerateCarProcessor{
		contentToProcess,
		fileNode,
		Processor{
			LightNode: ln,
		},
	}
}

func (r *GenerateCarProcessor) Info() error {
	panic("implement me")
}

func (r *GenerateCarProcessor) Run() error {
	// check if there are open bucket. if there are, generate the car file for the bucket.

	var buckets []core.CarBucket
	r.LightNode.DB.Model(&core.CarBucket{}).Where("status = ?", "open").Find(&buckets)

	// get all open buckets and process
	for _, bucket := range buckets {

		// only process if the bucket is more than 5GB
		var content []core.Content
		r.LightNode.DB.Model(&core.Content{}).Where("car_bucket_uuid = ?", bucket.Uuid).Find(&content)

		var totalSize int64
		for _, c := range content {
			fmt.Println(c.Cid, c.Size)
			totalSize += c.Size
		}
		fmt.Println("Total size: ", totalSize)
		fmt.Println("Total hit size: ", r.LightNode.Config.Common.AggregateSize)
		if totalSize > r.LightNode.Config.Common.AggregateSize && len(content) > 1 {
			bucket.Status = "processing"
			r.LightNode.DB.Save(&bucket)

			r.GenerateCarForBucket(bucket.Uuid)
			continue
		}

	}

	return nil
	//	panic("implement me")
}

func (r *GenerateCarProcessor) GenerateCarForBucket(bucketUuid string) {
	// [node4 > raw4, node3 > [raw3, node2 > [raw2, node1 > raw1]]]

	// create node and raw per file (layer them)
	var content []core.Content
	r.LightNode.DB.Model(&core.Content{}).Where("car_bucket_uuid = ?", bucketUuid).Find(&content)

	// for each content, generate a node and a raw
	dir := uio.NewDirectory(r.LightNode.Node.DAGService)
	dir.SetCidBuilder(GetCidBuilderDefault())
	for _, c := range content {

		cCid, err := cid.Decode(c.Cid)
		if err != nil {
			panic(err)
		}
		cData, errCData := r.LightNode.Node.Get(context.Background(), cCid)
		if errCData != nil {
			panic(errCData)
		}
		dir.AddChild(context.Background(), fmt.Sprintf("%d-%s", c.ID, c.Name), cData)

		// add a piece info per child.

		//cDataReader, err := uio.NewDagReader(context.Background(), cData, r.LightNode.Node.DAGService)
		//pieceInfo, err := core.FastCommp(cDataReader)
		pieceCid, padded, _, err := filclient.GeneratePieceCommitment(context.Background(), cData.Cid(), r.LightNode.Node.Blockstore)
		if err != nil {
			panic(err)
		}

		fmt.Println("Piece info: ", pieceCid)
		fmt.Println("Piece info: ", padded)

		c.PieceCid = pieceCid.String()
		c.PieceSize = int64(padded)

		r.LightNode.DB.Save(&c)

		if err != nil {
			panic(err)
		}

	}
	dirNd, err := dir.GetNode()
	if err != nil {
		panic(err)
	}
	dirSize, err := dirNd.Size()
	// add to the dag service
	r.LightNode.Node.DAGService.Add(context.Background(), dirNd)

	var bucket core.CarBucket
	r.LightNode.DB.Model(&core.CarBucket{}).Where("uuid = ?", bucketUuid).First(&bucket)
	bucket.Cid = dirNd.Cid().String()
	bucket.RequestingApiKey = r.Content.RequestingApiKey
	bucket.Name = dirNd.Cid().String()
	// fast comp of the bucket
	//cDataReader, err := uio.NewDagReader(context.Background(), dirNd, r.LightNode.Node.DAGService)
	//pieceInfoRoot, err := core.FastCommp(cDataReader)

	pieceCid, padded, _, err := filclient.GeneratePieceCommitment(context.Background(), dirNd.Cid(), r.LightNode.Node.Blockstore)
	if err != nil {
		panic(err)
	}

	fmt.Println("Piece info: ", pieceCid)
	fmt.Println("Piece info: ", padded)

	bucket.PieceCid = pieceCid.String()
	bucket.PieceSize = int64(padded)

	bucket.Size = int64(dirSize)
	r.LightNode.DB.Save(&bucket)

	// compute piece info per user
	// compute piece info per bucket

	fmt.Println("Bucket CID: ", bucket.Cid)
	fmt.Println("Bucket Size: ", bucket.Size)

	//abi.PieceInfo{
	//	Size:     uint64(padded),
	//	PieceCID: pieceCid,
	//}

	// create proofs HERE and persist that on the database.
	//aggregate, err := datasegment.NewAggregate(paddedPieceSize, []abi.PieceInfo{pieceInfo})

	// process the deal
	job := CreateNewDispatcher()
	job.AddJob(NewUploadCarToDeltaProcessor(r.LightNode, bucket, nil, bucket.Cid))
	job.Start(1)
}

func GetCidBuilderDefault() cid.Builder {
	cidBuilder, err := merkledag.PrefixForCidVersion(1)
	if err != nil {
		panic(err)
	}
	cidBuilder.MhType = uint64(multihash.SHA2_256)
	cidBuilder.MhLength = -1
	return cidBuilder
}
