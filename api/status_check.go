package api

import (
	"edge-ur/core"
	"github.com/labstack/echo/v4"
	"strings"
)

type StatusCheckResponse struct {
	Content struct {
		ID               uint   `json:"id"`
		EstuaryContentId string `json:"estuary_content_id,omitempty"`
		Status           string `json:"status"`
		Message          string `json:"message,omitempty"`
	} `json:"content"`
}

func ConfigureStatusCheckRouter(e *echo.Group, node *core.LightNode) {
	e.GET("/status/:id", func(c echo.Context) error {

		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

		var content core.Content
		node.DB.Raw("select c.id, c.estuary_content_id, c.status from contents as c, buckets as b where c.bucket_uuid = b.uuid and c.id = ? and requesting_api_key = ?", c.Param("id"), authParts[1]).Scan(&content)

		return c.JSON(200, StatusCheckResponse{
			Content: struct {
				ID               uint   `json:"id"`
				EstuaryContentId string `json:"estuary_content_id,omitempty"`
				Status           string `json:"status"`
				Message          string `json:"message,omitempty"`
			}{ID: content.ID, EstuaryContentId: content.EstuaryContentId, Status: content.Status},
		})

		return nil
	})
}
