# Phase 3: MCP Server (`jotform mcp start-server`)

## Goal

Expose Jotform operations as **MCP tools** so Claude Desktop, Claude Code, and other MCP-compatible agents can use Jotform natively — no shell commands required.

---

## 1. Add the MCP Library

```bash
go get github.com/mark3labs/mcp-go@latest
```

---

## 2. Tool Definitions

We expose 5 tools to the MCP host:

| Tool name | Description |
|---|---|
| `list_forms` | Returns all forms for the authenticated account |
| `get_form` | Returns full question structure of a specific form |
| `create_form` | Creates a form from a JSON schema |
| `list_submissions` | Returns recent submissions for a form |
| `generate_form_schema` | Generates a Jotform schema from a natural language prompt (AI) |

---

## 3. MCP Server (`internal/mcp/server.go`)

```go
package mcp

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/jotform/jotform-cli/internal/ai"
    "github.com/jotform/jotform-cli/internal/api"
    mcpgo "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

type Server struct {
    apiClient    *api.Client
    aiGenerator  *ai.Generator
}

func New(apiClient *api.Client, aiGenerator *ai.Generator) *Server {
    return &Server{
        apiClient:   apiClient,
        aiGenerator: aiGenerator,
    }
}

func (s *Server) Run() error {
    srv := server.NewMCPServer(
        "Jotform",
        "1.0.0",
        server.WithToolCapabilities(true),
    )

    s.registerTools(srv)

    // stdio transport — compatible with Claude Desktop and Claude Code
    return server.ServeStdio(srv)
}

func (s *Server) registerTools(srv *server.MCPServer) {
    // --- list_forms ---
    srv.AddTool(mcpgo.NewTool("list_forms",
        mcpgo.WithDescription("List all Jotform forms for the authenticated account"),
    ), s.handleListForms)

    // --- get_form ---
    srv.AddTool(mcpgo.NewTool("get_form",
        mcpgo.WithDescription("Get the full structure of a Jotform form"),
        mcpgo.WithString("form_id",
            mcpgo.Required(),
            mcpgo.Description("The numeric Jotform form ID"),
        ),
    ), s.handleGetForm)

    // --- create_form ---
    srv.AddTool(mcpgo.NewTool("create_form",
        mcpgo.WithDescription("Create a new Jotform form from a JSON schema"),
        mcpgo.WithString("schema",
            mcpgo.Required(),
            mcpgo.Description("JSON string containing the Jotform form schema"),
        ),
    ), s.handleCreateForm)

    // --- list_submissions ---
    srv.AddTool(mcpgo.NewTool("list_submissions",
        mcpgo.WithDescription("Retrieve recent submissions for a Jotform form"),
        mcpgo.WithString("form_id",
            mcpgo.Required(),
            mcpgo.Description("The numeric Jotform form ID"),
        ),
        mcpgo.WithNumber("limit",
            mcpgo.Description("Maximum number of submissions to return (default: 20)"),
        ),
    ), s.handleListSubmissions)

    // --- generate_form_schema ---
    srv.AddTool(mcpgo.NewTool("generate_form_schema",
        mcpgo.WithDescription("Generate a Jotform form schema from a natural language description"),
        mcpgo.WithString("prompt",
            mcpgo.Required(),
            mcpgo.Description("Natural language description of the form you want to create"),
        ),
    ), s.handleGenerateSchema)
}
```

---

## 4. Tool Handlers (`internal/mcp/handlers.go`)

