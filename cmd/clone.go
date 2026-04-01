package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jotform/jotform-cli/internal/config"
	"github.com/jotform/jotform-cli/internal/formcode"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone [form-id]",
	Short: "Clone a form into a new directory with project configuration",
	Long: `Clone a form into a new directory with .jotform.yaml configuration.
The directory name is derived from the form title (slugified) by default.

Examples:
  jotform clone 242753193847060                    # Creates directory from form title
  jotform clone 242753193847060 --name my-form     # Uses specified directory name
  jotform clone 242753193847060 --force            # Overwrites existing directory`,
	Args: cobra.ExactArgs(1),
	RunE: runClone,
}

func runClone(cmd *cobra.Command, args []string) error {
	formID := args[0]
	nameFlag, _ := cmd.Flags().GetString("name")
	force, _ := cmd.Flags().GetBool("force")

	// Fetch form data from API
	client, err := newClient()
	if err != nil {
		return err
	}

	form, err := client.GetForm(formID)
	if err != nil {
		return fmt.Errorf("failed to fetch form: %w", err)
	}

	// Determine target directory name
	targetDir := determineCloneDir(form.Title, formID, nameFlag)

	// Check if directory exists
	if _, err := os.Stat(targetDir); err == nil {
		if !force {
			return fmt.Errorf("directory already exists: %s/\nUse --force to overwrite or choose a different name with --name", targetDir)
		}
		if targetDir == "" || targetDir == "." || targetDir == string(os.PathSeparator) {
			return fmt.Errorf("refusing to overwrite unsafe directory target: %q", targetDir)
		}
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Define schema file path (relative to target directory)
	schemaFile := "form.yaml"
	schemaPath := filepath.Join(targetDir, schemaFile)

	// Export form schema to target directory
	schema, err := formcode.FormPropertiesToMap(form)
	if err != nil {
		return fmt.Errorf("failed to convert form schema: %w", err)
	}

	if err := formcode.WriteFile(schemaPath, schema); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	// Create .jotform.yaml in target directory
	cfg := &config.ProjectConfig{
		FormID: formID,
		Name:   form.Title,
		Schema: schemaFile, // Relative path within the project directory
	}

	if err := config.SaveProject(cfg, targetDir); err != nil {
		return fmt.Errorf("failed to create .jotform.yaml: %w", err)
	}

	// Display success messages
	fmt.Printf("Creating directory: %s/\n", targetDir)
	fmt.Printf("✔ Exported form → %s/%s\n", targetDir, schemaFile)
	fmt.Printf("✔ Created %s/%s\n", targetDir, config.ProjectFileName)

	return nil
}

// determineCloneDir determines the target directory name for cloning.
// Priority: --name flag > slugified form title (with collision handling)
func determineCloneDir(formTitle, formID string, nameFlag string) string {
	// If --name flag provided, use it directly
	if nameFlag != "" {
		return nameFlag
	}

	// Use slugified form title
	baseDir := config.Slugify(formTitle)

	// Check if directory exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return baseDir
	}

	// Handle collisions by appending numeric suffix
	for i := 2; i < 100; i++ {
		dir := fmt.Sprintf("%s-%d", baseDir, i)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return dir
		}
	}

	// Fallback to form ID if we can't find a unique name
	return formID
}

func init() {
	cloneCmd.Flags().String("name", "", "Custom directory name (overrides slugified form title)")
	cloneCmd.Flags().Bool("force", false, "Overwrite existing directory if it exists")

	rootCmd.AddCommand(cloneCmd)
}
