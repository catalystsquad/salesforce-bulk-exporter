package cmd

import (
	"os"

	sf "github.com/catalystsquad/salesforce-bulk-exporter/internal/salesforce"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// listJobsCmd represents the listJobs command
var listJobsCmd = &cobra.Command{
	Use:   "list-jobs",
	Short: "List current bulk jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		jobs, err := sf.GetAllBulkJobs()
		if err != nil {
			return err
		}

		// list all jobs in a nice table
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"ID", "Object", "State", "SystemModstamp"})
		for _, job := range jobs {
			t.AppendRow(table.Row{job.ID, job.Object, job.State, job.SystemModstamp})
		}
		t.Render()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listJobsCmd)
}
