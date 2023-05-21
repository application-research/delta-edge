package jobs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/application-research/edge-ur/core"
	"github.com/application-research/edge-ur/utils"
	"github.com/application-research/filclient"
	"github.com/filecoin-project/go-data-segment/datasegment"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
	uio "github.com/ipfs/go-unixfs/io"
	"io"
	"time"
)

type SplitterProcessor struct {
	Content core.Content `json:"content"`
	File    io.Reader    `json:"file"`
	Processor
}

func NewSplitterProcessor(ln *core.LightNode, contentToProcess core.Content, fileNode io.Reader) IProcessor {
	return &SplitterProcessor{
		contentToProcess,
		fileNode,
		Processor{
			LightNode: ln,
		},
	}
}

func (r *SplitterProcessor) Info() error {
	panic("implement me")
}

func (r *SplitterProcessor) Run() error {

	// split the file.
	fileSplitter := new(core.FileSplitter)
	fileSplitter.ChuckSize = r.LightNode.Config.Common.MaxSizeToSplit
	arrBts, err := fileSplitter.SplitFileFromReader(r.File) // nice split.
	if err != nil {
		panic(err)
	}

	// create a bucket
	bucketUuid, err := uuid.NewUUID()
	bucket := core.Bucket{
		Status:           "open",
		Name:             bucketUuid.String(),
		RequestingApiKey: r.Content.RequestingApiKey,
		Uuid:             bucketUuid.String(),
		Miner:            r.Content.Miner,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	r.LightNode.DB.Create(&bucket)

	// create a content for each split
	for i, b := range arrBts {
		bNd, err := r.LightNode.Node.AddPinFile(context.Background(), bytes.NewReader(b), nil)
		if err != nil {
			panic(err)
		}
		newContent := core.Content{
			Name:             "split-" + string(i),
			Size:             int64(len(b)),
			Cid:              bNd.Cid().String(),
			DeltaNodeUrl:     r.Content.DeltaNodeUrl,
			RequestingApiKey: r.Content.RequestingApiKey,
			Status:           utils.STATUS_PINNED,
			Miner:            r.Content.Miner,
			BucketUuid:       bucket.Uuid,
			MakeDeal:         true,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		r.LightNode.DB.Create(&newContent)
	}

	r.GenerateCarForBucket(bucket.Uuid)

	return nil
}

func (r *SplitterProcessor) GenerateCarForBucket(bucketUuid string) {
	// [node4 > raw4, node3 > [raw3, node2 > [raw2, node1 > raw1]]]

	// create node and raw per file (layer them)
	var content []core.Content
	r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid = ?", bucketUuid).Find(&content)

	// for each content, generate a node and a raw
	dir := uio.NewDirectory(r.LightNode.Node.DAGService)
	dir.SetCidBuilder(GetCidBuilderDefault())
	buf := new(bytes.Buffer)
	var subPieceInfos []abi.PieceInfo
	for _, c := range content {

		cCid, err := cid.Decode(c.Cid)
		if err != nil {
			panic(err)
		}
		cData, errCData := r.LightNode.Node.Get(context.Background(), cCid)
		if errCData != nil {
			panic(errCData)
		}
		dir.AddChild(context.Background(), c.Name, cData)
		_, err = io.Copy(buf, bytes.NewReader(cData.RawData()))
		if err != nil {
			panic(err)
		}
		pieceCid, payloadSize, unpadded, err := filclient.GeneratePieceCommitment(context.Background(), cData.Cid(), r.LightNode.Node.Blockstore)
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
	dirSize, err := dirNd.Size()
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
	bucket.Name = dirNd.Cid().String()

	pieceCid, _, unpadded, err := filclient.GeneratePieceCommitment(context.Background(), aggNd.Cid(), r.LightNode.Node.Blockstore)
	if err != nil {
		panic(err)
	}

	bucket.PieceCid = pieceCid.String()
	bucket.PieceSize = int64(unpadded.Padded())

	bucket.Size = int64(dirSize)
	r.LightNode.DB.Save(&bucket)

	fmt.Println("Bucket CID: ", bucket.Cid)
	fmt.Println("Bucket Size: ", bucket.Size)
	fmt.Println("Bucket Piece CID: ", bucket.PieceCid)
	fmt.Println("Bucket Piece Size: ", bucket.PieceSize)

	// dealSize := abi.PaddedPieceSize(1 << 20)

	// create proofs HERE and persist that on the database.
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
