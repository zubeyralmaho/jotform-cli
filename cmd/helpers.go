package cmd

import (
	"os"

	"github.com/jotform/jotform-cli/internal/api"
	"github.com/jotform/jotform-cli/internal/auth"
	"github.com/spf13/viper"
)

// resolveAPIKey returns the API key from (in order):
// 1. --api-key flag / JOTFORM_API_KEY env
// 2. System keychain
func resolveAPIKey() (string, error) {
	if key := viper.GetString("api_key"); key != "" {
		return key, nil
	}
	return auth.LoadAPIKey()
}

// newClient creates an authenticated API client.
func newClient() (*api.Client, error) {
	key, err := resolveAPIKey()
	if err != nil {
		return nil, err
	}
	c := api.New(key)
	if base := viper.GetString("base_url"); base != "" {
		c.BaseURL = base
	}
	return c, nil
}

// getAnthropicKey returns the Anthropic API key from config or env.
func getAnthropicKey() string {
	if key := viper.GetString("anthropic_api_key"); key != "" {
		return key
	}
	return os.Getenv("ANTHROPIC_API_KEY")
}
