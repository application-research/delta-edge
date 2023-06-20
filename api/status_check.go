package api

import (
	"context"
	"github.com/application-research/edge-ur/jobs"
	"github.com/ipfs/go-cid"
	"strings"

	"github.com/application-research/edge-ur/core"
	"github.com/labstack/echo/v4"
)

type StatusCheckResponse struct {
	Content struct {
		ID             int64  `json:"id"`
		Name           string `json:"name"`
		DeltaContentId int64  `json:"delta_content_id,omitempty"`
		Cid            string `json:"cid,omitempty"`
		Status         string `json:"status"`
		Message        string `json:"message,omitempty"`
	} `json:"content"`
}

func ConfigureStatusCheckRouter(e *echo.Group, node *core.LightNode) {
	e.GET("/status/cid/:id", func(c echo.Context) error {

		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

		var content core.Content
		node.DB.Raw("select * from contents as c where requesting_api_key = ? and id = ?", authParts[1], c.Param("id")).Scan(&content)
		content.RequestingApiKey = ""

		if content.ID == 0 {
			return c.JSON(404, map[string]interface{}{
				"message": "Content not found. Please check if you have the proper API key or if the content id is valid",
			})
		}

		// trigger status check
		job := jobs.CreateNewDispatcher()
		//job.AddJob(jobs.NewDealItemChecker(node, content))
		job.Start(1)

		return c.JSON(200, map[string]interface{}{
			"content": content,
		})
	})
	e.GET("/status/content/:contentId", func(c echo.Context) error {

		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

		var content core.Content
		node.DB.Raw("select * from contents as c where requesting_api_key = ? and id = ?", authParts[1], c.Param("id")).Scan(&content)
		content.RequestingApiKey = ""

		if content.ID == 0 {
			return c.JSON(404, map[string]interface{}{
				"message": "Content not found. Please check if you have the proper API key or if the content id is valid",
			})
		}

		// trigger status check
		job := jobs.CreateNewDispatcher()
		//job.AddJob(jobs.NewDealItemChecker(node, content))
		job.Start(1)

		return c.JSON(200, map[string]interface{}{
			"content": content,
		})
	})
	e.GET("/status/bucket/:bucketUuid", func(c echo.Context) error {

		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

		var bucket core.Bucket
		node.DB.Model(&core.Bucket{}).Where("requesting_api_key = ? and uuid = ?", authParts[1], c.Param("uuid")).Scan(&bucket)

		// get the cid
		bucketCid, err := cid.Decode(bucket.Cid)
		if err != nil {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket not found. Please check if you have the proper API key or if the bucket id is valid",
			})
		}
		dirNdRaw, err := node.Node.DAGService.Get(context.Background(), bucketCid)
		if err != nil {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket not found. Please check if you have the proper API key or if the bucket id is valid",
			})
		}

		var contents []core.Content
		node.DB.Model(&core.Content{}).Where("requesting_api_key = ? and bucket_uuid = ?", authParts[1], c.Param("uuid")).Scan(&contents)

		var contentResponse []core.Content
		for _, content := range contents {
			content.RequestingApiKey = ""

			for _, link := range dirNdRaw.Links() {
				cidFromDb, err := cid.Decode(content.Cid)
				if err != nil {
					continue
				}
				if link.Cid == cidFromDb {
					content.Cid = link.Cid.String()
					content.RequestingApiKey = ""
					contentResponse = append(contentResponse, content)
				}

			}
			//job := jobs.CreateNewDispatcher()
			//job.AddJob(jobs.NewDealItemChecker(node, content))

		}
		// trigger status check
		job := jobs.CreateNewDispatcher()
		//job.AddJob(jobs.NewCarDealItemChecker(node, bucket))
		job.Start(len(contents) + 1)

		bucket.RequestingApiKey = ""
		return c.JSON(200, map[string]interface{}{
			"bucket":        bucket,
			"content_links": contentResponse,
		})
	})
}
