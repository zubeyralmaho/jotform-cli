package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jotform/jotform-cli/internal/api"
	"github.com/jotform/jotform-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var submissionsCmd = &cobra.Command{
	Use:   "submissions",
	Short: "Manage form submissions",
}

var submissionsListCmd = &cobra.Command{
	Use:   "list [form-id]",
	Short: "List recent submissions for a form",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		client, err := newClient()
		if err != nil {
			return err
		}
		subs, err := client.GetSubmissions(args[0], 0, limit, "created_at", "DESC")
		if err != nil {
			return err
		}
		return output.Print(subs, output.Format(viper.GetString("output")))
	},
}

var submissionsWatchCmd = &cobra.Command{
	Use:   "watch [form-id]",
	Short: "Stream new submissions to stdout (newline-delimited JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		interval, _ := cmd.Flags().GetDuration("interval")
		client, err := newClient()
		if err != nil {
			return err
		}
		return watchSubmissions(client, args[0], interval)
	},
}

func watchSubmissions(client *api.Client, formID string, interval time.Duration) error {
	seen := map[string]bool{}
	enc := json.NewEncoder(os.Stdout)

	fmt.Fprintf(os.Stderr, "Watching form %s (interval: %s) — Ctrl+C to stop\n", formID, interval)

	for {
		subs, err := client.GetSubmissions(formID, 0, 50, "created_at", "DESC")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		} else {
			for i := len(subs) - 1; i >= 0; i-- {
				s := subs[i]
				if !seen[s.ID] {
					seen[s.ID] = true
					enc.Encode(s)
				}
			}
		}
		time.Sleep(interval)
	}
}

func init() {
	submissionsListCmd.Flags().Int("limit", 20, "Number of submissions to return")
	submissionsWatchCmd.Flags().Duration("interval", 5*time.Second, "Polling interval")

	submissionsCmd.AddCommand(submissionsListCmd, submissionsWatchCmd)
	rootCmd.AddCommand(submissionsCmd)
}
