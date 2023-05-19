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

type DealChecker struct {
	Processor
}
type DealRetry struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	NewContentID int    `json:"new_content_id"`
}
type DealResult struct {
	Content struct {
		ID                int    `json:"ID"`
		Name              string `json:"name"`
		Size              int    `json:"size"`
		Cid               string `json:"cid"`
		PieceCommitmentID int    `json:"piece_commitment_id"`
		Status            string `json:"status"`
		RequestType       string `json:"request_type"`
		ConnectionMode    string `json:"connection_mode"`
		AutoRetry         bool   `json:"auto_retry"`
		LastMessage       string `json:"last_message"`
		CreatedAt         string `json:"created_at"`
		UpdatedAt         string `json:"updated_at"`
	} `json:"content"`
	DealProposalParameters []struct {
		ID                 int    `json:"ID"`
		Content            int    `json:"content"`
		Label              string `json:"label"`
		Duration           int    `json:"duration"`
		RemoveUnsealedCopy bool   `json:"remove_unsealed_copy"`
		SkipIpniAnnounce   bool   `json:"skip_ipni_announce"`
		VerifiedDeal       bool   `json:"verified_deal"`
		CreatedAt          string `json:"created_at"`
		UpdatedAt          string `json:"updated_at"`
	} `json:"deal_proposal_parameters"`
	DealProposals []struct {
		ID        int    `json:"ID"`
		Content   int    `json:"content"`
		Unsigned  string `json:"unsigned"`
		Signed    string `json:"signed"`
		Meta      string `json:"meta"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	} `json:"deal_proposals"`
	Deals []struct {
		ID                  int       `json:"ID"`
		Content             int       `json:"content"`
		PropCid             string    `json:"propCid"`
		DealUUID            string    `json:"dealUuid"`
		Miner               string    `json:"miner"`
		DealID              int       `json:"dealId"`
		Failed              bool      `json:"failed"`
		Verified            bool      `json:"verified"`
		Slashed             bool      `json:"slashed"`
		FailedAt            time.Time `json:"failedAt"`
		DtChan              string    `json:"dtChan"`
		TransferStarted     time.Time `json:"transferStarted"`
		TransferFinished    time.Time `json:"transferFinished"`
		OnChainAt           time.Time `json:"onChainAt"`
		SealedAt            time.Time `json:"sealedAt"`
		LastMessage         string    `json:"lastMessage"`
		DealProtocolVersion string    `json:"deal_protocol_version"`
		CreatedAt           string    `json:"created_at"`
		UpdatedAt           string    `json:"updated_at"`
	} `json:"deals"`
	PieceCommitments []struct {
		ID                 int    `json:"ID"`
		Cid                string `json:"cid"`
		Piece              string `json:"piece"`
		Size               int    `json:"size"`
		PaddedPieceSize    int    `json:"padded_piece_size"`
		UnnpaddedPieceSize int    `json:"unnpadded_piece_size"`
		Status             string `json:"status"`
		LastMessage        string `json:"last_message"`
		CreatedAt          string `json:"created_at"`
		UpdatedAt          string `json:"updated_at"`
	} `json:"piece_commitments"`
}

func (r *DealChecker) Info() error {
	//TODO implement me
	panic("implement me")
}

func NewDealChecker(ln *core.LightNode) IProcessor {
	DELTA_UPLOAD_API = ln.Config.ExternalApi.ApiUrl
	return &DealChecker{
		Processor{
			LightNode: ln,
		},
	}
}

func (r *DealChecker) Run() error {

	// run thru the DIR contents and add them to the DB
	var content []core.Content
	// only get 25 at a time
	r.LightNode.DB.Raw("select * from contents where delta_content_id <> 0 and status not in (?)", "transfer-finished").Limit(25).Scan(&content)

	for _, c := range content {
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
		//}

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

	}

	return nil
}
