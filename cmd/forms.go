package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jotform/jotform-cli/internal/api"
	"github.com/jotform/jotform-cli/internal/config"
	"github.com/jotform/jotform-cli/internal/formcode"
	"github.com/jotform/jotform-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var formsCmd = &cobra.Command{
	Use:     "forms",
	Aliases: []string{"f", "form"},
	Short:   "Manage Jotform forms",
}

// ── LIST ────────────────────────────────────────────────────────────────

var formsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all forms",
	RunE:  runFormsList,
}

func runFormsList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}
	forms, err := client.ListForms(0, 100)
	if err != nil {
		return err
	}
	return output.Print(forms, output.Format(viper.GetString("output")))
}

// ── GET ─────────────────────────────────────────────────────────────────

var formsGetCmd = &cobra.Command{
	Use:   "get [form-id]",
	Short: "Fetch form structure",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFormsGet,
}

func runFormsGet(cmd *cobra.Command, args []string) error {
	formID, err := config.ResolveFormID(args)
	if err != nil {
		return err
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	form, err := client.GetForm(formID)
	if err != nil {
		return err
	}
	return output.Print(form, output.Format(viper.GetString("output")))
}

// ── CREATE ──────────────────────────────────────────────────────────────

var formsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a form from a local JSON or YAML file",
	RunE:  runFormsCreate,
}

func runFormsCreate(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	schemaFile, err := config.ResolveSchemaFile(filePath)
	if err != nil {
		return err
	}

	schema, err := formcode.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	// Schema validation before API call
	skipValidation, _ := cmd.Flags().GetBool("skip-validation")
	if !skipValidation {
		if errs := formcode.ValidateSchema(schema); len(errs) > 0 {
			fmt.Fprintln(os.Stderr, "Schema validation failed:")
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  ✗ %s\n", e)
			}
			fmt.Fprintln(os.Stderr, "\nUse --skip-validation to bypass.")
			return fmt.Errorf("%d validation error(s)", len(errs))
		}
	}

	client, err := newClient()
	if err != nil {
		return err
	}
	form, err := client.CreateForm(schema)
	if err != nil {
		return err
	}
	fmt.Printf("Form created: %s\nID:  %s\nURL: %s\n", form.Title, form.ID, form.URL)
	return nil
}

// ── UPDATE ──────────────────────────────────────────────────────────────

var formsUpdateCmd = &cobra.Command{
	Use:   "update [form-id]",
	Short: "Update an existing form from a local JSON or YAML file",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFormsUpdate,
}

func runFormsUpdate(cmd *cobra.Command, args []string) error {
	formID, err := config.ResolveFormID(args)
	if err != nil {
		return err
	}

	filePath, _ := cmd.Flags().GetString("file")
	schemaFile, err := config.ResolveSchemaFile(filePath)
	if err != nil {
		return err
	}

	schema, err := formcode.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	skipValidation, _ := cmd.Flags().GetBool("skip-validation")
	if !skipValidation {
		if errs := formcode.ValidateSchema(schema); len(errs) > 0 {
			fmt.Fprintln(os.Stderr, "Schema validation failed:")
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  ✗ %s\n", e)
			}
			fmt.Fprintln(os.Stderr, "\nUse --skip-validation to bypass.")
			return fmt.Errorf("%d validation error(s)", len(errs))
		}
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Printf("⚠️  Would update form %s from %s\n", formID, schemaFile)
		fmt.Println("Run without --dry-run to apply.")
		return nil
	}

	client, err := newClient()
	if err != nil {
		return err
	}
	form, err := client.UpdateForm(formID, schema)
	if err != nil {
		return err
	}
	fmt.Printf("Form updated: %s\nID:  %s\nURL: %s\n", form.Title, form.ID, form.URL)
	return nil
}

// ── DELETE ──────────────────────────────────────────────────────────────

var formsDeleteCmd = &cobra.Command{
	Use:   "delete [form-id]",
	Short: "Delete a form",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFormsDelete,
}

