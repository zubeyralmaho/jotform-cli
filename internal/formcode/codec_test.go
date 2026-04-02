package formcode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- ReadFile ----

func TestReadFile_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "form.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"title":"My Form","id":"1"}`), 0644))

	schema, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "My Form", schema["title"])
	assert.Equal(t, "1", schema["id"])
}

func TestReadFile_YAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "form.yaml")
	require.NoError(t, os.WriteFile(path, []byte("title: My Form\nid: \"1\"\n"), 0644))

	schema, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "My Form", schema["title"])
}

func TestReadFile_YML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "form.yml")
	require.NoError(t, os.WriteFile(path, []byte("title: YML Form\n"), 0644))

	schema, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "YML Form", schema["title"])
}

func TestReadFile_NotFound(t *testing.T) {
	_, err := ReadFile("/nonexistent/path/form.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read file")
}

func TestReadFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	require.NoError(t, os.WriteFile(path, []byte(`{invalid json`), 0644))

	_, err := ReadFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestReadFile_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	require.NoError(t, os.WriteFile(path, []byte("key: [unclosed"), 0644))

	_, err := ReadFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid YAML")
}

// ---- WriteFile ----

func TestWriteFile_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	data := map[string]interface{}{"title": "Test", "id": "42"}

	require.NoError(t, WriteFile(path, data))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), `"title"`)
	assert.Contains(t, string(content), `"Test"`)
	// JSON output should end with a newline
	assert.Equal(t, byte('\n'), content[len(content)-1])
}

func TestWriteFile_YAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.yaml")
	data := map[string]interface{}{"title": "Test", "id": "42"}

	require.NoError(t, WriteFile(path, data))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "title:")
	assert.Contains(t, string(content), "Test")
}

func TestWriteFile_YML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.yml")
	data := map[string]interface{}{"title": "YML Output"}

	require.NoError(t, WriteFile(path, data))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "title:")
}

func TestWriteFile_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "form.json")
	data := map[string]interface{}{"title": "Nested"}

	require.NoError(t, WriteFile(path, data))

	_, err := os.Stat(path)
	assert.NoError(t, err, "file should exist after WriteFile creates parent dirs")
}

func TestWriteFile_RoundTrip_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "form.json")
	original := map[string]interface{}{
		"title": "Round Trip Form",
		"id":    "999",
	}

	require.NoError(t, WriteFile(path, original))
	loaded, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, original["title"], loaded["title"])
	assert.Equal(t, original["id"], loaded["id"])
}

func TestWriteFile_RoundTrip_YAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "form.yaml")
	original := map[string]interface{}{
		"title": "YAML Round Trip",
		"id":    "777",
	}

	require.NoError(t, WriteFile(path, original))
	loaded, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, original["title"], loaded["title"])
}

// ---- FormPropertiesToMap ----

func TestFormPropertiesToMap_Basic(t *testing.T) {
	fp := FormProperties{
		ID:    "123",
		Title: "Contact Form",
		Questions: map[string]interface{}{
			"1": map[string]interface{}{"text": "Name"},
		},
		Properties: map[string]interface{}{
			"font": "Arial",
		},
	}

	m, err := FormPropertiesToMap(fp)
	require.NoError(t, err)
	assert.Equal(t, "123", m["id"])
	assert.Equal(t, "Contact Form", m["title"])
	assert.NotNil(t, m["questions"])
	assert.NotNil(t, m["properties"])
}

func TestFormPropertiesToMap_Empty(t *testing.T) {
	fp := FormProperties{}
	m, err := FormPropertiesToMap(fp)
	require.NoError(t, err)
	assert.NotNil(t, m)
}

func TestFormPropertiesToMap_NilQuestionsAndProperties(t *testing.T) {
	fp := FormProperties{ID: "1", Title: "Simple"}
	m, err := FormPropertiesToMap(fp)
	require.NoError(t, err)
	assert.Equal(t, "1", m["id"])
}
