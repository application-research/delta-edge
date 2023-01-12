package cmd

import (
	"context"
	"edge-ur/core"
	"fmt"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

// call the local APIs instead.
func PinCmd() []*cli.Command {
	var pinCommands []*cli.Command

	pinFileCmd := &cli.Command{
		Name:  "pin-file",
		Usage: "Pin a file on the Filecoin network.",
		Action: func(c *cli.Context) error {
			lightNode, _ := core.NewCliNode(c) // light node now
			value := c.Args().Get(0)
			r, err := os.Open(value)
			if err != nil {
				return nil
			}

			fileNode, err := lightNode.Node.AddPinFile(context.Background(), r, nil)
			size, err := fileNode.Size()
			content := core.Content{
				Name:             r.Name(),
				Size:             int64(size),
				Cid:              fileNode.Cid().String(),
				RequestingApiKey: viper.Get("API_KEY").(string),
				Created_at:       time.Now(),
				Updated_at:       time.Now(),
			}
			lightNode.DB.Create(&content)
			return nil
		},
	}

	pinDirCmd := &cli.Command{
		Name:  "pin-dir",
		Usage: "Pin a directory on the Filecoin network.",
		Action: func(c *cli.Context) error {
			lightNode, _ := core.NewCliNode(c) // light node now
			valuePath := c.Args().Get(0)
			fileNode, _ := lightNode.Node.AddPinDirectory(context.Background(), valuePath)
			size, err := fileNode.Size()
			if err != nil {
				return err
			}
			content := core.Content{
				Name:             valuePath,
				Size:             int64(size),
				Cid:              fileNode.Cid().String(),
				RequestingApiKey: viper.Get("API_KEY").(string),
				Created_at:       time.Now(),
				Updated_at:       time.Now(),
			}
			lightNode.DB.Create(&content)
			fmt.Println(fileNode.Cid().String())
			return nil
		},
	}

	pinCommands = append(pinCommands, pinFileCmd, pinDirCmd)
	return pinCommands
}
