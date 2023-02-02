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
	"os"
	"time"

	"github.com/application-research/edge-ur/core"
	"github.com/spf13/viper"
)

type CarUploadToEstuaryProcessor struct {
	Processor
}

func NewCarUploadToEstuaryProcessor(ln *core.LightNode) IProcessor {
	MODE = viper.Get("MODE").(string)
	PinEndpoint = viper.Get("REMOTE_PIN_ENDPOINT").(string)
	UploadEndpoint = viper.Get("REMOTE_UPLOAD_ENDPOINT").(string)
	return &CarUploadToEstuaryProcessor{
		Processor{
			LightNode: ln,
		},
	}
}

func (r *CarUploadToEstuaryProcessor) Info() error {
	//TODO implement me
	panic("implement me")
}

func (r *CarUploadToEstuaryProcessor) Run() error {

	// create a worker group.
	// run the content processor.

	// get open buckets and create a car for each content cid
	var buckets []core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("status = ?", "car-assigned").Find(&buckets)

	//	for each bucket, get the contents and all estuary add-ipfs endpoint
	for _, bucket := range buckets {

		if MODE == "remote-pin" {
			requestBody := IpfsPin{
				CID:     bucket.Cid,
				Name:    bucket.Name,
				Origins: r.LightNode.GetLocalhostOrigins(),
			}

			payloadBuf := new(bytes.Buffer)

			json.NewEncoder(payloadBuf).Encode(requestBody)
			req, _ := http.NewRequest("POST",
				PinEndpoint,
				payloadBuf)

			client := &http.Client{}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+bucket.RequestingApiKey)
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
				bucket.Updated_at = time.Now()
				bucket.Status = "uploaded-to-estuary"
				bucket.EstuaryContentId = addIpfsResponse.RequestID
				r.LightNode.DB.Updates(&bucket)
			}
		}

		if MODE == "remote-upload" {
			decodedCid, err := cid.Decode(bucket.Cid)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			fileNd, err := r.LightNode.Node.GetFile(context.Background(), decodedCid)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			file, err := os.Create(bucket.Name)
			fileNd.WriteTo(file)

			payload := &bytes.Buffer{}
			writer := multipart.NewWriter(payload)
			partFile, err := writer.CreateFormFile("data", bucket.Name)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			io.Copy(partFile, file)
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
			req.Header.Set("Authorization", "Bearer "+bucket.RequestingApiKey)
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

				fmt.Println(uploadIpfsStatusResponse)

				// connect to delegates
				bucket.Updated_at = time.Now()
				bucket.Status = "uploaded-to-estuary"
				bucket.EstuaryContentId = uploadIpfsStatusResponse.EstuaryID
				r.LightNode.DB.Updates(&bucket)
			}
		}

		// keep it open until every content is uploaded
		bucket.Updated_at = time.Now()
		bucket.Status = "completed"
		r.LightNode.DB.Save(&bucket)
	}

	return nil
}
