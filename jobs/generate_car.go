package jobs

import (
	"bytes"
	"context"
	"github.com/application-research/edge-ur/core"
	"github.com/ipfs/go-cid"
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

	// only process if the bucket is more than 5GB
	// get all open buckets and process
	for _, bucket := range buckets {
		r.GenerateCarForBucket(bucket.Uuid)
	}

	panic("implement me")
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
		cNode := merkledag.ProtoNode{}
		cRaw := merkledag.NewRawNode(cData.RawData())
		cNode.SetCidBuilder(GetCidBuilderDefault())

		// get last node from nodelayers
		if len(nodeLayers) == 0 {
			// add raw to the last node
			cNode.AddNodeLink("raw", cRaw)

			// then add the last node to the nodelayers
			nodeLayers = append(nodeLayers, cNode)
			continue
		} else {

			lastNodelayer := nodeLayers[len(nodeLayers)-1]

			// add raw to the last node
			lastNodelayer.AddNodeLink("raw", cRaw)

			// then add the last node to the nodelayers
			nodeLayers = append(nodeLayers, cNode)
		}

	}

	// get the selective car
	// get the last node
	lastNode := nodeLayers[len(nodeLayers)-1]
	sc := car.NewSelectiveCar(context.Background(), r.LightNode.Node.BlockStore(), []car.Dag{{Root: lastNode.Cid(), Selector: selectorparse.CommonSelector_ExploreAllRecursively}})
	buf := new(bytes.Buffer)
	var oneStepBlocks []car.Block
	err := sc.Write(buf, func(block car.Block) error {
		oneStepBlocks = append(oneStepBlocks, block)
		return nil
	})

	// load the car
	ch, err := car.LoadCar(context.Background(), r.LightNode.Node.BlockStore(), buf)
	if err != nil {
		panic(err)
	}

	var bucket core.CarBucket
	r.LightNode.DB.Model(&core.CarBucket{}).Where("uuid = ?", bucketUuid).First(&bucket)
	bucket.Cid = ch.Roots[0].String()
	bucket.Status = "car-generated"
	r.LightNode.DB.Save(&bucket)

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
