package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jotform/jotform-cli/internal/ai"
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

		gen := ai.NewGenerator(apiKey)
		schema, err := gen.GenerateSchema(context.Background(), prompt)
		if err != nil {
			return err
		}

		outFile, _ := cmd.Flags().GetString("out")
		if outFile != "" {
			data, _ := json.MarshalIndent(schema, "", "  ")
			if err := os.WriteFile(outFile, data, 0644); err != nil {
				return err
			}
			fmt.Printf("Schema written to %s\n", outFile)
			return nil
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(schema)
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
		form, err := client.GetForm(args[0])
		if err != nil {
			return err
		}

		formMap := map[string]interface{}{
			"title":     form.Title,
			"questions": form.Questions,
		}

		gen := ai.NewGenerator(apiKey)
		suggestions, err := gen.AnalyzeForm(context.Background(), formMap)
		if err != nil {
			return err
		}

		fmt.Println(suggestions)
		return nil
	},
}

func init() {
	aiGenerateCmd.Flags().StringP("out", "o", "", "Write schema to file instead of stdout")
	aiCmd.AddCommand(aiGenerateCmd, aiAnalyzeCmd)
	rootCmd.AddCommand(aiCmd)
}
