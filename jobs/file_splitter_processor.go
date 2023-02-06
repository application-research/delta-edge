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
	panic("implement me")
}
