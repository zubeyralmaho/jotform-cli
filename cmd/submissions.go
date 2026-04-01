package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/zubeyralmaho/jotform-cli/internal/api"
	"github.com/zubeyralmaho/jotform-cli/internal/config"
	"github.com/zubeyralmaho/jotform-cli/internal/output"
	"github.com/zubeyralmaho/jotform-cli/internal/ui"
	"github.com/zubeyralmaho/jotform-cli/internal/watch"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var submissionsCmd = &cobra.Command{
	Use:     "submissions",
	Aliases: []string{"subs", "sub"},
	Short:   "Manage form submissions",
}

var submissionsListCmd = &cobra.Command{
	Use:   "list [form-id]",
	Short: "List recent submissions for a form",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		formID, err := config.ResolveFormID(args)
		if err != nil {
			return err
		}
		limit, _ := cmd.Flags().GetInt("limit")
		client, err := newClient()
		if err != nil {
			return err
		}

		res, err := ui.RunWithSpinner("Loading submissions...", func() (interface{}, error) {
			return client.GetSubmissions(formID, 0, limit, "created_at", "DESC")
		})
		if err != nil {
			return err
		}
		subs := res.([]api.Submission)
		return output.Print(subs, output.Format(viper.GetString("output")))
	},
}

var submissionsWatchCmd = &cobra.Command{
	Use:   "watch [form-id]",
	Short: "Stream new submissions to stdout (newline-delimited JSON)",
	Long: `Long-polls the Jotform API and emits new submissions as newline-delimited JSON.
By default uses a checkpoint file (~/.jotform/watch-<formID>.cursor) to survive restarts.
Use --no-checkpoint to disable persistence and keep everything in memory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSubmissionsWatch,
}

func runSubmissionsWatch(cmd *cobra.Command, args []string) error {
	formID, err := config.ResolveFormID(args)
	if err != nil {
		return err
	}
	interval, _ := cmd.Flags().GetDuration("interval")
	noCheckpoint, _ := cmd.Flags().GetBool("no-checkpoint")
	client, err := newClient()
	if err != nil {
		return err
	}
	return watchSubmissions(client, formID, interval, !noCheckpoint)
}

func watchSubmissions(client *api.Client, formID string, interval time.Duration, useCheckpoint bool) error {
	enc := json.NewEncoder(os.Stdout)

	var cp *watch.Checkpoint
	seen := map[string]bool{}

	if useCheckpoint {
		var err error
		cp, err = watch.Load(formID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not load checkpoint: %v (starting fresh)\n", err)
			cp = nil
			useCheckpoint = false
		} else if cp.LastSeenID != "" {
			fmt.Fprintf(os.Stderr, "Resuming from checkpoint (last seen: %s at %s)\n", cp.LastSeenID, cp.LastCreatedAt)
		}
	}

	fmt.Fprintf(os.Stderr, "Watching form %s (interval: %s) — Ctrl+C to stop\n", formID, interval)

	for {
		subs, err := client.GetSubmissions(formID, 0, 50, "created_at", "DESC")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		} else {
			// Process oldest first
			for i := len(subs) - 1; i >= 0; i-- {
				s := subs[i]

				// Dedup check
				if useCheckpoint && cp != nil {
					if cp.HasSeen(s.ID, s.CreatedAt) {
						continue
					}
				} else {
					if seen[s.ID] {
						continue
					}
					seen[s.ID] = true
				}

				if err := enc.Encode(s); err != nil {
					fmt.Fprintf(os.Stderr, "encode error: %v\n", err)
				}

				// Update checkpoint
				if useCheckpoint && cp != nil {
					cp.Update(s.ID, s.CreatedAt)
				}
			}

			// Persist checkpoint after each batch
			if useCheckpoint && cp != nil {
				if err := cp.Save(); err != nil {
					fmt.Fprintf(os.Stderr, "warning: could not save checkpoint: %v\n", err)
				}
			}
		}
		time.Sleep(interval)
	}
}

func init() {
	submissionsListCmd.Flags().Int("limit", 20, "Number of submissions to return")
	submissionsWatchCmd.Flags().Duration("interval", 5*time.Second, "Polling interval")
	submissionsWatchCmd.Flags().Bool("no-checkpoint", false, "Disable checkpoint file (in-memory dedup only)")

	submissionsCmd.AddCommand(submissionsListCmd, submissionsWatchCmd)
	rootCmd.AddCommand(submissionsCmd)
}
