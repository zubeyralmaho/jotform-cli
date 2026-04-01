package formcode

import (
	"fmt"
	"testing"
	"time"
)

func TestChangeTypeString(t *testing.T) {
	tests := []struct {
		name     string
		ct       ChangeType
		expected string
	}{
		{"Added", ChangeAdded, "Added"},
		{"Modified", ChangeModified, "Modified"},
		{"Deleted", ChangeDeleted, "Deleted"},
		{"Unknown", ChangeType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ct.String()
			if result != tt.expected {
				t.Errorf("ChangeType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStatusReportStructure(t *testing.T) {
	// Test that StatusReport can be created with all fields
	now := time.Now()
	changes := []Change{
		{
			Type:        ChangeAdded,
			Path:        "questions.5",
			Description: "Added new field",
			NewValue:    "test value",
		},
		{
			Type:        ChangeModified,
			Path:        "questions.3.text",
			Description: "Modified field text",
			OldValue:    "Phone",
			NewValue:    "Mobile Phone",
		},
		{
			Type:        ChangeDeleted,
			Path:        "questions.7",
			Description: "Deleted field",
			OldValue:    "Fax Number",
		},
	}

	report := StatusReport{
		FormID:         "242753193847060",
		FormName:       "Contact Form",
		LocalPath:      "form.yaml",
		LocalModified:  now,
		RemoteModified: now.Add(-2 * time.Hour),
		Changes:        changes,
		HasChanges:     true,
	}

	// Verify all fields are accessible
	if report.FormID != "242753193847060" {
		t.Errorf("FormID = %v, want %v", report.FormID, "242753193847060")
	}
	if report.FormName != "Contact Form" {
		t.Errorf("FormName = %v, want %v", report.FormName, "Contact Form")
	}
	if report.LocalPath != "form.yaml" {
		t.Errorf("LocalPath = %v, want %v", report.LocalPath, "form.yaml")
	}
	if !report.HasChanges {
		t.Error("HasChanges should be true")
	}
	if len(report.Changes) != 3 {
		t.Errorf("len(Changes) = %v, want %v", len(report.Changes), 3)
	}
}

func TestChangeStructure(t *testing.T) {
	// Test Change with all fields
	change := Change{
		Type:        ChangeModified,
		Path:        "questions.3.text",
		Description: "Modified question text",
		OldValue:    "Phone",
		NewValue:    "Mobile Phone",
	}

	if change.Type != ChangeModified {
		t.Errorf("Type = %v, want %v", change.Type, ChangeModified)
	}
	if change.Path != "questions.3.text" {
		t.Errorf("Path = %v, want %v", change.Path, "questions.3.text")
	}
	if change.Description != "Modified question text" {
		t.Errorf("Description = %v, want %v", change.Description, "Modified question text")
	}
	if change.OldValue != "Phone" {
		t.Errorf("OldValue = %v, want %v", change.OldValue, "Phone")
	}
	if change.NewValue != "Mobile Phone" {
		t.Errorf("NewValue = %v, want %v", change.NewValue, "Mobile Phone")
	}
}

func TestEmptyStatusReport(t *testing.T) {
	// Test StatusReport with no changes
	report := StatusReport{
		FormID:     "123456789",
		FormName:   "Empty Form",
		Changes:    []Change{},
		HasChanges: false,
	}

	if report.HasChanges {
		t.Error("HasChanges should be false for empty changes")
	}
	if len(report.Changes) != 0 {
		t.Errorf("len(Changes) = %v, want 0", len(report.Changes))
	}
}

// Mock API client for testing
type mockAPIClient struct {
	form *FormProperties
	err  error
}

func (m *mockAPIClient) GetForm(id string) (*FormProperties, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.form, nil
}

func TestComputeStatus_IdenticalSchemas(t *testing.T) {
	// Create a temporary schema file
	tmpDir := t.TempDir()
	schemaFile := tmpDir + "/form.yaml"

	schema := map[string]interface{}{
		"id":    "123456789",
		"title": "Test Form",
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"text": "Name",
				"type": "control_textbox",
			},
		},
	}

	if err := WriteFile(schemaFile, schema); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Mock API client returning identical form
	mockClient := &mockAPIClient{
		form: &FormProperties{
			ID:    "123456789",
			Title: "Test Form",
			Questions: map[string]interface{}{
				"1": map[string]interface{}{
					"text": "Name",
					"type": "control_textbox",
				},
			},
		},
	}

	// Compute status
	report, err := ComputeStatus(mockClient, "123456789", schemaFile)
	if err != nil {
		t.Fatalf("ComputeStatus failed: %v", err)
	}

	// Verify no changes detected
	if report.HasChanges {
		t.Error("HasChanges should be false for identical schemas")
	}
	if len(report.Changes) != 0 {
		t.Errorf("Expected 0 changes, got %d", len(report.Changes))
	}
	if report.FormID != "123456789" {
		t.Errorf("FormID = %v, want %v", report.FormID, "123456789")
	}
	if report.FormName != "Test Form" {
		t.Errorf("FormName = %v, want %v", report.FormName, "Test Form")
	}
}

func TestComputeStatus_AddedQuestion(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := tmpDir + "/form.yaml"

	// Local schema has an extra question
	localSchema := map[string]interface{}{
		"id":    "123456789",
		"title": "Test Form",
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"text": "Name",
				"type": "control_textbox",
			},
			"2": map[string]interface{}{
				"text": "Email",
				"type": "control_email",
			},
		},
	}

	if err := WriteFile(schemaFile, localSchema); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Remote schema has only one question
	mockClient := &mockAPIClient{
		form: &FormProperties{
			ID:    "123456789",
			Title: "Test Form",
			Questions: map[string]interface{}{
				"1": map[string]interface{}{
					"text": "Name",
					"type": "control_textbox",
				},
			},
		},
	}

	report, err := ComputeStatus(mockClient, "123456789", schemaFile)
	if err != nil {
		t.Fatalf("ComputeStatus failed: %v", err)
	}

	// Verify change detected
	if !report.HasChanges {
		t.Error("HasChanges should be true")
	}
	if len(report.Changes) != 1 {
		t.Fatalf("Expected 1 change, got %d", len(report.Changes))
	}

	change := report.Changes[0]
	if change.Type != ChangeAdded {
		t.Errorf("Change type = %v, want %v", change.Type, ChangeAdded)
	}
	if change.Path != "questions.2" {
		t.Errorf("Change path = %v, want %v", change.Path, "questions.2")
	}
}

