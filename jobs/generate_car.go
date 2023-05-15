package jobs

import (
	"github.com/application-research/edge-ur/core"
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
		r.GenerateCarForBucket(bucket.Uuid)
	}

	panic("implement me")
}

func (r *GenerateCarProcessor) GenerateCarForBucket(bucketUuid string) {

	// save the car file, get CID and upload to delta.
}
