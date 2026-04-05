package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/zubeyralmaho/jotform-cli/internal/config"
	"github.com/zubeyralmaho/jotform-cli/internal/formcode"
	"github.com/zubeyralmaho/jotform-cli/internal/ui"
)

var devCmd = &cobra.Command{
	Use:     "dev",
	Aliases: []string{"watch-file"},
	Short:   "Watch local schema file and auto-push changes on save",
	Long: `Watches the local form schema file for changes and automatically
pushes updates to the remote Jotform form. Like nodemon for forms.

Requires a .jotform.yaml project context or explicit flags.`,
	RunE: runDev,
}

func runDev(cmd *cobra.Command, args []string) error {
	formID, err := config.ResolveFormID(nil)
	if err != nil {
		return fmt.Errorf("no project context — run `jotform init` first: %w", err)
	}

	filePath, _ := cmd.Flags().GetString("file")
	schemaFile, err := config.ResolveSchemaFile(filePath)
	if err != nil {
		return err
	}

	skipValidation, _ := cmd.Flags().GetBool("skip-validation")
	debounce, _ := cmd.Flags().GetDuration("debounce")

	// Resolve to absolute path for the watcher
	absPath, err := filepath.Abs(schemaFile)
	if err != nil {
		return err
	}

	// Verify file exists
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("schema file not found: %s", absPath)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	// Watch the directory (more reliable for editors that do atomic writes)
	dir := filepath.Dir(absPath)
	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	fmt.Println(ui.Title.Render("  Dev Mode"))
	fmt.Println(ui.Separator(60))
	fmt.Println(ui.KeyValuePairs([][2]string{
		{"  Form", formID},
		{"  File", schemaFile},
		{"  Debounce", debounce.String()},
	}))
	fmt.Println()
	fmt.Println(ui.Muted.Render("  Watching for changes — Ctrl+C to stop"))
	fmt.Println()

	var lastPush time.Time
	baseName := filepath.Base(absPath)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only react to writes on our schema file
			if filepath.Base(event.Name) != baseName {
				continue
			}
			if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
				continue
			}

			// Debounce rapid saves
			if time.Since(lastPush) < debounce {
				continue
			}
			lastPush = time.Now()

			fmt.Println(ui.Subtitle.Render("  Change detected") + ui.Muted.Render(fmt.Sprintf("  %s", time.Now().Format("15:04:05"))))
			pushChanges(formID, absPath, skipValidation)

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "  %s\n", ui.ErrorStyle.Render("watcher error: "+err.Error()))
		}
	}
}

func pushChanges(formID, schemaFile string, skipValidation bool) {
	// Read local schema
	local, err := formcode.ReadFile(schemaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s %s\n", ui.ErrorStyle.Render("Read error:"), err)
		return
	}

	// Validate
	if !skipValidation {
		if errs := formcode.ValidateSchema(local); len(errs) > 0 {
			fmt.Fprintf(os.Stderr, "  %s\n", ui.Warning.Render("Validation failed:"))
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "    %s %s\n", ui.ErrorStyle.Render("x"), e)
			}
			return
		}
	}

	client, err := newClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s %s\n", ui.ErrorStyle.Render("Auth error:"), err)
		return
	}

	// Fetch remote and check for changes
	form, err := client.GetForm(formID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s %s\n", ui.ErrorStyle.Render("Fetch error:"), err)
		return
	}

	remote, err := formcode.FormPropertiesToMap(form)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s %s\n", ui.ErrorStyle.Render("Convert error:"), err)
		return
	}

	if !formcode.HasChanges(remote, local) {
		fmt.Println(ui.Muted.Render("  No changes to push."))
		return
	}

	// Push
	_, err = client.UpdateForm(formID, local)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s %s\n", ui.ErrorStyle.Render("Push error:"), err)
		return
	}

	fmt.Println(ui.Success.Render("  Pushed successfully"))
}

func init() {
	devCmd.Flags().String("file", "", "Path to schema file to watch")
	devCmd.Flags().Bool("skip-validation", false, "Skip schema validation before pushing")
	devCmd.Flags().Duration("debounce", 500*time.Millisecond, "Debounce interval between pushes")

	rootCmd.AddCommand(devCmd)
}
