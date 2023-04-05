package api

import (
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
	e.GET("/status/:id", func(c echo.Context) error {

		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

		var content core.Content
		node.DB.Raw("select * from contents as c where requesting_api_key = ? and id = ?", authParts[1], c.Param("id")).Scan(&content)
		content.RequestingApiKey = ""
		return c.JSON(200, map[string]interface{}{
			"content": content,
		})
	})

	e.GET("/list", func(c echo.Context) error {

		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")

		var content []core.Content
		node.DB.Raw("select c.name, c.id, c.estuary_content_id, c.cid, c.status,c.created_at,c.updated_at from contents as c where requesting_api_key = ?", authParts[1]).Scan(&content)

		return c.JSON(200, content)

	})
}