func TestComputeStatus_DeletedQuestion(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := tmpDir + "/form.yaml"

	// Local schema has only one question
	localSchema := map[string]interface{}{
		"id":    "123456789",
		"title": "Test Form",
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"text": "Name",
				"type": "control_textbox",
			},
		},
	}

	if err := WriteFile(schemaFile, localSchema); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Remote schema has two questions
	mockClient := &mockAPIClient{
		form: &FormProperties{
			ID:    "123456789",
			Title: "Test Form",
			Questions: map[string]interface{}{
				"1": map[string]interface{}{
					"text": "Name",
					"type": "control_textbox",
				},
				"2": map[string]interface{}{
					"text": "Email",
					"type": "control_email",
				},
			},
		},
	}

	report, err := ComputeStatus(mockClient, "123456789", schemaFile)
	if err != nil {
		t.Fatalf("ComputeStatus failed: %v", err)
	}

	// Verify change detected
	if !report.HasChanges {
		t.Error("HasChanges should be true")
	}
	if len(report.Changes) != 1 {
		t.Fatalf("Expected 1 change, got %d", len(report.Changes))
	}

	change := report.Changes[0]
	if change.Type != ChangeDeleted {
		t.Errorf("Change type = %v, want %v", change.Type, ChangeDeleted)
	}
	if change.Path != "questions.2" {
		t.Errorf("Change path = %v, want %v", change.Path, "questions.2")
	}
}

