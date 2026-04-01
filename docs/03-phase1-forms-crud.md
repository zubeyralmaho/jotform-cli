# Phase 1b: Forms CRUD + Submissions

## Goal

Implement `jotform forms list/get/create/sync` and `jotform submissions list/watch`.

---

## 1. Forms API (`internal/api/forms.go`)

```go
package api

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type Form struct {
    ID       string `json:"id"`
    Title    string `json:"title"`
    URL      string `json:"url"`
    Status   string `json:"status"`
    Created  string `json:"created_at"`
    Updated  string `json:"updated_at"`
    Count    string `json:"count"` // submission count
}

type FormProperties struct {
    ID         string                 `json:"id"`
    Title      string                 `json:"title"`
    Questions  map[string]interface{} `json:"questions"`
    Properties map[string]interface{} `json:"properties"`
}

func (c *Client) ListForms(offset, limit int) ([]Form, error) {
    var resp apiResponse[[]Form]
    path := fmt.Sprintf("/user/forms?offset=%d&limit=%d", offset, limit)
    if err := c.get(path, &resp); err != nil {
        return nil, err
    }
    return resp.Content, nil
}

func (c *Client) GetForm(id string) (*FormProperties, error) {
    var resp apiResponse[FormProperties]
    if err := c.get("/form/"+id, &resp); err != nil {
        return nil, err
    }
    return &resp.Content, nil
}

func (c *Client) CreateForm(schema map[string]interface{}) (*Form, error) {
    body, err := json.Marshal(schema)
    if err != nil {
        return nil, err
    }

    url := fmt.Sprintf("%s/user/forms?apiKey=%s", c.BaseURL, c.APIKey)
    resp, err := c.http.Post(url, "application/json", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var apiResp apiResponse[Form]
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, err
    }
    return &apiResp.Content, nil
}
```

---

## 2. Output Formatter (`internal/output/formatter.go`)

```go
package output

import (
    "encoding/json"
    "fmt"
    "io"
    "os"

    "github.com/olekukonko/tablewriter"
    "gopkg.in/yaml.v3"
)

type Format string

const (
    FormatTable Format = "table"
    FormatJSON  Format = "json"
    FormatYAML  Format = "yaml"
)

func Print(data any, format Format) error {
    return PrintTo(os.Stdout, data, format)
}

func PrintTo(w io.Writer, data any, format Format) error {
    switch format {
    case FormatJSON:
        enc := json.NewEncoder(w)
        enc.SetIndent("", "  ")
        return enc.Encode(data)
    case FormatYAML:
        return yaml.NewEncoder(w).Encode(data)
    default:
        return printTable(w, data)
    }
}

func printTable(w io.Writer, data any) error {
    // Serialize to JSON then decode to []map for generic table rendering
    b, _ := json.Marshal(data)
    var rows []map[string]interface{}
    if err := json.Unmarshal(b, &rows); err != nil {
        fmt.Fprintln(w, data)
        return nil
    }
    if len(rows) == 0 {
        fmt.Fprintln(w, "(no results)")
        return nil
    }
    // Collect headers from first row
    headers := make([]string, 0)
    for k := range rows[0] {
        headers = append(headers, k)
    }
    table := tablewriter.NewWriter(w)
    table.SetHeader(headers)
    for _, row := range rows {
        vals := make([]string, len(headers))
        for i, h := range headers {
            vals[i] = fmt.Sprintf("%v", row[h])
        }
        table.Append(vals)
    }
    table.Render()
    return nil
}
```

---

## 3. Forms Command (`cmd/forms.go`)

