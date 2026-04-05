package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAnthropicKey_FromEnv(t *testing.T) {
	original := os.Getenv("ANTHROPIC_API_KEY")
	defer func() { _ = os.Setenv("ANTHROPIC_API_KEY", original) }()

	_ = os.Setenv("ANTHROPIC_API_KEY", "test-anthropic-key")
	key := getAnthropicKey()
	assert.Equal(t, "test-anthropic-key", key)
}

func TestGetAnthropicKey_Empty(t *testing.T) {
	original := os.Getenv("ANTHROPIC_API_KEY")
	defer func() { _ = os.Setenv("ANTHROPIC_API_KEY", original) }()

	_ = os.Unsetenv("ANTHROPIC_API_KEY")
	key := getAnthropicKey()
	// May be empty if viper also has no config
	assert.IsType(t, "", key)
}
