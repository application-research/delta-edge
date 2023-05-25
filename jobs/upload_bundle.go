package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/application-research/edge-ur/utils"
	"github.com/ipfs/go-cid"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/application-research/edge-ur/core"
)

type UploadBundleToDeltaProcessor struct {
	BundleReader io.Reader   `json:"bundle_reader"`
	Bundle       core.Bundle `json:"bundle"`
	Processor
}

func NewUploadBundleToDeltaProcessor(ln *core.LightNode, reader io.Reader, bundle core.Bundle) IProcessor {
	DELTA_UPLOAD_API = ln.Config.ExternalApi.ApiUrl
	REPLICATION_FACTOR = string(ln.Config.Common.ReplicationFactor)
	return &UploadBundleToDeltaProcessor{

		reader,
		bundle,
		Processor{
			LightNode: ln,
		},
	}
}

func (r *UploadBundleToDeltaProcessor) Info() error {
	panic("implement me")
}

func (r *UploadBundleToDeltaProcessor) Run() error {

	// if network connection is not available or delta node is not available, then we need to skip and
	// let the upload retry consolidate the content until it is available

	maxRetries := 5
	retryInterval := 5 * time.Second

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	partFile, err := writer.CreateFormFile("data", r.Bundle.Cid)
	if err != nil {
		fmt.Println("CreateFormFile error: ", err)
		return nil
	}
	cidToGet, err := cid.Decode(r.Bundle.Cid)
	if err != nil {
		fmt.Println("Error decoding cid: ", err)
		return nil
	}

	rootNd, err := r.LightNode.Node.GetFile(context.Background(), cidToGet)
	if err != nil {
		fmt.Println("Error getting root node: ", err)
		return nil
	}

	r.Bundle.Status = "uploading"
	r.LightNode.DB.Save(&r.Bundle)

	//rootNd.WriteTo(partFile)
	_, err = io.Copy(partFile, rootNd)
	if err != nil {
		fmt.Println("Copy error: ", err)
		return nil
	}

	repFactor := r.LightNode.Config.Common.ReplicationFactor
	partMetadata := fmt.Sprintf(`{"auto_retry":true,"miner":"%s","replication":%d}`, r.Bundle.Miner, repFactor)
	if partFile, err = writer.CreateFormField("metadata"); err != nil {
		fmt.Println("CreateFormField error: ", err)
		return nil
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
	req.Header.Set("Authorization", "Bearer "+r.Bundle.RequestingApiKey)
	client := &http.Client{}
	var res *http.Response
	for j := 0; j < maxRetries; j++ {
		// Create a new http.Request instance
		clonedReq := &http.Request{}

		// Copy the properties from the original request
		*clonedReq = *req

		// Copy the headers
		clonedReq.Header = make(http.Header)
		for k, v := range req.Header {
			clonedReq.Header[k] = append([]string(nil), v...)
		}
		res, err = client.Do(clonedReq)
		if err != nil || res.StatusCode != http.StatusOK {
			fmt.Println("Error uploading car to delta: ", err)
			r.Bundle.Status = "error"
			r.Bundle.LastMessage = err.Error()
			r.LightNode.DB.Save(&r.Bundle)
			time.Sleep(retryInterval)
			continue
		} else {
			if res.StatusCode == 200 {
				var dealE2EUploadResponse DealE2EUploadResponse
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					r.Bundle.Status = "error"
					r.Bundle.LastMessage = err.Error()
					r.LightNode.DB.Save(&r.Bundle)
					fmt.Println(err)
					continue
				}
				err = json.Unmarshal(body, &dealE2EUploadResponse)
				if err != nil {
					r.Bundle.Status = "error"
					r.Bundle.LastMessage = err.Error()
					r.LightNode.DB.Save(&r.Bundle)
					continue
				} else {
					if dealE2EUploadResponse.ContentID == 0 {
						continue
					} else {
						r.Bundle.UpdatedAt = time.Now()
						r.Bundle.LastMessage = utils.STATUS_UPLOADED_TO_DELTA
						r.Bundle.Status = utils.STATUS_UPLOADED_TO_DELTA
						r.Bundle.DeltaContentId = int64(dealE2EUploadResponse.ContentID)
						r.LightNode.DB.Save(&r.Bundle)
						break
					}
				}
			}
		}
	}
	// get all bucket and update status to bundled.
	var buckets []core.Bucket
	r.LightNode.DB.Where("bundle_uuid = ?", r.Bundle.Uuid).Find(&buckets)
	for _, bucket := range buckets {
		bucket.Status = "bundled"
		r.LightNode.DB.Save(&bucket)
	}
	return nil
}
