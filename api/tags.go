package api

import (
	"github.com/application-research/edge-ur/core"
	"github.com/labstack/echo/v4"
)

func ConfigureTagsRouter(e *echo.Group, node *core.LightNode) {
	//var DeltaUploadApi = node.Config.Delta.ApiUrl
	buckets := e.Group("/tags")
	buckets.GET("/create", handleCreateTag(node))
	buckets.POST("/modify", handleGetOpenBuckets(node))
	buckets.DELETE("/remove/:tagName", handleDeleteBucket(node))

}
func handleCreateTag(node *core.LightNode) func(c echo.Context) error {
	return func(c echo.Context) error {

		return nil
	}
}
