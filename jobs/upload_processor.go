package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-cid"
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
	Processor
}

func NewUploadToEstuaryProcessor(ln *core.LightNode) IProcessor {
	MODE = viper.Get("MODE").(string)
	PinEndpoint = viper.Get("REMOTE_PIN_ENDPOINT").(string)
	UploadEndpoint = viper.Get("REMOTE_UPLOAD_ENDPOINT").(string)
	return &UploadToEstuaryProcessor{
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

	// create a worker group.
	// run the content processor.

	// get open buckets and create a car for each content cid
	var buckets []core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("status = ?", "open").Find(&buckets)

	//	for each bucket, get the contents
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
					PinEndpoint,
					payloadBuf)

				client := &http.Client{}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+content.RequestingApiKey)
				res, err := client.Do(req)

				if err != nil {
					fmt.Println(err)
					return nil
				}

				if res.StatusCode != 202 {
					fmt.Println("error uploading to estuary", res.StatusCode)
					return nil
				}

				fmt.Println(res.StatusCode)
				if res.StatusCode == 202 {
					var addIpfsResponse IpfsPinStatusResponse
					body, err := ioutil.ReadAll(res.Body)
					if err != nil {
						fmt.Println(err)
						return nil
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
			if MODE == "remote-upload" {
				decodedCid, err := cid.Decode(content.Cid)
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
				requestFileByteArr, err := ioutil.ReadAll(fileNd)
				partFile, err := writer.CreateFormFile("data", content.Name)
				partFile.Write(requestFileByteArr)
				if err != nil {
					fmt.Println(err)
					return nil
				}
				err = writer.Close()
				if err != nil {
					fmt.Println(err)
					return nil
				}
				req, _ := http.NewRequest("POST",
					UploadEndpoint,
					payload)

				client := &http.Client{}
				req.Header.Add("content-type", "multipart/form-data")
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer "+content.RequestingApiKey)
				res, err := client.Do(req)

				if err != nil {
					fmt.Println(err)
					return nil
				}

				if res.StatusCode == 200 {
					var uploadIpfsStatusResponse IpfsUploadStatusResponse
					body, err := ioutil.ReadAll(res.Body)
					if err != nil {
						fmt.Println(err)
						return nil
					}
					json.Unmarshal(body, &uploadIpfsStatusResponse)

					// connect to delegates
					content.Updated_at = time.Now()
					content.Status = "uploaded-to-estuary"
					content.EstuaryContentId = uploadIpfsStatusResponse.EstuaryID
					r.LightNode.DB.Updates(&content)
				}
			}
		}

		// keep it open until every content has been uploaded
		bucket.Updated_at = time.Now()
		bucket.Status = "completed"
		r.LightNode.DB.Save(&bucket)
	}

	return nil
}
