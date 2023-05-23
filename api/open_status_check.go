package api

import (
	"context"
	"github.com/application-research/edge-ur/core"
	"github.com/application-research/edge-ur/jobs"
	"github.com/ipfs/go-cid"
	"github.com/labstack/echo/v4"
	"strconv"
)

func ConfigureOpenStatusCheckRouter(e *echo.Group, node *core.LightNode) {
	e.GET("/status/content/:id", func(c echo.Context) error {

		var content core.Content
		node.DB.Raw("select * from contents as c where id = ?", c.Param("id")).Scan(&content)
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

	e.GET("/status/bucket/:uuid", func(c echo.Context) error {
		var bucket core.Bucket
		node.DB.Model(&core.Bucket{}).Where("uuid = ?", c.Param("uuid")).Scan(&bucket)
		if bucket.ID == 0 {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket not found. Please check if you have the proper API key or if the bucket uuid is valid",
			})
		}

		bucket.RequestingApiKey = ""
		return c.JSON(200, map[string]interface{}{
			"bucket": bucket,
		})
		return nil
	})

	e.GET("/status/bucket/contents/:uuid", func(c echo.Context) error {
		var bucket core.Bucket
		node.DB.Model(&core.Bucket{}).Where("uuid = ?", c.Param("uuid")).Scan(&bucket)
		if bucket.ID == 0 {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket not found. Please check if you have the proper API key or if the bucket UUID is valid",
			})
		}

		// Get the query parameters for page and perPage
		pageStr := c.QueryParam("page")
		perPageStr := c.QueryParam("per_page")

		// Convert the query parameters to integers
		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			// If the page parameter is invalid, set it to the default value of 1
			page = 1
		}

		perPage, err := strconv.Atoi(perPageStr)
		if err != nil || perPage <= 0 {
			// If the perPage parameter is invalid, set it to the default value of 10
			perPage = 10
		}

		// Retrieve the total count of contents for the bucket
		var totalCount int64
		node.DB.Model(&core.Content{}).Where("bucket_uuid = ?", c.Param("uuid")).Count(&totalCount)

		// Calculate the offset and limit for the current page
		offset := (page - 1) * perPage
		limit := perPage

		// Retrieve the contents with pagination
		var contents []core.Content
		node.DB.Model(&core.Content{}).Where("bucket_uuid = ?", c.Param("uuid")).Offset(offset).Limit(limit).Scan(&contents)

		if len(contents) == 0 {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket has no contents",
			})
		}

		var contentResponse []core.Content
		for _, content := range contents {
			content.RequestingApiKey = ""
			contentResponse = append(contentResponse, content)
			job := jobs.CreateNewDispatcher()
			job.AddJob(jobs.NewDealItemChecker(node, content))
		}

		job := jobs.CreateNewDispatcher()
		job.AddJob(jobs.NewCarDealItemChecker(node, bucket))
		job.Start(len(contents) + 1)

		bucket.RequestingApiKey = ""
		return c.JSON(200, map[string]interface{}{
			"bucket":          bucket,
			"content_entries": contentResponse,
			"pagination": map[string]interface{}{
				"page":       page,
				"perPage":    perPage,
				"totalCount": totalCount,
			},
		})
	})

	e.GET("/status/bucket/dag/:uuid", func(c echo.Context) error {

		var bucket core.Bucket
		node.DB.Model(&core.Bucket{}).Where("uuid = ?", c.Param("uuid")).Scan(&bucket)

		if bucket.ID == 0 {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket not found or has no DAGs. Please check if you have the proper API key or if the bucket uuid is valid",
			})
		}

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

}
