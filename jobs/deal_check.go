package jobs

import (
	"edge-ur/core"
	cid2 "github.com/ipfs/go-cid"
	"github.com/spf13/viper"
)

type DealCheckProcessor struct {
	Processor
}

func NewDealCheckProcessor(ln *core.LightNode) DealCheckProcessor {
	MODE = viper.Get("MODE").(string)
	UPLOAD_ENDPOINT = viper.Get("REMOTE_PIN_ENDPOINT").(string)
	DELETE_AFTER_DEAL_MADE = viper.Get("DELETE_AFTER_DEAL_MADE").(string)
	return DealCheckProcessor{
		Processor{
			LightNode: ln,
		},
	}
}

func (r *DealCheckProcessor) Run() {
	// get the deal of the contents and update

	if DELETE_AFTER_DEAL_MADE == "true" {

	}
}
func (r *DealCheckProcessor) deleteCidOnLocalNode(cidParam string) {
	// delete the cid on the local node
	cid, error := cid2.Decode(cidParam)

	if error != nil {
		panic(error)
	}
	r.LightNode.Node.Blockstore.DeleteBlock(*r.context, cid) //
}
