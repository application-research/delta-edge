package jobs

import (
	"fmt"
	"github.com/application-research/edge-ur/core"
)

type RetryProcessor struct {
	Processor
}

func (r RetryProcessor) Info() error {
	//TODO implement me
	panic("implement me")
}

func (r RetryProcessor) Run() error {

	var buckets []core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("status = ?", "processing").Find(&buckets)

	// get all open buckets and process
	query := "bucket_uuid = ?"

	for _, bucket := range buckets {
		var content []core.Content
		r.LightNode.DB.Model(&core.Content{}).Where(query, bucket.Uuid).Find(&content)
		var totalSize int64
		var aggContent []core.Content
		for _, c := range content {

			var isContentExists bool
			for _, existingContent := range aggContent {
				if existingContent.Cid == c.Cid {
					isContentExists = true
					break
				}
			}
			if !isContentExists {
				totalSize += c.Size
				aggContent = append(aggContent, c)
			}
		}
		fmt.Println("Total size: ", totalSize)
		fmt.Println("Total hit size: ", r.LightNode.Config.Common.AggregateSize)
		if totalSize > r.LightNode.Config.Common.AggregateSize && len(content) > 1 {
			bucket.Status = "processing"
			r.LightNode.DB.Save(&bucket)

			// process the car generator
			job := CreateNewDispatcher()
			genCar := NewGenerateCarProcessor(r.LightNode, bucket)
			job.AddJob(genCar)
			job.Start(1)
			continue
		}
	}
	return nil
}

func NewRetryProcessor(ln *core.LightNode) IProcessor {
	return &RetryProcessor{
		Processor{
			LightNode: ln,
		},
	}

}
