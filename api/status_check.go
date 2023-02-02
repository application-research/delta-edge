package api

import (
	"strings"

	"github.com/application-research/edge-ur/core"
	"github.com/labstack/echo/v4"
)

type StatusCheckResponse struct {
	Content struct {
		ID               int64  `json:"id"`
		Name             string `json:"name"`
		EstuaryContentId int64  `json:"estuary_content_id,omitempty"`
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
				ID               int64  `json:"id"`
				Name             string `json:"name"`
				EstuaryContentId int64  `json:"estuary_content_id,omitempty"`
				Status           string `json:"status"`
				Message          string `json:"message,omitempty"`
			}{ID: content.ID, Name: content.Name, EstuaryContentId: content.EstuaryContentId, Status: content.Status},
		})

		return nil
	})

	e.GET("/list-all-cids", func(c echo.Context) error {

		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

		var content core.Content
		node.DB.Raw("select c.id, c.estuary_content_id, c.status from contents as c, buckets as b where c.bucket_uuid = b.uuid and requesting_api_key = ?", authParts[1]).Scan(&content)

		return c.JSON(200, StatusCheckResponse{
			Content: struct {
				ID               int64  `json:"id"`
				Name             string `json:"name"`
				EstuaryContentId int64  `json:"estuary_content_id,omitempty"`
				Status           string `json:"status"`
				Message          string `json:"message,omitempty"`
			}{ID: content.ID, Name: content.Name, EstuaryContentId: content.EstuaryContentId, Status: content.Status},
		})

		return nil
	})
}
