package jobs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
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
	File    io.Reader    `json:"file"`
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
	ContentID       int    `json:"content_id"`
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
		Status          string `json:"status"`
		Message         string `json:"message"`
		ContentID       int    `json:"content_id"`
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
	} `json:"replicated_contents"`
}

func NewUploadToEstuaryProcessor(ln *core.LightNode, contentToProcess core.Content, fileNode io.Reader) IProcessor {
	DELTA_UPLOAD_API = viper.Get("DELTA_NODE_API").(string)
	REPLICATION_FACTOR = viper.Get("REPLICATION_FACTOR").(string)
	return &UploadToEstuaryProcessor{
		contentToProcess,
		fileNode,
		Processor{
			LightNode: ln,
		},
	}
}

func (r *UploadToEstuaryProcessor) Info() error {
	panic("implement me")
}

func (r *UploadToEstuaryProcessor) Run() error {
	maxRetries := 5
	retryInterval := 5 * time.Second
	var content []core.Content
	r.LightNode.DB.Model(&core.Content{}).Where("id = ?", r.Content.ID).Find(&content)

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	partFile, err := writer.CreateFormFile("data", r.Content.Name)
	if err != nil {
		fmt.Println("CreateFormFile error: ", err)
		return nil
	}
	_, err = io.Copy(partFile, r.File)
	if err != nil {
		fmt.Println("Copy error: ", err)
		return nil
	}
	if partFile, err = writer.CreateFormField("metadata"); err != nil {
		fmt.Println("CreateFormField error: ", err)
		return nil
	}
	repFactor, err := strconv.Atoi(REPLICATION_FACTOR)
	if err != nil {
		fmt.Println("REPLICATION_FACTOR error: ", err)
		return nil
	}
	partMetadata := fmt.Sprintf(`{"auto_retry":true,"replication":%d}`, repFactor)
	if repFactor == 0 {
		partMetadata = fmt.Sprintf(`{"auto_retry":true}`)
	}

	if _, err = partFile.Write([]byte(partMetadata)); err != nil {
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
	var res *http.Response
	for j := 0; j < maxRetries; j++ {
		res, err = client.Do(req)
		if err != nil || res.StatusCode != http.StatusOK {
			fmt.Printf("Error sending request (attempt %d): %v\n", j+1, err)
			time.Sleep(retryInterval)
			continue
		} else {
			if res.StatusCode == 200 {
				var dealE2EUploadResponse DealE2EUploadResponse
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					fmt.Println(err)
					continue
				}
				err = json.Unmarshal(body, &dealE2EUploadResponse)
				if err != nil {
					fmt.Println(err)
					continue
				} else {
					if dealE2EUploadResponse.ContentID == 0 {
						continue
					} else {
						r.Content.UpdatedAt = time.Now()
						r.Content.Status = "uploaded-to-delta"
						r.Content.DeltaContentId = int64(dealE2EUploadResponse.ContentID)
						r.LightNode.DB.Save(&r.Content)

						// insert each replicated content into the database
						for _, replicatedContent := range dealE2EUploadResponse.ReplicatedContents {
							fmt.Println(replicatedContent)
							var replicatedContentModel core.Content
							//r.LightNode.DB.Model(&core.Content{}).Where("cid = ?", replicatedContent.Cid).Find(&replicatedContentModel)
							//if replicatedContentModel.ID == 0 {
							replicatedContentModel.Name = r.Content.Name
							replicatedContentModel.Cid = r.Content.Cid
							replicatedContentModel.Size = r.Content.Size
							replicatedContentModel.Status = replicatedContent.Status
							replicatedContentModel.LastMessage = replicatedContent.Message
							replicatedContentModel.DeltaNodeUrl = DELTA_UPLOAD_API
							replicatedContentModel.CreatedAt = time.Now()
							replicatedContentModel.UpdatedAt = time.Now()
							replicatedContentModel.RequestingApiKey = r.Content.RequestingApiKey
							replicatedContentModel.DeltaContentId = int64(replicatedContent.ContentID)
							r.LightNode.DB.Save(&replicatedContentModel)
							//}
						}
					}
				}
			}
		}
		break
	}

	return nil
}
