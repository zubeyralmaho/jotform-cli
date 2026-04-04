package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zubeyralmaho/jotform-cli/internal/config"
	"github.com/zubeyralmaho/jotform-cli/internal/formcode"
	"github.com/zubeyralmaho/jotform-cli/internal/ui"
)

var testCmd = &cobra.Command{
	Use:   "test [file]",
	Short: "Validate a form definition against best-practice rules",
	Long: `Runs a suite of validation rules against a local form schema file.

Rules include:
  labels          All input fields must have a text label
  unique-names    Field names must be unique
  email-validation  Email fields should use control_email type
  choice-options  Choice fields must have options
  submit-button   Form must have a submit button
  form-title      Form must have a title
  no-empty-options  No empty values in option lists
  required-fields   Important fields should be required
  field-ordering  Field order keys should be valid

Use --rules to run only specific rules, or --severity to filter by level.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
	var schemaFile string
	if len(args) > 0 {
		schemaFile = args[0]
	} else {
		filePath, _ := cmd.Flags().GetString("file")
		var err error
		schemaFile, err = config.ResolveSchemaFile(filePath)
		if err != nil {
			return err
		}
	}

	schema, err := formcode.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	// Get all rules
	allRules := formcode.BuiltinRules()

	// Filter by --rules flag
	rulesFilter, _ := cmd.Flags().GetString("rules")
	if rulesFilter != "" {
		selected := map[string]bool{}
		for _, r := range strings.Split(rulesFilter, ",") {
			selected[strings.TrimSpace(r)] = true
		}
		var filtered []formcode.Rule
		for _, r := range allRules {
			if selected[r.Name] {
				filtered = append(filtered, r)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("no matching rules found for: %s", rulesFilter)
		}
		allRules = filtered
	}

	// Run rules
	findings := formcode.RunRules(schema, allRules)

	// Filter by severity
	minSeverity, _ := cmd.Flags().GetString("severity")
	if minSeverity != "" {
		var minLevel formcode.Severity
		switch strings.ToLower(minSeverity) {
		case "error":
			minLevel = formcode.SeverityError
		case "warning":
			minLevel = formcode.SeverityWarning
		case "info":
			minLevel = formcode.SeverityInfo
		default:
			return fmt.Errorf("unknown severity %q (use: error, warning, info)", minSeverity)
		}
		var filtered []formcode.RuleFinding
		for _, f := range findings {
			if f.Severity <= minLevel {
				filtered = append(filtered, f)
			}
		}
		findings = filtered
	}

	// List mode
	listRules, _ := cmd.Flags().GetBool("list")
	if listRules {
		fmt.Println(ui.Title.Render("  Available Rules"))
		fmt.Println(ui.Separator(60))
		for _, r := range formcode.BuiltinRules() {
			fmt.Printf("  %s  %s\n", ui.Subtitle.Render(fmt.Sprintf("%-18s", r.Name)), ui.Muted.Render(r.Description))
		}
		return nil
	}

	// Display results
	fmt.Println(ui.Title.Render("  Test Results"))
	fmt.Println(ui.Separator(60))
	fmt.Println(ui.Muted.Render(fmt.Sprintf("  File: %s  |  Rules: %d", schemaFile, len(allRules))))
	fmt.Println()

	if len(findings) == 0 {
		fmt.Println(ui.Success.Render("  All checks passed!"))
		return nil
	}

	errors := 0
	warnings := 0
	infos := 0

	for _, f := range findings {
		var icon string
		var style func(...string) string
		switch f.Severity {
		case formcode.SeverityError:
			icon = "x"
			style = ui.ErrorStyle.Render
			errors++
		case formcode.SeverityWarning:
			icon = "!"
			style = ui.Warning.Render
			warnings++
		case formcode.SeverityInfo:
			icon = "i"
			style = ui.Subtitle.Render
			infos++
		}

		field := ""
		if f.Field != "" {
			field = ui.Muted.Render(fmt.Sprintf("[%s] ", f.Field))
		}
		ruleName := ui.Muted.Render(fmt.Sprintf("(%s)", f.Rule))
		fmt.Printf("  %s %s%s %s\n", style(icon), field, f.Message, ruleName)
	}

	fmt.Println()
	parts := []string{}
	if errors > 0 {
		parts = append(parts, ui.ErrorStyle.Render(fmt.Sprintf("%d error(s)", errors)))
	}
	if warnings > 0 {
		parts = append(parts, ui.Warning.Render(fmt.Sprintf("%d warning(s)", warnings)))
	}
	if infos > 0 {
		parts = append(parts, ui.Subtitle.Render(fmt.Sprintf("%d info(s)", infos)))
	}
	fmt.Println("  " + strings.Join(parts, "  "))

	if errors > 0 {
		os.Exit(1)
	}
	return nil
}

func init() {
	testCmd.Flags().String("file", "", "Path to form schema file")
	testCmd.Flags().String("rules", "", "Comma-separated list of rules to run")
	testCmd.Flags().String("severity", "", "Minimum severity to show: error, warning, info")
	testCmd.Flags().Bool("list", false, "List all available rules")

	rootCmd.AddCommand(testCmd)
}
