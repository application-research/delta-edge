package jobs

import (
	"github.com/application-research/edge-ur/core"
	"github.com/spf13/viper"
	"strconv"
)

type FileSplitterProcessor struct {
	Processor
}

func NewFileSplitterProcessor(ln *core.LightNode) IProcessor {
	BucketSizeThreshold = viper.Get("CAR_GENERATOR_SIZE").(string)
	BucketSizeTh, _ = strconv.Atoi(BucketSizeThreshold)
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
