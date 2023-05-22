package jobs

//
//import (
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"io"
//	"io/ioutil"
//	"math"
//	"mime/multipart"
//	"net/http"
//	"strconv"
//	"time"
//
//	"github.com/application-research/edge-ur/core"
//	"github.com/spf13/viper"
//)
//
//const (
//	maxRetries    = 5
//	retryInterval = 5 * time.Second
//	maxChunkSize  = 5 * 1024 * 1024 * 1024 // 5 GB
//)
//
//type ChunkUploadToEstuaryProcessor struct {
//	Content core.Content `json:"content"`
//	File    io.Reader    `json:"file"`
//	Processor
//}
//
//func NewChunkUploadToEstuaryProcessor(ln *core.LightNode, contentToProcess core.Content, fileNode io.Reader) IProcessor {
//	DELTA_UPLOAD_API = viper.Get("DELTA_NODE_API").(string)
//	REPLICATION_FACTOR = viper.Get("REPLICATION_FACTOR").(string)
//	return &ChunkUploadToEstuaryProcessor{
//		contentToProcess,
//		fileNode,
//		Processor{
//			LightNode: ln,
//		},
//	}
//}
//
//func (r *ChunkUploadToEstuaryProcessor) Info() error {
//	panic("implement me")
//}
//
//func (r *ChunkUploadToEstuaryProcessor) Run() error {
//	// Get the total size of the file
//	fileInfo, err := r.File.Stat()
//	if err != nil {
//		fmt.Println(err)
//		return nil
//	}
//	fileSize := fileInfo.Size()
//
//	// Calculate the number of chunks needed
//	numChunks := int(fileSize/maxChunkSize) + 1
//
//	// Create a buffer to store the chunk data
//	buffer := make([]byte, maxChunkSize)
//
//	// Read and upload each chunk
//	var chunkIndex int64
//	for i := 0; i < numChunks; i++ {
//		// Open a new file reader
//		fileReader := io.LimitReader(r.File, maxChunkSize)
//
//		// Read the chunk data into the buffer
//		_, err = io.ReadFull(fileReader, buffer)
//		if err != nil {
//			fmt.Println(err)
//			return nil
//		}
//
//		// Create a new multipart payload for the chunk
//		payload := &bytes.Buffer{}
//		writer := multipart.NewWriter(payload)
//
//		// Add the chunk data to the payload
//		partFile, err := writer.CreateFormFile("data", r.Content.Name)
//		if err != nil {
//			fmt.Println("CreateFormFile error: ", err)
//			return nil
//		}
//		_, err = partFile.Write(buffer)
//		if err != nil {
//			fmt.Println("Copy error: ", err)
//			return nil
//		}
//
//		// Add the chunk metadata to the payload
//		if partFile, err = writer.CreateFormField("metadata"); err != nil {
//			fmt.Println("CreateFormField error: ", err)
//		}
//		repFactor, err := strconv.Atoi(REPLICATION_FACTOR)
//		if err != nil {
//			fmt.Println("REPLICATION_FACTOR error: ", err)
//		}
//		partMetadata := fmt.Sprintf(`{"auto_retry":true,"replication":%d}`, repFactor)
//		if repFactor == 0 {
//			partMetadata = fmt.Sprintf(`{"auto_retry":true}`)
//		}
//		if _, err = partFile.Write([]byte(partMetadata)); err != nil {
//			fmt.Println("Write error: ", err)
//			return nil
//		}
//
//		// Close the multipart writer
//		if err = writer.Close(); err != nil {
//			fmt.Println("Close error: ", err)
//			return nil
//		}
//
//		// Create
//		// calculate total number of chunks
//		numChunks := int64(math.Ceil(float64(r.Content.Size) / float64(maxChunkSize)))
//
//		// initialize byte slice to store chunk data
//		chunk := make([]byte, maxChunkSize)
//
//		// loop through the file and read the data in chunks
//		for i := int64(0); i < numChunks; i++ {
//			// read chunk
//			_, err = io.ReadFull(r.File, chunk)
//			if err != nil && err != io.EOF {
//				fmt.Println("Read error: ", err)
//				return nil
//			}
//
//			// create a new buffer to store chunk data
//			chunkBuffer := &bytes.Buffer{}
//			chunkWriter := multipart.NewWriter(chunkBuffer)
//
//			// create form data field for chunk data
//			chunkFileWriter, err := chunkWriter.CreateFormFile("data", r.Content.Name)
//			if err != nil {
//				fmt.Println("CreateFormFile error: ", err)
//				return nil
//			}
//			_, err = chunkFileWriter.Write(chunk)
//			if err != nil {
//				fmt.Println("Write error: ", err)
//				return nil
//			}
//
//			// add metadata to chunk
//			chunkMetadata := fmt.Sprintf(`{"auto_retry":true,"replication":%d}`, repFactor)
//			if repFactor == 0 {
//				chunkMetadata = fmt.Sprintf(`{"auto_retry":true}`)
//			}
//			chunkMetadataWriter, err := chunkWriter.CreateFormField("metadata")
//			if err != nil {
//				fmt.Println("CreateFormField error: ", err)
//				return nil
//			}
//			_, err = chunkMetadataWriter.Write([]byte(chunkMetadata))
//			if err != nil {
//				fmt.Println("Write error: ", err)
//				return nil
//			}
//
//			// close the chunk writer
//			err = chunkWriter.Close()
//			if err != nil {
//				fmt.Println("Close error: ", err)
//				return nil
//			}
//
//			// create a new request for the chunk
//			chunkReq, err := http.NewRequest("POST", DELTA_UPLOAD_API+"/api/v1/deal/end-to-end", chunkBuffer)
//			if err != nil {
//				fmt.Println("NewRequest error: ", err)
//				return nil
//			}
//			chunkReq.Header.Set("Content-Type", chunkWriter.FormDataContentType())
//			chunkReq.Header.Set("Authorization", "Bearer "+r.Content.RequestingApiKey)
//
//			// send the chunk request and retry if necessary
//			for j := 0; j < maxRetries; j++ {
//				chunkRes, err := client.Do(chunkReq)
//				if err != nil || chunkRes.StatusCode != http.StatusOK {
//					fmt.Printf("Error sending chunk request (attempt %d): %v\n", j+1, err)
//					time.Sleep(retryInterval)
//					continue
//				} else {
//					if chunkRes.StatusCode == 200 {
//						var dealE2EUploadResponse DealE2EUploadResponse
//						body, err := ioutil.ReadAll(chunkRes.Body)
//						if err != nil {
//							fmt.Println(err)
//							continue
//						}
//						err = json.Unmarshal(body, &dealE2EUploadResponse)
//						if err != nil {
//							fmt.Println(err)
//							continue
//						} else {
//							if dealE2EUploadResponse.ContentID == 0 {
//								continue
//							} else {
//								r.Content.UpdatedAt = time.Now()
//								r.Content.Status = "uploaded-to-delta"
//								r.Content.DeltaContentId = int64(dealE2EUploadResponse.ContentID)
//								r.LightNode.DB.Save(&r.Content)
//
//								// insert each replicated content into the database
//								for _, replicatedContent := range dealE2EUploadResponse.ReplicatedContents {
//									r.LightNode.DB.Create(&models.Content{
//										Name:           r.Content.Name,
//										Hash:           r.Content.Hash,
//										Size:           r.Content.Size,
//										Replication:    r.Content.Replication,
//										UpdatedAt:      r.Content.UpdatedAt,
//										Status:         r.Content.Status,
//										LightNodeID:    r.Content.LightNodeID,
//										DeltaContentId: int64(replicatedContent.ContentID),
//									})
//								}
//							}
//						}
//					}
//					break
//				}
//			}
//		}
//	}
//
//	return nil
//}
