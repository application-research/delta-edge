package jobs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/application-research/edge-ur/core"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipld/go-car"
	selectorparse "github.com/ipld/go-ipld-prime/traversal/selector/parse"
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
		fmt.Println("Total hit size: ", 5*1024*1024)
		if totalSize > 5*1024*1024 {
			r.GenerateCarForBucket(bucket.Uuid)
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
	var nodeLayers []merkledag.ProtoNode
	for _, c := range content {

		cCid, err := cid.Decode(c.Cid)
		if err != nil {
			panic(err)
		}
		cData, errCData := r.LightNode.Node.Get(context.Background(), cCid)
		if errCData != nil {
			panic(errCData)
		}
		cNode := &merkledag.ProtoNode{}
		cRaw := merkledag.NewRawNode(cData.RawData())
		cNode.SetCidBuilder(GetCidBuilderDefault())
		cNode.AddNodeLink("raw", cRaw)

		if len(nodeLayers) == 0 {
			nodeLayers = append(nodeLayers, *cNode)
			continue
		}
		lastNodelayer := nodeLayers[len(nodeLayers)-1]
		cNode.AddNodeLink("node", &lastNodelayer)
		nodeLayers = append(nodeLayers, *cNode)

		// add to the dag service
		r.LightNode.Node.DAGService.Add(context.Background(), &lastNodelayer)
		r.LightNode.Node.DAGService.Add(context.Background(), cNode)
		r.LightNode.Node.DAGService.Add(context.Background(), cRaw)

	}

	// get the selective car
	// get the last node
	fmt.Println("Node layers: " + fmt.Sprint(len(nodeLayers)))
	for _, n := range nodeLayers {
		fmt.Println(n.Cid())
	}

	lastNode := nodeLayers[len(nodeLayers)-1]
	sc := car.NewSelectiveCar(context.Background(), r.LightNode.Node.BlockStore(), []car.Dag{{Root: lastNode.Cid(), Selector: selectorparse.CommonSelector_ExploreAllRecursively}})
	buf := new(bytes.Buffer)
	blockCount := 0
	var oneStepBlocks []car.Block
	err := sc.Write(buf, func(block car.Block) error {
		oneStepBlocks = append(oneStepBlocks, block)
		blockCount++
		return nil
	})

	// load the car
	ch, err := car.LoadCar(context.Background(), r.LightNode.Node.BlockStore(), buf)
	if err != nil {
		panic(err)
	}

	var combinedData []byte
	for _, c := range ch.Roots {
		fmt.Println("Root CID before: ", c.String())
		rootCid, err := r.LightNode.Node.Get(context.Background(), c)
		fmt.Println("Root CID after: ", rootCid.String())
		if err != nil {
			panic(err)
		}
		//traverseLinks(context.Background(), r.LightNode.Node.DAGService, rootCid)
		combinedData, err = traverseLinks(context.Background(), r.LightNode.Node.DAGService, rootCid)
		if err != nil {
			panic(err)
		}
	}

	var bucket core.CarBucket
	r.LightNode.DB.Model(&core.CarBucket{}).Where("uuid = ?", bucketUuid).First(&bucket)
	bucket.Cid = ch.Roots[0].String()
	bucket.RequestingApiKey = r.Content.RequestingApiKey
	bucket.Size = int64(len(combinedData))
	r.LightNode.DB.Save(&bucket)

	// upload to delta
	fmt.Println("combinedData: ", len(combinedData))
	job := CreateNewDispatcher()
	job.AddJob(NewUploadCarToDeltaProcessor(r.LightNode, bucket, bytes.NewBuffer(combinedData)))
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

func traverseLinks(ctx context.Context, ds format.DAGService, nd format.Node) ([]byte, error) {
	var combinedData []byte
	fmt.Println("Node: ", nd.Cid().String())
	for _, link := range nd.Links() {
		fmt.Println("Link: ", link.Cid.String())
		node, err := link.GetNode(ctx, ds)
		if err != nil {
			return nil, err
		}
		combinedData = append(combinedData, node.RawData()...)
		data, err := traverseLinks(ctx, ds, node)
		if err != nil {
			return nil, err
		}
		combinedData = append(combinedData, data...)
	}

	return combinedData, nil
}
