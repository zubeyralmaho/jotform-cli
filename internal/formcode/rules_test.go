package formcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuleNoDuplicateNames(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"type": "control_textbox", "text": "Name", "name": "name"},
			"2": map[string]interface{}{"type": "control_textbox", "text": "Also Name", "name": "name"},
		},
	}
	rule := ruleNoDuplicateNames()
	findings := rule.Check(schema)
	assert.Len(t, findings, 1)
	assert.Equal(t, SeverityError, findings[0].Severity)
	assert.Contains(t, findings[0].Message, "duplicate name")
}

func TestRuleEmailValidation(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"type": "control_textbox", "text": "Your Email", "name": "email"},
		},
	}
	rule := ruleEmailFieldsHaveValidation()
	findings := rule.Check(schema)
	assert.Len(t, findings, 1)
	assert.Equal(t, SeverityWarning, findings[0].Severity)
	assert.Contains(t, findings[0].Message, "control_email")
}

func TestRuleEmailValidation_Correct(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"type": "control_email", "text": "Your Email", "name": "email"},
		},
	}
	rule := ruleEmailFieldsHaveValidation()
	findings := rule.Check(schema)
	assert.Empty(t, findings)
}

func TestRuleSubmitButton_Missing(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"type": "control_textbox", "text": "Name"},
		},
	}
	rule := ruleSubmitButtonExists()
	findings := rule.Check(schema)
	assert.Len(t, findings, 1)
	assert.Equal(t, SeverityError, findings[0].Severity)
}

func TestRuleSubmitButton_Present(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"type": "control_button", "text": "Submit"},
		},
	}
	rule := ruleSubmitButtonExists()
	findings := rule.Check(schema)
	assert.Empty(t, findings)
}

func TestRunRules_ValidSchema(t *testing.T) {
	schema := map[string]interface{}{
		"properties": map[string]interface{}{"title": "Test Form"},
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"type": "control_head", "text": "Header"},
			"2": map[string]interface{}{"type": "control_email", "text": "Email", "name": "email", "required": "Yes"},
			"3": map[string]interface{}{"type": "control_button", "text": "Submit"},
		},
	}
	findings := RunRules(schema, BuiltinRules())
	// Should have no errors
	for _, f := range findings {
		assert.NotEqual(t, SeverityError, f.Severity, "unexpected error: %s", f.Message)
	}
}

func TestRuleNoEmptyOptions(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"type": "control_radio", "text": "Pick one", "options": "A||C"},
		},
	}
	rule := ruleNoEmptyOptions()
	findings := rule.Check(schema)
	assert.Len(t, findings, 1)
	assert.Equal(t, SeverityWarning, findings[0].Severity)
}