```go
package mcp

import (
    "context"
    "encoding/json"
    "fmt"

    mcpgo "github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleListForms(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
    forms, err := s.apiClient.ListForms(0, 50)
    if err != nil {
        return mcpgo.NewToolResultError(err.Error()), nil
    }
    data, _ := json.MarshalIndent(forms, "", "  ")
    return mcpgo.NewToolResultText(string(data)), nil
}

func (s *Server) handleGetForm(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
    formID, ok := req.Params.Arguments["form_id"].(string)
    if !ok || formID == "" {
        return mcpgo.NewToolResultError("form_id is required"), nil
    }
    form, err := s.apiClient.GetForm(formID)
    if err != nil {
        return mcpgo.NewToolResultError(err.Error()), nil
    }
    data, _ := json.MarshalIndent(form, "", "  ")
    return mcpgo.NewToolResultText(string(data)), nil
}

func (s *Server) handleCreateForm(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
    schemaStr, ok := req.Params.Arguments["schema"].(string)
    if !ok {
        return mcpgo.NewToolResultError("schema is required"), nil
    }
    var schema map[string]interface{}
    if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
        return mcpgo.NewToolResultError(fmt.Sprintf("invalid JSON schema: %s", err)), nil
    }
    form, err := s.apiClient.CreateForm(schema)
    if err != nil {
        return mcpgo.NewToolResultError(err.Error()), nil
    }
    result := fmt.Sprintf("Form created successfully!\nID: %s\nTitle: %s\nURL: %s", form.ID, form.Title, form.URL)
    return mcpgo.NewToolResultText(result), nil
}

func (s *Server) handleListSubmissions(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
    formID, _ := req.Params.Arguments["form_id"].(string)
    limit := 20
    if l, ok := req.Params.Arguments["limit"].(float64); ok {
        limit = int(l)
    }
    subs, err := s.apiClient.GetSubmissions(formID, 0, limit, "created_at", "DESC")
    if err != nil {
        return mcpgo.NewToolResultError(err.Error()), nil
    }
    data, _ := json.MarshalIndent(subs, "", "  ")
    return mcpgo.NewToolResultText(string(data)), nil
}

func (s *Server) handleGenerateSchema(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
    prompt, _ := req.Params.Arguments["prompt"].(string)
    schema, err := s.aiGenerator.GenerateSchema(ctx, prompt)
    if err != nil {
        return mcpgo.NewToolResultError(err.Error()), nil
    }
    data, _ := json.MarshalIndent(schema, "", "  ")
    return mcpgo.NewToolResultText(string(data)), nil
}
```

---

## 5. MCP Command (`cmd/mcp.go`)

```go
package cmd

import (
    "github.com/jotform/jotform-cli/internal/ai"
    "github.com/jotform/jotform-cli/internal/mcp"
    "github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
    Use:   "mcp",
    Short: "Model Context Protocol server",
}

var mcpStartCmd = &cobra.Command{
    Use:   "start-server",
    Short: "Start the MCP server (stdio transport)",
    RunE: func(cmd *cobra.Command, args []string) error {
        apiClient, err := newClient()
        if err != nil {
            return err
        }
        aiGen := ai.NewGenerator(getAnthropicKey())
        srv := mcp.New(apiClient, aiGen)
        return srv.Run()
    },
}

func init() {
    mcpCmd.AddCommand(mcpStartCmd)
    rootCmd.AddCommand(mcpCmd)
}
```

---

## 6. Connecting to Claude Desktop

Add to `~/.config/claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "jotform": {
      "command": "/usr/local/bin/jotform",
      "args": ["mcp", "start-server"],
      "env": {
        "JOTFORM_API_KEY": "your-api-key-here",
        "ANTHROPIC_API_KEY": "sk-ant-..."
      }
    }
  }
}
```

After restarting Claude Desktop, you'll see Jotform tools available in the tool picker.

---

## 6. Connecting to Claude Code

```bash
# Add to your project's .mcp.json
{
  "mcpServers": {
    "jotform": {
      "command": "jotform",
      "args": ["mcp", "start-server"]
    }
  }
}
```

Or add globally via Claude Code settings.

---

## 7. Agent Workflow Example

With the MCP server running, a Claude agent can:

```
User: "Create a customer satisfaction survey with NPS score and a comment box"

Claude: [calls generate_form_schema with that prompt]
        [receives valid Jotform JSON]
        [calls create_form with the schema]
        → "Created form 'Customer Satisfaction Survey' at https://form.jotform.com/..."
```

---

## Acceptance Criteria

- [ ] `jotform mcp start-server` starts without error and waits on stdin
- [ ] Claude Desktop can see `list_forms`, `get_form`, `create_form`, `list_submissions`, `generate_form_schema`
- [ ] `create_form` round-trip: generate schema → create form → verify in Jotform dashboard
- [ ] All tool errors return structured `error` responses (not panics)

---

## Next Step

→ [06-phase4-release.md](06-phase4-release.md)
