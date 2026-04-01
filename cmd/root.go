package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "jotform",
	Short: "Jotform CLI — AI-native data collection at the terminal",
	Long: `Jotform CLI lets developers and AI agents create, manage,
and stream Jotform data directly from the terminal or CI/CD pipelines.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/jotform/config.yaml)")
	rootCmd.PersistentFlags().String("api-key", "", "Jotform API key (overrides keychain)")
	rootCmd.PersistentFlags().String("output", "table", "Output format: table | json | yaml")
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
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
