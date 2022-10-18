package cmd

import (
	"fmt"
	"strings"
	"time"

	sf "github.com/catalystsquad/salesforce-bulk-exporter/internal/salesforce"
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export object_name",
	Short: "Exports all data from an object",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// cobra should require just one argument, so hardcode the index reference
		object := args[0]
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
		// generate the query to use in the job
		var query string
		if len(exportCmdFields) == 0 {
			query, err = sf.GenerateQueryWithAllFields(object)
			if err != nil {
				return err
			}
		} else {
			query = fmt.Sprintf("SELECT %s FROM %s", strings.Join(exportCmdFields[:], ", "), object)
		}

		// submit the query
		jobID, err := sf.SubmitBulkQueryJob(query)
		if err != nil {
			return err
		}
		fmt.Printf("Submitted bulk query job with ID: %s\n", jobID)
		// wait for completion
		if exportCmdDownload {
			err = sf.WaitUntilJobComplete(jobID, exportCmdWaitInterval)
			if err != nil {
				return err
			}
			filenames, err := sf.SaveAllResults(jobID, exportCmdFilePrefix, exportCmdFileExt)
			if err != nil {
				return err
			}
			fmt.Printf("Saved export to files: %s\n", strings.Join(filenames[:], ","))
		}
		return nil
	},
}

var (
	exportCmdDownload     bool
	exportCmdWaitInterval time.Duration
	exportCmdFields       []string
	exportCmdFilePrefix   string
	exportCmdFileExt      string
)

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().BoolVarP(&exportCmdDownload, "download", "w", false, "Wait for the job to complete and download")
	exportCmd.Flags().DurationVarP(&exportCmdWaitInterval, "wait-interval", "i", 10*time.Second, "Time to wait in between polls of job status")
	exportCmd.Flags().StringSliceVar(&exportCmdFields, "fields", []string{}, "Specify which fields to export, by default all fields are discovered")
	exportCmd.Flags().StringVarP(&exportCmdFilePrefix, "filename-prefix", "f", "export", "Filename prefix for the downloaded files from Salesforce")
	exportCmd.Flags().StringVarP(&exportCmdFileExt, "file-extension", "e", "csv", "Filename extension for the downloaded files from Salesforce")
}
