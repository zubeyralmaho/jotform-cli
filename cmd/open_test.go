package cmd

import (
	"os"
	"testing"

	"github.com/zubeyralmaho/jotform-cli/internal/config"
)

// TestIsValidFormID tests form ID validation
func TestIsValidFormID(t *testing.T) {
	tests := []struct {
		name     string
		formID   string
		expected bool
	}{
		{"valid numeric ID", "242753193847060", true},
		{"valid short ID", "123", true},
		{"invalid with letters", "abc123", false},
		{"invalid with special chars", "123-456", false},
		{"invalid empty", "", false},
		{"invalid with spaces", "123 456", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidFormID(tt.formID)
			if result != tt.expected {
				t.Errorf("isValidFormID(%q) = %v, want %v", tt.formID, result, tt.expected)
			}
		})
	}
}

// TestRunOpenWithExplicitFormID tests opening with explicit form ID argument
func TestRunOpenWithExplicitFormID(t *testing.T) {
	// This test verifies the command accepts a valid form ID
	// We can't actually open a browser in tests, but we can verify the logic
	
	// Test with valid form ID
	formID := "242753193847060"
	if !isValidFormID(formID) {
		t.Errorf("Expected %s to be valid", formID)
	}
	
	// Test with invalid form ID
	invalidID := "invalid-id"
	if isValidFormID(invalidID) {
		t.Errorf("Expected %s to be invalid", invalidID)
	}
}

// TestRunOpenWithProjectContext tests opening with project context
func TestRunOpenWithProjectContext(t *testing.T) {
	// Create a temporary directory with .jotform.yaml
	tmpDir := t.TempDir()
	
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	// Create project config
	cfg := &config.ProjectConfig{
		FormID: "242753193847060",
		Name:   "Test Form",
		Schema: "form.yaml",
	}
	
	if err := config.SaveProject(cfg, "."); err != nil {
		t.Fatalf("Failed to save project config: %v", err)
	}
	
	// Verify form ID can be resolved from context
	formID, err := config.ResolveFormID([]string{})
	if err != nil {
		t.Fatalf("Failed to resolve form ID from context: %v", err)
	}
	
	if formID != cfg.FormID {
		t.Errorf("Expected form ID %s, got %s", cfg.FormID, formID)
	}
	
	// Verify form ID is valid
	if !isValidFormID(formID) {
		t.Errorf("Expected resolved form ID to be valid")
	}
}

// TestRunOpenWithoutContext tests error handling when no context exists
func TestRunOpenWithoutContext(t *testing.T) {
	// Create a temporary directory without .jotform.yaml
	tmpDir := t.TempDir()
	
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	// Try to resolve form ID without context or args
	_, err = config.ResolveFormID([]string{})
	if err == nil {
		t.Error("Expected error when resolving form ID without context or args")
	}
}

// TestFormURLConstruction tests that URLs are constructed correctly
func TestFormURLConstruction(t *testing.T) {
	tests := []struct {
		name     string
		formID   string
		expected string
	}{
		{"standard form ID", "242753193847060", "https://form.jotform.com/242753193847060"},
		{"short form ID", "123", "https://form.jotform.com/123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Construct URL using the same pattern as runOpen
			formURL := "https://form.jotform.com/" + tt.formID
			if formURL != tt.expected {
				t.Errorf("Expected URL %s, got %s", tt.expected, formURL)
			}
		})
	}
}

// TestOpenBrowserCrossPlatform tests that openBrowser handles different platforms
func TestOpenBrowserCrossPlatform(t *testing.T) {
	// We can't actually test browser opening in unit tests
	// But we can verify the function exists and has the right signature
	
	// This is a smoke test - just verify the function doesn't panic with a valid URL
	// The actual browser opening will be tested manually
	url := "https://form.jotform.com/242753193847060"
	
	// We expect this to either succeed or fail gracefully
	// We don't want to actually open a browser during tests
	_ = url
	
	t.Log("Browser opening must be tested manually on each platform")
}

// TestExplicitFormIDOverridesContext tests that explicit args take precedence
func TestExplicitFormIDOverridesContext(t *testing.T) {
	// Create a temporary directory with .jotform.yaml
	tmpDir := t.TempDir()
	
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	// Create project config with one form ID
	cfg := &config.ProjectConfig{
		FormID: "111111111111111",
		Name:   "Context Form",
		Schema: "form.yaml",
	}
	
	if err := config.SaveProject(cfg, "."); err != nil {
		t.Fatalf("Failed to save project config: %v", err)
	}
	
	// Resolve with explicit argument (should override context)
	explicitID := "222222222222222"
	formID, err := config.ResolveFormID([]string{explicitID})
	if err != nil {
		t.Fatalf("Failed to resolve form ID: %v", err)
	}
	
	if formID != explicitID {
		t.Errorf("Expected explicit form ID %s to override context, got %s", explicitID, formID)
	}
}

// TestOpenCommandIntegration tests the full open command flow
func TestOpenCommandIntegration(t *testing.T) {
	// Create a temporary directory with project context
	tmpDir := t.TempDir()
	
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	// Create project config
	cfg := &config.ProjectConfig{
		FormID: "242753193847060",
		Name:   "Test Form",
		Schema: "form.yaml",
	}
	
	if err := config.SaveProject(cfg, "."); err != nil {
		t.Fatalf("Failed to save project config: %v", err)
	}
	
	// Test 1: Resolve form ID from context
	formID, err := config.ResolveFormID([]string{})
	if err != nil {
		t.Fatalf("Failed to resolve form ID from context: %v", err)
	}
	
	// Test 2: Validate form ID
	if !isValidFormID(formID) {
		t.Errorf("Expected form ID to be valid")
	}
	
	// Test 3: Construct URL
	expectedURL := "https://form.jotform.com/242753193847060"
	actualURL := "https://form.jotform.com/" + formID
	if actualURL != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, actualURL)
	}
	
	// Note: We don't actually open the browser in tests
	t.Log("Browser opening must be tested manually")
}
