package api

import (
	"edge-ur/core"
	"github.com/labstack/echo/v4"
)

type StatusCheckResponse struct {
	Status  string       `json:"status"`
	Content core.Content `json:"content"`
}

func ConfigureStatusCheckRouter(e *echo.Group, node *core.LightNode) {
	e.GET("/status/:id", func(c echo.Context) error {

		// check status of the bucket of the cid
		//select c.id, b.status from contents as c, buckets as b
		//	where c.bucket_uuid = b.uuid;
		var content core.Content
		node.DB.Raw("select c.id, c.estuary_content_id, b.status from contents as c, buckets as b where c.bucket_uuid = b.uuid and c.id = ?", c.Param("id")).Scan(&content)

		return c.JSON(200, StatusCheckResponse{
			Status: "success",
			Content: core.Content{
				ID:               content.ID,
				EstuaryContentId: content.EstuaryContentId,
				Status:           content.Status,
			},
		})

		return nil
	})
}
