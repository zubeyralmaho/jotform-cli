package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type Generator struct {
	client *anthropic.Client
	model  anthropic.Model
}

func NewGenerator(apiKey string, opts ...option.RequestOption) *Generator {
	allOpts := append([]option.RequestOption{option.WithAPIKey(apiKey)}, opts...)
	client := anthropic.NewClient(allOpts...)
	return &Generator{
		client: &client,
		model:  anthropic.ModelClaudeSonnet4_5,
	}
}

// GenerateSchema converts a natural language prompt into a Jotform JSON schema.
func (g *Generator) GenerateSchema(ctx context.Context, prompt string) (map[string]interface{}, error) {
	msg, err := g.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     g.model,
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: "You are a Jotform form architect. " + jotformSchemaContract},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock(fmt.Sprintf("Create a Jotform schema for: %s", prompt)),
			),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("claude API error: %w", err)
	}

	raw := extractText(msg)
	// Strip markdown fences if present
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		return nil, fmt.Errorf("claude returned invalid JSON: %w\nRaw: %s", err, raw)
	}
	return schema, nil
}

// AnalyzeForm sends a form structure to Claude for UX review.
func (g *Generator) AnalyzeForm(ctx context.Context, form map[string]interface{}) (string, error) {
	formJSON, _ := json.MarshalIndent(form, "", "  ")

	msg, err := g.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     g.model,
		MaxTokens: 2048,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock(fmt.Sprintf(
					"Analyze this Jotform structure for UX issues, confusing labels, missing fields, "+
						"and completion-rate problems. Be concise and actionable.\n\n%s",
					formJSON,
				)),
			),
		},
	})
	if err != nil {
		return "", fmt.Errorf("claude API error: %w", err)
	}
	return extractText(msg), nil
}

func extractText(msg *anthropic.Message) string {
	var parts []string
	for _, block := range msg.Content {
		if block.Type == "text" {
			parts = append(parts, block.Text)
		}
	}
	return strings.Join(parts, "\n")
}
