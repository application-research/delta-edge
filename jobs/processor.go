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
var CONTENT_STATUS_CHECK_ENDPOINT = ""

type JobExecutable func() error
type IProcessor interface {
	Run() error
}
type Processor struct {
	context   *context.Context
	LightNode *core.LightNode
}
