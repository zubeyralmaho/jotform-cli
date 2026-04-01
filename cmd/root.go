package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zubeyralmaho/jotform-cli/internal/ui"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:           "jotform",
	Short:         "AI-native data collection at the terminal",
	Long:          "Manage Jotform forms, stream submissions, and generate schemas with AI — directly from the terminal or CI/CD pipelines.",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(err.Error()))
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initAPIKeyCache)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/jotform/config.yaml)")
	rootCmd.PersistentFlags().String("api-key", "", "Jotform API key (overrides keychain)")
	rootCmd.PersistentFlags().String("base-url", "", "Jotform API base URL")
	rootCmd.PersistentFlags().String("output", "table", "Output format: table | json | yaml")
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("base-url"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))

	// Apply branded help to root command
	ui.SetCustomHelp(rootCmd)
}

// initAPIKeyCache pre-loads the API key into cache at app startup to avoid
// repeated keychain prompts during command execution
func initAPIKeyCache() {
	// Only attempt to load if not already provided via flag or env
	if viper.GetString("api_key") == "" {
		// Ignore errors - API key might not be in keychain yet on first login
		resolveAPIKey()
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, _ := os.UserHomeDir()
		viper.AddConfigPath(home + "/.config/jotform")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}
	viper.SetEnvPrefix("JOTFORM")
	viper.AutomaticEnv()
	viper.ReadInConfig()
}
