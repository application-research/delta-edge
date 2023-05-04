package api

import (
	"context"
	"github.com/application-research/edge-ur/core"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"net/http"
)

// ConfigureNodeInfoRouter It configures the router to handle requests for node information
func ConfigureNodeInfoRouter(e *echo.Group, node *core.LightNode) {

	nodeGroup := e.Group("/node")
	nodeGroup.GET("/info", handleNodeInfo(node))
	//nodeGroup.GET("/addr", handleNodeAddr(node))
	nodeGroup.GET("/peers", handleNodePeers(node))
}

// handleNodeInfo is the handler for the /node/info endpoint
func handleNodeInfo(node *core.LightNode) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, node.Node.Host.ID())
	}
}

func handleNodePeers(node *core.LightNode) func(c echo.Context) error {
	return func(c echo.Context) error {
		_, span := otel.Tracer("handleNodePeers").Start(context.Background(), "handleNodeHostApiKey")
		defer span.End()

		span.SetName("ConfigureNodeInfoRouter")
		span.SetAttributes(attribute.String("user-agent", c.Request().UserAgent()))
		span.SetAttributes(attribute.String("path", c.Path()))
		span.SetAttributes(attribute.String("method", c.Request().Method))
		return c.JSON(200, node.Node.Host.Network().Peers())
	}
}
