package jobs

import (
	"edge-ur/core"
)

type DealCheckProcessor struct {
	Processor
}

func NewDealCheckProcessor(ln *core.LightNode) DealCheckProcessor {
	return DealCheckProcessor{
		Processor{
			LightNode: ln,
		},
	}
}

func (r *DealCheckProcessor) Run() {
	// get the deal of the contents and update
}
