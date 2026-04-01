package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// commandGroup groups commands by category for display.
type commandGroup struct {
	Title    string
	Commands []*cobra.Command
}

// categorizeCommands groups subcommands into logical sections.
func categorizeCommands(cmd *cobra.Command) []commandGroup {
	groups := map[string][]*cobra.Command{
		"Core":     {},
		"Workflow": {},
		"AI":       {},
		"Other":    {},
	}

	// Map command names to groups
	coreNames := map[string]bool{
		"dashboard": true,
		"auth":      true, "login": true, "logout": true, "whoami": true,
		"forms": true, "ls": true, "list": true, "get": true,
		"new": true, "create": true, "rm": true, "delete": true,
		"submissions": true, "watch": true,
	}
	workflowNames := map[string]bool{
		"init": true, "clone": true, "open": true,
		"pull": true, "push": true, "diff": true, "status": true,
	}
	aiNames := map[string]bool{
		"ai": true, "generate": true, "gen": true, "mcp": true,
	}

	for _, sub := range cmd.Commands() {
		if sub.Hidden || sub.Name() == "help" || sub.Name() == "completion" {
			continue
		}
		name := sub.Name()
		switch {
		case coreNames[name]:
			groups["Core"] = append(groups["Core"], sub)
		case workflowNames[name]:
			groups["Workflow"] = append(groups["Workflow"], sub)
		case aiNames[name]:
			groups["AI"] = append(groups["AI"], sub)
		default:
			groups["Other"] = append(groups["Other"], sub)
		}
	}

	// Build ordered output
	order := []string{"Core", "Workflow", "AI", "Other"}
	var result []commandGroup
	for _, title := range order {
		cmds := groups[title]
		if len(cmds) > 0 {
			sort.Slice(cmds, func(i, j int) bool {
				if title == "Core" {
					if cmds[i].Name() == "dashboard" {
						return true
					}
					if cmds[j].Name() == "dashboard" {
						return false
					}
				}
				return cmds[i].Name() < cmds[j].Name()
			})
			result = append(result, commandGroup{Title: title, Commands: cmds})
		}
	}
	return result
}

// SetCustomHelp replaces the default cobra help template on cmd with a branded one.
func SetCustomHelp(cmd *cobra.Command) {
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(renderHelp(cmd))
	})
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Println(renderHelp(cmd))
		return nil
	})
}

func renderHelp(cmd *cobra.Command) string {
	var b strings.Builder

	// Logo + tagline
	b.WriteString(LogoWithText())
	b.WriteString("\n\n")

	// Description
	if cmd.Long != "" {
		b.WriteString("  " + Muted.Render(cmd.Long))
	} else if cmd.Short != "" {
		b.WriteString("  " + Muted.Render(cmd.Short))
	}
	b.WriteString("\n\n")

	// Usage line
	usageStyle := lipgloss.NewStyle().Foreground(BrandBlue).Bold(true)
	b.WriteString("  " + Label.Render("Usage:") + "  " + usageStyle.Render(cmd.UseLine()))
	if cmd.HasAvailableSubCommands() {
		b.WriteString(usageStyle.Render(" [command]"))
	}
	b.WriteString("\n\n")

	// Command groups (for root/parent commands)
	if cmd.HasAvailableSubCommands() {
		groups := categorizeCommands(cmd)
		for _, g := range groups {
			b.WriteString(renderCommandGroup(g))
			b.WriteString("\n")
		}
	}

	// Flags
	if cmd.HasAvailableLocalFlags() || cmd.HasAvailablePersistentFlags() {
		b.WriteString(renderFlags(cmd))
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("  " + Muted.Render("Use") + " " + Value.Render(cmd.CommandPath()+" [command] --help") + " " + Muted.Render("for more information about a command."))
	b.WriteString("\n")

	return b.String()
}

func renderCommandGroup(g commandGroup) string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(BrandOrange).
		Bold(true)

	b.WriteString("  " + headerStyle.Render(g.Title) + "\n")

	// Find max command name length for alignment
	maxLen := 0
	for _, c := range g.Commands {
		nameStr := c.Name()
		if len(c.Aliases) > 0 {
			nameStr += ", " + strings.Join(c.Aliases, ", ")
		}
		if len(nameStr) > maxLen {
			maxLen = len(nameStr)
		}
	}

	cmdStyle := lipgloss.NewStyle().Foreground(BrandBlue)
	for _, c := range g.Commands {
		nameStr := c.Name()
		if len(c.Aliases) > 0 {
			nameStr += ", " + strings.Join(c.Aliases, ", ")
		}
		padded := fmt.Sprintf("%-*s", maxLen+2, nameStr)
		b.WriteString("    " + cmdStyle.Render(padded) + " " + Muted.Render(c.Short) + "\n")
	}

	return b.String()
}

func renderFlags(cmd *cobra.Command) string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(BrandOrange).
		Bold(true)

	b.WriteString("  " + headerStyle.Render("Flags") + "\n")

	flagStyle := lipgloss.NewStyle().Foreground(BrandYellow)

	// Render flags manually for styling
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		name := "--" + f.Name
		shorthand := ""
		if f.Shorthand != "" {
			shorthand = "-" + f.Shorthand + ", "
		}
		defVal := ""
		if f.DefValue != "" && f.DefValue != "false" {
			defVal = Muted.Render(fmt.Sprintf(" (default: %s)", f.DefValue))
		}
		_, _ = fmt.Fprintf(&b, "    %s%s  %s%s\n",
			flagStyle.Render(shorthand),
			flagStyle.Render(name),
			Muted.Render(f.Usage),
			defVal,
		)
	})

	return b.String()
}
