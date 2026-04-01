package formcode

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ComputeDiff produces a unified-style text diff between remote and local form schemas.
// Both inputs are generic maps (from JSON/YAML). Returns empty string if identical.
func ComputeDiff(remote, local map[string]interface{}) (string, error) {
	remoteJSON, err := json.MarshalIndent(remote, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshalling remote: %w", err)
	}
	localJSON, err := json.MarshalIndent(local, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshalling local: %w", err)
	}

	remoteLines := strings.Split(string(remoteJSON), "\n")
	localLines := strings.Split(string(localJSON), "\n")

	if string(remoteJSON) == string(localJSON) {
		return "", nil
	}

	return unifiedDiff("remote", "local", remoteLines, localLines), nil
}

// HasChanges returns true if remote and local differ.
func HasChanges(remote, local map[string]interface{}) bool {
	remoteJSON, _ := json.Marshal(remote)
	localJSON, _ := json.Marshal(local)
	return string(remoteJSON) != string(localJSON)
}

// unifiedDiff produces a simple unified diff output.
// Uses a basic LCS-based approach for readable diffs.
func unifiedDiff(oldName, newName string, oldLines, newLines []string) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("--- %s\n", oldName))
	buf.WriteString(fmt.Sprintf("+++ %s\n", newName))

	// Simple line-by-line comparison with context
	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	type change struct {
		lineNo int
		old    string
		new    string
	}

	var changes []change
	i, j := 0, 0
	for i < len(oldLines) && j < len(newLines) {
		if oldLines[i] == newLines[j] {
			i++
			j++
			continue
		}
		changes = append(changes, change{lineNo: i + 1, old: oldLines[i], new: newLines[j]})
		i++
		j++
	}
	for ; i < len(oldLines); i++ {
		changes = append(changes, change{lineNo: i + 1, old: oldLines[i]})
	}
	for ; j < len(newLines); j++ {
		changes = append(changes, change{lineNo: j + 1, new: newLines[j]})
	}

	if len(changes) == 0 {
		return ""
	}

	for _, c := range changes {
		if c.old != "" {
			buf.WriteString(fmt.Sprintf("-%s\n", c.old))
		}
		if c.new != "" {
			buf.WriteString(fmt.Sprintf("+%s\n", c.new))
		}
	}

	return buf.String()
}