func TestComputeStatus_ModifiedQuestion(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := tmpDir + "/form.yaml"

	// Local schema has modified question text
	localSchema := map[string]interface{}{
		"id":    "123456789",
		"title": "Test Form",
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"text": "Full Name",
				"type": "control_textbox",
			},
		},
	}

	if err := WriteFile(schemaFile, localSchema); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Remote schema has different text
	mockClient := &mockAPIClient{
		form: &FormProperties{
			ID:    "123456789",
			Title: "Test Form",
			Questions: map[string]interface{}{
				"1": map[string]interface{}{
					"text": "Name",
					"type": "control_textbox",
				},
			},
		},
	}

	report, err := ComputeStatus(mockClient, "123456789", schemaFile)
	if err != nil {
		t.Fatalf("ComputeStatus failed: %v", err)
	}

	// Verify change detected
	if !report.HasChanges {
		t.Error("HasChanges should be true")
	}
	if len(report.Changes) != 1 {
		t.Fatalf("Expected 1 change, got %d", len(report.Changes))
	}

	change := report.Changes[0]
	if change.Type != ChangeModified {
		t.Errorf("Change type = %v, want %v", change.Type, ChangeModified)
	}
	if change.Path != "questions.1.text" {
		t.Errorf("Change path = %v, want %v", change.Path, "questions.1.text")
	}
	if change.OldValue != "Name" {
		t.Errorf("OldValue = %v, want %v", change.OldValue, "Name")
	}
	if change.NewValue != "Full Name" {
		t.Errorf("NewValue = %v, want %v", change.NewValue, "Full Name")
	}
}

func TestComputeStatus_MultipleChanges(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := tmpDir + "/form.yaml"

	// Local schema with multiple changes
	localSchema := map[string]interface{}{
		"id":    "123456789",
		"title": "Test Form",
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"text": "Full Name", // Modified
				"type": "control_textbox",
			},
			"3": map[string]interface{}{ // Added
				"text": "Phone",
				"type": "control_phone",
			},
		},
	}

	if err := WriteFile(schemaFile, localSchema); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Remote schema
	mockClient := &mockAPIClient{
		form: &FormProperties{
			ID:    "123456789",
			Title: "Test Form",
			Questions: map[string]interface{}{
				"1": map[string]interface{}{
					"text": "Name",
					"type": "control_textbox",
				},
				"2": map[string]interface{}{ // Deleted
					"text": "Email",
					"type": "control_email",
				},
			},
		},
	}

	report, err := ComputeStatus(mockClient, "123456789", schemaFile)
	if err != nil {
		t.Fatalf("ComputeStatus failed: %v", err)
	}

	// Verify multiple changes detected
	if !report.HasChanges {
		t.Error("HasChanges should be true")
	}
	if len(report.Changes) != 3 {
		t.Errorf("Expected 3 changes, got %d", len(report.Changes))
	}

	// Check that we have one of each type
	changeTypes := make(map[ChangeType]int)
	for _, change := range report.Changes {
		changeTypes[change.Type]++
	}

	if changeTypes[ChangeModified] != 1 {
		t.Errorf("Expected 1 modified change, got %d", changeTypes[ChangeModified])
	}
	if changeTypes[ChangeAdded] != 1 {
		t.Errorf("Expected 1 added change, got %d", changeTypes[ChangeAdded])
	}
	if changeTypes[ChangeDeleted] != 1 {
		t.Errorf("Expected 1 deleted change, got %d", changeTypes[ChangeDeleted])
	}
}

