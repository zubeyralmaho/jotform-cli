package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyValue(t *testing.T) {
	result := KeyValue("Name", "John")
	assert.Contains(t, result, "Name")
	assert.Contains(t, result, "John")
}

func TestKeyValuePairs(t *testing.T) {
	pairs := [][2]string{
		{"Name", "John"},
		{"Email", "john@example.com"},
		{"Role", "Admin"},
	}

	result := KeyValuePairs(pairs)
	assert.Contains(t, result, "Name")
	assert.Contains(t, result, "John")
	assert.Contains(t, result, "Email")
	assert.Contains(t, result, "john@example.com")

	lines := strings.Split(result, "\n")
	assert.Len(t, lines, 3)
}

func TestKeyValuePairs_Empty(t *testing.T) {
	result := KeyValuePairs([][2]string{})
	assert.Empty(t, result)
}

func TestKeyValuePairs_SinglePair(t *testing.T) {
	result := KeyValuePairs([][2]string{{"Key", "Val"}})
	assert.Contains(t, result, "Key")
	assert.Contains(t, result, "Val")
}

func TestSeparator(t *testing.T) {
	sep := Separator(10)
	assert.NotEmpty(t, sep)
}

func TestSeparator_ZeroWidth(t *testing.T) {
	sep := Separator(0)
	// Zero-width separator produces empty string from Repeat
	assert.Empty(t, sep)
}

func TestSuccessBanner(t *testing.T) {
	banner := SuccessBanner("Operation complete!")
	assert.NotEmpty(t, banner)
	assert.Contains(t, banner, "Operation complete!")
}

func TestErrorBanner(t *testing.T) {
	banner := ErrorBanner("Something went wrong")
	assert.NotEmpty(t, banner)
	assert.Contains(t, banner, "Something went wrong")
}

func TestNewSpinner(t *testing.T) {
	s := NewSpinner()
	assert.NotNil(t, s)
}
