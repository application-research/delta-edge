package cmd

import (
	"context"
	"edge-ur/api"
	"edge-ur/core"
	"edge-ur/jobs"
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
		Usage: "A light version of Estuary that allows users to upload and download data from the Filecoin network.",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "repo",
			},
			&cli.StringFlag{
				Name: "db-location",
			},
		},
		Action: func(c *cli.Context) error {

			repo := c.String("repo")

			if repo == "" {
				repo = "."
			}

			ln, err := core.NewLightNode(context.Background(), repo)
			if err != nil {
				return err
			}

			//	launch the jobs
			go runProcessors(ln)

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

func runProcessors(ln *core.LightNode) {

	// run the job every 10 seconds.
	bucketAssignFreq, err := strconv.Atoi(viper.Get("BUCKET_ASSIGN").(string))
	uploadFreq, err := strconv.Atoi(viper.Get("UPLOAD_PROCESS").(string))
	dealCheckFreq, err := strconv.Atoi(viper.Get("DEAL_CHECK").(string))

	if err != nil {
		bucketAssignFreq = 10
		uploadFreq = 10
	}

	bucketAssignFreqTick := time.NewTicker(time.Duration(bucketAssignFreq) * time.Second)
	uploadFreqTick := time.NewTicker(time.Duration(uploadFreq) * time.Second)
	dealCheckFreqTick := time.NewTicker(time.Duration(dealCheckFreq) * time.Second)
	for {
		select {
		case <-bucketAssignFreqTick.C:
			go func() {
				bucketAssignRun := jobs.NewBucketAssignProcessor(ln)
				bucketAssignRun.Run()
			}()

		case <-uploadFreqTick.C:
			go func() {
				uploadToEstuaryRun := jobs.NewUploadToEstuaryProcessor(ln)
				uploadToEstuaryRun.Run()
			}()
		case <-dealCheckFreqTick.C:
			go func() {
				dealCheck := jobs.NewDealCheckProcessor(ln)
				dealCheck.Run()
			}()
		}
	}
}