func runFormsDelete(cmd *cobra.Command, args []string) error {
	formID, err := config.ResolveFormID(args)
	if err != nil {
		return err
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Printf("⚠️  Would delete form %s\n", formID)
		fmt.Println("Run without --dry-run to confirm.")
		return nil
	}

	// Interactive confirmation unless --force
	force, _ := cmd.Flags().GetBool("force")
	if !force {
		if !confirmPrompt(fmt.Sprintf("Delete form %s? This cannot be undone.", formID)) {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	client, err := newClient()
	if err != nil {
		return err
	}
	if err := client.DeleteForm(formID); err != nil {
		return err
	}
	fmt.Printf("Form %s deleted.\n", formID)
	return nil
}

// ── SYNC ────────────────────────────────────────────────────────────────

var formsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Download all forms to ~/.jotform/",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		forms, err := client.ListForms(0, 200)
		if err != nil {
			return err
		}

		home, _ := os.UserHomeDir()
		dir := home + "/.jotform"
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		for _, f := range forms {
			props, err := client.GetForm(f.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skipping %s: %v\n", f.ID, err)
				continue
			}
			b, _ := json.MarshalIndent(props, "", "  ")
			path := fmt.Sprintf("%s/%s.json", dir, f.ID)
			os.WriteFile(path, b, 0644)
		}
		fmt.Printf("Synced %d forms to %s\n", len(forms), dir)
		return nil
	},
}

// ── EXPORT / PULL ───────────────────────────────────────────────────────

var formsExportCmd = &cobra.Command{
	Use:   "export [form-id]",
	Short: "Export a single form to a local YAML or JSON file",
	Long:  `Downloads the full form structure and writes it to a local file for version control.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFormsExport,
}

func runFormsExport(cmd *cobra.Command, args []string) error {
	formID, err := config.ResolveFormID(args)
	if err != nil {
		return err
	}

	outPath, _ := cmd.Flags().GetString("out")
	if outPath == "" {
		// Check project context for default schema path
		cfg, _ := config.LoadProject()
		if cfg != nil && cfg.Schema != "" {
			outPath = cfg.Schema
		} else {
			outPath = formID + ".yaml"
		}
	}

	client, err := newClient()
	if err != nil {
		return err
	}
	form, err := client.GetForm(formID)
	if err != nil {
		return err
	}

	data, err := formcode.FormPropertiesToMap(form)
	if err != nil {
		return err
	}

	if err := formcode.WriteFile(outPath, data); err != nil {
		return err
	}
	fmt.Printf("Exported form %s → %s\n", formID, outPath)
	return nil
}

// ── IMPORT ──────────────────────────────────────────────────────────────

var formsImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a form from a local JSON or YAML file (creates a new form)",
	Long:  `Reads a local form definition and creates it on Jotform. Alias for 'create' with Form-as-Code terminology.`,
	RunE:  runFormsCreate, // Same logic as create
}

// ── DIFF ────────────────────────────────────────────────────────────────

var formsDiffCmd = &cobra.Command{
	Use:   "diff [form-id]",
	Short: "Show differences between remote form and local file",
	Long:  `Compares the remote form structure with a local file and displays a unified diff.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFormsDiff,
}

func runFormsDiff(cmd *cobra.Command, args []string) error {
	formID, err := config.ResolveFormID(args)
	if err != nil {
		return err
	}

	filePath, _ := cmd.Flags().GetString("file")
	schemaFile, err := config.ResolveSchemaFile(filePath)
	if err != nil {
		return err
	}

	local, err := formcode.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	client, err := newClient()
	if err != nil {
		return err
	}
	form, err := client.GetForm(formID)
	if err != nil {
		return err
	}

	remote, err := formcode.FormPropertiesToMap(form)
	if err != nil {
		return err
	}

	if !formcode.HasChanges(remote, local) {
		fmt.Println("No changes detected.")
		return nil
	}

	diff, err := formcode.ComputeDiff(remote, local)
	if err != nil {
		return err
	}
	fmt.Print(diff)
	return nil
}

// ── STATUS ──────────────────────────────────────────────────────────────

