package ui

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "jotform",
		Short: "Jotform CLI",
		Long:  "Form-as-Code CLI for Jotform",
	}

	root.AddCommand(&cobra.Command{Use: "dashboard", Short: "Interactive dashboard"})
	root.AddCommand(&cobra.Command{Use: "auth", Short: "Manage authentication"})
	root.AddCommand(&cobra.Command{Use: "forms", Short: "Manage forms"})
	root.AddCommand(&cobra.Command{Use: "submissions", Short: "View submissions"})
	root.AddCommand(&cobra.Command{Use: "init", Short: "Initialize project"})
	root.AddCommand(&cobra.Command{Use: "clone", Short: "Clone a form"})
	root.AddCommand(&cobra.Command{Use: "pull", Short: "Pull remote changes"})
	root.AddCommand(&cobra.Command{Use: "push", Short: "Push local changes"})
	root.AddCommand(&cobra.Command{Use: "diff", Short: "Show differences"})
	root.AddCommand(&cobra.Command{Use: "status", Short: "Show sync status"})
	root.AddCommand(&cobra.Command{Use: "ai", Short: "AI form generation"})
	root.AddCommand(&cobra.Command{Use: "mcp", Short: "MCP server"})
	root.AddCommand(&cobra.Command{Use: "version", Short: "Print version"})
	root.AddCommand(&cobra.Command{Use: "share", Short: "Share a form"})
	root.AddCommand(&cobra.Command{Use: "template", Short: "Form templates"})
	root.AddCommand(&cobra.Command{Use: "webhooks", Short: "Manage webhooks"})

	return root
}

func TestCategorizeCommands(t *testing.T) {
	root := newTestRootCmd()
	groups := categorizeCommands(root)

	assert.NotEmpty(t, groups)

	groupMap := map[string]commandGroup{}
	for _, g := range groups {
		groupMap[g.Title] = g
	}

	// Core group should contain dashboard, auth, forms, submissions
	core, ok := groupMap["Core"]
	assert.True(t, ok)
	assert.NotEmpty(t, core.Commands)

	coreNames := map[string]bool{}
	for _, c := range core.Commands {
		coreNames[c.Name()] = true
	}
	assert.True(t, coreNames["dashboard"])
	assert.True(t, coreNames["auth"])
	assert.True(t, coreNames["forms"])

	// Workflow group
	workflow, ok := groupMap["Workflow"]
	assert.True(t, ok)
	workflowNames := map[string]bool{}
	for _, c := range workflow.Commands {
		workflowNames[c.Name()] = true
	}
	assert.True(t, workflowNames["init"])
	assert.True(t, workflowNames["clone"])
	assert.True(t, workflowNames["pull"])

	// AI group
	ai, ok := groupMap["AI"]
	assert.True(t, ok)
	aiNames := map[string]bool{}
	for _, c := range ai.Commands {
		aiNames[c.Name()] = true
	}
	assert.True(t, aiNames["ai"])
	assert.True(t, aiNames["mcp"])
}

func TestCategorizeCommands_DashboardFirst(t *testing.T) {
	root := newTestRootCmd()
	groups := categorizeCommands(root)

	for _, g := range groups {
		if g.Title == "Core" {
			assert.Equal(t, "dashboard", g.Commands[0].Name())
			break
		}
	}
}

func TestCategorizeCommands_HiddenCommandsExcluded(t *testing.T) {
	root := &cobra.Command{Use: "jotform"}
	hidden := &cobra.Command{Use: "secret", Short: "hidden", Hidden: true}
	root.AddCommand(hidden)
	root.AddCommand(&cobra.Command{Use: "auth", Short: "Auth"})

	groups := categorizeCommands(root)
	for _, g := range groups {
		for _, c := range g.Commands {
			assert.NotEqual(t, "secret", c.Name())
		}
	}
}

func TestRenderHelp(t *testing.T) {
	root := newTestRootCmd()
	output := renderHelp(root)

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "jotform")
	assert.Contains(t, output, "Usage:")
}

func TestSetCustomHelp(t *testing.T) {
	root := newTestRootCmd()
	SetCustomHelp(root)
	// Just verify it doesn't panic
	assert.NotNil(t, root)
}

func TestRenderCommandGroup(t *testing.T) {
	g := commandGroup{
		Title: "Test",
		Commands: []*cobra.Command{
			{Use: "foo", Short: "Do foo"},
			{Use: "bar", Short: "Do bar"},
		},
	}

	output := renderCommandGroup(g)
	assert.Contains(t, output, "Test")
	assert.Contains(t, output, "foo")
	assert.Contains(t, output, "bar")
}

func TestRenderFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("file", "f", "", "Schema file path")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	output := renderFlags(cmd)
	assert.Contains(t, output, "Flags")
	assert.Contains(t, output, "--file")
	assert.Contains(t, output, "--verbose")
}
