package jobs

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/application-research/edge-ur/core"
	"github.com/filecoin-project/go-data-segment/datasegment"
	"github.com/filecoin-project/go-data-segment/util"
	commcid "github.com/filecoin-project/go-fil-commcid"
	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/filecoin-project/go-fil-markets/shared"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	logging "github.com/ipfs/go-log/v2"
	uio "github.com/ipfs/go-unixfs/io"
	"github.com/ipld/go-car"
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

	// create node and raw per file (layer them)
	var contentsToUpdateWithPieceInfo []core.Content
	r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid = ?", bucketUuid).Find(&contentsToUpdateWithPieceInfo)

	// for each content, generate a node and a raw
	dir := uio.NewDirectory(r.LightNode.Node.DAGService)
	dir.SetCidBuilder(GetCidBuilderDefault())

	// get the subPieceInfos
	var subPieceInfos []abi.PieceInfo
	var intTotalSize int64
	var totalUnpaddedSize int64
	var bucketCommpPa string
	var bucketSizePa int64
	for _, c := range contentsToUpdateWithPieceInfo {
		cCid, err := cid.Decode(c.Cid)
		if err != nil {
			log.Errorf("error decoding cid: %s", err)
			c.Status = "error"
			c.LastMessage = err.Error()
			return err
		}

		pieceCid, _, unpadded, buf, err := GeneratePieceCommitment(context.Background(), cCid, r.LightNode.Node.Blockstore)
		c.PieceCid = pieceCid.String()
		if err != nil {
			log.Errorf("error generating piece commitment: %s", err)
			c.Status = "error"
			c.LastMessage = err.Error()
			return err
		}

		c.PieceSize = int64(unpadded.Padded())

		// add to the array
		subPieceInfos = append(subPieceInfos, abi.PieceInfo{
			Size:     unpadded.Padded(),
			PieceCID: pieceCid,
		})

		intTotalSize += int64(unpadded.Padded())
		totalUnpaddedSize += int64(unpadded)
		// write to blockstore
		ch, err := car.LoadCar(context.Background(), r.LightNode.Node.Blockstore, &buf)
		if err != nil {
			log.Errorf("error loading car: %s", err)
			c.Status = "error"
			c.LastMessage = err.Error()
			return err
		}
		selectiveCarNode, err := r.LightNode.Node.AddPinFile(context.Background(), &buf, nil)
		if err != nil {
			log.Errorf("error adding pin file: %s", err)
			c.Status = "error"
			c.LastMessage = err.Error()
			return err
		}

		if len(ch.Roots) > 0 {
			selectiveCarNodeCid := selectiveCarNode.Cid()
			c.SelectiveCarCid = selectiveCarNodeCid.String()
		}
		r.LightNode.DB.Save(&c)
	}

	// generate the aggregate using the subpieceinfos
	totalSizePow2, err := util.CeilPow2(uint64(intTotalSize * 2))
	if err != nil {
		log.Errorf("error generating ceil pow2: %s", err)
		return err
	}
	agg, err := datasegment.NewAggregate(abi.PaddedPieceSize(totalSizePow2), subPieceInfos)
	if err != nil {
		log.Errorf("error generating aggregate: %s", err)
		return err
	}

	var aggReaders []io.Reader
	var updateContentsForAgg []core.Content
	r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid = ?", bucketUuid).Find(&updateContentsForAgg)

	for _, cAgg := range updateContentsForAgg {
		fmt.Println("cAgg", cAgg.Cid, bucketUuid)
		cCidAgg, err := cid.Decode(cAgg.Cid)
		if err != nil {
			log.Errorf("error decoding cid: %s", err)
			return err
		}
		cDataAgg, errCData := r.LightNode.Node.GetFile(context.Background(), cCidAgg) // get the node
		if errCData != nil {
			log.Errorf("error getting file: %s", errCData)
			return errCData
		}
		aggReaders = append(aggReaders, cDataAgg)
	}

	rootReader, err := agg.AggregateObjectReader(aggReaders)
	if err != nil {
		log.Errorf("error aggregating object reader: %s", err)
		return err
	}

	aggNd, err := r.LightNode.Node.AddPinFile(context.Background(), rootReader, nil)
	if err != nil {
		log.Errorf("error adding pin file: %s", err)
		return err
	}

	filcPiece, _, unpaddedAgg, _, err := GeneratePieceCommitment(context.Background(), aggNd.Cid(), r.LightNode.Node.Blockstore)
	//aggBufNode, err := r.LightNode.Node.AddPinFile(context.Background(), &aggBufR, nil)
	if err != nil {
		log.Errorf("error generating piece commitment: %s", err)
		return err
	}

	var bucket core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("uuid = ?", bucketUuid).First(&bucket)
	bucket.FilCPieceCid = filcPiece.String()
	bucket.FilCPieceSize = int64(unpaddedAgg.Padded())
	bucket.RequestingApiKey = r.Bucket.RequestingApiKey
	bucket.Miner = "t017840"
	aggCid, err := agg.PieceCID()
	if err != nil {
		log.Errorf("error generating piece cid: %s", err)
		return err
	}

	bucket.PieceCid = aggCid.String()
	bucket.Cid = aggNd.Cid().String()
	//bucket.PieceSize = int64(unpaddedAgg.Padded())
	bucket.Status = "filled"
	bucket.Size = totalUnpaddedSize

	// get the proof for each piece
	var updatedContents []core.Content
	r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid = ?", bucketUuid).Find(&updatedContents)

	for _, cProof := range updatedContents {
		pieceCidStr, err := cid.Decode(cProof.PieceCid)
		if err != nil {
			log.Errorf("error decoding cid: %s", err)
			return err
		}

		pieceInfo := abi.PieceInfo{
			Size:     abi.PaddedPieceSize(cProof.PieceSize),
			PieceCID: pieceCidStr,
		}
		proofForEach, err := agg.ProofForPieceInfo(pieceInfo)
		verifierDataForEach := datasegment.VerifierDataForPieceInfo(pieceInfo)
		if err != nil {
			log.Errorf("error generating proof for piece info: %s", err)
			return err
		}
		aux, err := proofForEach.ComputeExpectedAuxData(datasegment.VerifierDataForPieceInfo(pieceInfo))
		if err != nil {
			log.Errorf("error computing expected aux data: %s", err)
			return err
		}

		bucketCid, _ := cid.Decode(bucket.PieceCid)
		if aux.CommPa.String() != bucketCid.String() {
			log.Error("commPa does not match")
			return fmt.Errorf("commPa does not match")
		}

		incW := &bytes.Buffer{}
		verifierDataW := &bytes.Buffer{}

		proofForEach.MarshalCBOR(incW)
		verifierDataForEach.MarshalCBOR(verifierDataW)

		cProof.InclusionProof = incW.Bytes()
		cProof.VerifierData = verifierDataW.Bytes()
		cProof.CommPa = aux.CommPa.String()

		cProof.SizePa = int64(aux.SizePa)
		cProof.CommPc = verifierDataForEach.CommPc.String()
		cProof.SizePc = int64(verifierDataForEach.SizePc)

		r.LightNode.DB.Save(&cProof)

		bucketCommpPa = aux.CommPa.String()
		bucketSizePa = int64(aux.SizePa)
	}
	bucket.PieceCid = bucketCommpPa
	bucket.PieceSize = bucketSizePa
	r.LightNode.DB.Save(&bucket)

	job := CreateNewDispatcher()
	job.AddJob(NewUploadCarToDeltaProcessor(r.LightNode, bucket.Uuid))
	job.Start(1)

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
