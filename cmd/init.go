package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/zubeyralmaho/jotform-cli/internal/config"
	"github.com/zubeyralmaho/jotform-cli/internal/formcode"
	"github.com/zubeyralmaho/jotform-cli/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a project directory with .jotform.yaml",
	Long: `Initialize a project directory by creating a .jotform.yaml configuration file.
This enables context-aware commands that don't require explicit form IDs.

Interactive mode (default):
  jotform init

Non-interactive mode:
  jotform init --form-id 123456789
  jotform init --new --title "My Form"`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if .jotform.yaml already exists
	if _, err := os.Stat(config.ProjectFileName); err == nil {
		return fmt.Errorf("%s already exists in current directory", config.ProjectFileName)
	}

	formID, _ := cmd.Flags().GetString("form-id")
	newForm, _ := cmd.Flags().GetBool("new")
	formTitle, _ := cmd.Flags().GetString("title")
	schemaFile, _ := cmd.Flags().GetString("schema")

	// Determine if we're in interactive mode
	interactive := formID == "" && !newForm

	if interactive {
		return runInteractiveInit()
	}

	// Non-interactive mode validation
	if newForm && formTitle == "" {
		return fmt.Errorf("--title is required when using --new")
	}
	if !newForm && formID == "" {
		return fmt.Errorf("--form-id is required when not using --new")
	}

	// Default schema file
	if schemaFile == "" {
		schemaFile = "form.yaml"
	}

	// Execute initialization
	if newForm {
		return initNewForm(formTitle, schemaFile)
	}
	return initExistingForm(formID, schemaFile)
}

func runInteractiveInit() error {
	reader := bufio.NewReader(os.Stdin)

	// Step 1: Ask for mode
	fmt.Print("Link to existing form or create new? [existing/new]: ")
	modeInput, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("input cancelled")
	}
	mode := strings.TrimSpace(strings.ToLower(modeInput))

	var formID, formTitle, schemaFile string

	if mode == "new" {
		// Step 2a: Prompt for form title
		for {
			fmt.Print("Form title: ")
			titleInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("input cancelled")
			}
			formTitle = strings.TrimSpace(titleInput)
			if formTitle != "" {
				break
			}
			fmt.Fprintln(os.Stderr, "Error: form title cannot be empty")
		}
	} else if mode == "existing" {
		// Step 2b: Prompt for form ID with validation
		formIDPattern := regexp.MustCompile(`^\d+$`)
		for {
			fmt.Print("Form ID: ")
			idInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("input cancelled")
			}
			formID = strings.TrimSpace(idInput)
			if formIDPattern.MatchString(formID) {
				break
			}
			fmt.Fprintln(os.Stderr, "Error: form ID must be numeric")
		}
	} else {
		return fmt.Errorf("invalid mode: must be 'existing' or 'new'")
	}

	// Step 3: Prompt for schema file path with default
	fmt.Print("Local schema file [form.yaml]: ")
	schemaInput, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("input cancelled")
	}
	schemaFile = strings.TrimSpace(schemaInput)
	if schemaFile == "" {
		schemaFile = "form.yaml"
	}

	// Execute initialization based on mode
	if mode == "new" {
		return initNewForm(formTitle, schemaFile)
	}
	return initExistingForm(formID, schemaFile)
}

func initExistingForm(formID, schemaFile string) error {
	// Fetch form data from API
	client, err := newClient()
	if err != nil {
		return err
	}

	form, err := client.GetForm(formID)
	if err != nil {
		return fmt.Errorf("failed to fetch form: %w", err)
	}

	// Export form schema to local file
	schema, err := formcode.FormPropertiesToMap(form)
	if err != nil {
		return fmt.Errorf("failed to convert form schema: %w", err)
	}

	if err := formcode.WriteFile(schemaFile, schema); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	// Create .jotform.yaml configuration
	cfg := &config.ProjectConfig{
		FormID: formID,
		Name:   form.Title,
		Schema: schemaFile,
	}

	if err := config.SaveProject(cfg, "."); err != nil {
		return fmt.Errorf("failed to create .jotform.yaml: %w", err)
	}

	// Display success message
	fmt.Println(ui.SuccessBanner("Project initialized"))
	fmt.Println()
	fmt.Println(ui.KeyValuePairs([][2]string{
		{"Schema", schemaFile},
		{"Config", config.ProjectFileName},
	}))
	fmt.Println()
	printInitHints()

	return nil
}

func initNewForm(formTitle, schemaFile string) error {
	// Create a minimal form schema
	client, err := newClient()
	if err != nil {
		return err
	}

	// Create new form with minimal schema
	minimalSchema := map[string]interface{}{
		"properties": map[string]interface{}{
			"title": formTitle,
		},
	}

	form, err := client.CreateForm(minimalSchema)
	if err != nil {
		return fmt.Errorf("failed to create form: %w", err)
	}

	// Fetch the full form data to export
	fullForm, err := client.GetForm(form.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch created form: %w", err)
	}

	// Export form schema to local file
	schema, err := formcode.FormPropertiesToMap(fullForm)
	if err != nil {
		return fmt.Errorf("failed to convert form schema: %w", err)
	}

	if err := formcode.WriteFile(schemaFile, schema); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	// Create .jotform.yaml configuration
	cfg := &config.ProjectConfig{
		FormID: form.ID,
		Name:   formTitle,
		Schema: schemaFile,
	}

	if err := config.SaveProject(cfg, "."); err != nil {
		return fmt.Errorf("failed to create .jotform.yaml: %w", err)
	}

	// Display success message
	fmt.Println(ui.SuccessBanner("Form created & project initialized"))
	fmt.Println()
	fmt.Println(ui.KeyValuePairs([][2]string{
		{"Form", fmt.Sprintf("%s (ID: %s)", formTitle, form.ID)},
		{"Schema", schemaFile},
		{"Config", config.ProjectFileName},
	}))
	fmt.Println()
	printInitHints()

	return nil
}

func printInitHints() {
	fmt.Println(ui.Muted.Render("Now you can use:"))
	fmt.Println("  " + ui.Value.Render("jotform diff") + "     " + ui.Muted.Render("compare local vs remote"))
	fmt.Println("  " + ui.Value.Render("jotform push") + "     " + ui.Muted.Render("apply local changes"))
	fmt.Println("  " + ui.Value.Render("jotform pull") + "     " + ui.Muted.Render("download latest remote"))
	fmt.Println("  " + ui.Value.Render("jotform watch") + "    " + ui.Muted.Render("stream submissions"))
}

func init() {
	initCmd.Flags().String("form-id", "", "Form ID for existing form (non-interactive mode)")
	initCmd.Flags().Bool("new", false, "Create a new form (non-interactive mode)")
	initCmd.Flags().String("title", "", "Form title when creating new form (requires --new)")
	initCmd.Flags().String("schema", "", "Local schema file path (default: form.yaml)")

	rootCmd.AddCommand(initCmd)
}
