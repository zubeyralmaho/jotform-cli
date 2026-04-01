package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type row struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func TestPrintTo_JSON(t *testing.T) {
	data := []row{{ID: "1", Title: "Test"}}
	var buf bytes.Buffer
	err := PrintTo(&buf, data, FormatJSON)
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "1", result[0]["id"])
	assert.Equal(t, "Test", result[0]["title"])
}

func TestPrintTo_YAML(t *testing.T) {
	data := []row{{ID: "1", Title: "Test"}}
	var buf bytes.Buffer
	err := PrintTo(&buf, data, FormatYAML)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "id: \"1\"")
	assert.Contains(t, buf.String(), "title: Test")
}

func TestPrintTo_Table(t *testing.T) {
	data := []row{{ID: "1", Title: "Survey"}}
	var buf bytes.Buffer
	err := PrintTo(&buf, data, FormatTable)
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, strings.ToUpper(out), "TITLE")
	assert.Contains(t, out, "Survey")
}

func TestPrintTo_Table_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := PrintTo(&buf, []row{}, FormatTable)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "no results")
}
