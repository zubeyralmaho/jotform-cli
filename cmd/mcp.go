package cmd

import (
	"github.com/zubeyralmaho/jotform-cli/internal/ai"
	mcpserver "github.com/zubeyralmaho/jotform-cli/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol server",
}

var mcpStartCmd = &cobra.Command{
	Use:   "start-server",
	Short: "Start the MCP server (stdio transport)",
	Long: `Launches an MCP server over stdio, exposing Jotform tools
for use by Claude Desktop, Claude Code, and other MCP-compatible hosts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := newClient()
		if err != nil {
			return err
		}

		var aiGen *ai.Generator
		if key := getAnthropicKey(); key != "" {
			aiGen = ai.NewGenerator(key)
		}

		srv := mcpserver.New(apiClient, aiGen)
		return srv.Run()
	},
}

func init() {
	mcpCmd.AddCommand(mcpStartCmd)
	rootCmd.AddCommand(mcpCmd)
}
