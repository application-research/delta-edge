package api

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/application-research/edge-ur/core"
	"github.com/ipfs/go-cid"
	"github.com/labstack/echo/v4"
)

type CidRequest struct {
	Cids []string `json:"cids"`
}

type UploadSplitResponse struct {
	Status        string              `json:"status"`
	Message       string              `json:"message"`
	RootCid       string              `json:"rootCid,omitempty"`
	RootContentId int64               `json:"rootContentId,omitempty"`
	Splits        []core.UploadSplits `json:"splits,omitempty"`
}

type UploadResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	ID      int64  `json:"id,omitempty"`
}

func ConfigurePinningRouter(e *echo.Group, node *core.LightNode) {

	content := e.Group("/content")
	content.POST("/add-car", func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")
		file, err := c.FormFile("data")
		if err != nil {
			return err
		}
		src, err := file.Open()
		if err != nil {
			return err
		}

		addNode, err := node.Node.AddPinFile(c.Request().Context(), src, nil)

		// save the file to the database.
		content := core.Content{
			Name:             file.Filename,
			Size:             file.Size,
			Cid:              addNode.Cid().String(),
			RequestingApiKey: authParts[1],
			Status:           "pinned",
			Created_at:       time.Now(),
			Updated_at:       time.Now(),
		}

		node.DB.Create(&content)

		if err != nil {
			return err
		}

		c.JSON(200, UploadResponse{
			Status:  "success",
			Message: "Car uploaded and pinned successfully to the network.",
			ID:      content.ID,
		})
		return nil
	})

	content.POST("/add-split", func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

		chunkSizeParam := c.FormValue("chunkSize")
		chunkSizeParamInt64, err := strconv.ParseInt(chunkSizeParam, 10, 64)
		if err != nil {
			return err
		}
		splitter := core.NewFileSplitter(struct {
			ChuckSize int64
			LightNode *core.LightNode
		}{ChuckSize: chunkSizeParamInt64, LightNode: node}) // parameterize split

		file, err := c.FormFile("data")
		if err != nil {
			return err
		}
		src, err := file.Open()
		if err != nil {
			return err
		}

		// split the file.
		splitChunk, err := splitter.SplitFileFromReaderIntoBlockstore(src)

		//	 add the json split to whypfs and return it
		splitResult, err := json.Marshal(splitChunk)
		reader := bytes.NewReader(splitResult)
		nodeSplitResult, err := node.Node.AddPinFile(c.Request().Context(), reader, nil)
		if err != nil {
			return err
		}

		// save the root
		rootContent := core.Content{
			Name:             file.Filename,
			Size:             int64(len(splitResult)),
			Cid:              nodeSplitResult.Cid().String(),
			RequestingApiKey: authParts[1],
			Status:           "pinned",
			Created_at:       time.Now(),
			Updated_at:       time.Now(),
		}

		node.DB.Create(&rootContent)

		// create a new content record for each split so we can track them and put them
		// on estuary.
		var uploadSplitResponse UploadSplitResponse
		var uploadSplits []core.UploadSplits
		for _, split := range splitChunk {
			content := core.Content{
				Name:             strconv.Itoa(split.Index),
				Size:             int64(split.Size),
				Cid:              split.Cid,
				RequestingApiKey: authParts[1],
				Status:           "pinned",
				Created_at:       time.Now(),
				Updated_at:       time.Now(),
			}

			node.DB.Create(&content)
			uploadSplits = append(uploadSplits, core.UploadSplits{
				Cid:       split.Cid,
				Index:     split.Index,
				ContentId: content.ID,
			})
		}

		uploadSplitResponse.Status = "success"
		uploadSplitResponse.Message = "File split and pinned successfully to the network."
		uploadSplitResponse.RootCid = nodeSplitResult.Cid().String()
		uploadSplitResponse.RootContentId = rootContent.ID
		uploadSplitResponse.Splits = uploadSplits

		if err != nil {
			return err
		}

		c.JSON(200, uploadSplitResponse)
		return nil
	})

	content.POST("/add", func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")
		file, err := c.FormFile("data")
		if err != nil {
			return err
		}
		src, err := file.Open()
		if err != nil {
			return err
		}

		addNode, err := node.Node.AddPinFile(c.Request().Context(), src, nil)

		// get available staging buckets.
		// save the file to the database.
		content := core.Content{
			Name:             file.Filename,
			Size:             file.Size,
			Cid:              addNode.Cid().String(),
			RequestingApiKey: authParts[1],
			Status:           "pinned",
			Created_at:       time.Now(),
			Updated_at:       time.Now(),
		}

		node.DB.Create(&content)

		if err != nil {
			c.JSON(500, UploadResponse{
				Status:  "error",
				Message: "Error pinning the file" + err.Error(),
			})
		}

		c.JSON(200, UploadResponse{
			Status:  "success",
			Message: "File uploaded and pinned successfully",
			ID:      content.ID,
		})
		return nil
	})

	content.POST("/cid/:cid", func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")
		cidFromForm := c.Param("cid")
		cidNode, err := cid.Decode(cidFromForm)
		if err != nil {
			return err
		}

		//	 get the node
		addNode, err := node.Node.Get(c.Request().Context(), cidNode)

		// get available staging buckets.
		// save the file to the database.
		size, err := addNode.Size()

		content := core.Content{
			Name:             addNode.Cid().String(),
			Size:             int64(size),
			Cid:              addNode.Cid().String(),
			RequestingApiKey: authParts[1],
			Status:           "pinned",
			Created_at:       time.Now(),
			Updated_at:       time.Now(),
		}

		node.DB.Create(&content)

		if err != nil {
			c.JSON(500, UploadResponse{
				Status:  "error",
				Message: "Error pinning the cid" + err.Error(),
			})
		}

		c.JSON(200, UploadResponse{
			Status:  "success",
			Message: "CID uploaded and pinned successfully",
			ID:      content.ID,
		})
		return nil
	})

	content.POST("/cids", func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")
		var cidRequest CidRequest
		c.Bind(&cidRequest)
		for _, cidFromForm := range cidRequest.Cids {
			cidNode, err := cid.Decode(cidFromForm)
			if err != nil {
				return err
			}

			//	 get the node and save on the database
			addNode, err := node.Node.Get(c.Request().Context(), cidNode)

			// get available staging buckets.
			// save the file to the database.
			size, err := addNode.Size()

			content := core.Content{
				Name:             addNode.Cid().String(),
				Size:             int64(size),
				Cid:              addNode.Cid().String(),
				RequestingApiKey: authParts[1],
				Status:           "pinned",
				Created_at:       time.Now(),
				Updated_at:       time.Now(),
			}

			node.DB.Create(&content)
		}
		return nil
	})
}
