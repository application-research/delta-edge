// It creates a new Echo instance, adds some middleware, creates a new WhyPFS node, creates a new GatewayHandler, and then
// adds a route to the Echo instance
package main

import (
	_ "net/http"
	"os"

	"github.com/application-research/edge-ur/cmd"
	logging "github.com/ipfs/go-log/v2"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

var (
	log = logging.Logger("edge-ur")
)

func main() {

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Error(err)
	}
	// get all the commands
	var commands []*cli.Command

	// commands
	commands = append(commands, cmd.DaemonCmd()...)
	commands = append(commands, cmd.PinCmd()...)

	app := &cli.App{
		Commands: commands,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
