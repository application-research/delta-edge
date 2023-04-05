package api

import (
	"github.com/application-research/edge-ur/core"
	"github.com/labstack/echo/v4"
	"net/http"
)

func ConfigureRetrieveRouter(e *echo.Group, node *core.LightNode) {

	e.GET("/retrieve/split", func(c echo.Context) error {
		return RetrieveSplitHandler(c, node)
	})
}

// RetrieveSplitHandler is the handler for the /retrieve/split endpoint
func RetrieveSplitHandler(c echo.Context, node *core.LightNode) error {

	reassembler := core.NewSplitReassembler(struct {
		LightNode *core.LightNode
	}{LightNode: node})

	splitCid := c.QueryParam("split-cid")
	file, err := reassembler.ReassembleFileFromCid(splitCid)
	// split the file.
	if err != nil {
		return err
	}

	writer := c.Response().Writer
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/octet-stream")
	writer.Header().Set("Content-Disposition", "attachment; filename=\""+file.Name()+"\"")
	writer.Header().Set("Content-Length", "1000000000")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("Accept-Ranges", "bytes")
	writer.Write([]byte("hello world"))
	return nil
}
