package jobs

import (
	"context"
	"edge-ur/core"
)

var workerPool = make(chan struct{}, 10)
var MODE = "remote-pin"
var UPLOAD_ENDPOINT = ""
var API_KEY = ""
var DELETE_AFTER_DEAL_MADE = "false"

type Processor struct {
	ProcessorInterface
	context   *context.Context
	LightNode *core.LightNode
}

type ProcessorInterface interface {
	PreProcess() Processor
	PostProcess() Processor
	Run() Processor
	Verify() Processor
}
