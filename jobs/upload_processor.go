package jobs

import (
	"bytes"
	"edge-ur/core"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type IpfsPin struct {
	CID     string                 `json:"cid"`
	Name    string                 `json:"name"`
	Origins []string               `json:"origins"`
	Meta    map[string]interface{} `json:"meta"`
}
type PinningStatus string

type IpfsPinStatusResponse struct {
	RequestID string                 `json:"requestid"`
	Status    PinningStatus          `json:"status"`
	Created   time.Time              `json:"created"`
	Delegates []string               `json:"delegates"`
	Info      map[string]interface{} `json:"info"`
	Pin       IpfsPin                `json:"pin"`
}

type UploadToEstuaryProcessor struct {
	Processor
}

func NewUploadToEstuaryProcessor(ln *core.LightNode) UploadToEstuaryProcessor {
	return UploadToEstuaryProcessor{
		Processor{
			LightNode: ln,
		},
	}
}

// will improve to worker
func (r *UploadToEstuaryProcessor) Run() {
	// get open buckets and create a car for each content cid
	var buckets []core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("status = ?", "open").Find(&buckets)

	//	for each bucket, get the contents and all estuary add-ipfs endpoint
	for _, bucket := range buckets {
		var contents []core.Content
		r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid = ?", bucket.UUID).Find(&contents)

		// call the api to upload cid
		// update bucket cid and status
		for _, content := range contents {
			requestBody := IpfsPin{
				CID:  content.Cid,
				Name: content.Name,
			}
			payloadBuf := new(bytes.Buffer)
			json.NewEncoder(payloadBuf).Encode(requestBody)
			req, _ := http.NewRequest("POST",
				"https://api.estuary.tech/content/add-ipfs",
				payloadBuf)

			client := &http.Client{}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+content.RequestingApiKey)
			res, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer res.Body.Close()
			var addIpfsResponse IpfsPinStatusResponse
			body, err := ioutil.ReadAll(res.Body)
			json.Unmarshal(body, &addIpfsResponse)

			content.Updated_at = time.Now()
			content.EstuaryContentId = addIpfsResponse.RequestID
			r.LightNode.DB.Updates(&content)

		}
		bucket.Updated_at = time.Now()
		bucket.Status = "completed"
		r.LightNode.DB.Save(&bucket)
	}
}
