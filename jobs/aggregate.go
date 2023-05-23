package jobs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/application-research/edge-ur/core"
	"github.com/application-research/filclient"
	"github.com/filecoin-project/go-data-segment/datasegment"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-merkledag"
	uio "github.com/ipfs/go-unixfs/io"
	"github.com/multiformats/go-multihash"
	"io"
)

type AggregateProcessor struct {
	Content core.Content `json:"content"`
	File    io.Reader    `json:"file"`
	Processor
}

func NewAggregateProcessor(ln *core.LightNode, contentToProcess core.Content, fileNode io.Reader) IProcessor {
	return &AggregateProcessor{
		contentToProcess,
		fileNode,
		Processor{
			LightNode: ln,
		},
	}
}

func (r *AggregateProcessor) Info() error {
	panic("implement me")
}

func (r *AggregateProcessor) Run() error {
	// check if there are open bucket. if there are, generate the car file for the bucket.

	var buckets []core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("status = ?", "open").Find(&buckets)

	// get all open buckets and process
	query := "bucket_uuid = ?"
	if r.LightNode.Config.Common.AggregatePerApiKey && r.Content.RequestingApiKey != "" {
		query += " AND requesting_api_key = ?"
	}

	for _, bucket := range buckets {
		var content []core.Content
		r.LightNode.DB.Model(&core.Content{}).Where(query, bucket.Uuid, r.Content.RequestingApiKey).Find(&content)
		var totalSize int64
		var aggContent []core.Content
		for _, c := range content {
			fmt.Println(c.Cid, c.Size)
			totalSize += c.Size
			aggContent = append(aggContent, c)
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

func (r *AggregateProcessor) GenerateCarForBucket(bucketUuid string) {
	// [node4 > raw4, node3 > [raw3, node2 > [raw2, node1 > raw1]]]

	// create node and raw per file (layer them)
	var content []core.Content
	r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid = ?", bucketUuid).Find(&content)

	// for each content, generate a node and a raw
	dir := uio.NewDirectory(r.LightNode.Node.DAGService)
	dir.SetCidBuilder(GetCidBuilderDefault())
	buf := &bytes.Buffer{}
	var subPieceInfos []abi.PieceInfo
	for _, c := range content {

		cCid, err := cid.Decode(c.Cid)
		if err != nil {
			panic(err)
		}
		cData, errCData := r.LightNode.Node.GetFile(context.Background(), cCid) // get the node
		cDataNode, errCData := r.LightNode.Node.Get(context.Background(), cCid) // get the file
		if errCData != nil {
			panic(errCData)
		}
		dir.AddChild(context.Background(), c.Name, cDataNode)

		cData.WriteTo(buf)
		if err != nil {
			panic(err)
		}
		pieceCid, payloadSize, unpadded, err := filclient.GeneratePieceCommitment(context.Background(), cCid, r.LightNode.Node.Blockstore)
		if err != nil {
			panic(err)
		}

		fmt.Println("Piece cid: ", pieceCid)
		fmt.Println("Payload size: ", payloadSize)
		fmt.Println("Padded Piece: ", unpadded.Padded())

		c.PieceCid = pieceCid.String()
		c.PieceSize = int64(unpadded.Padded())

		r.LightNode.DB.Save(&c)

		// if pieceCid is already on the subPieceInfos, skip
		// if not, add to the subPieceInfos
		for _, subPieceInfo := range subPieceInfos {
			if subPieceInfo.PieceCID == pieceCid {
				continue
			}

			subPieceInfos = append(subPieceInfos, abi.PieceInfo{
				Size:     unpadded.Padded(),
				PieceCID: pieceCid,
			})
		}

		if err != nil {
			panic(err)
		}

	}
	dirNd, err := dir.GetNode()
	if err != nil {
		panic(err)
	}
	//	dirSize, err := dirNd.Size()
	// add to the dag service
	aggNd, err := r.LightNode.Node.AddPinFile(context.Background(), buf, nil)
	if err != nil {
		panic(err)
	}
	r.LightNode.Node.DAGService.Add(context.Background(), dirNd)
	r.LightNode.Node.DAGService.Add(context.Background(), aggNd)

	var bucket core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("uuid = ?", bucketUuid).First(&bucket)
	bucket.Cid = dirNd.Cid().String()
	bucket.RequestingApiKey = r.Content.RequestingApiKey

	pieceCid, _, unpadded, err := filclient.GeneratePieceCommitment(context.Background(), aggNd.Cid(), r.LightNode.Node.Blockstore)
	if err != nil {
		panic(err)
	}

	bucket.PieceCid = pieceCid.String()
	bucket.PieceSize = int64(unpadded.Padded())

	bucket.Size = int64(buf.Len())
	r.LightNode.DB.Save(&bucket)

	fmt.Println("Bucket CID: ", bucket.Cid)
	fmt.Println("Bucket Size: ", bucket.Size)
	fmt.Println("Bucket Piece CID: ", bucket.PieceCid)
	fmt.Println("Bucket Piece Size: ", bucket.PieceSize)

	aggregate, err := datasegment.NewAggregate(abi.PaddedPieceSize(bucket.PieceSize), subPieceInfos)
	if err != nil {
		fmt.Println("Err", err.Error())
	}
	fmt.Println("Aggregate: ", aggregate)
	aggCid, _ := aggregate.PieceCID()
	indexPieceCid, _ := aggregate.IndexPieceCID()
	aggregate.Index.ValidEntries()
	fmt.Println("Aggregate: ", aggCid.String())
	fmt.Println("OwnPiece: ", pieceCid.String())
	fmt.Println("aggregate.IndexPieceCID(): ", indexPieceCid.String())

	// process the deal
	job := CreateNewDispatcher()
	job.AddJob(NewUploadCarToDeltaProcessor(r.LightNode, bucket, buf, bucket.Cid))
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
