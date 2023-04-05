package jobs

import (
	"github.com/application-research/edge-ur/core"
)

type FileSplitterProcessor struct {
	Processor
}

func NewFileSplitterProcessor(ln *core.LightNode) IProcessor {
	return &FileSplitterProcessor{
		Processor{
			LightNode: ln,
		},
	}
}

func (r *FileSplitterProcessor) Info() error {
	panic("implement me")
}

func (r *FileSplitterProcessor) Run() error {
	// get the cid from the table

	// open the file

	// split and generate car

	// create a content request for each split

	// generate a metadata entry for each
	return nil
}
