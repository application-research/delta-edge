package api

import (
	"github.com/application-research/edge-ur/core"
	"github.com/labstack/echo/v4"
	"net/http"
)

// ConfigureNodeInfoRouter It configures the router to handle requests for node information
func ConfigureNodeInfoRouter(e *echo.Group, node *core.LightNode) {

	nodeGroup := e.Group("/node")
	nodeGroup.GET("/info", handleNodeInfo(node))
	//nodeGroup.GET("/addr", handleNodeAddr(node))
	//nodeGroup.GET("/peers", handleNodePeers(node))
}

// handleNodeInfo is the handler for the /node/info endpoint
func handleNodeInfo(node *core.LightNode) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, node.Config.APINodeAddress)
	}
}
