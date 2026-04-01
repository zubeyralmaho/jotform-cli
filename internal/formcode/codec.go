package formcode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ReadFile loads a form schema from a JSON or YAML file, auto-detecting format by extension.
func ReadFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	var schema map[string]interface{}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &schema); err != nil {
			return nil, fmt.Errorf("invalid YAML: %w", err)
		}
	default:
		if err := json.Unmarshal(data, &schema); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	}
	return schema, nil
}

// WriteFile writes a form schema to a file in JSON or YAML format.
// Format is auto-detected from the file extension (.yaml/.yml → YAML, else JSON).
func WriteFile(path string, data map[string]interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(path))
	var out []byte
	var err error

	switch ext {
	case ".yaml", ".yml":
		out, err = yaml.Marshal(data)
	default:
		out, err = json.MarshalIndent(data, "", "  ")
		if err == nil {
			out = append(out, '\n')
		}
	}
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	return os.WriteFile(path, out, 0644)
}

// FormPropertiesToMap converts a FormProperties-like structure to a generic map
// for serialization. Accepts the raw API response.
func FormPropertiesToMap(fp interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(fp)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