var formsStatusCmd = &cobra.Command{
	Use:   "status [form-id]",
	Short: "Show differences between local and remote form",
	Long:  `Displays a summary of changes between the local schema file and the remote form, similar to git status.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFormsStatus,
}

func runFormsStatus(cmd *cobra.Command, args []string) error {
	formID, err := config.ResolveFormID(args)
	if err != nil {
		return err
	}

	filePath, _ := cmd.Flags().GetString("file")
	schemaFile, err := config.ResolveSchemaFile(filePath)
	if err != nil {
		return err
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	// Create an adapter that wraps the API client
	adapter := &apiClientAdapter{client: client}
	report, err := formcode.ComputeStatus(adapter, formID, schemaFile)
	if err != nil {
		return err
	}

	// Check for --summary flag
	summary, _ := cmd.Flags().GetBool("summary")
	if summary {
		displayStatusSummary(report)
	} else {
		displayStatusReport(report)
	}

	return nil
}

// apiClientAdapter adapts api.Client to formcode.APIClient interface
type apiClientAdapter struct {
	client interface {
		GetForm(id string) (*api.FormProperties, error)
	}
}

func (a *apiClientAdapter) GetForm(id string) (*formcode.FormProperties, error) {
	apiForm, err := a.client.GetForm(id)
	if err != nil {
		return nil, err
	}
	
	// Convert api.FormProperties to formcode.FormProperties
	return &formcode.FormProperties{
		ID:         apiForm.ID,
		Title:      apiForm.Title,
		Questions:  apiForm.Questions,
		Properties: apiForm.Properties,
	}, nil
}

// displayStatusReport shows the full status report with all changes
func displayStatusReport(report *formcode.StatusReport) {
	fmt.Printf("Form: %s (%s)\n", report.FormName, report.FormID)
	fmt.Printf("Local schema: %s (modified %s)\n", 
		filepath.Base(report.LocalPath), 
		formatRelativeTime(report.LocalModified))
	fmt.Printf("Remote: last updated %s\n\n", formatRelativeTime(report.RemoteModified))

	if !report.HasChanges {
		fmt.Println("No changes detected.")
		return
	}

	fmt.Println("Changes:")
	for _, change := range report.Changes {
		displayChange(change)
	}

	fmt.Println()
	// Suggest next actions based on timestamps
	if report.LocalModified.After(report.RemoteModified) {
		fmt.Println("Run 'jotform push' to apply local changes.")
	} else {
		fmt.Println("Run 'jotform pull' to download remote changes.")
	}
	fmt.Println("Run 'jotform diff' to see detailed differences.")
}

// displayStatusSummary shows only change counts
func displayStatusSummary(report *formcode.StatusReport) {
	if !report.HasChanges {
		fmt.Println("No changes detected.")
		return
	}

	added := 0
	modified := 0
	deleted := 0

	for _, change := range report.Changes {
		switch change.Type {
		case formcode.ChangeAdded:
			added++
		case formcode.ChangeModified:
			modified++
		case formcode.ChangeDeleted:
			deleted++
		}
	}

	fmt.Printf("%d changes: %d modified, %d added, %d deleted\n", 
		len(report.Changes), modified, added, deleted)
}

// displayChange formats a single change for display
func displayChange(change formcode.Change) {
	var indicator string
	switch change.Type {
	case formcode.ChangeAdded:
		indicator = "+"
	case formcode.ChangeModified:
		indicator = "~"
	case formcode.ChangeDeleted:
		indicator = "-"
	}

	// Format the change based on type
	if change.Type == formcode.ChangeModified && change.OldValue != nil && change.NewValue != nil {
		// Show old → new for modifications
		oldStr := formatValue(change.OldValue)
		newStr := formatValue(change.NewValue)
		fmt.Printf("  %s %s: %s → %s\n", indicator, change.Path, oldStr, newStr)
	} else if change.Type == formcode.ChangeAdded {
		// Show new value for additions
		newStr := formatValue(change.NewValue)
		fmt.Printf("  %s %s: %s\n", indicator, change.Path, newStr)
	} else if change.Type == formcode.ChangeDeleted {
		// Show old value for deletions
		oldStr := formatValue(change.OldValue)
		fmt.Printf("  %s %s: %s\n", indicator, change.Path, oldStr)
	} else {
		// Fallback to description
		fmt.Printf("  %s %s\n", indicator, change.Description)
	}
}

// formatValue converts an interface{} to a readable string for display
func formatValue(val interface{}) string {
	if val == nil {
		return "<nil>"
	}

	switch v := val.(type) {
	case string:
		// Truncate long strings
		if len(v) > 50 {
			return fmt.Sprintf("\"%s...\"", v[:47])
		}
		return fmt.Sprintf("\"%s\"", v)
	case map[string]interface{}:
		// For objects, show a summary
		if text, ok := v["text"].(string); ok {
			return fmt.Sprintf("field \"%s\"", text)
		}
		if name, ok := v["name"].(string); ok {
			return fmt.Sprintf("field \"%s\"", name)
		}
		return "object"
	case []interface{}:
		return fmt.Sprintf("array[%d]", len(v))
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatRelativeTime converts a timestamp to a human-readable relative time
func formatRelativeTime(t time.Time) string {
	duration := time.Since(t)
	
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
}

// ── APPLY / PUSH ────────────────────────────────────────────────────────

var formsApplyCmd = &cobra.Command{
	Use:   "apply [form-id]",
	Short: "Apply local changes to a remote form (diff + update)",
	Long: `Compares the local file with the remote form; if changes exist,
displays the diff and applies the update.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runFormsApply,
}

