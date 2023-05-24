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
	uio "github.com/ipfs/go-unixfs/io"
)

type GenerateCarProcessor struct {
	Bucket core.Bucket
	Processor
}

func (g GenerateCarProcessor) Info() error {
	panic("implement me")
}

func (g GenerateCarProcessor) Run() error {
	g.GenerateCarForBucket(g.Bucket.Uuid)
	return nil
}

func NewGenerateCarProcessor(ln *core.LightNode, bucketToProcess core.Bucket) IProcessor {
	return &GenerateCarProcessor{
		bucketToProcess,
		Processor{
			LightNode: ln,
		},
	}
}

func (r *GenerateCarProcessor) GenerateCarForBucket(bucketUuid string) {
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

		// subPieceInfo
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
	//dirNd, err := dir.GetNode()
	//if err != nil {
	//	panic(err)
	//}
	//	dirSize, err := dirNd.Size()
	// add to the dag service
	aggNd, err := r.LightNode.Node.AddPinFile(context.Background(), buf, nil)
	if err != nil {
		panic(err)
	}

	var bucket core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("uuid = ?", bucketUuid).First(&bucket)
	bucket.Cid = aggNd.Cid().String()
	bucket.RequestingApiKey = r.Bucket.RequestingApiKey

	pieceCid, _, unpadded, err := filclient.GeneratePieceCommitment(context.Background(), aggNd.Cid(), r.LightNode.Node.Blockstore)
	if err != nil {
		panic(err)
	}

	bucket.PieceCid = pieceCid.String()
	bucket.PieceSize = int64(unpadded.Padded())

	bucket.Size = int64(unpadded)
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
	job.AddJob(NewUploadCarToDeltaProcessor(r.LightNode, bucket, bucket.Cid))
	job.Start(1)
}
