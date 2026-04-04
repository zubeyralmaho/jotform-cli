package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zubeyralmaho/jotform-cli/internal/api"
	"github.com/zubeyralmaho/jotform-cli/internal/config"
	"github.com/zubeyralmaho/jotform-cli/internal/output"
	"github.com/zubeyralmaho/jotform-cli/internal/ui"
)

var webhooksCmd = &cobra.Command{
	Use:     "webhooks",
	Aliases: []string{"webhook", "wh"},
	Short:   "Manage form webhook endpoints",
}

// ── LIST ────────────────────────────────────────────────────────────────

var webhooksListCmd = &cobra.Command{
	Use:   "list [form-id]",
	Short: "List all webhooks for a form",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runWebhooksList,
}

func runWebhooksList(cmd *cobra.Command, args []string) error {
	formID, err := config.ResolveFormID(args)
	if err != nil {
		return err
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	res, err := ui.RunWithSpinner("Loading webhooks...", func() (interface{}, error) {
		return client.GetFormWebhooks(formID)
	})
	if err != nil {
		return err
	}
	webhooks := res.([]api.Webhook)

	format := output.Format(viper.GetString("output"))
	if format != output.FormatTable {
		return output.Print(webhooks, format)
	}

	if len(webhooks) == 0 {
		fmt.Println(ui.Muted.Render("  No webhooks configured for form " + formID))
		fmt.Println()
		fmt.Println(ui.Muted.Render("  Add one with: ") + ui.Value.Render("jotform webhooks add <url>"))
		return nil
	}

	fmt.Println(ui.Title.Render("  Webhooks") + ui.Muted.Render(fmt.Sprintf("  (form %s, %d total)", formID, len(webhooks))))
	fmt.Println(ui.Separator(60))

	rows := make([]string, 0, len(webhooks))
	for _, w := range webhooks {
		id := ui.Subtitle.Render(fmt.Sprintf("%-6s", w.ID))
		url := ui.Value.Render(w.URL)
		rows = append(rows, fmt.Sprintf("  %s  %s", id, url))
	}

	ui.RenderStaggered(rows, 20*time.Millisecond)
	return nil
}

// ── ADD ─────────────────────────────────────────────────────────────────

var webhooksAddCmd = &cobra.Command{
	Use:   "add <url> [form-id]",
	Short: "Add a webhook endpoint to a form",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runWebhooksAdd,
}

func runWebhooksAdd(cmd *cobra.Command, args []string) error {
	webhookURL := args[0]

	// If second arg provided, it's the form ID
	var formArgs []string
	if len(args) > 1 {
		formArgs = args[1:]
	}
	formID, err := config.ResolveFormID(formArgs)
	if err != nil {
		return err
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	_, err = ui.RunWithSpinner("Adding webhook...", func() (interface{}, error) {
		return nil, client.CreateFormWebhook(formID, webhookURL)
	})
	if err != nil {
		return err
	}

	fmt.Println(ui.SuccessBanner("Webhook added"))
	fmt.Println(ui.KeyValuePairs([][2]string{
		{"  Form", formID},
		{"  URL", webhookURL},
	}))
	return nil
}

// ── REMOVE ──────────────────────────────────────────────────────────────

var webhooksRemoveCmd = &cobra.Command{
	Use:     "remove <webhook-id> [form-id]",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a webhook from a form",
	Args:    cobra.RangeArgs(1, 2),
	RunE:    runWebhooksRemove,
}

func runWebhooksRemove(cmd *cobra.Command, args []string) error {
	webhookID := args[0]

	var formArgs []string
	if len(args) > 1 {
		formArgs = args[1:]
	}
	formID, err := config.ResolveFormID(formArgs)
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force {
		if !confirmPrompt(fmt.Sprintf("Remove webhook %s from form %s?", webhookID, formID)) {
			fmt.Println(ui.Muted.Render("  Cancelled."))
			return nil
		}
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	_, err = ui.RunWithSpinner("Removing webhook...", func() (interface{}, error) {
		return nil, client.DeleteFormWebhook(formID, webhookID)
	})
	if err != nil {
		return err
	}

	fmt.Println(ui.SuccessBanner("Webhook removed"))
	return nil
}

// ── TEST ────────────────────────────────────────────────────────────────

var webhooksTestCmd = &cobra.Command{
	Use:   "test <url>",
	Short: "Send a test POST request to a webhook URL",
	Args:  cobra.ExactArgs(1),
	RunE:  runWebhooksTest,
}

func runWebhooksTest(cmd *cobra.Command, args []string) error {
	webhookURL := args[0]

	fmt.Println(ui.Muted.Render("  Sending test payload to ") + ui.Value.Render(webhookURL))

	res, err := ui.RunWithSpinner("Testing webhook...", func() (interface{}, error) {
		status, err := api.TestWebhook(webhookURL)
		return status, err
	})
	if err != nil {
		return fmt.Errorf("webhook test failed: %w", err)
	}
	statusCode := res.(int)

	if statusCode >= 200 && statusCode < 300 {
		fmt.Println(ui.SuccessBanner(fmt.Sprintf("Webhook responded with %d", statusCode)))
	} else {
		fmt.Println(ui.ErrorBanner(fmt.Sprintf("Webhook responded with %d", statusCode)))
	}
	return nil
}

func init() {
	webhooksRemoveCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	webhooksCmd.AddCommand(webhooksListCmd, webhooksAddCmd, webhooksRemoveCmd, webhooksTestCmd)
	rootCmd.AddCommand(webhooksCmd)
}
