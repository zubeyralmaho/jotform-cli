package cmd

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"

	"github.com/jotform/jotform-cli/internal/config"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open [form-id]",
	Short: "Open a form in the default browser",
	Long: `Open a Jotform form in the system default browser.
The form ID can be provided as an argument or resolved from project context.

Examples:
  jotform open                    # Uses form ID from .jotform.yaml
  jotform open 242753193847060    # Opens specific form`,
	Args: cobra.MaximumNArgs(1),
	RunE: runOpen,
}

func runOpen(cmd *cobra.Command, args []string) error {
	// Resolve form ID from args or project context
	formID, err := config.ResolveFormID(args)
	if err != nil {
		return err
	}

	// Validate form ID format (must be numeric)
	if !isValidFormID(formID) {
		return fmt.Errorf("invalid form ID format: %s (must be numeric)", formID)
	}

	// Construct Jotform form URL
	formURL := fmt.Sprintf("https://form.jotform.com/%s", formID)

	// Launch browser
	if err := openBrowser(formURL); err != nil {
		// Display URL for manual copying when browser launch fails
		fmt.Printf("Error: failed to open browser: %v\n", err)
		fmt.Printf("Form URL: %s\n", formURL)
		return err
	}

	fmt.Printf("Opening %s\n", formURL)
	return nil
}

// isValidFormID checks if the form ID is numeric
func isValidFormID(formID string) bool {
	matched, _ := regexp.MatchString(`^\d+$`, formID)
	return matched
}

// openBrowser opens the given URL in the system default browser
// Handles cross-platform browser launching (macOS, Linux, Windows)
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

func init() {
	rootCmd.AddCommand(openCmd)
}
