package jobs

import "edge-ur/core"

var workerPool = make(chan struct{}, 10)

type Processor struct {
	ProcessorInterface
	LightNode *core.LightNode
}

type ProcessorInterface interface {
	PreProcess()
	PostProcess()
	Run()
	Verify()
}
