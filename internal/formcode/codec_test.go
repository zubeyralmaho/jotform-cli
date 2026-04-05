package formcode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadFile_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.json")
	err := os.WriteFile(path, []byte(`{"title":"Test","questions":{"1":{"type":"control_head"}}}`), 0644)
	require.NoError(t, err)

	schema, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "Test", schema["title"])
	assert.NotNil(t, schema["questions"])
}

func TestReadFile_YAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.yaml")
	content := "title: Test\nquestions:\n  \"1\":\n    type: control_head\n"
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	schema, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "Test", schema["title"])
}

func TestReadFile_YML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.yml")
	content := "title: YML Test\n"
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	schema, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "YML Test", schema["title"])
}

func TestReadFile_NotFound(t *testing.T) {
	_, err := ReadFile("/nonexistent/path/file.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read file")
}

func TestReadFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	err := os.WriteFile(path, []byte(`{invalid json`), 0644)
	require.NoError(t, err)

	_, err = ReadFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestReadFile_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	err := os.WriteFile(path, []byte(":\n  :\n    - :\n  invalid: ["), 0644)
	require.NoError(t, err)

	_, err = ReadFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid YAML")
}

func TestWriteFile_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	data := map[string]interface{}{"title": "Written"}
	err := WriteFile(path, data)
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), `"title"`)
	assert.Contains(t, string(content), `"Written"`)
	// JSON files should end with newline
	assert.Equal(t, byte('\n'), content[len(content)-1])
}

func TestWriteFile_YAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.yaml")

	data := map[string]interface{}{"title": "Written"}
	err := WriteFile(path, data)
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "title: Written")
}

func TestWriteFile_CreatesSubdirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "out.json")

	err := WriteFile(path, map[string]interface{}{"ok": true})
	require.NoError(t, err)

	_, err = os.Stat(path)
	require.NoError(t, err)
}

func TestWriteFile_ReadFile_Roundtrip_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip.json")

	original := map[string]interface{}{
		"title": "Roundtrip",
		"properties": map[string]interface{}{
			"title": "Roundtrip",
		},
	}

	err := WriteFile(path, original)
	require.NoError(t, err)

	loaded, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "Roundtrip", loaded["title"])
}

func TestWriteFile_ReadFile_Roundtrip_YAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip.yaml")

	original := map[string]interface{}{
		"title": "YAML Roundtrip",
	}

	err := WriteFile(path, original)
	require.NoError(t, err)

	loaded, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "YAML Roundtrip", loaded["title"])
}

func TestFormPropertiesToMap(t *testing.T) {
	type FP struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	result, err := FormPropertiesToMap(FP{ID: "123", Title: "Test"})
	require.NoError(t, err)
	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "Test", result["title"])
}

func TestFormPropertiesToMap_Map(t *testing.T) {
	input := map[string]interface{}{"id": "999", "title": "Map Input"}
	result, err := FormPropertiesToMap(input)
	require.NoError(t, err)
	assert.Equal(t, "999", result["id"])
}
