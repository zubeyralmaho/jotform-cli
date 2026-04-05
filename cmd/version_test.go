package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionVariables(t *testing.T) {
	assert.NotEmpty(t, Version)
	assert.NotEmpty(t, Commit)
	assert.NotEmpty(t, BuildDate)
}

func TestVersionCmd_Exists(t *testing.T) {
	assert.NotNil(t, versionCmd)
	assert.Equal(t, "version", versionCmd.Use)
	assert.NotEmpty(t, versionCmd.Short)
}
