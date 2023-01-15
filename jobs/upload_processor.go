package jobs

import (
	"bytes"
	"context"
	"edge-ur/core"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
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
	MODE = viper.Get("MODE").(string)
	UPLOAD_ENDPOINT = viper.Get("REMOTE_PIN_ENDPOINT").(string)
	return UploadToEstuaryProcessor{
		Processor{
			LightNode: ln,
		},
	}
}

// TODO: WORKER GROUPS!!!!!
func (r *UploadToEstuaryProcessor) Run() {

	// create a worker group.
	// run the content processor.

	// get open buckets and create a car for each content cid
	var buckets []core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("status = ?", "open").Find(&buckets)

	//	for each bucket, get the contents and all estuary add-ipfs endpoint
	for _, bucket := range buckets {

		var contents []core.Content
		r.LightNode.DB.Model(&core.Content{}).Where("bucket_uuid = ?", bucket.UUID).Or("estuary_content_id = ''").Find(&contents)

		// call the api to upload cid
		// update bucket cid and status
		for _, content := range contents {

			if MODE == "remote-pin" {
				requestBody := IpfsPin{
					CID:     content.Cid,
					Name:    content.Name,
					Origins: r.LightNode.GetLocalhostOrigins(),
				}

				payloadBuf := new(bytes.Buffer)

				json.NewEncoder(payloadBuf).Encode(requestBody)
				req, _ := http.NewRequest("POST",
					UPLOAD_ENDPOINT,
					payloadBuf)

				client := &http.Client{}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+content.RequestingApiKey)
				res, err := client.Do(req)
				defer res.Body.Close()
				if err != nil {
					fmt.Println(err)
					return
				}

				if res.StatusCode != 202 {
					fmt.Println("error uploading to estuary", res.StatusCode)
					return
				}

				if res.StatusCode == 202 {
					var addIpfsResponse IpfsPinStatusResponse
					body, err := ioutil.ReadAll(res.Body)
					if err != nil {
						fmt.Println(err)
						return
					}
					json.Unmarshal(body, &addIpfsResponse)

					// connect to delegates
					r.LightNode.ConnectToDelegates(context.Background(), addIpfsResponse.Pin.Origins)
					content.Updated_at = time.Now()
					content.Status = "uploaded-to-estuary"
					content.EstuaryContentId = addIpfsResponse.RequestID
					r.LightNode.DB.Updates(&content)
				}
			}
		}

		// keep it open until every content is uploaded
		bucket.Updated_at = time.Now()
		bucket.Status = "completed"
		r.LightNode.DB.Save(&bucket)
	}
}
