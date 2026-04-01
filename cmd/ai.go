package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jotform/jotform-cli/internal/ai"
	"github.com/jotform/jotform-cli/internal/ui"
	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI-powered form generation and analysis",
}

var aiGenerateCmd = &cobra.Command{
	Use:   "generate-schema [prompt]",
	Short: "Generate a Jotform schema from a natural language prompt",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := strings.Join(args, " ")
		apiKey := getAnthropicKey()
		if apiKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY is required — set it via env or config")
		}

		gen := newAIGenerator(cmd, apiKey)

		res, err := ui.RunWithSpinner("Generating schema...", func() (interface{}, error) {
			return gen.GenerateSchema(context.Background(), prompt)
		})
		if err != nil {
			return err
		}
		result := res.(*ai.GenerateResult)

		showUsage, _ := cmd.Flags().GetBool("show-usage")
		if showUsage {
			fmt.Fprintln(os.Stderr, ui.Muted.Render(fmt.Sprintf("[tokens: in=%d out=%d]", result.Usage.InputTokens, result.Usage.OutputTokens)))
		}

		outFile, _ := cmd.Flags().GetString("out")
		if outFile != "" {
			data, _ := json.MarshalIndent(result.Schema, "", "  ")
			if err := os.WriteFile(outFile, data, 0644); err != nil {
				return err
			}
			fmt.Println(ui.SuccessBanner("Schema written to " + outFile))
			return nil
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result.Schema)
	},
}

var aiAnalyzeCmd = &cobra.Command{
	Use:   "analyze [form-id]",
	Short: "Get AI-powered UX suggestions for a form",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey := getAnthropicKey()
		if apiKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY is required — set it via env or config")
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		type analyzeResult struct {
			suggestions string
			usage       *ai.Usage
		}

		res, err := ui.RunWithSpinner("Analyzing form...", func() (interface{}, error) {
			form, err := client.GetForm(args[0])
			if err != nil {
				return nil, err
			}
			formMap := map[string]interface{}{
				"title":     form.Title,
				"questions": form.Questions,
			}
			gen := newAIGenerator(cmd, apiKey)
			suggestions, usage, err := gen.AnalyzeForm(context.Background(), formMap)
			if err != nil {
				return nil, err
			}
			return &analyzeResult{suggestions: suggestions, usage: usage}, nil
		})
		if err != nil {
			return err
		}
		result := res.(*analyzeResult)

		fmt.Println(ui.Title.Render("  Analysis"))
		fmt.Println(ui.Separator(60))
		fmt.Println(result.suggestions)

		showUsage, _ := cmd.Flags().GetBool("show-usage")
		if showUsage && result.usage != nil {
			fmt.Fprintln(os.Stderr, ui.Muted.Render(fmt.Sprintf("[tokens: in=%d out=%d]", result.usage.InputTokens, result.usage.OutputTokens)))
		}
		return nil
	},
}

// newAIGenerator creates an AI generator with options from command flags.
func newAIGenerator(cmd *cobra.Command, apiKey string) *ai.Generator {
	opts := ai.DefaultOptions()

	if model, _ := cmd.Flags().GetString("model"); model != "" {
		opts.Model = model
	}
	if maxTokens, _ := cmd.Flags().GetInt("max-tokens"); maxTokens > 0 {
		opts.MaxTokens = maxTokens
	}
	if timeout, _ := cmd.Flags().GetDuration("timeout"); timeout > 0 {
		opts.Timeout = timeout
	}
	if retries, _ := cmd.Flags().GetInt("max-retries"); retries >= 0 {
		opts.MaxRetries = retries
	}

	return ai.NewGeneratorWithOptions(apiKey, opts)
}

func init() {
	// AI sub-command shared flags
	for _, cmd := range []*cobra.Command{aiGenerateCmd, aiAnalyzeCmd} {
		cmd.Flags().String("model", "", "Anthropic model to use (default: claude-sonnet-4-5-20250514)")
		cmd.Flags().Int("max-tokens", 0, "Maximum output tokens (default: 4096)")
		cmd.Flags().Duration("timeout", 0, "Request timeout (default: 60s)")
		cmd.Flags().Int("max-retries", -1, "Max retries on transient errors (default: 2)")
		cmd.Flags().Bool("show-usage", false, "Show token usage after completion")
	}

	aiGenerateCmd.Flags().StringP("out", "o", "", "Write schema to file instead of stdout")
	aiCmd.AddCommand(aiGenerateCmd, aiAnalyzeCmd)
	rootCmd.AddCommand(aiCmd)
}
