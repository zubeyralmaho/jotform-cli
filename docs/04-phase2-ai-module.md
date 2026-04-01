# Phase 2: AI Module (`jotform ai`)

## Goal

- `jotform ai generate-schema "prompt"` → valid Jotform JSON schema
- `jotform ai analyze [form-id]` → LLM-powered UX suggestions

Uses the **Claude API** (claude-sonnet-4-6 by default).

---

## 1. Add the Anthropic SDK

```bash
go get github.com/anthropics/anthropic-sdk-go@latest
```

---

## 2. Jotform Schema Contract

Before prompting Claude, we need a stable schema definition to include in the system prompt.

**internal/ai/schema_contract.go**

```go
package ai

// jotformSchemaContract is injected into every generate-schema prompt
// so Claude knows exactly what structure Jotform's API expects.
const jotformSchemaContract = `
A valid Jotform form schema is a JSON object with this structure:

{
  "questions": {
    "<order>": {
      "type": "<field_type>",
      "text": "<label>",
      "name": "<snake_case_name>",
      "order": "<integer_as_string>",
      "required": "No" | "Yes"
    }
  },
  "properties": {
    "title": "<form title>"
  }
}

Supported field types: control_textbox, control_textarea, control_email,
control_phone, control_number, control_radio, control_checkbox,
control_dropdown, control_fileupload, control_date, control_address.

For radio/checkbox/dropdown, add an "answers" map:
  "answers": { "1": {"text": "Option A"}, "2": {"text": "Option B"} }

Return ONLY the JSON object, no markdown fences, no explanation.
`
```

---

## 3. Schema Generator (`internal/ai/generator.go`)

```go
package ai

import (
    "context"
    "encoding/json"
    "fmt"

    anthropic "github.com/anthropics/anthropic-sdk-go"
)

type Generator struct {
    client *anthropic.Client
    model  string
}

func NewGenerator(apiKey string) *Generator {
    client := anthropic.NewClient(anthropic.WithAPIKey(apiKey))
    return &Generator{
        client: &client,
        model:  "claude-sonnet-4-6",
    }
}

// GenerateSchema converts a plain-text prompt into a Jotform form JSON schema.
func (g *Generator) GenerateSchema(ctx context.Context, prompt string) (map[string]interface{}, error) {
    msg, err := g.client.Messages.New(ctx, anthropic.MessageNewParams{
        Model:     anthropic.F(g.model),
        MaxTokens: anthropic.F(int64(4096)),
        System: anthropic.F([]anthropic.TextBlockParam{
            {Text: anthropic.F("You are a Jotform form architect. " + jotformSchemaContract)},
        }),
        Messages: anthropic.F([]anthropic.MessageParam{
            anthropic.UserMessageParam(anthropic.ContentPart(
                anthropic.TextPart(fmt.Sprintf("Create a Jotform schema for: %s", prompt)),
            )),
        }),
    })
    if err != nil {
        return nil, fmt.Errorf("claude API error: %w", err)
    }

    raw := msg.Content[0].Text
    var schema map[string]interface{}
    if err := json.Unmarshal([]byte(raw), &schema); err != nil {
        return nil, fmt.Errorf("claude returned invalid JSON: %w\nRaw: %s", err, raw)
    }
    return schema, nil
}

// AnalyzeForm sends a form's question structure to Claude for UX review.
func (g *Generator) AnalyzeForm(ctx context.Context, form map[string]interface{}) (string, error) {
    formJSON, _ := json.MarshalIndent(form, "", "  ")

    msg, err := g.client.Messages.New(ctx, anthropic.MessageNewParams{
        Model:     anthropic.F(g.model),
        MaxTokens: anthropic.F(int64(2048)),
        Messages: anthropic.F([]anthropic.MessageParam{
            anthropic.UserMessageParam(anthropic.ContentPart(
                anthropic.TextPart(fmt.Sprintf(
                    "Analyze this Jotform structure for UX issues, confusing labels, missing fields, "+
                        "and completion-rate problems. Be concise and actionable.\n\n%s",
                    formJSON,
                )),
            )),
        }),
    })
    if err != nil {
        return "", err
    }
    return msg.Content[0].Text, nil
}
```

