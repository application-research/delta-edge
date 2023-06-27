package jobs

import (
	"fmt"
	"io"

	"github.com/application-research/edge-ur/core"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-merkledag"
	"github.com/multiformats/go-multihash"
)

type BucketAggregator struct {
	Force   bool         `json:"force"`
	Content core.Content `json:"content"`
	File    io.Reader    `json:"file"`
	Processor
}

func NewBucketAggregator(ln *core.LightNode, contentToProcess core.Content, fileNode io.Reader, force bool) IProcessor {
	return &BucketAggregator{
		force,
		contentToProcess,
		fileNode,
		Processor{
			LightNode: ln,
		},
	}
}

func (r *BucketAggregator) Info() error {
	panic("implement me")
}

// Run is the main function of the BucketAggregator struct. It is responsible for aggregating the contents of a bucket
func (r *BucketAggregator) Run() error {
	// check if there are open bucket. if there are, generate the car file for the bucket.

	var buckets []core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("status = ?", "open").Find(&buckets)

	// for each bucket, get all the contents and check if the total size is greater than the aggregate size limit (default 1GB)
	// if it is, generate a car file for the bucket and update the bucket status to processing
	for _, bucket := range buckets {
		var content []core.Content
		// get all open buckets and process
		query := "bucket_uuid = ?"
		if r.LightNode.Config.Common.AggregatePerApiKey && r.Content.RequestingApiKey != "" {
			query += " AND requesting_api_key = ?"
			r.LightNode.DB.Model(&core.Content{}).Where(query, bucket.Uuid, r.Content.RequestingApiKey).Find(&content)
		} else {
			r.LightNode.DB.Model(&core.Content{}).Where(query, bucket.Uuid).Find(&content)
		}
		var totalSize int64
		var aggContent []core.Content
		for _, c := range content {
			totalSize += c.Size
			aggContent = append(aggContent, c)
		}

		// get bucket policy
		var policy core.Policy
		r.LightNode.DB.Model(&core.Policy{}).Where("id = ?", bucket.PolicyId).First(&policy)

		if totalSize > policy.BucketSize && len(content) > 1 {
			fmt.Println("Generating car file for bucket: ", bucket.Uuid)
			bucket.Status = "processing"
			r.LightNode.DB.Save(&bucket)

			// process the car generator
			job := CreateNewDispatcher()
			genCar := NewBucketCarGenerator(r.LightNode, bucket)
			job.AddJob(genCar)
			job.Start(1)
			continue
		}
	}

	return nil
	//	panic("implement me")
}

// GetCidBuilderDefault is a helper function that returns a default cid builder
func GetCidBuilderDefault() cid.Builder {
	cidBuilder, err := merkledag.PrefixForCidVersion(1)
	if err != nil {
		panic(err)
	}
	cidBuilder.MhType = uint64(multihash.SHA2_256)
	cidBuilder.MhLength = -1
	return cidBuilder
}
