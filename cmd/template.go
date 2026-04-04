package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/zubeyralmaho/jotform-cli/internal/formcode"
	"github.com/zubeyralmaho/jotform-cli/internal/templates"
	"github.com/zubeyralmaho/jotform-cli/internal/ui"
)

var templateCmd = &cobra.Command{
	Use:     "template",
	Aliases: []string{"tpl"},
	Short:   "Browse and use curated form starter templates",
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available starter templates",
	RunE:  runTemplateList,
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	all := templates.Builtin()

	fmt.Println(ui.Title.Render("  Starter Templates") + ui.Muted.Render(fmt.Sprintf("  (%d available)", len(all))))
	fmt.Println(ui.Separator(60))

	// Group by category
	categories := map[string][]templates.Template{}
	var order []string
	for _, t := range all {
		if _, seen := categories[t.Category]; !seen {
			order = append(order, t.Category)
		}
		categories[t.Category] = append(categories[t.Category], t)
	}

	var rows []string
	for _, cat := range order {
		rows = append(rows, "")
		rows = append(rows, ui.Subtitle.Render(fmt.Sprintf("  %s", cat)))
		for _, t := range categories[cat] {
			name := ui.Value.Render(fmt.Sprintf("  %-18s", t.Name))
			desc := ui.Muted.Render(t.Description)
			rows = append(rows, fmt.Sprintf("  %s %s", name, desc))
		}
	}

	ui.RenderStaggered(rows, 15*time.Millisecond)
	fmt.Println()
	fmt.Println(ui.Muted.Render("  Use: ") + ui.Value.Render("jotform template use <name>") + ui.Muted.Render(" to scaffold a form"))
	return nil
}

var templateUseCmd = &cobra.Command{
	Use:   "use <template-name>",
	Short: "Scaffold a form from a starter template",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateUse,
}

func runTemplateUse(cmd *cobra.Command, args []string) error {
	name := args[0]
	tpl := templates.Get(name)
	if tpl == nil {
		fmt.Println(ui.ErrorStyle.Render("  Unknown template: " + name))
		fmt.Println()
		fmt.Println(ui.Muted.Render("  Available templates:"))
		for _, t := range templates.Builtin() {
			fmt.Printf("    %s  %s\n", ui.Value.Render(t.Name), ui.Muted.Render(t.Description))
		}
		return fmt.Errorf("template %q not found", name)
	}

	outFile, _ := cmd.Flags().GetString("out")
	if outFile == "" {
		outFile = "form.yaml"
	}

	if err := formcode.WriteFile(outFile, tpl.Schema); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	fmt.Println(ui.SuccessBanner(fmt.Sprintf("Template '%s' scaffolded", name)))
	fmt.Println(ui.KeyValuePairs([][2]string{
		{"  Template", tpl.Name},
		{"  Category", tpl.Category},
		{"  File", outFile},
	}))
	fmt.Println()
	fmt.Println(ui.Muted.Render("  Next steps:"))
	fmt.Println(ui.Muted.Render("    1. Edit ") + ui.Value.Render(outFile) + ui.Muted.Render(" to customize your form"))
	fmt.Println(ui.Muted.Render("    2. Run ") + ui.Value.Render("jotform test") + ui.Muted.Render(" to validate"))
	fmt.Println(ui.Muted.Render("    3. Run ") + ui.Value.Render("jotform new") + ui.Muted.Render(" to create on Jotform"))
	return nil
}

var templateShowCmd = &cobra.Command{
	Use:   "show <template-name>",
	Short: "Preview a template's schema without creating a file",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateShow,
}

func runTemplateShow(cmd *cobra.Command, args []string) error {
	name := args[0]
	tpl := templates.Get(name)
	if tpl == nil {
		return fmt.Errorf("template %q not found — run `jotform template list` to see available templates", name)
	}

	fmt.Println(ui.Title.Render(fmt.Sprintf("  %s", tpl.Name)) + ui.Muted.Render(fmt.Sprintf("  (%s)", tpl.Category)))
	fmt.Println(ui.Muted.Render("  " + tpl.Description))
	fmt.Println(ui.Separator(60))

	questions, ok := tpl.Schema["questions"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Display questions in order
	for i := 1; i <= len(questions); i++ {
		key := fmt.Sprintf("%d", i)
		qRaw, ok := questions[key]
		if !ok {
			continue
		}
		q, ok := qRaw.(map[string]interface{})
		if !ok {
			continue
		}
		qType, _ := q["type"].(string)
		text, _ := q["text"].(string)
		required, _ := q["required"].(string)

		typeLabel := ui.Muted.Render(fmt.Sprintf("%-20s", qType))
		textLabel := ui.Value.Render(text)
		reqLabel := ""
		if required == "Yes" {
			reqLabel = ui.Warning.Render(" *required")
		}

		fmt.Printf("  %s  %s  %s%s\n", ui.Subtitle.Render(fmt.Sprintf("#%s", key)), typeLabel, textLabel, reqLabel)
	}

	return nil
}

func init() {
	templateUseCmd.Flags().StringP("out", "o", "", "Output file path (default: form.yaml)")

	templateCmd.AddCommand(templateListCmd, templateUseCmd, templateShowCmd)
	rootCmd.AddCommand(templateCmd)
}
