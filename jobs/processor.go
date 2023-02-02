package jobs

import (
	"context"

	"github.com/application-research/edge-ur/core"
)

var workerPool = make(chan struct{}, 10)
var MODE = "remote-pin"
var PinEndpoint = ""
var UploadEndpoint = ""
var API_KEY = ""
var DELETE_AFTER_DEAL_MADE = "false"
var CONTENT_STATUS_CHECK_ENDPOINT = ""

type JobExecutable func() error
type IProcessor interface {
	Info() error
	Run() error
}

type ProcessorInfo struct {
	Name string
}
type Processor struct {
	context   *context.Context
	LightNode *core.LightNode
}
