package api

import (
	"edge-ur/core"
	"github.com/labstack/echo/v4"
	"strings"
)

type StatusCheckResponse struct {
	Status  string `json:"status"`
	Content struct {
		ID               uint   `json:"id"`
		EstuaryContentId string `json:"estuary_content_id"`
		Status           string `json:"status"`
		Message          string `json:"message"`
	}
}

func ConfigureStatusCheckRouter(e *echo.Group, node *core.LightNode) {
	e.GET("/status/:id", func(c echo.Context) error {

		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

		var content core.Content
		node.DB.Raw("select c.id, c.estuary_content_id, b.status from contents as c, buckets as b where c.bucket_uuid = b.uuid and c.id = ? and requesting_api_key = ?", c.Param("id"), authParts[1]).Scan(&content)

		return c.JSON(200, StatusCheckResponse{
			Status: "success",
			Content: struct {
				ID               uint   `json:"id"`
				EstuaryContentId string `json:"estuary_content_id"`
				Status           string `json:"status"`
				Message          string `json:"message"`
			}{ID: content.ID, EstuaryContentId: content.EstuaryContentId, Status: content.Status, Message: "Content is now being processed by Estuary."},
		})

		return nil
	})
}
