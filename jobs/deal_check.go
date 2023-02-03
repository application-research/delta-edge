package jobs

import (
	"encoding/json"
	"fmt"
	"github.com/application-research/edge-ur/core"
	cid2 "github.com/ipfs/go-cid"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type ContentStatusResponse struct {
	Content struct {
		ID            int64     `gorm:"primaryKey"`
		CreatedAt     time.Time `json:"createdAt"`
		UpdatedAt     time.Time `json:"updatedAt"`
		Cid           string    `json:"cid"`
		Name          string    `json:"name"`
		UserID        int       `json:"userId"`
		Description   string    `json:"description"`
		Size          int       `json:"size"`
		Type          int       `json:"type"`
		Active        bool      `json:"active"`
		Offloaded     bool      `json:"offloaded"`
		Replication   int       `json:"replication"`
		AggregatedIn  int       `json:"aggregatedIn"`
		Aggregate     bool      `json:"aggregate"`
		Pinning       bool      `json:"pinning"`
		PinMeta       string    `json:"pinMeta"`
		Replace       bool      `json:"replace"`
		Origins       string    `json:"origins"`
		Failed        bool      `json:"failed"`
		Location      string    `json:"location"`
		DagSplit      bool      `json:"dagSplit"`
		SplitFrom     int       `json:"splitFrom"`
		PinningStatus string    `json:"pinningStatus"`
		DealStatus    string    `json:"dealStatus"`
	} `json:"content"`
	Deals []struct {
		Deal         ContentDealResponse `json:"deal"`
		Transfer     interface{}         `json:"transfer"`
		OnChainState struct {
			SectorStartEpoch int `json:"sectorStartEpoch"`
			LastUpdatedEpoch int `json:"lastUpdatedEpoch"`
			SlashEpoch       int `json:"slashEpoch"`
		} `json:"onChainState"`
	} `json:"deals"`
	FailuresCount int `json:"failuresCount"`
}

type ContentDealResponse struct {
	ID                  int         `json:"ID"`
	CreatedAt           time.Time   `json:"CreatedAt"`
	UpdatedAt           time.Time   `json:"UpdatedAt"`
	DeletedAt           interface{} `json:"DeletedAt"`
	Content             int         `json:"content"`
	UserID              int         `json:"user_id"`
	PropCid             string      `json:"propCid"`
	DealUUID            string      `json:"dealUuid"`
	Miner               string      `json:"miner"`
	DealID              int         `json:"dealId"`
	Failed              bool        `json:"failed"`
	Verified            bool        `json:"verified"`
	Slashed             bool        `json:"slashed"`
	FailedAt            time.Time   `json:"failedAt"`
	DtChan              string      `json:"dtChan"`
	TransferStarted     time.Time   `json:"transferStarted"`
	TransferFinished    time.Time   `json:"transferFinished"`
	OnChainAt           time.Time   `json:"onChainAt"`
	SealedAt            time.Time   `json:"sealedAt"`
	DealProtocolVersion string      `json:"deal_protocol_version"`
	MinerVersion        string      `json:"miner_version"`
}

type DealCheckProcessor struct {
	Processor
}

func NewDealCheckProcessor(ln *core.LightNode) IProcessor {
	MODE = viper.Get("MODE").(string)
	UploadEndpoint = viper.Get("REMOTE_PIN_ENDPOINT").(string)
	DELETE_AFTER_DEAL_MADE = viper.Get("DELETE_AFTER_DEAL_MADE").(string)
	CONTENT_STATUS_CHECK_ENDPOINT = viper.Get("CONTENT_STATUS_CHECK_ENDPOINT").(string)
	return &DealCheckProcessor{
		Processor{
			LightNode: ln,
		},
	}
}

func (r *DealCheckProcessor) Info() error {
	//TODO implement me
	panic("implement me")
}

func (r *DealCheckProcessor) Run() error {
	// get the deal of the contents and update

	// get the contents that has estuary_request_id from the DB
	var contents []core.Content

	r.LightNode.DB.Model(&core.Content{}).Where("estuary_content_id IS NOT NULL").Find(&contents)

	for _, content := range contents {

		req, _ := http.NewRequest("GET",
			CONTENT_STATUS_CHECK_ENDPOINT+"/"+fmt.Sprint(content.EstuaryContentId), nil)

		client := &http.Client{}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+content.RequestingApiKey)
		res, err := client.Do(req)

		var contentStatus ContentStatusResponse
		err = json.NewDecoder(res.Body).Decode(&contentStatus)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if res.StatusCode == 200 {
			fmt.Println("", res.StatusCode)
			fmt.Println("content status", contentStatus)
			continue
		}

		// save the content status
		r.LightNode.DB.Transaction(func(tx *gorm.DB) error {
			//tx.Model(&ContentStatus{}).Save(&contentStatus)
			//tx.Model(&ContentDeal{}).Save(&contentStatus.Deals)
			return nil
		})

	}
	return nil
}
func (r *DealCheckProcessor) deleteCidOnLocalNode(cidParam string) {
	// delete the cid on the local node
	cid, error := cid2.Decode(cidParam)

	if error != nil {
		panic(error)
	}
	r.LightNode.Node.Blockstore.DeleteBlock(*r.context, cid) //
}

func (r *DealCheckProcessor) convertContentStatusResponseToModel(response ContentStatusResponse) (core.ContentStatus, error) {
	return core.ContentStatus{}, nil
}

func (r *DealCheckProcessor) convertContentDealResponseToModel(response ContentDealResponse) (core.ContentDeal, error) {

	return core.ContentDeal{}, nil
}
