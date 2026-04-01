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
	apiClient   *api.Client
	aiGenerator *ai.Generator
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

	return server.ServeStdio(srv)
}

func (s *Server) registerTools(srv *server.MCPServer) {
	srv.AddTool(mcpgo.NewTool("list_forms",
		mcpgo.WithDescription("List all Jotform forms for the authenticated account"),
	), s.handleListForms)

	srv.AddTool(mcpgo.NewTool("get_form",
		mcpgo.WithDescription("Get the full structure (questions, properties) of a Jotform form"),
		mcpgo.WithString("form_id",
			mcpgo.Required(),
			mcpgo.Description("The numeric Jotform form ID"),
		),
	), s.handleGetForm)

	srv.AddTool(mcpgo.NewTool("create_form",
		mcpgo.WithDescription("Create a new Jotform form from a JSON schema"),
		mcpgo.WithString("schema",
			mcpgo.Required(),
			mcpgo.Description("JSON string containing the Jotform form schema with questions and properties"),
		),
	), s.handleCreateForm)

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

	if s.aiGenerator != nil {
		srv.AddTool(mcpgo.NewTool("generate_form_schema",
			mcpgo.WithDescription("Generate a Jotform form schema from a natural language description using AI"),
			mcpgo.WithString("prompt",
				mcpgo.Required(),
				mcpgo.Description("Natural language description of the form to create"),
			),
		), s.handleGenerateSchema)
	}
}

func (s *Server) handleListForms(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	forms, err := s.apiClient.ListForms(0, 50)
	if err != nil {
		return mcpgo.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(forms, "", "  ")
	return mcpgo.NewToolResultText(string(data)), nil
}

func (s *Server) handleGetForm(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	formID := req.GetString("form_id", "")
	if formID == "" {
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
	schemaStr := req.GetString("schema", "")
	if schemaStr == "" {
		return mcpgo.NewToolResultError("schema is required"), nil
	}
	var schema map[string]any
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
	formID := req.GetString("form_id", "")
	if formID == "" {
		return mcpgo.NewToolResultError("form_id is required"), nil
	}
	limit := req.GetInt("limit", 20)
	subs, err := s.apiClient.GetSubmissions(formID, 0, limit, "created_at", "DESC")
	if err != nil {
		return mcpgo.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(subs, "", "  ")
	return mcpgo.NewToolResultText(string(data)), nil
}

func (s *Server) handleGenerateSchema(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	prompt := req.GetString("prompt", "")
	if prompt == "" {
		return mcpgo.NewToolResultError("prompt is required"), nil
	}
	schema, err := s.aiGenerator.GenerateSchema(ctx, prompt)
	if err != nil {
		return mcpgo.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(schema, "", "  ")
	return mcpgo.NewToolResultText(string(data)), nil
}
