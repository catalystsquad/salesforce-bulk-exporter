package cmd

import (
	"fmt"
	"strings"
	"time"

	sf "github.com/catalystsquad/salesforce-bulk-exporter/internal/salesforce"
	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download job_id",
	Short: "Downloads the results of a bulk job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// cobra should require just one argument, so hardcode the index reference
		jobID := args[0]
		// initialize the salesforce utils client
		err := sf.InitSFClient(
			config.baseURL,
			config.apiVersion,
			config.clientID,
			config.clientSecret,
			config.username,
			config.password,
			config.grantType,
		)
		if err != nil {
			return err
		}
		// check if or wait until job is done
		if downloadCmdWait {
			err = sf.WaitUntilJobComplete(jobID, downloadCmdWaitInterval)
			if err != nil {
				return err
			}
		} else {
			complete, state, err := sf.CheckIfJobComplete(jobID)
			if err != nil {
				return err
			}
			if !complete {
				return fmt.Errorf("Job not complete, current state is: %s\n", state)
			}
		}
		// download files
		filenames, err := sf.SaveAllResults(jobID, downloadCmdFilePrefix, downloadCmdFileExt)
		if err != nil {
			return err
		}
		fmt.Printf("Saved export to files: %s\n", strings.Join(filenames[:], ","))
		return nil
	},
}

var (
	downloadCmdWait         bool
	downloadCmdWaitInterval time.Duration
	downloadCmdFilePrefix   string
	downloadCmdFileExt      string
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().BoolVarP(&downloadCmdWait, "wait", "w", false, "Wait for the job to complete")
	downloadCmd.Flags().DurationVarP(&downloadCmdWaitInterval, "wait-interval", "i", 10*time.Second, "Time to wait in between polls of job status")
	downloadCmd.Flags().StringVarP(&downloadCmdFilePrefix, "filename-prefix", "f", "export", "Filename prefix for the downloaded files from Salesforce")
	downloadCmd.Flags().StringVarP(&downloadCmdFileExt, "file-extension", "e", "csv", "Filename extension for the downloaded files from Salesforce")
}