---

## 4. AI Command (`cmd/ai.go`)

```go
package cmd

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "github.com/jotform/jotform-cli/internal/ai"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var aiCmd = &cobra.Command{
    Use:   "ai",
    Short: "AI-powered form generation and analysis",
}

var aiGenerateCmd = &cobra.Command{
    Use:   "generate-schema [prompt]",
    Short: "Generate a Jotform schema from a natural language prompt",
    Args:  cobra.MinimumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        prompt := strings.Join(args, " ")

        gen := ai.NewGenerator(getAnthropicKey())
        schema, err := gen.GenerateSchema(context.Background(), prompt)
        if err != nil {
            return err
        }

        outFile, _ := cmd.Flags().GetString("out")
        if outFile != "" {
            data, _ := json.MarshalIndent(schema, "", "  ")
            os.WriteFile(outFile, data, 0644)
            fmt.Printf("Schema written to %s\n", outFile)
            return nil
        }

        enc := json.NewEncoder(os.Stdout)
        enc.SetIndent("", "  ")
        return enc.Encode(schema)
    },
}

var aiAnalyzeCmd = &cobra.Command{
    Use:   "analyze [form-id]",
    Short: "Get AI-powered UX suggestions for a form",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        client, err := newClient()
        if err != nil {
            return err
        }
        form, err := client.GetForm(args[0])
        if err != nil {
            return err
        }

        // Convert to generic map for analysis
        formMap := map[string]interface{}{
            "title":     form.Title,
            "questions": form.Questions,
        }

        gen := ai.NewGenerator(getAnthropicKey())
        suggestions, err := gen.AnalyzeForm(context.Background(), formMap)
        if err != nil {
            return err
        }

        fmt.Println(suggestions)
        return nil
    },
}

func getAnthropicKey() string {
    if key := viper.GetString("anthropic_api_key"); key != "" {
        return key
    }
    return os.Getenv("ANTHROPIC_API_KEY")
}

func init() {
    aiGenerateCmd.Flags().StringP("out", "o", "", "Write schema to file instead of stdout")
    aiCmd.AddCommand(aiGenerateCmd, aiAnalyzeCmd)
    rootCmd.AddCommand(aiCmd)
}
```

---

## 5. End-to-End Workflow

```bash
# Generate a schema and immediately deploy it
jotform ai generate-schema "Onboarding survey for SaaS app users" --out onboarding.json
jotform forms create --file onboarding.json
```

```bash
# Analyze an existing form
jotform ai analyze 12345678
```

Sample output:
```
## UX Analysis: "Customer Feedback Form"

**Issues Found:**
1. Question 3 ("How satisfied are you?") uses a text box — a 1-5 rating scale
   would reduce friction and enable quantitative reporting.
2. No progress indicator — consider splitting 12 questions into 2 pages.
3. "Other (please specify)" is missing from Q7 (checkbox field).

**Estimated Impact:** +12–18% completion rate if issues 1–2 are addressed.
```

---

## Configuration

Add to `~/.config/jotform/config.yaml`:
```yaml
anthropic_api_key: "sk-ant-..."
```

Or via environment:
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

---

## Acceptance Criteria

- [ ] `jotform ai generate-schema "..."` returns valid Jotform JSON
- [ ] `jotform ai generate-schema "..." | jotform forms create --file /dev/stdin` deploys in one pipe
- [ ] `jotform ai analyze [id]` returns actionable suggestions
- [ ] Missing `ANTHROPIC_API_KEY` gives a clear error message

---

## Next Step

→ [05-phase3-mcp-server.md](05-phase3-mcp-server.md)
