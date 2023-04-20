package api

import (
	"context"
	"fmt"
	"github.com/application-research/edge-ur/jobs"
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
	Replication        int    `json:"replication,omitempty"`
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
	var DeltaUploadApi = node.Config.Delta.ApiUrl
	content := e.Group("/content")
	content.POST("/add", handlePinAddToNode(node, DeltaUploadApi))
	content.POST("/add-large", handlePinAddToNodeLarge(node, DeltaUploadApi))
	content.POST("/add-car", handlePinAddCarToNode(node, DeltaUploadApi))
}

func handlePinAddToNodeLarge(node *core.LightNode, DeltaUploadApi string) func(c echo.Context) error {
	return func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")
		fmt.Println("authParts: ", authParts[1])
		return nil
	}
}

func handlePinAddToNode(node *core.LightNode, DeltaUploadApi string) func(c echo.Context) error {
	return func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

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
		fmt.Println("addNode: ", addNode.Cid().String())
		newContent := core.Content{
			Name:             file.Filename,
			Size:             file.Size,
			Cid:              addNode.Cid().String(),
			DeltaNodeUrl:     DeltaUploadApi,
			RequestingApiKey: authParts[1],
			Status:           "pinned",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		node.DB.Create(&newContent)

		job := jobs.CreateNewDispatcher()
		job.AddJob(jobs.NewUploadToEstuaryProcessor(node, newContent, srcR))
		job.Start(1)

		if err != nil {
			c.JSON(500, UploadResponse{
				Status:  "error",
				Message: "Error pinning the file" + err.Error(),
			})
		}

		c.JSON(200, UploadResponse{
			Status:  "success",
			Message: "File uploaded and pinned successfully. Please take note of the id.",
			ID:      newContent.ID,
			Cid:     newContent.Cid,
		})
		return nil
	}
}
func handlePinAddCarToNode(node *core.LightNode, DeltaUploadApi string) func(c echo.Context) error {
	return func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

		file, err := c.FormFile("data")
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

		// insert a new record
		newContent := core.Content{
			Name:             file.Filename,
			Size:             file.Size,
			Cid:              rootCid,
			DeltaNodeUrl:     DeltaUploadApi,
			RequestingApiKey: authParts[1],
			Status:           "pinned",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		node.DB.Create(&newContent)

		job := jobs.CreateNewDispatcher()
		job.AddJob(jobs.NewUploadToEstuaryProcessor(node, newContent, srcR))
		job.Start(1)

		return nil
	}
}
