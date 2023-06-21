package jobs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/application-research/edge-ur/core"
	commcid "github.com/filecoin-project/go-fil-commcid"
	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/filecoin-project/go-fil-markets/shared"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	logging "github.com/ipfs/go-log/v2"
	uio "github.com/ipfs/go-unixfs/io"
	"github.com/ipld/go-car"
	"io"
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
	buf := new(bytes.Buffer)
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

		_, err = io.Copy(buf, bytes.NewReader(cDataAgg.RawData()))
		if err != nil {
			panic(err)
		}

		//aggReaders = append(aggReaders, cDataAgg)
		dir.AddChild(context.Background(), cAgg.Name, cDataAgg)
	}
	dirNode, err := dir.GetNode()
	if err != nil {
		log.Errorf("error getting directory node: %s", err)
		return err
	}
	r.LightNode.Node.Add(context.Background(), dirNode)
	_, err = r.LightNode.Node.AddPinFile(context.Background(), buf, nil)
	if err != nil {
		log.Errorf("error adding file: %s", err)
		return err
	}

	pieceCid, carSize, unpaddedPieceSize, bufFile, err := GeneratePieceCommitment(context.Background(), dirNode.Cid(), r.LightNode.Node.Blockstore)
	bufFileN, err := r.LightNode.Node.AddPinFile(context.Background(), &bufFile, nil)

	if err != nil {
		log.Errorf("error generating piece commitment: %s", err)
	}
	bucket.PieceCid = pieceCid.String()
	bucket.PieceSize = int64(unpaddedPieceSize.Padded())
	bucket.DirCid = dirNode.Cid().String()
	bucket.Size = int64(carSize)
	bucket.Cid = bufFileN.Cid().String()
	bucket.Status = "ready-for-deal-making"
	r.LightNode.DB.Save(&bucket)

	return nil
}

func GeneratePieceCommitment(ctx context.Context, payloadCid cid.Cid, bstore blockstore.Blockstore) (cid.Cid, uint64, abi.UnpaddedPieceSize, bytes.Buffer, error) {
	selectiveCar := car.NewSelectiveCar(
		context.Background(),
		bstore,
		[]car.Dag{{Root: payloadCid, Selector: shared.AllSelector()}},
		car.MaxTraversalLinks(maxTraversalLinks),
		car.TraverseLinksOnlyOnce(),
	)

	buf := new(bytes.Buffer)
	blockCount := 0
	var oneStepBlocks []car.Block
	err := selectiveCar.Write(buf, func(block car.Block) error {
		oneStepBlocks = append(oneStepBlocks, block)
		blockCount++
		return nil
	})
	if err != nil {
		return cid.Undef, 0, 0, *buf, err
	}

	preparedCar, err := selectiveCar.Prepare()
	if err != nil {
		return cid.Undef, 0, 0, *buf, err
	}

	writer := new(commp.Calc)
	carWriter := &bytes.Buffer{}
	err = preparedCar.Dump(ctx, writer)
	if err != nil {
		return cid.Undef, 0, 0, *buf, err
	}
	commpc, size, err := writer.Digest()
	if err != nil {
		return cid.Undef, 0, 0, *buf, err
	}
	err = preparedCar.Dump(ctx, carWriter)
	if err != nil {
		return cid.Undef, 0, 0, *buf, err
	}

	commCid, err := commcid.DataCommitmentV1ToCID(commpc)
	if err != nil {
		return cid.Undef, 0, 0, *buf, err
	}

	return commCid, preparedCar.Size(), abi.PaddedPieceSize(size).Unpadded(), *buf, nil
}
