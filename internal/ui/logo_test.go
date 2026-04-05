package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildLogo_NotEmpty(t *testing.T) {
	logo := buildLogo()
	assert.NotEmpty(t, logo)
}

func TestLogo_NotEmpty(t *testing.T) {
	logo := Logo()
	assert.NotEmpty(t, logo)
}

func TestLogoCompact_NotEmpty(t *testing.T) {
	compact := LogoCompact()
	assert.NotEmpty(t, compact)
}

func TestLogoWithText_NotEmpty(t *testing.T) {
	full := LogoWithText()
	assert.NotEmpty(t, full)
}

func TestLogoCompact_ContainsSymbols(t *testing.T) {
	compact := LogoCompact()
	assert.Contains(t, compact, "▰")
}
