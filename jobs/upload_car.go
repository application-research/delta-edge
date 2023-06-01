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

type UploadCarToDeltaProcessor struct {
	BucketUuid string `json:"bucket_uuid"`
	Processor
}

func NewUploadCarToDeltaProcessor(ln *core.LightNode, bucketUuid string) IProcessor {
	DELTA_UPLOAD_API = ln.Config.ExternalApi.DeltaNodeApiUrl
	REPLICATION_FACTOR = string(ln.Config.Common.ReplicationFactor)
	return &UploadCarToDeltaProcessor{
		bucketUuid,
		Processor{
			LightNode: ln,
		},
	}
}

func (r *UploadCarToDeltaProcessor) Info() error {
	panic("implement me")
}

func (r *UploadCarToDeltaProcessor) Run() error {

	// if network connection is not available or delta node is not available, then we need to skip and
	// let the upload retry consolidate the content until it is available
	var bucket core.Bucket
	r.LightNode.DB.Model(&core.Bucket{}).Where("uuid = ?", r.BucketUuid).First(&bucket)

	maxRetries := 5
	retryInterval := 5 * time.Second

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	partFile, err := writer.CreateFormFile("data", bucket.Cid)
	if err != nil {
		fmt.Println("CreateFormFile error: ", err)
		return nil
	}
	cidToGet, err := cid.Decode(bucket.Cid)
	if err != nil {
		fmt.Println("Error decoding cid: ", err)
		return nil
	}

	rootNd, err := r.LightNode.Node.GetFile(context.Background(), cidToGet)
	if err != nil {
		fmt.Println("Error getting root node: ", err)
		return nil
	}

	bucket.Status = "uploading"
	r.LightNode.DB.Save(&bucket)

	//rootNd.WriteTo(partFile)
	_, err = io.Copy(partFile, rootNd)
	if err != nil {
		fmt.Println("Copy error: ", err)
		return nil
	}

	repFactor := r.LightNode.Config.Common.ReplicationFactor
	fmt.Println("Replication factor: ", repFactor)
	fmt.Println("Piece cid: ", bucket.PieceCid)
	fmt.Println("Piece size: ", bucket.PieceSize)
	fmt.Println("Miner: ", bucket.Miner)

	//partMetadata := fmt.Sprintf(`{"auto_retry":true,"size":%d,"miner":"%s","replication":%d,"piece_commitment":{"piece_cid":"%s","padded_piece_size":%d}}`, bucket.Size, bucket.Miner, repFactor, bucket.FilCPieceCid, bucket.FilCPieceSize)
	partMetadata := fmt.Sprintf(`{"auto_retry":true,"size":%d,"miner":"%s","replication":%d}`, bucket.Size, bucket.Miner, repFactor)

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
	req.Header.Set("Authorization", "Bearer "+bucket.RequestingApiKey)
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
			bucket.Status = "error"
			if err != nil {
				bucket.LastMessage = err.Error()
				continue
			}
			body, errorRead := ioutil.ReadAll(res.Body)
			if errorRead != nil {
				bucket.LastMessage = errorRead.Error()
				continue
			}
			// close response body
			res.Body.Close()

			// print response body
			fmt.Println(string(body))
			bucket.LastMessage = string(body)
			r.LightNode.DB.Save(&bucket)
			time.Sleep(retryInterval)
			continue
		} else {
			if res.StatusCode == 200 {
				var dealE2EUploadResponse DealE2EUploadResponse
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					bucket.Status = "error"
					bucket.LastMessage = err.Error()
					r.LightNode.DB.Save(&bucket)
					fmt.Println(err)
					continue
				}
				err = json.Unmarshal(body, &dealE2EUploadResponse)
				if err != nil {
					bucket.Status = "error"
					bucket.LastMessage = err.Error()
					r.LightNode.DB.Save(&bucket)
					continue
				} else {
					if dealE2EUploadResponse.ContentID == 0 {
						continue
					} else {
						bucket.UpdatedAt = time.Now()
						bucket.LastMessage = utils.STATUS_UPLOADED_TO_DELTA
						bucket.Status = utils.STATUS_UPLOADED_TO_DELTA
						bucket.DeltaContentId = int64(dealE2EUploadResponse.ContentID)
						r.LightNode.DB.Save(&bucket)
						break
					}
				}
			}
		}
	}

	return nil
}
