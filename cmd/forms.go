package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jotform/jotform-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var formsCmd = &cobra.Command{
	Use:   "forms",
	Short: "Manage Jotform forms",
}

var formsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all forms",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		forms, err := client.ListForms(0, 100)
		if err != nil {
			return err
		}
		return output.Print(forms, output.Format(viper.GetString("output")))
	},
}

var formsGetCmd = &cobra.Command{
	Use:   "get [form-id]",
	Short: "Fetch form structure",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		form, err := client.GetForm(args[0])
		if err != nil {
			return err
		}
		return output.Print(form, output.Format(viper.GetString("output")))
	},
}

var formsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a form from a local JSON or YAML file",
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("cannot read file: %w", err)
		}

		var schema map[string]interface{}
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(data, &schema); err != nil {
				return fmt.Errorf("invalid YAML: %w", err)
			}
		default:
			if err := json.Unmarshal(data, &schema); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
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
	},
}

var formsDeleteCmd = &cobra.Command{
	Use:   "delete [form-id]",
	Short: "Delete a form",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		if err := client.DeleteForm(args[0]); err != nil {
			return err
		}
		fmt.Printf("Form %s deleted.\n", args[0])
		return nil
	},
}

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

func init() {
	formsCreateCmd.Flags().String("file", "", "Path to JSON or YAML form schema file")
	formsCreateCmd.MarkFlagRequired("file")

	formsCmd.AddCommand(formsListCmd, formsGetCmd, formsCreateCmd, formsDeleteCmd, formsSyncCmd)
	rootCmd.AddCommand(formsCmd)
}
