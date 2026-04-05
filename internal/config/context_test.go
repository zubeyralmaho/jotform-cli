package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "Contact Form", "contact-form"},
		{"special chars", "My Form!!! @#$%", "my-form"},
		{"numbers", "Form 123", "form-123"},
		{"already slug", "my-form", "my-form"},
		{"extra spaces", "  Spaced  Out  ", "spaced-out"},
		{"empty", "", "form"},
		{"only special", "!@#$%", "form"},
		{"long title", "This is a really long form title that should be truncated at fifty characters limit here", "this-is-a-really-long-form-title-that-should-be-tr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len(result), 50)
		})
	}
}

func TestSaveProject(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{
		FormID: "12345",
		Name:   "test-form",
		Schema: "schema.json",
	}

	err := SaveProject(cfg, dir)
	require.NoError(t, err)

	path := filepath.Join(dir, ProjectFileName)
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "form_id: \"12345\"")
	assert.Contains(t, content, "name: test-form")
	assert.Contains(t, content, "schema: schema.json")
	assert.Contains(t, content, "# Jotform CLI project configuration")
}

func TestSaveProject_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub", "dir")
	cfg := &ProjectConfig{FormID: "111"}

	err := SaveProject(cfg, dir)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, ProjectFileName))
	require.NoError(t, err)
}

func TestResolveFormID_FromArgs(t *testing.T) {
	id, err := ResolveFormID([]string{"12345"})
	require.NoError(t, err)
	assert.Equal(t, "12345", id)
}

func TestResolveFormID_EmptyArgs(t *testing.T) {
	// Change to a temp dir that has no .jotform.yaml
	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()
	_ = os.Chdir(t.TempDir())

	_, err := ResolveFormID([]string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "form ID required")
}

func TestResolveFormID_FromProjectConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{FormID: "99999", Name: "test"}
	err := SaveProject(cfg, dir)
	require.NoError(t, err)

	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()
	_ = os.Chdir(dir)

	id, err := ResolveFormID([]string{})
	require.NoError(t, err)
	assert.Equal(t, "99999", id)
}

func TestResolveSchemaFile_FromFlag(t *testing.T) {
	path, err := ResolveSchemaFile("/path/to/schema.json")
	require.NoError(t, err)
	assert.Equal(t, "/path/to/schema.json", path)
}

func TestResolveSchemaFile_NoFlag(t *testing.T) {
	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()
	_ = os.Chdir(t.TempDir())

	_, err := ResolveSchemaFile("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema file required")
}

func TestResolveSchemaFile_FromProjectConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{FormID: "111", Schema: "form.yaml"}
	err := SaveProject(cfg, dir)
	require.NoError(t, err)

	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()
	_ = os.Chdir(dir)

	path, err := ResolveSchemaFile("")
	require.NoError(t, err)
	assert.Contains(t, path, "form.yaml")
}

func TestLoadProject_NoFile(t *testing.T) {
	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()
	_ = os.Chdir(t.TempDir())

	cfg, err := LoadProject()
	assert.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestLoadProject_WithFile(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{FormID: "777", Name: "loaded", Schema: "schema.json"}
	err := SaveProject(cfg, dir)
	require.NoError(t, err)

	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()
	_ = os.Chdir(dir)

	loaded, err := LoadProject()
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, "777", loaded.FormID)
	assert.Equal(t, "loaded", loaded.Name)
	// Schema should be resolved to absolute path
	assert.True(t, filepath.IsAbs(loaded.Schema))
}

func TestLoadProject_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ProjectFileName)
	err := os.WriteFile(path, []byte(":\n  invalid: [yaml"), 0644)
	require.NoError(t, err)

	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()
	_ = os.Chdir(dir)

	_, err = LoadProject()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}
