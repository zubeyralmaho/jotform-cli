# Testing Strategy

## Layers

```
Unit Tests        → internal/api, internal/auth, internal/ai, internal/output
Integration Tests → cmd/ (uses real or stub HTTP server)
E2E Tests         → Jotform sandbox account + real API key
```

---

## 1. Unit Tests

### API Client — Mock HTTP Server

```go
// internal/api/client_test.go
func TestListForms(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Contains(t, r.URL.RawQuery, "apiKey=test-key")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "responseCode": 200,
            "content": []map[string]interface{}{
                {"id": "123", "title": "Test Form"},
            },
        })
    }))
    defer ts.Close()

    client := api.New("test-key")
    client.BaseURL = ts.URL

    forms, err := client.ListForms(0, 10)
    require.NoError(t, err)
    assert.Len(t, forms, 1)
    assert.Equal(t, "Test Form", forms[0].Title)
}
```

### AI Generator — Mocked Claude Response

```go
// internal/ai/generator_test.go
func TestGenerateSchema_ValidJSON(t *testing.T) {
    // Use a test server to mock Anthropic API
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "content": []map[string]interface{}{
                {"type": "text", "text": `{"questions":{"1":{"type":"control_textbox","text":"Name","order":"1"}},"properties":{"title":"Test"}}`},
            },
        })
    }))
    defer ts.Close()

    // ... inject test server URL into generator
    schema, err := gen.GenerateSchema(context.Background(), "simple name form")
    require.NoError(t, err)
    assert.Contains(t, schema, "questions")
}
```

---

## 2. Output Formatter Tests

```go
func TestPrintForms_TableFormat(t *testing.T) {
    forms := []api.Form{
        {ID: "1", Title: "Survey", Status: "ENABLED"},
    }
    var buf bytes.Buffer
    err := output.PrintTo(&buf, forms, output.FormatTable)
    require.NoError(t, err)
    assert.Contains(t, buf.String(), "Survey")
    assert.Contains(t, buf.String(), "ENABLED")
}

func TestPrintForms_JSONFormat(t *testing.T) {
    forms := []api.Form{{ID: "1", Title: "Survey"}}
    var buf bytes.Buffer
    output.PrintTo(&buf, forms, output.FormatJSON)
    var result []map[string]interface{}
    json.Unmarshal(buf.Bytes(), &result)
    assert.Equal(t, "Survey", result[0]["title"])
}
```

---

## 3. Integration Tests (Command Layer)

Use `cobra`'s built-in test helpers:

```go
func TestFormsListCommand(t *testing.T) {
    ts := newMockJotformServer(t)
    t.Setenv("JOTFORM_API_KEY", "test-key")
    t.Setenv("JOTFORM_BASE_URL", ts.URL)

    root := newRootCmd() // factory fn for testable root
    root.SetArgs([]string{"forms", "list", "--output", "json"})
    var out bytes.Buffer
    root.SetOut(&out)

    err := root.Execute()
    require.NoError(t, err)

    var forms []map[string]interface{}
    json.Unmarshal(out.Bytes(), &forms)
    assert.NotEmpty(t, forms)
}
```

---

## 4. MCP Tool Tests

```go
func TestMCPListFormsHandler(t *testing.T) {
    apiClient := &api.Client{...} // mock
    srv := mcp.New(apiClient, nil)

    req := mcpgo.CallToolRequest{
        Params: mcpgo.CallToolRequestParams{
            Name: "list_forms",
        },
    }
    result, err := srv.HandleTool(context.Background(), req)
    require.NoError(t, err)
    assert.False(t, result.IsError)
}
```

---

## 5. E2E Tests (Jotform Sandbox)

Tag with `//go:build e2e` so they only run when explicitly requested:

```go
//go:build e2e

func TestE2E_CreateAndDeleteForm(t *testing.T) {
    key := os.Getenv("JOTFORM_API_KEY")
    require.NotEmpty(t, key, "JOTFORM_API_KEY required for e2e tests")

    client := api.New(key)

    schema := map[string]interface{}{
        "questions": map[string]interface{}{
            "1": map[string]interface{}{
                "type": "control_textbox",
                "text": "Test Question",
                "order": "1",
            },
        },
        "properties": map[string]interface{}{
            "title": "E2E Test Form",
        },
    }

    form, err := client.CreateForm(schema)
    require.NoError(t, err)
    assert.NotEmpty(t, form.ID)

    // Cleanup
    t.Cleanup(func() {
        client.DeleteForm(form.ID)
    })
}
```

Run with:
```bash
JOTFORM_API_KEY=xxx go test ./... -tags e2e -run TestE2E
```

---

## 6. Coverage Targets

| Package | Target |
|---|---|
| `internal/api` | ≥ 80% |
| `internal/auth` | ≥ 70% |
| `internal/output` | ≥ 90% |
| `internal/mcp` | ≥ 75% |
| `cmd/` | ≥ 60% |

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```
