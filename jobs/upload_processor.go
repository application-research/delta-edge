package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-cid"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/application-research/edge-ur/core"
	"github.com/spf13/viper"
)

type IpfsPin struct {
	CID     string                 `json:"cid"`
	Name    string                 `json:"name"`
	Origins []string               `json:"origins"`
	Meta    map[string]interface{} `json:"meta"`
}
type PinningStatus string

type IpfsPinStatusResponse struct {
	RequestID int64                  `json:"requestid"`
	Status    PinningStatus          `json:"status"`
	Created   time.Time              `json:"created"`
	Delegates []string               `json:"delegates"`
	Info      map[string]interface{} `json:"info"`
	Pin       IpfsPin                `json:"pin"`
}

type IpfsUploadStatusResponse struct {
	Cid                 string   `json:"cid"`
	RetrievalURL        string   `json:"retrieval_url"`
	EstuaryRetrievalURL string   `json:"estuary_retrieval_url"`
	EstuaryID           int64    `json:"estuaryId"`
	Providers           []string `json:"providers"`
}

type UploadToEstuaryProcessor struct {
	Content core.Content `json:"content"`
	Processor
}

type DealE2EUploadRequest struct {
	Cid            string `json:"cid,omitempty"`
	Miner          string `json:"miner,omitempty"`
	Duration       int64  `json:"duration,omitempty"`
	DurationInDays int64  `json:"duration_in_days,omitempty"`
	//Wallet             WalletRequest          `json:"wallet,omitempty"`
	//PieceCommitment    PieceCommitmentRequest `json:"piece_commitment,omitempty"`
	ConnectionMode     string `json:"connection_mode,omitempty"`
	Size               int64  `json:"size,omitempty"`
	StartEpoch         int64  `json:"start_epoch,omitempty"`
	StartEpochInDays   int64  `json:"start_epoch_in_days,omitempty"`
	Replication        int    `json:"replication,omitempty"`
	RemoveUnsealedCopy bool   `json:"remove_unsealed_copy"`
	SkipIPNIAnnounce   bool   `json:"skip_ipni_announce"`
	AutoRetry          bool   `json:"auto_retry"`
	Label              string `json:"label,omitempty"`
	DealVerifyState    string `json:"deal_verify_state,omitempty"`
}

// DealResponse Creating a new struct called DealResponse and then returning it.
type DealE2EUploadResponse struct {
	Status          string `json:"status"`
	Message         string `json:"message"`
	ContentId       int64  `json:"content_id,omitempty"`
	DealRequestMeta struct {
		Cid    string `json:"cid"`
		Miner  string `json:"miner"`
		Wallet struct {
		} `json:"wallet"`
		PieceCommitment struct {
		} `json:"piece_commitment"`
		ConnectionMode     string `json:"connection_mode"`
		Replication        int    `json:"replication"`
		RemoveUnsealedCopy bool   `json:"remove_unsealed_copy"`
		SkipIpniAnnounce   bool   `json:"skip_ipni_announce"`
		AutoRetry          bool   `json:"auto_retry"`
	} `json:"deal_request_meta"`
	DealProposalParameterRequestMeta struct {
		ID                 int    `json:"ID"`
		Content            int    `json:"content"`
		Label              string `json:"label"`
		Duration           int    `json:"duration"`
		RemoveUnsealedCopy bool   `json:"remove_unsealed_copy"`
		SkipIpniAnnounce   bool   `json:"skip_ipni_announce"`
		VerifiedDeal       bool   `json:"verified_deal"`
		CreatedAt          string `json:"created_at"`
		UpdatedAt          string `json:"updated_at"`
	} `json:"deal_proposal_parameter_request_meta"`
	ReplicatedContents []struct {
		ID               int    `json:"ID"`
		Name             string `json:"name"`
		Size             int    `json:"size"`
		Cid              string `json:"cid"`
		RequestingAPIKey string `json:"requesting_api_key"`
		Status           string `json:"status"`
		RequestType      string `json:"request_type"`
		ConnectionMode   string `json:"connection_mode"`
		AutoRetry        bool   `json:"auto_retry"`
		LastMessage      string `json:"last_message"`
		CreatedAt        string `json:"created_at"`
		UpdatedAt        string `json:"updated_at"`
	} `json:"replicated_contents"`
}

func NewUploadToEstuaryProcessor(ln *core.LightNode, contentToProcess core.Content) IProcessor {
	DELTA_UPLOAD_API = viper.Get("DELTA_NODE_API").(string)
	return &UploadToEstuaryProcessor{
		contentToProcess,
		Processor{
			LightNode: ln,
		},
	}
}

func (r *UploadToEstuaryProcessor) Info() error {
	//TODO implement me
	panic("implement me")
}

func (r *UploadToEstuaryProcessor) Run() error {

	var content []core.Content
	r.LightNode.DB.Model(&core.Content{}).Where("id = ?", r.Content.ID).Find(&content)

	decodedCid, err := cid.Decode(r.Content.Cid)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fileNd, err := r.LightNode.Node.GetFile(context.Background(), decodedCid)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	partFile, err := writer.CreateFormFile("data", r.Content.Name)
	if err != nil {
		fmt.Println("CreateFormFile error: ", err)
		return nil
	}
	_, err = io.Copy(partFile, fileNd)
	if err != nil {
		fmt.Println("Copy error: ", err)
		return nil
	}
	if partFile, err = writer.CreateFormField("metadata"); err != nil {
		fmt.Println("CreateFormField error: ", err)
		return nil
	}
	if _, err = partFile.Write([]byte(`{"auto_retry":true,"replication":3}`)); err != nil {
		fmt.Println("Write error: ", err)
		return nil
	}
	if err = writer.Close(); err != nil {
		fmt.Println("Close error: ", err)
		return nil
	}

	req, err := http.NewRequest("POST",
		DELTA_UPLOAD_API+"/api/v1/deal/end-to-end",
		payload)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+r.Content.RequestingApiKey)
	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		var dealE2EUploadResponse DealE2EUploadResponse
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		json.Unmarshal(body, &dealE2EUploadResponse)

		// connect to delegates
		r.Content.UpdatedAt = time.Now()
		r.Content.Status = "uploaded-to-delta"
		r.Content.DeltaContentId = dealE2EUploadResponse.ContentId
		r.LightNode.DB.Updates(&r.Content)
	}

	return nil
}