func TestComputeStatus_FileNotFound(t *testing.T) {
	mockClient := &mockAPIClient{
		form: &FormProperties{
			ID:    "123456789",
			Title: "Test Form",
		},
	}

	_, err := ComputeStatus(mockClient, "123456789", "/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestComputeStatus_APIError(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := tmpDir + "/form.yaml"

	schema := map[string]interface{}{
		"id":    "123456789",
		"title": "Test Form",
	}

	if err := WriteFile(schemaFile, schema); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Mock API client returning error
	mockClient := &mockAPIClient{
		err: fmt.Errorf("API error"),
	}

	_, err := ComputeStatus(mockClient, "123456789", schemaFile)
	if err == nil {
		t.Error("Expected error from API")
	}
}

func TestComputeStatus_RemoteTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := tmpDir + "/form.yaml"

	schema := map[string]interface{}{
		"id":    "123456789",
		"title": "Test Form",
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"text": "Name",
				"type": "control_textbox",
			},
		},
	}

	if err := WriteFile(schemaFile, schema); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Mock API client with timestamp
	expectedTime := "2024-01-15T10:30:00Z"
	mockClient := &mockAPIClient{
		form: &FormProperties{
			ID:    "123456789",
			Title: "Test Form",
			Questions: map[string]interface{}{
				"1": map[string]interface{}{
					"text": "Name",
					"type": "control_textbox",
				},
			},
			Properties: map[string]interface{}{
				"updated_at": expectedTime,
			},
		},
	}

	report, err := ComputeStatus(mockClient, "123456789", schemaFile)
	if err != nil {
		t.Fatalf("ComputeStatus failed: %v", err)
	}

	// Verify remote timestamp was parsed
	parsedTime, _ := time.Parse(time.RFC3339, expectedTime)
	if !report.RemoteModified.Equal(parsedTime) {
		t.Errorf("RemoteModified = %v, want %v", report.RemoteModified, parsedTime)
	}
}

func TestDetectChanges_EmptySchemas(t *testing.T) {
	local := map[string]interface{}{}
	remote := map[string]interface{}{}

	changes := detectChanges(local, remote)
	if len(changes) != 0 {
		t.Errorf("Expected 0 changes for empty schemas, got %d", len(changes))
	}
}

func TestGetQuestionsMap(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]interface{}
		expected int
	}{
		{
			name:     "No questions field",
			schema:   map[string]interface{}{},
			expected: 0,
		},
		{
			name: "Valid questions",
			schema: map[string]interface{}{
				"questions": map[string]interface{}{
					"1": map[string]interface{}{"text": "Name"},
					"2": map[string]interface{}{"text": "Email"},
				},
			},
			expected: 2,
		},
		{
			name: "Invalid questions type",
			schema: map[string]interface{}{
				"questions": "not a map",
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getQuestionsMap(tt.schema)
			if len(result) != tt.expected {
				t.Errorf("Expected %d questions, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestGetQuestionText(t *testing.T) {
	tests := []struct {
		name     string
		question map[string]interface{}
		expected string
	}{
		{
			name:     "Has text field",
			question: map[string]interface{}{"text": "Name"},
			expected: "Name",
		},
		{
			name:     "Has name field",
			question: map[string]interface{}{"name": "email"},
			expected: "email",
		},
		{
			name:     "No text or name",
			question: map[string]interface{}{"type": "textbox"},
			expected: "field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getQuestionText(tt.question)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDeepEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "Equal strings",
			a:        "test",
			b:        "test",
			expected: true,
		},
		{
			name:     "Different strings",
			a:        "test",
			b:        "other",
			expected: false,
		},
		{
			name:     "Equal maps",
			a:        map[string]interface{}{"key": "value"},
			b:        map[string]interface{}{"key": "value"},
			expected: true,
		},
		{
			name:     "Different maps",
			a:        map[string]interface{}{"key": "value1"},
			b:        map[string]interface{}{"key": "value2"},
			expected: false,
		},
		{
			name:     "Equal numbers",
			a:        42,
			b:        42,
			expected: true,
		},
		{
			name:     "Different numbers",
			a:        42,
			b:        43,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deepEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("deepEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
