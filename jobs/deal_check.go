package jobs

import (
	"fmt"
	"github.com/application-research/edge-ur/core"
)

type DealChecker struct {
	Processor
}

func (r *DealChecker) Info() error {
	//TODO implement me
	panic("implement me")
}

func NewDealChecker(ln *core.LightNode) IProcessor {
	return &DealChecker{
		Processor{
			LightNode: ln,
		},
	}
}

func (r *DealChecker) Run() error {

	// run thru the DIR contents and add them to the DB
	var content []core.Content
	r.LightNode.DB.Raw("select * from contents where status = 'uploaded-to-delta'").Scan(&content)

	for _, c := range content {
		fmt.Println(c)
		// check the deal status
		// get the delta id and check deal.
		// if deal is active, set status to 'deal-active'
		// if deal is not active, set status to 'deal-failed'
	}

	return nil
}
