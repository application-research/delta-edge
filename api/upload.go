package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/application-research/edge-ur/jobs"
	"github.com/application-research/edge-ur/utils"
	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car"
	"strings"
	"time"

	"github.com/application-research/edge-ur/core"
	"github.com/labstack/echo/v4"
)

type CidRequest struct {
	Cids []string `json:"cids"`
}

type DealE2EUploadRequest struct {
	Cid                string `json:"cid,omitempty"`
	Miner              string `json:"miner,omitempty"`
	Duration           int64  `json:"duration,omitempty"`
	DurationInDays     int64  `json:"duration_in_days,omitempty"`
	ConnectionMode     string `json:"connection_mode,omitempty"`
	Size               int64  `json:"size,omitempty"`
	StartEpoch         int64  `json:"start_epoch,omitempty"`
	StartEpochInDays   int64  `json:"start_epoch_in_days,omitempty"`
	RemoveUnsealedCopy bool   `json:"remove_unsealed_copy"`
	SkipIPNIAnnounce   bool   `json:"skip_ipni_announce"`
	AutoRetry          bool   `json:"auto_retry"`
	Label              string `json:"label,omitempty"`
	DealVerifyState    string `json:"deal_verify_state,omitempty"`
}

// DealE2EUploadResponse DealResponse Creating a new struct called DealResponse and then returning it.
type DealE2EUploadResponse struct {
	Status                       string      `json:"status"`
	Message                      string      `json:"message"`
	ContentId                    int64       `json:"content_id,omitempty"`
	DealRequest                  interface{} `json:"deal_request_meta,omitempty"`
	DealProposalParameterRequest interface{} `json:"deal_proposal_parameter_request_meta,omitempty"`
	ReplicatedContents           interface{} `json:"replicated_contents,omitempty"`
}

type UploadSplitResponse struct {
	Status        string              `json:"status"`
	Message       string              `json:"message"`
	RootCid       string              `json:"rootCid,omitempty"`
	RootContentId int64               `json:"rootContentId,omitempty"`
	Splits        []core.UploadSplits `json:"splits,omitempty"`
}

type UploadResponse struct {
	Status       string      `json:"status"`
	Message      string      `json:"message"`
	ID           int64       `json:"id,omitempty"`
	Cid          string      `json:"cid,omitempty"`
	DeltaContent interface{} `json:"delta_content,omitempty"`
	ContentUrl   string      `json:"content_url,omitempty"`
}

func ConfigurePinningRouter(e *echo.Group, node *core.LightNode) {
	var DeltaUploadApi = node.Config.ExternalApi.ApiUrl
	content := e.Group("/content")
	content.POST("/add", handleUploadToCarBucketAndMiners(node, DeltaUploadApi))
	content.POST("/add-car", handlePinAddCarToNodeToMiners(node, DeltaUploadApi))
	content.POST("/fetch-pin", handleFetchPinToNodeToMiners(node, DeltaUploadApi)) // foreign cids

}

// upload a file
// check an open bucket, if none create one, if there's open get the ID
// save content with the bucket ID
// run bucket upload checker.

