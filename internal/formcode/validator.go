package formcode

import (
	"fmt"
	"strings"
)

// ValidationError represents a single schema validation issue.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) String() string {
	if e.Field != "" {
		return fmt.Sprintf("[%s] %s", e.Field, e.Message)
	}
	return e.Message
}

// ValidFieldTypes lists all Jotform question types recognized by the API.
var ValidFieldTypes = map[string]bool{
	"control_textbox":    true,
	"control_textarea":   true,
	"control_email":      true,
	"control_phone":      true,
	"control_number":     true,
	"control_radio":      true,
	"control_checkbox":   true,
	"control_dropdown":   true,
	"control_fileupload": true,
	"control_date":       true,
	"control_address":    true,
	"control_head":       true,
	"control_divider":    true,
	"control_button":     true,
	"control_fullname":   true,
	"control_scale":      true,
	"control_matrix":     true,
	"control_time":       true,
	"control_spinner":    true,
	"control_rating":     true,
	"control_signature":  true,
	"control_widget":     true,
	"control_image":      true,
	"control_text":       true,
	"control_collapse":   true,
	"control_pagebreak":  true,
	"control_section":    true,
}

// ValidateSchema checks a form schema for common issues before API submission.
// Returns a slice of validation errors (empty = valid).
func ValidateSchema(schema map[string]interface{}) []ValidationError {
	var errs []ValidationError

	// Check questions key exists
	questionsRaw, ok := schema["questions"]
	if !ok {
		errs = append(errs, ValidationError{
			Field:   "questions",
			Message: "missing 'questions' key — a form must have at least one question",
		})
		return errs // Can't validate further without questions
	}

	questions, ok := questionsRaw.(map[string]interface{})
	if !ok {
		errs = append(errs, ValidationError{
			Field:   "questions",
			Message: "'questions' must be an object (map of order → question)",
		})
		return errs
	}

	if len(questions) == 0 {
		errs = append(errs, ValidationError{
			Field:   "questions",
			Message: "'questions' is empty — a form must have at least one question",
		})
	}

	hasSubmitButton := false

	for order, qRaw := range questions {
		q, ok := qRaw.(map[string]interface{})
		if !ok {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("questions.%s", order),
				Message: "question must be an object",
			})
			continue
		}

		// Check type
		qType, _ := q["type"].(string)
		if qType == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("questions.%s.type", order),
				Message: "missing 'type' field",
			})
		} else if !ValidFieldTypes[qType] {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("questions.%s.type", order),
				Message: fmt.Sprintf("unknown field type '%s'", qType),
			})
		}

		// Check text (all question types should have a label)
		if qType != "control_divider" && qType != "control_pagebreak" {
			text, _ := q["text"].(string)
			if text == "" {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("questions.%s.text", order),
					Message: "missing 'text' (question label)",
				})
			}
		}

		// Check options for choice fields
		if qType == "control_radio" || qType == "control_checkbox" || qType == "control_dropdown" {
			opts, _ := q["options"].(string)
			if opts == "" {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("questions.%s.options", order),
					Message: fmt.Sprintf("'%s' requires pipe-separated 'options' (e.g. \"A|B|C\")", qType),
				})
			}
		}

		if qType == "control_button" {
			hasSubmitButton = true
		}
	}

	// Check properties.title
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		title, _ := props["title"].(string)
		if strings.TrimSpace(title) == "" {
			errs = append(errs, ValidationError{
				Field:   "properties.title",
				Message: "form title is empty",
			})
		}
	} else {
		errs = append(errs, ValidationError{
			Field:   "properties",
			Message: "missing 'properties' with at least a 'title'",
		})
	}

	if !hasSubmitButton {
		errs = append(errs, ValidationError{
			Field:   "questions",
			Message: "no submit button found (add a 'control_button' question)",
		})
	}

	return errs
}
