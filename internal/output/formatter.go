package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

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
	b, _ := json.Marshal(data)

	// Try as array of objects
	var rows []map[string]interface{}
	if err := json.Unmarshal(b, &rows); err == nil {
		if len(rows) == 0 {
			_, _ = fmt.Fprintln(w, "(no results)")
			return nil
		}
		headers := sortedKeys(rows[0])
		table := tablewriter.NewTable(w)
		table.Header(headers)
		for _, row := range rows {
			vals := make([]string, len(headers))
			for i, h := range headers {
				vals[i] = fmt.Sprintf("%v", row[h])
			}
			_ = table.Append(vals)
		}
		return table.Render()
	}

	// Single object
	var obj map[string]interface{}
	if err := json.Unmarshal(b, &obj); err == nil {
		table := tablewriter.NewTable(w)
		table.Header([]string{"Field", "Value"})
		for _, k := range sortedKeys(obj) {
			_ = table.Append([]string{k, fmt.Sprintf("%v", obj[k])})
		}
		return table.Render()
	}

	_, _ = fmt.Fprintln(w, data)
	return nil
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
