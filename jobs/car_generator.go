package jobs

import "github.com/application-research/edge-ur/core"

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

func (c CarGeneratorProcessor) Run() error {

	//	get the content for each bucket

	// 	generate a car file

	// 	save the car file CID to the bucket table

	return nil
}
