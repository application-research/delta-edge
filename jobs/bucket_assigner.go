package jobs

import (
	"edge-ur/core"
	"github.com/google/uuid"
	"time"
)

type BucketAssignProcessor struct {
	Processor
}

func NewBucketAssignProcessor(ln *core.LightNode) BucketAssignProcessor {
	return BucketAssignProcessor{
		Processor{
			LightNode: ln,
		},
	}
}

func (r *BucketAssignProcessor) Run() {
	// run the content processor.
	var contents []core.Content
	r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid is ''").Find(&contents)

	// get range of content ids and assign a bucket
	// if there are contents, create a new bucket and assign it to the contents
	uuid, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	if len(contents) > 0 {
		// create a new bucket
		bucket := core.Bucket{
			Status:     "open",        // open, car-assigned, piece-assigned, storage-deal-done
			Name:       uuid.String(), // same as uuid
			UUID:       uuid.String(),
			Created_at: time.Now(), // log it.
		}
		r.LightNode.DB.Create(&bucket)

		// assign bucket to contents
		r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid is ''").Update("bucket_uuid", bucket.UUID).Update("status", "bucket-assigned")

	}

}
