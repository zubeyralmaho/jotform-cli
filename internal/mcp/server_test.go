package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jotform/jotform-cli/internal/api"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAPI(t *testing.T, handler http.HandlerFunc) *api.Client {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c := api.New("test-key")
	c.BaseURL = ts.URL
	return c
}

func makeReq(args map[string]any) mcpgo.CallToolRequest {
	return mcpgo.CallToolRequest{
		Params: mcpgo.CallToolParams{
			Arguments: args,
		},
	}
}

func TestHandleListForms(t *testing.T) {
	client := newTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"responseCode": 200,
			"content": []map[string]any{
				{"id": "111", "title": "Form A"},
				{"id": "222", "title": "Form B"},
			},
		})
	})

	srv := New(client, nil)
	result, err := srv.handleListForms(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := mcpgo.GetTextFromContent(result.Content)
	assert.Contains(t, text, "Form A")
	assert.Contains(t, text, "Form B")
}

func TestHandleGetForm(t *testing.T) {
	client := newTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/form/999")
		json.NewEncoder(w).Encode(map[string]any{
			"responseCode": 200,
			"content": map[string]any{
				"id":    "999",
				"title": "My Form",
				"questions": map[string]any{
					"1": map[string]any{"type": "control_textbox", "text": "Name"},
				},
			},
		})
	})

	srv := New(client, nil)
	result, err := srv.handleGetForm(context.Background(), makeReq(map[string]any{"form_id": "999"}))
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := mcpgo.GetTextFromContent(result.Content)
	assert.Contains(t, text, "My Form")
	assert.Contains(t, text, "control_textbox")
}

func TestHandleGetForm_MissingID(t *testing.T) {
	client := newTestAPI(t, func(w http.ResponseWriter, r *http.Request) {})
	srv := New(client, nil)
	result, err := srv.handleGetForm(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestHandleCreateForm(t *testing.T) {
	client := newTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"responseCode": 200,
			"content": map[string]any{
				"id":    "555",
				"title": "New Form",
				"url":   "https://form.jotform.com/555",
			},
		})
	})

	schema := `{"questions":{"1":{"type":"control_textbox","text":"Name"}},"properties":{"title":"New Form"}}`
	srv := New(client, nil)
	result, err := srv.handleCreateForm(context.Background(), makeReq(map[string]any{"schema": schema}))
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := mcpgo.GetTextFromContent(result.Content)
	assert.Contains(t, text, "555")
	assert.Contains(t, text, "New Form")
}

func TestHandleCreateForm_InvalidJSON(t *testing.T) {
	client := newTestAPI(t, func(w http.ResponseWriter, r *http.Request) {})
	srv := New(client, nil)
	result, err := srv.handleCreateForm(context.Background(), makeReq(map[string]any{"schema": "not json"}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestHandleListSubmissions(t *testing.T) {
	client := newTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/form/123/submissions")
		json.NewEncoder(w).Encode(map[string]any{
			"responseCode": 200,
			"content": []map[string]any{
				{"id": "s1", "created_at": "2025-01-01"},
			},
		})
	})

	srv := New(client, nil)
	result, err := srv.handleListSubmissions(context.Background(), makeReq(map[string]any{"form_id": "123", "limit": float64(10)}))
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := mcpgo.GetTextFromContent(result.Content)
	assert.Contains(t, text, "s1")
}