func runFormsApply(cmd *cobra.Command, args []string) error {
	formID, err := config.ResolveFormID(args)
	if err != nil {
		return err
	}

	filePath, _ := cmd.Flags().GetString("file")
	schemaFile, err := config.ResolveSchemaFile(filePath)
	if err != nil {
		return err
	}

	local, err := formcode.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	skipValidation, _ := cmd.Flags().GetBool("skip-validation")
	if !skipValidation {
		if errs := formcode.ValidateSchema(local); len(errs) > 0 {
			fmt.Fprintln(os.Stderr, "Schema validation failed:")
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  ✗ %s\n", e)
			}
			fmt.Fprintln(os.Stderr, "\nUse --skip-validation to bypass.")
			return fmt.Errorf("%d validation error(s)", len(errs))
		}
	}

	client, err := newClient()
	if err != nil {
		return err
	}
	form, err := client.GetForm(formID)
	if err != nil {
		return err
	}

	remote, err := formcode.FormPropertiesToMap(form)
	if err != nil {
		return err
	}

	if !formcode.HasChanges(remote, local) {
		fmt.Println("No changes to apply.")
		return nil
	}

	diff, err := formcode.ComputeDiff(remote, local)
	if err != nil {
		return err
	}
	fmt.Println("Changes to apply:")
	fmt.Print(diff)
	fmt.Println()

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Println("No changes applied (--dry-run).")
		return nil
	}

	updated, err := client.UpdateForm(formID, local)
	if err != nil {
		return err
	}
	fmt.Printf("Form updated: %s\nID:  %s\nURL: %s\n", updated.Title, updated.ID, updated.URL)
	return nil
}

// ── INIT ────────────────────────────────────────────────────────────────

func init() {
	// create flags
	formsCreateCmd.Flags().String("file", "", "Path to JSON or YAML form schema file")
	formsCreateCmd.Flags().Bool("skip-validation", false, "Skip schema validation before creating")

	// update flags
	formsUpdateCmd.Flags().String("file", "", "Path to JSON or YAML form schema file")
	formsUpdateCmd.Flags().Bool("skip-validation", false, "Skip schema validation before updating")
	formsUpdateCmd.Flags().Bool("dry-run", false, "Preview changes without applying")

	// delete flags
	formsDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	formsDeleteCmd.Flags().Bool("dry-run", false, "Preview action without executing")

	// export flags
	formsExportCmd.Flags().StringP("out", "o", "", "Output file path (default: <form-id>.yaml)")

	// import flags (same as create)
	formsImportCmd.Flags().String("file", "", "Path to JSON or YAML form schema file")
	formsImportCmd.Flags().Bool("skip-validation", false, "Skip schema validation before importing")

	// diff flags
	formsDiffCmd.Flags().String("file", "", "Path to local form file to compare against remote")

	// apply flags
	formsApplyCmd.Flags().String("file", "", "Path to local form file to apply to remote")
	formsApplyCmd.Flags().Bool("skip-validation", false, "Skip schema validation before applying")
	formsApplyCmd.Flags().Bool("dry-run", false, "Preview changes without applying")

	// status flags
	formsStatusCmd.Flags().String("file", "", "Path to local form file to compare")
	formsStatusCmd.Flags().Bool("summary", false, "Show only change counts without details")

	formsCmd.AddCommand(
		formsListCmd,
		formsGetCmd,
		formsCreateCmd,
		formsUpdateCmd,
		formsDeleteCmd,
		formsSyncCmd,
		formsExportCmd,
		formsImportCmd,
		formsDiffCmd,
		formsApplyCmd,
		formsStatusCmd,
	)
	rootCmd.AddCommand(formsCmd)
}