func handleUploadToCarBucketAndMiners(node *core.LightNode, DeltaUploadApi string) func(c echo.Context) error {
	return func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")
		minersString := c.FormValue("miners") // comma-separated list of miners to pin to
		makeDeal := c.FormValue("make_deal")  // whether to make a deal with the miners or not

		if makeDeal == "" {
			makeDeal = "true"
		}

		miners := make(map[string]bool)
		for _, miner := range strings.Split(minersString, ",") {
			miners[miner] = true
		}

		file, err := c.FormFile("data")
		if err != nil {
			return err
		}
		src, err := file.Open()
		srcR, err := file.Open()
		if err != nil {
			return err
		}

		addNode, err := node.Node.AddPinFile(c.Request().Context(), src, nil)
		if err != nil {
			return c.JSON(500, UploadResponse{
				Status:  "error",
				Message: "Error adding the file to IPFS",
			})
		}

		// check open bucket
		var contentList []core.Content

		for miner := range miners {
			if file.Size > node.Config.Common.AggregateSize {
				newContent := core.Content{
					Name:             file.Filename,
					Size:             file.Size,
					Cid:              addNode.Cid().String(),
					DeltaNodeUrl:     DeltaUploadApi,
					RequestingApiKey: authParts[1],
					Status:           utils.STATUS_PINNED,
					Miner:            miner,
					MakeDeal: func() bool {
						if makeDeal == "true" {
							return true
						}
						return false
					}(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				node.DB.Create(&newContent)

				if makeDeal == "true" {
					job := jobs.CreateNewDispatcher()
					job.AddJob(jobs.NewUploadToDeltaProcessor(node, newContent, srcR))
					job.Start(1)
				}

				if err != nil {
					c.JSON(500, UploadResponse{
						Status:  "error",
						Message: "Error pinning the file" + err.Error(),
					})
				}
				newContent.RequestingApiKey = ""
				contentList = append(contentList, newContent)
			} else if file.Size > node.Config.Common.AggregateSize && file.Size < node.Config.Common.MaxSizeToSplit {
				var bucket core.Bucket
				node.DB.Where("status = ? and miner = ?", "open", miner).First(&bucket)
				if bucket.ID == 0 {
					// create a new bucket
					bucketUuid, err := uuid.NewUUID()
					if err != nil {
						return c.JSON(500, UploadResponse{
							Status:  "error",
							Message: "Error creating bucket",
						})
					}
					bucket = core.Bucket{
						Status:           "open",
						Name:             bucketUuid.String(),
						RequestingApiKey: authParts[1],
						Uuid:             bucketUuid.String(),
						Miner:            miner, // blank
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}
					node.DB.Create(&bucket)
				}

				newContent := core.Content{
					Name:             file.Filename,
					Size:             file.Size,
					Cid:              addNode.Cid().String(),
					DeltaNodeUrl:     DeltaUploadApi,
					RequestingApiKey: authParts[1],
					Status:           utils.STATUS_PINNED,
					Miner:            miner,
					BucketUuid:       bucket.Uuid,
					MakeDeal: func() bool {
						if makeDeal == "true" {
							return true
						}
						return false
					}(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				node.DB.Create(&newContent)

				if makeDeal == "true" {
					job := jobs.CreateNewDispatcher()
					job.AddJob(jobs.NewAggregateProcessor(node, newContent, srcR))
					job.Start(1)
				}

				if err != nil {
					c.JSON(500, UploadResponse{
						Status:  "error",
						Message: "Error pinning the file" + err.Error(),
					})
				}
				newContent.RequestingApiKey = ""
				contentList = append(contentList, newContent)
			} else if file.Size > node.Config.Common.MaxSizeToSplit {
				newContent := core.Content{
					Name:             file.Filename,
					Size:             file.Size,
					Cid:              addNode.Cid().String(),
					DeltaNodeUrl:     DeltaUploadApi,
					RequestingApiKey: authParts[1],
					Status:           utils.STATUS_PINNED,
					Miner:            miner,
					MakeDeal: func() bool {
						if makeDeal == "true" {
							return true
						}
						return false
					}(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				node.DB.Create(&newContent)

				if makeDeal == "true" {
					job := jobs.CreateNewDispatcher()
					job.AddJob(jobs.NewSplitterProcessor(node, newContent, srcR))
					job.Start(1)
				}

				if err != nil {
					c.JSON(500, UploadResponse{
						Status:  "error",
						Message: "Error pinning the file" + err.Error(),
					})
				}
				newContent.RequestingApiKey = ""
				contentList = append(contentList, newContent)
			}
		}

		c.JSON(200, struct {
			Status   string         `json:"status"`
			Message  string         `json:"message"`
			Contents []core.Content `json:"contents"`
		}{
			Status:   "success",
			Message:  "File uploaded and pinned successfully to miners. Please take note of the ids.",
			Contents: contentList,
		})

		return nil
	}
}
func handlePinAddToNodeToMiners(node *core.LightNode, DeltaUploadApi string) func(c echo.Context) error {
	return func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")
		minersString := c.FormValue("miners") // comma-separated list of miners to pin to
		makeDeal := c.FormValue("make_deal")  // whether to make a deal with the miners or not

		if makeDeal == "" {
			makeDeal = "true"
		}

		miners := make(map[string]bool)
		for _, miner := range strings.Split(minersString, ",") {
			miners[miner] = true
		}

		file, err := c.FormFile("data")
		if err != nil {
			return err
		}
		src, err := file.Open()
		srcR, err := file.Open()
		if err != nil {
			return err
		}

		addNode, err := node.Node.AddPinFile(c.Request().Context(), src, nil)
		if err != nil {
			return c.JSON(500, UploadResponse{
				Status:  "error",
				Message: "Error adding the file to IPFS",
			})
		}

		var contentList []core.Content

		for miner := range miners {
			newContent := core.Content{
				Name:             file.Filename,
				Size:             file.Size,
				Cid:              addNode.Cid().String(),
				DeltaNodeUrl:     DeltaUploadApi,
				RequestingApiKey: authParts[1],
				Status:           utils.STATUS_PINNED,
				Miner:            miner,
				MakeDeal: func() bool {
					if makeDeal == "true" {
						return true
					}
					return false
				}(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			node.DB.Create(&newContent)

			if makeDeal == "true" {
				job := jobs.CreateNewDispatcher()
				job.AddJob(jobs.NewUploadToDeltaProcessor(node, newContent, srcR))
				job.Start(1)
			}

			if err != nil {
				c.JSON(500, UploadResponse{
					Status:  "error",
					Message: "Error pinning the file" + err.Error(),
				})
			}
			newContent.RequestingApiKey = ""
			contentList = append(contentList, newContent)
		}

		c.JSON(200, struct {
			Status   string         `json:"status"`
			Message  string         `json:"message"`
			Contents []core.Content `json:"contents"`
		}{
			Status:   "success",
			Message:  "File uploaded and pinned successfully to miners. Please take note of the ids.",
			Contents: contentList,
		})

		return nil
	}
}
func handleFetchPinToNodeToMiners(node *core.LightNode, DeltaUploadApi string) func(c echo.Context) error {
	return func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")
		cidToFetch := c.FormValue("cid")
		minersString := c.FormValue("miners")
		makeDeal := c.FormValue("make_deal")

		if makeDeal == "" {
			makeDeal = "true"
		}
		if minersString == "" {
			return c.JSON(500, UploadResponse{
				Status:  "error",
				Message: "Missing miners",
			})
		}

		miners := make(map[string]bool)

		for _, miner := range strings.Split(minersString, ",") {
			miners[miner] = true
		}

		source := c.FormValue("source") // multiaddress where to pull. This can be empty.

		// if the source is given, peer to it. (optional but recommended)
		node.ConnectToDelegates(context.Background(), []string{source}) // connects to the specified IPFS node multiaddress
		cidToFetchCid, err := cid.Decode(cidToFetch)
		if err != nil {
			return c.JSON(500, UploadResponse{
				Status:  "error",
				Message: "Invalid CID",
			})
		}

		addNode, err := node.Node.Get(c.Request().Context(), cidToFetchCid)
		if err != nil {
			return c.JSON(500, UploadResponse{
				Status:  "error",
				Message: "Error getting the file from the source or network",
			})
		}
		addNodeSize, err := addNode.Size()
		if err != nil {
			return c.JSON(500, UploadResponse{
				Status:  "error",
				Message: "Error getting the file size",
			})
		}

		fmt.Println("addNode: ", addNode.Cid().String())

		var contentList []core.Content
		for miner := range miners {
			newContent := core.Content{
				Name:             addNode.Cid().String(),
				Size:             int64(addNodeSize),
				Cid:              addNode.Cid().String(),
				DeltaNodeUrl:     DeltaUploadApi,
				RequestingApiKey: authParts[1],
				Status:           "fetching",
				Miner:            miner,
				MakeDeal: func() bool {
					if makeDeal == "true" {
						return true
					}
					return false
				}(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			node.DB.Create(&newContent)

			if makeDeal == "true" {
				srcR := bytes.NewReader(addNode.RawData())
				job := jobs.CreateNewDispatcher()
				job.AddJob(jobs.NewUploadToDeltaProcessor(node, newContent, srcR))
				job.Start(1)
			}

			if err != nil {
				c.JSON(500, UploadResponse{
					Status:  "error",
					Message: "Error pinning the file" + err.Error(),
				})
			}
			newContent.RequestingApiKey = ""
			contentList = append(contentList, newContent)
		}

		c.JSON(200, struct {
			Status   string         `json:"status"`
			Message  string         `json:"message"`
			Contents []core.Content `json:"contents"`
		}{
			Status:   "success",
			Message:  "File uploaded and pinned successfully to miners. Please take note of the ids.",
			Contents: contentList,
		})

		return nil
	}
}
func handlePinAddCarToNodeToMiners(node *core.LightNode, DeltaUploadApi string) func(c echo.Context) error {
	return func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

		file, err := c.FormFile("data")
		minersString := c.FormValue("miners") // comma-separated list of miners to pin to
		makeDeal := c.FormValue("make_deal")  // whether to make a deal with the miners or not
		if makeDeal == "" {
			makeDeal = "true"
		}
		miners := make(map[string]bool)
		for _, miner := range strings.Split(minersString, ",") {
			miners[miner] = true
		}
		if err != nil {
			c.JSON(500, UploadResponse{
				Status:  "error",
				Message: "Error pinning the car file:" + err.Error(),
			})
		}

		src, err := file.Open()
		srcR := src

		if err != nil {
			c.JSON(500, UploadResponse{
				Status:  "error",
				Message: "Error pinning the car file:" + err.Error(),
			})
		}

		carHeader, err := car.LoadCar(context.Background(), node.Node.Blockstore, src)
		if err != nil {
			c.JSON(500, UploadResponse{
				Status:  "error",
				Message: "Error loading car file: " + err.Error(),
			})
		}

		rootCid := carHeader.Roots[0].String()
		var contentList []core.Content
		for miner := range miners {
			newContent := core.Content{
				Name:             file.Filename,
				Size:             file.Size,
				Cid:              rootCid,
				DeltaNodeUrl:     DeltaUploadApi,
				RequestingApiKey: authParts[1],
				Status:           utils.STATUS_PINNED,
				MakeDeal: func() bool {
					if makeDeal == "true" {
						return true
					}
					return false
				}(),
				Miner:     miner,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			node.DB.Create(&newContent)

			if makeDeal == "true" {
				job := jobs.CreateNewDispatcher()
				job.AddJob(jobs.NewUploadToDeltaProcessor(node, newContent, srcR))
				job.Start(1)
			}

			if err != nil {
				c.JSON(500, UploadResponse{
					Status:  "error",
					Message: "Error pinning the file" + err.Error(),
				})
			}

			contentList = append(contentList, newContent)
		}

		c.JSON(200, struct {
			Status   string         `json:"status"`
			Message  string         `json:"message"`
			Contents []core.Content `json:"contents"`
		}{
			Status:   "success",
			Message:  "File uploaded and pinned successfully to miners. Please take note of the ids.",
			Contents: contentList,
		})

		return nil
	}
}
