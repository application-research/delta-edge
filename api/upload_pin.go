package api

import (
	"edge-ur/core"
	"github.com/ipfs/go-cid"
	"github.com/labstack/echo/v4"
	"strings"
	"time"
)

func ConfigurePinningRouter(e *echo.Group, node *core.LightNode) {

	content := e.Group("/content")
	content.POST("/add", func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")
		file, err := c.FormFile("data")
		if err != nil {
			return err
		}
		src, err := file.Open()
		if err != nil {
			return err
		}

		addNode, err := node.Node.AddPinFile(c.Request().Context(), src, nil)

		// get availabel staging buckets.
		// save the file to the database.
		content := core.Content{
			Name:             file.Filename,
			Size:             file.Size,
			Cid:              addNode.Cid().String(),
			RequestingApiKey: authParts[1],
			Created_at:       time.Now(),
			Updated_at:       time.Now(),
		}

		node.DB.Create(&content)

		if err != nil {
			return err
		}
		c.Response().Write([]byte(addNode.Cid().String()))
		return nil
	})

	content.POST("/cid", func(c echo.Context) error {
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")
		cidFromForm := c.FormValue("cid")
		cidNode, err := cid.Decode(cidFromForm)
		if err != nil {
			return err
		}

		//	 get the node
		addNode, err := node.Node.Get(c.Request().Context(), cidNode)

		// get availabel staging buckets.
		// save the file to the database.
		size, err := addNode.Size()

		content := core.Content{
			Name:             addNode.Cid().String(),
			Size:             int64(size),
			Cid:              addNode.Cid().String(),
			RequestingApiKey: authParts[1],
			Created_at:       time.Now(),
			Updated_at:       time.Now(),
		}

		node.DB.Create(&content)
		return nil
	})

	content.POST("/cids", func(c echo.Context) error {
		cids := c.FormValue("cids")

		// process each cids
		cidsArray := strings.Split(cids, ",")
		for _, cidt := range cidsArray {
			cidNode, err := cid.Decode(cidt)
			if err != nil {
				return err
			}

			//	 get the node and save on the database
			addNode, err := node.Node.Get(c.Request().Context(), cidNode)

			// get availabel staging buckets.
			// save the file to the database.
			size, err := addNode.Size()

			content := core.Content{
				Name:       addNode.Cid().String(),
				Size:       int64(size),
				Cid:        addNode.Cid().String(),
				Created_at: time.Now(),
				Updated_at: time.Now(),
			}

			node.DB.Create(&content)
		}
		return nil
	})
}
