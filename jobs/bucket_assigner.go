package jobs

import (
	"github.com/spf13/viper"
	"strconv"
	"time"

	"github.com/application-research/edge-ur/core"
	"github.com/google/uuid"
)

var BucketSizeTh int

type BucketAssignProcessor struct {
	Processor
}

func NewBucketAssignProcessor(ln *core.LightNode) IProcessor {
	BucketSizeThreshold = viper.Get("BUCKET_SIZE_THRESHOLD").(string)
	BucketSizeTh, _ = strconv.Atoi(BucketSizeThreshold)
	return &BucketAssignProcessor{
		Processor{
			LightNode: ln,
		},
	}
}

func (r *BucketAssignProcessor) Info() error {
	panic("implement me")
}

func (r *BucketAssignProcessor) Run() error {
	// run the content processor.
	var contents []core.Content
	var contentCollectionToBucket []core.Content
	r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid is ''").Find(&contents)

	// get range of content ids and assign a bucket
	for _, content := range contents {
		if r.GetContentCollectionToBucket(contentCollectionToBucket) < int64(BucketSizeTh) {
			contentCollectionToBucket = append(contentCollectionToBucket, content)
		}
	}
	if r.GetContentCollectionToBucket(contentCollectionToBucket) >= int64(BucketSizeTh) {
		// if there are contents, create a new bucket and assign it to the contents
		uuid, err := uuid.NewUUID()
		if err != nil {
			panic(err)
		}
		if len(contentCollectionToBucket) > 0 {
			// create a new bucket
			bucket := core.Bucket{
				Status:     "open",        // open, car-assigned, piece-assigned, storage-deal-done
				Name:       uuid.String(), // same as uuid
				UUID:       uuid.String(),
				Created_at: time.Now(), // log it.
			}
			r.LightNode.DB.Create(&bucket)

			// assign bucket to contents
			//r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid is ''").Update("bucket_uuid", bucket.UUID).Update("status", "bucket-assigned")
			r.UpdateContentCollectionToBucket(contentCollectionToBucket, bucket)
		}

		contentCollectionToBucket = []core.Content{}

	}

	return nil
}

func (r *BucketAssignProcessor) GetContentCollectionToBucket(contentCollectionToBucket []core.Content) int64 {
	// get size of bucket
	var size int64
	for _, content := range contentCollectionToBucket {
		size += content.Size
	}

	return size
}

func (r *BucketAssignProcessor) UpdateContentCollectionToBucket(contentCollectionToBucket []core.Content, bucket core.Bucket) int64 {
	// get size of bucket
	var size int64
	for _, content := range contentCollectionToBucket {
		content.BucketUuid = bucket.UUID
		content.Status = "bucket-assigned"
		r.LightNode.DB.Save(&content)
	}

	return size
}
