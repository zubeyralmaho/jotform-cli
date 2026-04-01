package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const ProjectFileName = ".jotform.yaml"

// ProjectConfig represents a per-project Jotform configuration file.
type ProjectConfig struct {
	FormID string `yaml:"form_id"`
	Name   string `yaml:"name"`
	Schema string `yaml:"schema"` // relative path to local schema file
}

// LoadProject looks for .jotform.yaml in the current directory and parent directories.
// Returns nil, nil if no project file is found (not an error — just no context).
func LoadProject() (*ProjectConfig, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, nil
	}

	for {
		path := filepath.Join(dir, ProjectFileName)
		data, err := os.ReadFile(path)
		if err == nil {
			var cfg ProjectConfig
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("invalid %s: %w", path, err)
			}
			// Resolve schema path relative to the project file location
			if cfg.Schema != "" && !filepath.IsAbs(cfg.Schema) {
				cfg.Schema = filepath.Join(dir, cfg.Schema)
			}
			return &cfg, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil, nil
}

// SaveProject writes a .jotform.yaml file to the given directory.
func SaveProject(cfg *ProjectConfig, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	header := "# Jotform CLI project configuration\n# See: https://github.com/zubeyralmaho/jotform-cli\n\n"
	path := filepath.Join(dir, ProjectFileName)
	return os.WriteFile(path, []byte(header+string(data)), 0644)
}

// ResolveFormID determines the form ID from (in priority order):
//  1. Positional argument (if provided)
//  2. .jotform.yaml form_id
//
// Returns an error with helpful guidance if neither is available.
func ResolveFormID(args []string) (string, error) {
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}
	cfg, err := LoadProject()
	if err != nil {
		return "", err
	}
	if cfg != nil && cfg.FormID != "" {
		return cfg.FormID, nil
	}
	return "", fmt.Errorf("form ID required — pass as argument or run 'jotform init' to set up project context")
}

// ResolveSchemaFile determines the schema file path from (in priority order):
//  1. --file flag (if provided)
//  2. .jotform.yaml schema path
//
// Returns an error if neither is available.
func ResolveSchemaFile(flagFile string) (string, error) {
	if flagFile != "" {
		return flagFile, nil
	}
	cfg, err := LoadProject()
	if err != nil {
		return "", err
	}
	if cfg != nil && cfg.Schema != "" {
		return cfg.Schema, nil
	}
	return "", fmt.Errorf("schema file required — pass --file or run 'jotform init' to set up project context")
}

// Slugify converts a form title to a filesystem-friendly directory name.
func Slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	// Replace non-alphanumeric characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "form"
	}
	// Limit length
	if len(s) > 50 {
		s = s[:50]
		s = strings.TrimRight(s, "-")
	}
	return s
}
