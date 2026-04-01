package cmd

import (
	"testing"
	"time"

	"github.com/zubeyralmaho/jotform-cli/internal/formcode"
)

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "just now",
			time:     now.Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "1 minute ago",
			time:     now.Add(-1 * time.Minute),
			expected: "1 minute ago",
		},
		{
			name:     "5 minutes ago",
			time:     now.Add(-5 * time.Minute),
			expected: "5 minutes ago",
		},
		{
			name:     "1 hour ago",
			time:     now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "3 hours ago",
			time:     now.Add(-3 * time.Hour),
			expected: "3 hours ago",
		},
		{
			name:     "1 day ago",
			time:     now.Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "3 days ago",
			time:     now.Add(-3 * 24 * time.Hour),
			expected: "3 days ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRelativeTime(tt.time)
			if result != tt.expected {
				t.Errorf("formatRelativeTime(%v) = %q, want %q", tt.time, result, tt.expected)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "nil value",
			value:    nil,
			expected: "<nil>",
		},
		{
			name:     "short string",
			value:    "Hello",
			expected: `"Hello"`,
		},
		{
			name:     "long string",
			value:    "This is a very long string that should be truncated because it exceeds the maximum length",
			expected: `"This is a very long string that should be trunc..."`,
		},
		{
			name:     "map with text",
			value:    map[string]interface{}{"text": "Name", "type": "textbox"},
			expected: `field "Name"`,
		},
		{
			name:     "map with name",
			value:    map[string]interface{}{"name": "email", "type": "email"},
			expected: `field "email"`,
		},
		{
			name:     "map without text or name",
			value:    map[string]interface{}{"type": "submit"},
			expected: "object",
		},
		{
			name:     "array",
			value:    []interface{}{"a", "b", "c"},
			expected: "array[3]",
		},
		{
			name:     "number",
			value:    42,
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value)
			if result != tt.expected {
				t.Errorf("formatValue(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestDisplayChange(t *testing.T) {
	// This test just ensures displayChange doesn't panic with various change types
	changes := []formcode.Change{
		{
			Type:     formcode.ChangeAdded,
			Path:     "questions.5",
			NewValue: map[string]interface{}{"text": "New Field"},
		},
		{
			Type:     formcode.ChangeModified,
			Path:     "questions.3.text",
			OldValue: "Phone",
			NewValue: "Mobile Phone",
		},
		{
			Type:     formcode.ChangeDeleted,
			Path:     "questions.7",
			OldValue: map[string]interface{}{"text": "Fax"},
		},
	}

	for _, change := range changes {
		// Just ensure it doesn't panic
		displayChange(change)
	}
}

func TestDisplayStatusSummary(t *testing.T) {
	report := &formcode.StatusReport{
		FormID:   "123456789",
		FormName: "Test Form",
		Changes: []formcode.Change{
			{Type: formcode.ChangeAdded, Path: "questions.1"},
			{Type: formcode.ChangeModified, Path: "questions.2"},
			{Type: formcode.ChangeModified, Path: "questions.3"},
			{Type: formcode.ChangeDeleted, Path: "questions.4"},
		},
		HasChanges: true,
	}

	// Just ensure it doesn't panic
	displayStatusSummary(report)
}

func TestDisplayStatusReport(t *testing.T) {
	report := &formcode.StatusReport{
		FormID:         "123456789",
		FormName:       "Test Form",
		LocalPath:      "/path/to/form.yaml",
		LocalModified:  time.Now().Add(-2 * time.Hour),
		RemoteModified: time.Now().Add(-5 * time.Hour),
		Changes: []formcode.Change{
			{
				Type:     formcode.ChangeModified,
				Path:     "questions.3.text",
				OldValue: "Phone",
				NewValue: "Mobile Phone",
			},
		},
		HasChanges: true,
	}

	// Just ensure it doesn't panic
	displayStatusReport(report)
}

func TestDisplayStatusReport_NoChanges(t *testing.T) {
	report := &formcode.StatusReport{
		FormID:         "123456789",
		FormName:       "Test Form",
		LocalPath:      "/path/to/form.yaml",
		LocalModified:  time.Now(),
		RemoteModified: time.Now(),
		Changes:        []formcode.Change{},
		HasChanges:     false,
	}

	// Just ensure it doesn't panic
	displayStatusReport(report)
}
