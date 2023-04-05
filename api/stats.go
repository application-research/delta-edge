package api

import (
	"github.com/application-research/edge-ur/core"
	"github.com/labstack/echo/v4"
)

type Stats struct {
	TotalContentCount int `json:"total_content_count"`
	TotalSize         int `json:"total_size"`
}

func ConfigureStatsRouter(e *echo.Group, node *core.LightNode) {
	e.GET("/stats", func(c echo.Context) error {

		var stats Stats
		node.DB.Raw("select count(*) as total_content_count, sum(size) as total_size from contents").Scan(&stats)

		return c.JSON(200, stats)
		return nil
	})

}
