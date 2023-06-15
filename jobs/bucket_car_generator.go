package jobs

import (
	"context"
	"fmt"
	"github.com/application-research/edge-ur/core"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	uio "github.com/ipfs/go-unixfs/io"
)

// The log constant is a logging.Logger that is used to log messages for the jobs package.
var log = logging.Logger("jobs")

// The maxTraversalLinks constant is an int that represents the maximum number of traversal links.
const maxTraversalLinks = 32 * (1 << 20)

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

// The Run method of the BucketCarGenerator struct takes no parameters and returns an error. It is used to run the
// BucketCarGenerator struct.
func (g BucketCarGenerator) Run() error {
	if err := g.GenerateCarForBucket(g.Bucket.Uuid); err != nil {
		log.Errorf("error generating car for bucket: %s", err)
		return err
	}
	return nil
}

// NewBucketCarGenerator is a function that takes a LightNode and a bucketToProcess as parameters and returns a
// BucketCarGenerator struct. It is used to create a new BucketCarGenerator struct.
func NewBucketCarGenerator(ln *core.LightNode, bucketToProcess core.Bucket) IProcessor {
	return &BucketCarGenerator{
		bucketToProcess,
		Processor{
			LightNode: ln,
		},
	}
}

// GenerateCarForBucket is a method of the BucketCarGenerator struct. It takes a bucketUuid string as a parameter and
// returns nothing. It is used to generate a car with aggregated contents for a bucket
func (r *BucketCarGenerator) GenerateCarForBucket(bucketUuid string) error {

	// get the main bucket
	var bucket core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("uuid = ?", bucketUuid).First(&bucket)

	var updateContentsForAgg []core.Content
	r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid = ?", bucketUuid).Find(&updateContentsForAgg)

	// for each content, generate a node and a raw
	dir := uio.NewDirectory(r.LightNode.Node.DAGService)
	dir.SetCidBuilder(GetCidBuilderDefault())

	for _, cAgg := range updateContentsForAgg {
		fmt.Println("cAgg", cAgg.Cid, bucketUuid)
		cCidAgg, err := cid.Decode(cAgg.Cid)
		if err != nil {
			log.Errorf("error decoding cid: %s", err)
			return err
		}
		cDataAgg, errCData := r.LightNode.Node.Get(context.Background(), cCidAgg) // get the node
		if errCData != nil {
			log.Errorf("error getting file: %s", errCData)
			return errCData
		}

		//aggReaders = append(aggReaders, cDataAgg)
		dir.AddChild(context.Background(), cAgg.Name, cDataAgg)
	}
	dirNode, err := dir.GetNode()
	if err != nil {
		log.Errorf("error getting directory node: %s", err)
		return err
	}
	bucket.Cid = dirNode.Cid().String()
	r.LightNode.DB.Save(&bucket)

	job := CreateNewDispatcher()
	job.AddJob(NewUploadCarToDeltaProcessor(r.LightNode, bucket.Uuid))
	job.Start(1)

	return nil
}
