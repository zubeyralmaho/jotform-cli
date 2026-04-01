package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockAnthropicServer(t *testing.T, responseText string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/messages", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":   "msg_test",
			"type": "message",
			"role": "assistant",
			"content": []map[string]any{
				{"type": "text", "text": responseText},
			},
			"model":         "claude-sonnet-4-5",
			"stop_reason":   "end_turn",
			"stop_sequence": nil,
			"usage":         map[string]any{"input_tokens": 10, "output_tokens": 50},
		})
	}))
	t.Cleanup(ts.Close)
	return ts
}

func newTestGenerator(serverURL string) *Generator {
	return NewGenerator("test-key", option.WithBaseURL(serverURL))
}

func TestGenerateSchema_ValidJSON(t *testing.T) {
	schema := `{"questions":{"1":{"type":"control_head","text":"Feedback","order":"1"}},"properties":{"title":"Feedback Form"}}`
	ts := newMockAnthropicServer(t, schema)
	gen := newTestGenerator(ts.URL)

	result, err := gen.GenerateSchema(context.Background(), "a feedback form")
	require.NoError(t, err)
	assert.Contains(t, result.Schema, "questions")
	assert.Contains(t, result.Schema, "properties")
	assert.Greater(t, result.Usage.InputTokens, 0)

	props, ok := result.Schema["properties"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Feedback Form", props["title"])
}

func TestGenerateSchema_StripsMarkdownFences(t *testing.T) {
	schema := "```json\n{\"questions\":{},\"properties\":{\"title\":\"Test\"}}\n```"
	ts := newMockAnthropicServer(t, schema)
	gen := newTestGenerator(ts.URL)

	result, err := gen.GenerateSchema(context.Background(), "test form")
	require.NoError(t, err)
	assert.Contains(t, result.Schema, "properties")
}

func TestGenerateSchema_InvalidJSON(t *testing.T) {
	ts := newMockAnthropicServer(t, "this is not json at all")
	gen := newTestGenerator(ts.URL)

	_, err := gen.GenerateSchema(context.Background(), "something")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestAnalyzeForm(t *testing.T) {
	analysis := "## Issues\n1. Missing email validation\n2. Too many required fields"
	ts := newMockAnthropicServer(t, analysis)
	gen := newTestGenerator(ts.URL)

	form := map[string]any{
		"title": "Test Form",
		"questions": map[string]any{
			"1": map[string]any{"type": "control_textbox", "text": "Name"},
		},
	}

	result, usage, err := gen.AnalyzeForm(context.Background(), form)
	require.NoError(t, err)
	assert.Contains(t, result, "Missing email validation")
	assert.Contains(t, result, "required fields")
	assert.NotNil(t, usage)
	assert.Greater(t, usage.InputTokens, 0)
}
