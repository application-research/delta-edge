package cmd

import (
	"context"
	"edge-ur/api"
	"edge-ur/core"
	"edge-ur/jobs"
	"github.com/urfave/cli/v2"
	"time"
)

func DaemonCmd() []*cli.Command {
	// add a command to run API node
	var daemonCommands []*cli.Command

	daemonCmd := &cli.Command{
		Name:  "daemon",
		Usage: "A light version of Estuary that allows users to upload and download data from the Filecoin network.",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "enable-api",
			},
		},
		Action: func(c *cli.Context) error {

			ln, err := core.NewLightNode(context.Background())
			if err != nil {
				return err
			}

			//	launch the jobs
			go runJobs(ln)

			// launch the API node
			api.InitializeEchoRouterConfig(ln)
			api.LoopForever()

			return nil
		},
	}

	// add commands.
	daemonCommands = append(daemonCommands, daemonCmd)

	return daemonCommands

}

func runJobs(ln *core.LightNode) {
	// run the job every 10 seconds.
	tick10 := time.NewTicker(10 * time.Second)
	tick30 := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-tick10.C:
			// run the job.

			go func() {
				bucketAssignRun := jobs.NewBucketAssignProcessor(ln)
				bucketAssignRun.Run()
			}()

			go func() {
				uploadToEstuaryRun := jobs.NewUploadToEstuaryProcessor(ln)
				uploadToEstuaryRun.Run()
			}()

		case <-tick30.C:
			go func() {
				uploadToEstuaryRun := jobs.NewUploadToEstuaryProcessor(ln)
				uploadToEstuaryRun.Run()
			}()

		}
	}
}
