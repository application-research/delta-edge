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
	"time"
)

var MiniBucketThresholdSize = 1000000000

const UnixfsLinksPerLevel = 1 << 10
const UnixfsChunkSize uint64 = 1 << 20

type CarGeneratorProcessor struct {
	Processor
}

func NewCarGeneratorProcessor(ln *core.LightNode) IProcessor {
	return &CarGeneratorProcessor{
		Processor{
			LightNode: ln,
		},
	}
}

func (c CarGeneratorProcessor) Info() error {
	//TODO implement me
	panic("implement me")
}

//func (c CarGeneratorProcessor) GenerateCar(bucketUuid string, contents []core.Content) {
//	dagServ := merkledag.NewDAGService(blockservice.New(c.LightNode.Node.Blockstore, offline.Exchange(c.LightNode.Node.Blockstore)))
//	cidBuilder, err := merkledag.PrefixForCidVersion(1)
//	if err != nil {
//		logger.Warn(err)
//		return
//	}
//	var layers []interface{}
//	rootNode := uio.NewDirectory(dagServ)
//	rootNode.SetCidBuilder(cidBuilder)
//	layers = append(layers, &rootNode)
//}
//
//func (c CarGeneratorProcessor) BuildFileNode(ctx context.Context, cidBuilder cid.Builder) (node ipld.Node, err error) {
//
//	params := ihelper.DagBuilderParams{
//		Maxlinks:   UnixfsLinksPerLevel,
//		RawLeaves:  true,
//		CidBuilder: cidBuilder,
//		Dagserv:    c.LightNode.Node.DAGService,
//		NoCopy:     true,
//	}
//	db, err := params.New(chunker.NewSizeSplitter(r, int64(UnixfsChunkSize)))
//	db.SetOffset(uint64(item.Start))
//	if err != nil {
//		logger.Warn(err)
//		return
//	}
//	node, err = balanced.Layout(db)
//	if err != nil {
//		logger.Warn(err)
//		return
//	}
//	return
//}

func (c CarGeneratorProcessor) Run() error {

	//	get the content for each bucket
	// get open buckets and create a car for each content cid
	var buckets []core.Bucket
	c.LightNode.DB.Model(&core.Bucket{}).Where("status = ?", "open").Find(&buckets)

	//	for each bucket, get the contents and all estuary add-ipfs endpoint
	for _, bucket := range buckets {
		var contents []core.Content
		c.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid = ?", bucket.UUID).Or("estuary_content_id = ''").Find(&contents)

		rootCarCid, err := c.buildCarForListOfContents(bucket.UUID, contents)
		if err != nil {
			panic(err)
		}

		// update bucket cid and status
		bucket.Cid = rootCarCid.String()
		bucket.Status = "car-assigned"
		bucket.Updated_at = time.Now()
		c.LightNode.DB.Updates(&bucket)
	}

	return nil
}

func (c *CarGeneratorProcessor) buildCarForListOfContents(bucketUuid string, contents []core.Content) (cid.Cid, error) {
	var rootCid cid.Cid

	// if there's only 1 content on the bucket, we just process the content itself.
	if len(contents) == 1 {
		// get the node and return the cid
		nd, err := c.getNodeForCid(contents[0])
		if err != nil {
			return rootCid, err
		}
		return nd.Cid(), nil
	}

	//	 if more than 1, pack it into a car.
	baseNode := merkledag.NewRawNode([]byte(bucketUuid))
	//var nodes []merkledag.ProtoNode
	for i, content := range contents {
		node := &merkledag.ProtoNode{}
		nodeFromCid, err := c.getNodeForCid(content)
		if err != nil {
			return cid.Undef, err
		}

		// link the first record to baseNode
		if i == 0 {
			node.AddNodeLink(nodeFromCid.String(), baseNode)

			data, err := c.LightNode.Node.Get(context.Background(), nodeFromCid.Cid())
			node.SetData(data.RawData())

			if err != nil {
				return cid.Undef, err
			}
		}

		node.AddNodeLink(nodeFromCid.String(), nodeFromCid)
		data, err := c.LightNode.Node.Get(context.Background(), nodeFromCid.Cid())
		node.SetData(data.RawData())

		// when last index
		if i == len(contents)-1 {
			rootCid = node.Cid()
		}

		c.addToBlockstore(c.LightNode.Node.DAGService, node)
	}

	rootNodeFromP, err := c.LightNode.Node.Get(context.Background(), rootCid)
	if err != nil {
		return cid.Undef, err
	}

	buf := new(bytes.Buffer)
	if err := car.WriteCar(context.Background(), c.LightNode.Node, []cid.Cid{rootCid}, buf); err != nil {
		panic(err)
	}
	fmt.Println("CAR file size: ", buf.Len())
	_, err = car.LoadCar(context.Background(), c.LightNode.Node.Blockstore, buf)
	if err != nil {
		panic(err)
	}

	c.traverseLinks(context.Background(), c.LightNode.Node.DAGService, rootNodeFromP)

	fmt.Println("Root CID: ", rootCid.String())
	return rootCid, nil
}

func (r *CarGeneratorProcessor) addToBlockstore(ds format.DAGService, nds ...format.Node) {
	for _, nd := range nds {
		if err := ds.Add(context.Background(), nd); err != nil {
			panic(err)
		}
	}
}

func (r *CarGeneratorProcessor) getNodeForCid(content core.Content) (format.Node, error) {
	decodedCid, err := cid.Decode(content.Cid)
	if err != nil {
		return nil, err
	}
	return r.LightNode.Node.Get(context.Background(), decodedCid)
}

// function to traverse all links
func (r *CarGeneratorProcessor) traverseLinks(ctx context.Context, ds format.DAGService, nd format.Node) {
	for _, link := range nd.Links() {
		node, err := link.GetNode(ctx, ds)
		if err != nil {
			panic(err)
		}
		fmt.Println("Node CID: ", node.Cid().String())
		r.traverseLinks(ctx, ds, node)
	}
}
