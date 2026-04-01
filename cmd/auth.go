package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jotform/jotform-cli/internal/api"
	"github.com/jotform/jotform-cli/internal/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication credentials",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store your Jotform API key securely",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("Enter your Jotform API key: ")
		reader := bufio.NewReader(os.Stdin)
		key, _ := reader.ReadString('\n')
		key = strings.TrimSpace(key)

		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		client := api.New(key)
		user, err := client.GetUser()
		if err != nil {
			return fmt.Errorf("invalid API key: %w", err)
		}

		if err := auth.SaveAPIKey(key); err != nil {
			return err
		}
		fmt.Printf("Logged in as %s (%s)\n", user.Name, user.Email)
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.DeleteAPIKey(); err != nil {
			return err
		}
		fmt.Println("Logged out.")
		return nil
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current authenticated user",
	RunE: func(cmd *cobra.Command, args []string) error {
		key, err := resolveAPIKey()
		if err != nil {
			return err
		}
		client := api.New(key)
		user, err := client.GetUser()
		if err != nil {
			return err
		}
		fmt.Printf("User:  %s\nEmail: %s\nPlan:  %s\n", user.Username, user.Email, user.AccountType)
		return nil
	},
}

func init() {
	authCmd.AddCommand(loginCmd, logoutCmd, whoamiCmd)
	rootCmd.AddCommand(authCmd)
}
