package cmd

import (
	"github.com/spf13/cobra"
)

// ── Top-Level Shortcuts ─────────────────────────────────────────────────
// These provide git/unix-style short commands at the root level.
// Each delegates to the corresponding subcommand's RunE function.
// The original grouped commands (forms list, auth login, etc.) remain fully
// functional for backward compatibility.

// ── Auth shortcuts ──────────────────────────────────────────────────────

var shortLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store your Jotform API key (shortcut for: auth login)",
	RunE:  loginCmd.RunE,
}

var shortLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials (shortcut for: auth logout)",
	RunE:  logoutCmd.RunE,
}

var shortWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current authenticated user (shortcut for: auth whoami)",
	RunE:  whoamiCmd.RunE,
}

// ── Forms shortcuts ─────────────────────────────────────────────────────

var shortLsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all forms (shortcut for: forms list)",
	RunE:    runFormsList,
}

var shortGetCmd = &cobra.Command{
	Use:   "get [form-id]",
	Short: "Fetch form structure (shortcut for: forms get)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFormsGet,
}

var shortNewCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"create"},
	Short:   "Create a form from a local file (shortcut for: forms create)",
	RunE:    runFormsCreate,
}

var shortRmCmd = &cobra.Command{
	Use:     "rm [form-id]",
	Aliases: []string{"remove", "delete"},
	Short:   "Delete a form (shortcut for: forms delete)",
	Args:    cobra.MaximumNArgs(1),
	RunE:    runFormsDelete,
}

// ── Git-like workflow shortcuts ─────────────────────────────────────────

var shortPullCmd = &cobra.Command{
	Use:     "pull [form-id]",
	Aliases: []string{"export"},
	Short:   "Download remote form to local file (shortcut for: forms export)",
	Args:    cobra.MaximumNArgs(1),
	RunE:    runFormsExport,
}

var shortPushCmd = &cobra.Command{
	Use:   "push [form-id]",
	Short: "Apply local changes to remote form (shortcut for: forms apply)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFormsApply,
}

var shortDiffCmd = &cobra.Command{
	Use:   "diff [form-id]",
	Short: "Compare local file with remote form (shortcut for: forms diff)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFormsDiff,
}

// ── Submissions shortcut ────────────────────────────────────────────────

var shortWatchCmd = &cobra.Command{
	Use:   "watch [form-id]",
	Short: "Stream new submissions to stdout (shortcut for: submissions watch)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSubmissionsWatch,
}

// ── AI shortcut ─────────────────────────────────────────────────────────

var shortGenerateCmd = &cobra.Command{
	Use:     "generate [prompt]",
	Aliases: []string{"gen"},
	Short:   "Generate a form schema from a prompt (shortcut for: ai generate-schema)",
	Args:    cobra.MinimumNArgs(1),
	RunE:    aiGenerateCmd.RunE,
}

func init() {
	// Copy flags from originals to shortcuts

	// new/create shortcuts
	shortNewCmd.Flags().String("file", "", "Path to JSON or YAML form schema file")
	shortNewCmd.Flags().Bool("skip-validation", false, "Skip schema validation")

	// rm/delete shortcuts
	shortRmCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	shortRmCmd.Flags().Bool("dry-run", false, "Preview action without executing")

	// pull/export shortcuts
	shortPullCmd.Flags().StringP("out", "o", "", "Output file path (default: <form-id>.yaml)")

	// push/apply shortcuts
	shortPushCmd.Flags().String("file", "", "Path to local form file to apply")
	shortPushCmd.Flags().Bool("skip-validation", false, "Skip schema validation")
	shortPushCmd.Flags().Bool("dry-run", false, "Preview changes without applying")

	// diff shortcuts
	shortDiffCmd.Flags().String("file", "", "Path to local form file to compare")

	// watch shortcuts
	shortWatchCmd.Flags().Duration("interval", 5*1e9, "Polling interval") // 5s
	shortWatchCmd.Flags().Bool("no-checkpoint", false, "Disable checkpoint persistence")

	// generate shortcuts — copy output flag
	shortGenerateCmd.Flags().StringP("out", "o", "", "Write schema to file instead of stdout")
	shortGenerateCmd.Flags().String("model", "", "Anthropic model to use")
	shortGenerateCmd.Flags().Int("max-tokens", 0, "Maximum output tokens")
	shortGenerateCmd.Flags().Duration("timeout", 0, "Request timeout")
	shortGenerateCmd.Flags().Int("max-retries", -1, "Max retries on transient errors")
	shortGenerateCmd.Flags().Bool("show-usage", false, "Show token usage after completion")

	// Register all shortcuts at root level
	rootCmd.AddCommand(
		shortLoginCmd,
		shortLogoutCmd,
		shortWhoamiCmd,
		shortLsCmd,
		shortGetCmd,
		shortNewCmd,
		shortRmCmd,
		shortPullCmd,
		shortPushCmd,
		shortDiffCmd,
		shortWatchCmd,
		shortGenerateCmd,
	)
}