```go
package cmd

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/jotform/jotform-cli/internal/api"
    "github.com/jotform/jotform-cli/internal/auth"
    "github.com/jotform/jotform-cli/internal/output"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var formsCmd = &cobra.Command{
    Use:   "forms",
    Short: "Manage Jotform forms",
}

var formsListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all forms",
    RunE: func(cmd *cobra.Command, args []string) error {
        client, err := newClient()
        if err != nil {
            return err
        }
        forms, err := client.ListForms(0, 100)
        if err != nil {
            return err
        }
        return output.Print(forms, output.Format(viper.GetString("output")))
    },
}

var formsGetCmd = &cobra.Command{
    Use:   "get [form-id]",
    Short: "Fetch form structure as JSON",
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
        return output.Print(form, output.Format(viper.GetString("output")))
    },
}

var formsCreateCmd = &cobra.Command{
    Use:   "create --file [path]",
    Short: "Create a form from a local JSON/YAML file",
    RunE: func(cmd *cobra.Command, args []string) error {
        filePath, _ := cmd.Flags().GetString("file")
        data, err := os.ReadFile(filePath)
        if err != nil {
            return fmt.Errorf("cannot read file: %w", err)
        }
        var schema map[string]interface{}
        if err := json.Unmarshal(data, &schema); err != nil {
            return fmt.Errorf("invalid JSON: %w", err)
        }
        client, err := newClient()
        if err != nil {
            return err
        }
        form, err := client.CreateForm(schema)
        if err != nil {
            return err
        }
        fmt.Printf("Form created: %s (%s)\n", form.Title, form.ID)
        fmt.Printf("URL: %s\n", form.URL)
        return nil
    },
}

// newClient is a shared helper used by all commands.
func newClient() (*api.Client, error) {
    key := viper.GetString("api_key")
    if key == "" {
        var err error
        key, err = auth.LoadAPIKey()
        if err != nil {
            return nil, err
        }
    }
    return api.New(key), nil
}

func init() {
    formsCreateCmd.Flags().String("file", "", "Path to JSON/YAML form schema file")
    formsCreateCmd.MarkFlagRequired("file")

    formsCmd.AddCommand(formsListCmd, formsGetCmd, formsCreateCmd)
    rootCmd.AddCommand(formsCmd)
}
```

---

## 4. Submissions Watch (`cmd/submissions.go`)

The `watch` sub-command long-polls and writes newline-delimited JSON to stdout — designed for piping:

```bash
jotform submissions watch 12345678 | jq '.answers'
```

```go
var submissionsWatchCmd = &cobra.Command{
    Use:   "watch [form-id]",
    Short: "Stream new submissions to stdout (newline-delimited JSON)",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        client, err := newClient()
        if err != nil {
            return err
        }
        interval, _ := cmd.Flags().GetDuration("interval")
        return watchSubmissions(client, args[0], interval)
    },
}

func watchSubmissions(client *api.Client, formID string, interval time.Duration) error {
    seen := map[string]bool{}
    enc := json.NewEncoder(os.Stdout)

    for {
        subs, err := client.GetSubmissions(formID, 0, 50, "created_at", "DESC")
        if err != nil {
            fmt.Fprintln(os.Stderr, "error:", err)
        } else {
            for _, s := range subs {
                if !seen[s.ID] {
                    seen[s.ID] = true
                    enc.Encode(s)
                }
            }
        }
        time.Sleep(interval)
    }
}
```

---

## 5. Form-as-Code: `forms sync`

Sync downloads all forms and writes them as `~/.jotform/<id>.json`:

```bash
jotform forms sync
# Synced 12 forms to ~/.jotform/
```

This enables `git diff ~/.jotform/` to see what changed between runs.

---

## Acceptance Criteria

- [ ] `jotform forms list` renders a table with ID, Title, Submissions, Updated
- [ ] `jotform forms list --output json` outputs raw JSON
- [ ] `jotform forms get [id]` prints full question tree
- [ ] `jotform forms create --file schema.json` creates and prints the new form URL
- [ ] `jotform submissions watch [id]` streams newline-delimited JSON
- [ ] `jotform submissions watch [id] | jq .` works without error

---

## Next Step

→ [04-phase2-ai-module.md](04-phase2-ai-module.md)
