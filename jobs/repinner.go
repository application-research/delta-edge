package jobs

import "github.com/application-research/edge-ur/core"

type RepinnerProcessor struct {
	Processor
}

func NewRepinnerProcessor(ln *core.LightNode) IProcessor {
	return &RepinnerProcessor{
		Processor{
			LightNode: ln,
		},
	}
}

func (c RepinnerProcessor) Info() error {
	//TODO implement me
	panic("implement me")
}

func (c RepinnerProcessor) Run() error {

	//	get the content for each bucket

	// 	generate a car file

	// 	save the car file CID to the bucket table

	return nil
}
