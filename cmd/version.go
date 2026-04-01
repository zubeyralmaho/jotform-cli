package cmd

import (
	"fmt"

	"github.com/zubeyralmaho/jotform-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.LogoWithText())
		fmt.Println()
		fmt.Println(ui.KeyValuePairs([][2]string{
			{"Version", Version},
			{"Commit", Commit},
			{"Built", BuildDate},
		}))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
