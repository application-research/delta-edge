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

type BucketChecker struct {
	Bucket core.Bucket `json:"bucket"`
	Processor
}

func (r *BucketChecker) Info() error {
	return nil
}

func NewBucketChecker(ln *core.LightNode, bucket core.Bucket) IProcessor {
	DELTA_UPLOAD_API = ln.Config.ExternalApi.DeltaNodeApiUrl
	return &BucketChecker{
		Bucket: bucket,
		Processor: Processor{
			LightNode: ln,
		},
	}
}

func (r *BucketChecker) Run() error {

	// run thru the DIR contents and add them to the DB
	var bucket core.Bucket
	r.LightNode.DB.Raw("select * from buckets where id = ?", r.Bucket.ID).Scan(&bucket)

	//c := bucket
	//for _, c := range content {
	contentId := strconv.Itoa(int(bucket.DeltaContentId))
	resp, err := http.Get(bucket.DeltaNodeUrl + "/open/stats/content/" + contentId)
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
	bucket.LastMessage = dealResult.Content.LastMessage
	if len(dealResult.Deals) > 0 {
		bucket.Miner = dealResult.Deals[len(dealResult.Deals)-1].Miner
	}
	bucket.Status = dealResult.Content.Status
	bucket.DealId = int64(dealResult.Deals[0].DealID)
	r.LightNode.DB.Save(&bucket)

	// if the updated date is 1 day old, then we should just retry the request
	if bucket.Status == "transfer-started" {
		if bucket.UpdatedAt.Before(time.Now().Add(-24 * time.Hour)) {
			fmt.Println("Content is transfer-started, but has not been updated in 24 hours, retrying")
			bucket.Status = "retry"
			r.LightNode.DB.Save(&bucket)

			contentIdFromDelta := strconv.Itoa(int(bucket.DeltaContentId))
			respRetry, err := http.Get(bucket.DeltaNodeUrl + "/api/v1/retry/deal/end-to-end/" + contentIdFromDelta)
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
			bucket.LastMessage = "Retrying"
			bucket.DeltaContentId = int64(dealRetry.NewContentID)
			r.LightNode.DB.Save(&bucket)

		}
	}

	var contents []core.Content
	r.LightNode.DB.Raw("select * from contents where bucket_uuid = ?", bucket.Uuid).Scan(&contents)
	// check each content
	for _, c := range contents {
		c.DeltaContentId = bucket.DeltaContentId
		c.DeltaNodeUrl = bucket.DeltaNodeUrl
		c.DealId = bucket.DealId
		r.LightNode.DB.Save(&c)
		job := CreateNewDispatcher()
		job.AddJob(NewDealItemChecker(r.LightNode, c))
		job.Start(1)
	}

	return nil
}
