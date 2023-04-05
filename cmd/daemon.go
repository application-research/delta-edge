package cmd

import (
	"context"
	"fmt"
	"github.com/application-research/edge-ur/api"
	"github.com/application-research/edge-ur/core"
	"github.com/application-research/edge-ur/jobs"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"strconv"
	"time"
)

func DaemonCmd() []*cli.Command {
	// add a command to run API node
	var daemonCommands []*cli.Command

	daemonCmd := &cli.Command{
		Name:  "daemon",
		Usage: "Edge gateway daemon that allows users to upload and download data to/from the Filecoin network.",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "repo",
			},
		},

		Action: func(c *cli.Context) error {

			repo := c.String("repo")

			if repo == "" {
				repo = viper.Get("REPO").(string)
			}

			ln, err := core.NewEdgeNode(context.Background(), repo)
			if err != nil {
				return err
			}

			//	launch the jobs
			go runProcessors(ln)

			// launch the API node
			fmt.Println(`
 _______    ________   ________   _______                    ___  ___   ________     
|\  ___ \  |\   ___ \ |\   ____\ |\  ___ \                  |\  \|\  \ |\   __  \    
\ \   __/| \ \  \_|\ \\ \  \___| \ \   __/|    ____________ \ \  \\\  \\ \  \|\  \   
 \ \  \_|/__\ \  \ \\ \\ \  \  ___\ \  \_|/__ |\____________\\ \  \\\  \\ \   _  _\  
  \ \  \_|\ \\ \  \_\\ \\ \  \|\  \\ \  \_|\ \\|____________| \ \  \\\  \\ \  \\  \| 
   \ \_______\\ \_______\\ \_______\\ \_______\                \ \_______\\ \__\\ _\ 
    \|_______| \|_______| \|_______| \|_______|                 \|_______| \|__|\|__|
`)
			fmt.Println("Starting API server")
			fmt.Println("API server up and running on port 1313")
			api.InitializeEchoRouterConfig(ln)
			api.LoopForever()

			return nil
		},
	}

	// add commands.
	daemonCommands = append(daemonCommands, daemonCmd)

	return daemonCommands

}

func runProcessors(ln *core.LightNode) {
	dealCheckFreq, err := strconv.Atoi(viper.Get("DEAL_CHECK").(string))
	if err != nil {
		dealCheckFreq = 10
	}
	dealCheckFreqTick := time.NewTicker(time.Duration(dealCheckFreq) * time.Second)

	for {
		select {
		case <-dealCheckFreqTick.C:
			go func() {
				dealCheck := jobs.NewDealChecker(ln)
				d := jobs.CreateNewDispatcher() // dispatch jobs
				d.AddJob(dealCheck)
				d.Start(1)

				for {
					if d.Finished() {
						break
					}
				}
			}()
		}
	}
}
