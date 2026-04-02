package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- SaveProject / LoadProject ----

func TestSaveProject_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{
		FormID: "123456",
		Name:   "my-form",
		Schema: "form.yaml",
	}

	require.NoError(t, SaveProject(cfg, dir))

	path := filepath.Join(dir, ProjectFileName)
	_, err := os.Stat(path)
	require.NoError(t, err, "project file should exist after SaveProject")
}

func TestSaveProject_FileContents(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{
		FormID: "999888",
		Name:   "test-form",
		Schema: "schema.yaml",
	}

	require.NoError(t, SaveProject(cfg, dir))

	content, err := os.ReadFile(filepath.Join(dir, ProjectFileName))
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "999888")
	assert.Contains(t, s, "test-form")
	assert.Contains(t, s, "schema.yaml")
	// Should have comment header
	assert.Contains(t, s, "# Jotform CLI")
}

func TestSaveProject_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")
	cfg := &ProjectConfig{FormID: "1"}

	require.NoError(t, SaveProject(cfg, nested))

	_, err := os.Stat(filepath.Join(nested, ProjectFileName))
	assert.NoError(t, err)
}

func TestSaveAndLoadProject_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	original := &ProjectConfig{
		FormID: "42424242",
		Name:   "round-trip",
		Schema: "form.json",
	}

	require.NoError(t, SaveProject(original, dir))

	// Change cwd to dir so LoadProject can find the file
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	loaded, err := LoadProject()
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, original.FormID, loaded.FormID)
	assert.Equal(t, original.Name, loaded.Name)
}

func TestLoadProject_NotFound(t *testing.T) {
	// Use a fresh empty temp dir with no .jotform.yaml
	dir := t.TempDir()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	cfg, err := LoadProject()
	require.NoError(t, err)
	assert.Nil(t, cfg, "LoadProject should return nil when no project file exists")
}

func TestLoadProject_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ProjectFileName)
	require.NoError(t, os.WriteFile(path, []byte("invalid: [yaml: content"), 0644))

	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	_, err = LoadProject()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestLoadProject_SchemaResolution(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{
		FormID: "1",
		Schema: "relative/schema.yaml",
	}
	require.NoError(t, SaveProject(cfg, dir))

	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	loaded, err := LoadProject()
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.True(t, filepath.IsAbs(loaded.Schema), "schema path should be resolved to absolute")
	assert.Equal(t, filepath.Join(dir, "relative/schema.yaml"), loaded.Schema)
}

// ---- ResolveFormID ----

func TestResolveFormID_FromArgs(t *testing.T) {
	id, err := ResolveFormID([]string{"123456789"})
	require.NoError(t, err)
	assert.Equal(t, "123456789", id)
}

func TestResolveFormID_FromProjectFile(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{FormID: "from-project"}
	require.NoError(t, SaveProject(cfg, dir))

	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	id, err := ResolveFormID([]string{})
	require.NoError(t, err)
	assert.Equal(t, "from-project", id)
}

func TestResolveFormID_NoArgNoProject(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	_, err = ResolveFormID([]string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "form ID required")
}

func TestResolveFormID_EmptyArgFallsThrough(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	_, err = ResolveFormID([]string{""})
	require.Error(t, err)
}

// ---- ResolveSchemaFile ----

func TestResolveSchemaFile_FromFlag(t *testing.T) {
	path, err := ResolveSchemaFile("/some/path/form.yaml")
	require.NoError(t, err)
	assert.Equal(t, "/some/path/form.yaml", path)
}

func TestResolveSchemaFile_FromProjectFile(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{FormID: "1", Schema: "form.yaml"}
	require.NoError(t, SaveProject(cfg, dir))

	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	path, err := ResolveSchemaFile("")
	require.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestResolveSchemaFile_NoFlagNoProject(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	_, err = ResolveSchemaFile("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema file required")
}

// ---- Slugify ----

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My Contact Form", "my-contact-form"},
		{"UPPERCASE TITLE", "uppercase-title"},
		{"  leading trailing  ", "leading-trailing"},
		{"hello---world", "hello-world"},
		{"form!@#$%special", "form-special"},
		{"", "form"},
		{"123 Numbers", "123-numbers"},
		{"already-slugified", "already-slugified"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Slugify(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSlugify_LongTitle(t *testing.T) {
	long := "this is a very long form title that exceeds the fifty character limit by a lot"
	result := Slugify(long)
	assert.LessOrEqual(t, len(result), 50)
	assert.NotEqual(t, '-', result[len(result)-1], "should not end with a hyphen")
}

func TestSlugify_OnlySpecialChars(t *testing.T) {
	result := Slugify("!@#$%^&*()")
	assert.Equal(t, "form", result)
}
