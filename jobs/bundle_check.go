package jobs

import (
	"encoding/json"
	"fmt"
	"github.com/application-research/edge-ur/core"
	"io"
	"net/http"
	"strconv"
	"time"
)

type BundleChecker struct {
	Bundle core.Bundle `json:"bundle"`
	Processor
}

func (r *BundleChecker) Info() error {
	return nil
}

func NewBundleChecker(ln *core.LightNode, bucket core.Bundle) IProcessor {
	DELTA_UPLOAD_API = ln.Config.ExternalApi.DeltaNodeApiUrl
	return &BundleChecker{
		Bundle: bucket,
		Processor: Processor{
			LightNode: ln,
		},
	}
}

func (r *BundleChecker) Run() error {

	// run thru the DIR contents and add them to the DB
	var bundle core.Bundle
	r.LightNode.DB.Raw("select * from bundles where id = ?", r.Bundle.ID).Scan(&bundle)

	c := bundle
	//for _, c := range content {
	contentId := strconv.Itoa(int(c.DeltaContentId))
	resp, err := http.Get(c.DeltaNodeUrl + "/open/stats/content/" + contentId)
	if err != nil {
		fmt.Println("Get error: ", err)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ReadAll error: ", err)
	}

	if resp.StatusCode != 200 {
		fmt.Println("Status code error: ", resp.StatusCode)
		return nil
	}
	var dealResult DealResult
	json.Unmarshal(body, &dealResult)
	c.LastMessage = dealResult.Content.LastMessage
	if len(dealResult.Deals) > 0 {
		c.Miner = dealResult.Deals[len(dealResult.Deals)-1].Miner
	}
	c.Status = dealResult.Content.Status
	r.LightNode.DB.Save(&c)

	// if the updated date is 1 day old, then we should just retry the request
	if c.Status == "transfer-started" {
		if c.UpdatedAt.Before(time.Now().Add(-24 * time.Hour)) {
			fmt.Println("Content is transfer-started, but has not been updated in 24 hours, retrying")
			c.Status = "retry"
			r.LightNode.DB.Save(&c)

			contentIdFromDelta := strconv.Itoa(int(c.DeltaContentId))
			respRetry, err := http.Get(c.DeltaNodeUrl + "/api/v1/retry/deal/end-to-end/" + contentIdFromDelta)
			if err != nil {
				fmt.Println("Get error: ", err)
				return nil
			}

			bodyRetry, err := io.ReadAll(respRetry.Body)
			if err != nil {
				fmt.Println("ReadAll error: ", err)
			}
			if respRetry.StatusCode != 200 {
				fmt.Println("Status code error: ", respRetry.StatusCode)
				return nil
			}

			var dealRetry DealRetry
			json.Unmarshal(bodyRetry, &dealRetry)
			c.LastMessage = "Retrying"
			c.DeltaContentId = int64(dealRetry.NewContentID)
			r.LightNode.DB.Save(&c)

		}
	}

	//}

	return nil
}
