package api

import (
	"context"
	"github.com/application-research/edge-ur/jobs"
	"github.com/ipfs/go-cid"
	"strconv"
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
	e.GET("/status/content/:id", func(c echo.Context) error {

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
		job.AddJob(jobs.NewDealItemChecker(node, content))
		job.Start(1)

		return c.JSON(200, map[string]interface{}{
			"content": content,
		})
	})
	e.GET("/status/bucket/contents/:uuid", func(c echo.Context) error {

		var bucket core.Bucket
		node.DB.Model(&core.Bucket{}).Where("uuid = ?", c.Param("uuid")).Scan(&bucket)

		var contents []core.Content
		node.DB.Model(&core.Content{}).Where("bucket_uuid = ?", c.Param("uuid")).Scan(&contents)

		var contentResponse []core.Content
		for _, content := range contents {
			content.RequestingApiKey = ""
			contentResponse = append(contentResponse, content)
			job := jobs.CreateNewDispatcher()
			job.AddJob(jobs.NewDealItemChecker(node, content))

		}

		// trigger status check of the bucket
		job := jobs.CreateNewDispatcher()
		job.AddJob(jobs.NewCarDealItemChecker(node, bucket))
		job.Start(len(contents) + 1)

		bucket.RequestingApiKey = ""
		return c.JSON(200, map[string]interface{}{
			"bucket":          bucket,
			"content_entries": contentResponse,
		})
	})
	e.GET("/status/bucket/dag/:uuid", func(c echo.Context) error {

		var bucket core.Bucket
		node.DB.Model(&core.Bucket{}).Where("uuid = ?", c.Param("uuid")).Scan(&bucket)

		// get the cid
		bucketCid, err := cid.Decode(bucket.Cid)
		if err != nil {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket not found. Please check if you have the proper API key or if the bucket uuid is valid",
			})
		}
		dirNdRaw, err := node.Node.DAGService.Get(context.Background(), bucketCid)
		if err != nil {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket not found. Please check if you have the proper API key or if the bucket id is valid",
			})
		}

		var contents []core.Content
		node.DB.Model(&core.Content{}).Where("bucket_uuid = ?", c.Param("uuid")).Scan(&contents)

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
			job := jobs.CreateNewDispatcher()
			job.AddJob(jobs.NewDealItemChecker(node, content))

		}
		// trigger status check
		job := jobs.CreateNewDispatcher()
		job.AddJob(jobs.NewCarDealItemChecker(node, bucket))
		job.Start(len(contents) + 1)

		bucket.RequestingApiKey = ""
		return c.JSON(200, map[string]interface{}{
			"bucket":        bucket,
			"content_links": contentResponse,
		})
	})
	e.GET("/list", func(c echo.Context) error {

		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

		// get page number
		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil {
			page = 1
		}

		// get page size
		pageSize, err := strconv.Atoi(c.QueryParam("page_size"))
		if err != nil {
			pageSize = 10
		}

		var totalContents int64
		node.DB.Raw("select count(*) from contents where requesting_api_key = ?", authParts[1]).Scan(&totalContents)

		var contents []core.Content
		offset := (page - 1) * pageSize

		// Execute query with LIMIT and OFFSET clauses for paging
		err = node.DB.Select("name, id, delta_content_id, cid, status, last_message, created_at, updated_at").
			Where("requesting_api_key = ?", authParts[1]).
			Order("created_at DESC").
			Limit(pageSize).
			Offset(offset).
			Find(&contents).Error

		if err != nil {
			return c.JSON(500, map[string]interface{}{
				"message": "Error while fetching contents",
				"error":   err.Error(),
			})
		}

		return c.JSON(200, map[string]interface{}{
			"total":    totalContents,
			"page":     page,
			"contents": contents,
		})

	})
}
