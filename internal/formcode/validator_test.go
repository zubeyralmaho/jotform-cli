package formcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSchema_Valid(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"type":  "control_head",
				"text":  "Contact Form",
				"order": "1",
			},
			"2": map[string]interface{}{
				"type":  "control_textbox",
				"text":  "Full Name",
				"name":  "fullName",
				"order": "2",
			},
			"3": map[string]interface{}{
				"type":  "control_email",
				"text":  "Email Address",
				"name":  "email",
				"order": "3",
			},
			"4": map[string]interface{}{
				"type":  "control_button",
				"text":  "Submit",
				"order": "4",
			},
		},
		"properties": map[string]interface{}{
			"title": "Contact Form",
		},
	}

	errs := ValidateSchema(schema)
	assert.Empty(t, errs, "valid schema should produce no errors")
}

func TestValidateSchema_MissingQuestions(t *testing.T) {
	schema := map[string]interface{}{
		"properties": map[string]interface{}{
			"title": "Empty Form",
		},
	}

	errs := ValidateSchema(schema)
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Message, "missing 'questions'")
}

func TestValidateSchema_UnknownFieldType(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"type": "control_unknown_widget",
				"text": "Test",
			},
			"2": map[string]interface{}{
				"type": "control_button",
				"text": "Submit",
			},
		},
		"properties": map[string]interface{}{
			"title": "Test Form",
		},
	}

	errs := ValidateSchema(schema)
	found := false
	for _, e := range errs {
		if e.Field == "questions.1.type" {
			found = true
			assert.Contains(t, e.Message, "unknown field type")
		}
	}
	assert.True(t, found, "should report unknown field type")
}

func TestValidateSchema_MissingTitle(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"type": "control_button",
				"text": "Submit",
			},
		},
		"properties": map[string]interface{}{},
	}

	errs := ValidateSchema(schema)
	found := false
	for _, e := range errs {
		if e.Field == "properties.title" {
			found = true
		}
	}
	assert.True(t, found, "should report missing title")
}

func TestValidateSchema_MissingProperties(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"type": "control_button",
				"text": "Submit",
			},
		},
	}

	errs := ValidateSchema(schema)
	found := false
	for _, e := range errs {
		if e.Field == "properties" {
			found = true
		}
	}
	assert.True(t, found, "should report missing properties")
}

func TestValidateSchema_NoSubmitButton(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"type": "control_textbox",
				"text": "Name",
			},
		},
		"properties": map[string]interface{}{
			"title": "Test",
		},
	}

	errs := ValidateSchema(schema)
	found := false
	for _, e := range errs {
		if e.Message == "no submit button found (add a 'control_button' question)" {
			found = true
		}
	}
	assert.True(t, found, "should report missing submit button")
}

func TestValidateSchema_ChoiceFieldMissingOptions(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"type": "control_radio",
				"text": "Pick one",
			},
			"2": map[string]interface{}{
				"type": "control_button",
				"text": "Submit",
			},
		},
		"properties": map[string]interface{}{
			"title": "Test",
		},
	}

	errs := ValidateSchema(schema)
	found := false
	for _, e := range errs {
		if e.Field == "questions.1.options" {
			found = true
			assert.Contains(t, e.Message, "requires pipe-separated 'options'")
		}
	}
	assert.True(t, found, "should report missing options for radio field")
}

func TestValidateSchema_MissingType(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"text": "No type field",
			},
			"2": map[string]interface{}{
				"type": "control_button",
				"text": "Submit",
			},
		},
		"properties": map[string]interface{}{
			"title": "Test",
		},
	}

	errs := ValidateSchema(schema)
	found := false
	for _, e := range errs {
		if e.Field == "questions.1.type" {
			found = true
			assert.Contains(t, e.Message, "missing 'type'")
		}
	}
	assert.True(t, found, "should report missing type")
}
