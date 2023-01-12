package jobs

import (
	"edge-ur/core"
)

type DirectoryProcessor struct {
	Processor
}

func NewDirectoryProcessor(ln *core.LightNode) DirectoryProcessor {
	return DirectoryProcessor{
		Processor{
			LightNode: ln,
		},
	}
}

func (r *DirectoryProcessor) Run() {
	// run thru the DIR contents and add them to the DB
}
