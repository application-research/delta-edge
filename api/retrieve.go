package api

import (
	"github.com/application-research/edge-ur/core"
	"github.com/labstack/echo/v4"
)

func ConfigureRetrieveRouter(e *echo.Group, node *core.LightNode) {

	//	api
	gatewayHandler.node = node.Node
	e.GET("/retrieve/split", RetrieveSplitHandler)
}

// RetrieveSplitHandler is the handler for the /retrieve/split endpoint
func RetrieveSplitHandler(e echo.Context) error {

	////authorizationString := c.Request().Header.Get("Authorization")
	////authParts := strings.Split(authorizationString, " ")
	//
	//splitter := core.NewFileSplitter(struct {
	//	ChuckSize int
	//	LightNode *core.LightNode
	//}{ChuckSize: 1024 * 1024, LightNode: node}) // parameterize split
	//
	//splitCid := e.QueryParam("split-cid")
	//
	//// split the file.
	//splitChunk, err := splitter.ReassembleFile(src)
	//
	////	 add the json split to whypfs and return it
	//splitResult, err := json.Marshal(splitChunk)
	//reader := bytes.NewReader(splitResult)
	//nodeSplitResult, err := node.Node.AddPinFile(c.Request().Context(), reader, nil)
	//if err != nil {
	//
	//}
	//c.JSON(200, nodeSplitResult.Cid().String())
	return nil
}
