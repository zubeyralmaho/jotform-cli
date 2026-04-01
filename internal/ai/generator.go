package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// GeneratorOptions configures the AI generator behavior.
type GeneratorOptions struct {
	Model      string        // Anthropic model name (default: claude-sonnet-4-5-20250514)
	MaxTokens  int           // Max output tokens (default: 4096)
	Timeout    time.Duration // Per-request timeout (default: 60s)
	MaxRetries int           // Number of retries on transient errors (default: 2)
}

// DefaultOptions returns sensible production defaults.
func DefaultOptions() GeneratorOptions {
	return GeneratorOptions{
		Model:      string(anthropic.ModelClaudeSonnet4_5),
		MaxTokens:  4096,
		Timeout:    60 * time.Second,
		MaxRetries: 2,
	}
}

// Usage tracks token consumption for cost visibility.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// GenerateResult wraps a schema generation outcome with usage stats.
type GenerateResult struct {
	Schema map[string]interface{} `json:"schema"`
	Usage  Usage                  `json:"usage"`
}

type Generator struct {
	client  *anthropic.Client
	options GeneratorOptions
}

func NewGenerator(apiKey string, opts ...option.RequestOption) *Generator {
	return NewGeneratorWithOptions(apiKey, DefaultOptions(), opts...)
}

func NewGeneratorWithOptions(apiKey string, genOpts GeneratorOptions, opts ...option.RequestOption) *Generator {
	allOpts := append([]option.RequestOption{option.WithAPIKey(apiKey)}, opts...)
	client := anthropic.NewClient(allOpts...)

	// Apply defaults for zero values
	if genOpts.Model == "" {
		genOpts.Model = string(anthropic.ModelClaudeSonnet4_5)
	}
	if genOpts.MaxTokens <= 0 {
		genOpts.MaxTokens = 4096
	}
	if genOpts.Timeout <= 0 {
		genOpts.Timeout = 60 * time.Second
	}
	if genOpts.MaxRetries < 0 {
		genOpts.MaxRetries = 0
	}

	return &Generator{
		client:  &client,
		options: genOpts,
	}
}

// GenerateSchema converts a natural language prompt into a Jotform JSON schema.
func (g *Generator) GenerateSchema(ctx context.Context, prompt string) (*GenerateResult, error) {
	msg, err := g.callWithRetry(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(g.options.Model),
		MaxTokens: int64(g.options.MaxTokens),
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

	return &GenerateResult{
		Schema: schema,
		Usage: Usage{
			InputTokens:  int(msg.Usage.InputTokens),
			OutputTokens: int(msg.Usage.OutputTokens),
		},
	}, nil
}

// AnalyzeForm sends a form structure to Claude for UX review.
func (g *Generator) AnalyzeForm(ctx context.Context, form map[string]interface{}) (string, *Usage, error) {
	formJSON, _ := json.MarshalIndent(form, "", "  ")

	msg, err := g.callWithRetry(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(g.options.Model),
		MaxTokens: int64(g.options.MaxTokens),
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
		return "", nil, fmt.Errorf("claude API error: %w", err)
	}

	usage := &Usage{
		InputTokens:  int(msg.Usage.InputTokens),
		OutputTokens: int(msg.Usage.OutputTokens),
	}
	return extractText(msg), usage, nil
}

// callWithRetry wraps the Anthropic Messages.New call with timeout and exponential backoff.
func (g *Generator) callWithRetry(ctx context.Context, params anthropic.MessageNewParams) (*anthropic.Message, error) {
	var lastErr error
	attempts := g.options.MaxRetries + 1

	for attempt := range attempts {
		callCtx, cancel := context.WithTimeout(ctx, g.options.Timeout)
		msg, err := g.client.Messages.New(callCtx, params)
		cancel()

		if err == nil {
			return msg, nil
		}
		lastErr = err

		// Don't retry on the last attempt
		if attempt < attempts-1 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
	return nil, lastErr
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
