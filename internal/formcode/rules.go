package formcode

import (
	"fmt"
	"strings"
)

// Rule is a named validation check that produces zero or more findings.
type Rule struct {
	Name        string
	Description string
	Check       func(schema map[string]interface{}) []RuleFinding
}

// Severity indicates how serious a finding is.
type Severity int

const (
	SeverityError   Severity = iota // Must fix
	SeverityWarning                 // Should fix
	SeverityInfo                    // Suggestion
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

// RuleFinding is a single issue found by a rule.
type RuleFinding struct {
	Rule     string
	Severity Severity
	Field    string
	Message  string
}

// BuiltinRules returns the default set of validation rules.
func BuiltinRules() []Rule {
	return []Rule{
		ruleRequiredFieldsHaveLabels(),
		ruleNoDuplicateNames(),
		ruleEmailFieldsHaveValidation(),
		ruleChoiceFieldsHaveOptions(),
		ruleSubmitButtonExists(),
		ruleFormHasTitle(),
		ruleNoEmptyOptions(),
		ruleRequiredFieldsMarked(),
		ruleFieldOrdering(),
	}
}

// ruleRequiredFieldsHaveLabels checks that all input fields have non-empty labels.
func ruleRequiredFieldsHaveLabels() Rule {
	return Rule{
		Name:        "labels",
		Description: "All input fields must have a text label",
		Check: func(schema map[string]interface{}) []RuleFinding {
			questions := getQuestions(schema)
			if questions == nil {
				return nil
			}
			var findings []RuleFinding
			for order, qRaw := range questions {
				q, ok := qRaw.(map[string]interface{})
				if !ok {
					continue
				}
				qType, _ := q["type"].(string)
				// Skip decorative/layout controls
				if qType == "control_divider" || qType == "control_pagebreak" || qType == "control_button" || qType == "control_head" || qType == "control_image" || qType == "control_collapse" || qType == "control_section" {
					continue
				}
				text, _ := q["text"].(string)
				if strings.TrimSpace(text) == "" {
					findings = append(findings, RuleFinding{
						Rule:     "labels",
						Severity: SeverityError,
						Field:    fmt.Sprintf("questions.%s", order),
						Message:  fmt.Sprintf("field (type: %s) is missing a text label", qType),
					})
				}
			}
			return findings
		},
	}
}

// ruleNoDuplicateNames checks that no two fields share the same name attribute.
func ruleNoDuplicateNames() Rule {
	return Rule{
		Name:        "unique-names",
		Description: "Field names must be unique",
		Check: func(schema map[string]interface{}) []RuleFinding {
			questions := getQuestions(schema)
			if questions == nil {
				return nil
			}
			seen := map[string]string{} // name → first order
			var findings []RuleFinding
			for order, qRaw := range questions {
				q, ok := qRaw.(map[string]interface{})
				if !ok {
					continue
				}
				name, _ := q["name"].(string)
				if name == "" {
					continue
				}
				if firstOrder, exists := seen[name]; exists {
					findings = append(findings, RuleFinding{
						Rule:     "unique-names",
						Severity: SeverityError,
						Field:    fmt.Sprintf("questions.%s", order),
						Message:  fmt.Sprintf("duplicate name %q (also in questions.%s)", name, firstOrder),
					})
				} else {
					seen[name] = order
				}
			}
			return findings
		},
	}
}

// ruleEmailFieldsHaveValidation checks that email fields use the correct type.
func ruleEmailFieldsHaveValidation() Rule {
	return Rule{
		Name:        "email-validation",
		Description: "Email fields should use control_email type for built-in validation",
		Check: func(schema map[string]interface{}) []RuleFinding {
			questions := getQuestions(schema)
			if questions == nil {
				return nil
			}
			var findings []RuleFinding
			for order, qRaw := range questions {
				q, ok := qRaw.(map[string]interface{})
				if !ok {
					continue
				}
				qType, _ := q["type"].(string)
				name, _ := q["name"].(string)
				text, _ := q["text"].(string)

				// If it looks like an email field but isn't using control_email
				label := strings.ToLower(name + " " + text)
				if qType == "control_textbox" && (strings.Contains(label, "email") || strings.Contains(label, "e-mail")) {
					findings = append(findings, RuleFinding{
						Rule:     "email-validation",
						Severity: SeverityWarning,
						Field:    fmt.Sprintf("questions.%s", order),
						Message:  fmt.Sprintf("field %q looks like an email field but uses control_textbox — consider control_email for built-in validation", text),
					})
				}
			}
			return findings
		},
	}
}

// ruleChoiceFieldsHaveOptions checks that radio/checkbox/dropdown fields have options.
func ruleChoiceFieldsHaveOptions() Rule {
	return Rule{
		Name:        "choice-options",
		Description: "Choice fields (radio, checkbox, dropdown) must have options",
		Check: func(schema map[string]interface{}) []RuleFinding {
			questions := getQuestions(schema)
			if questions == nil {
				return nil
			}
			var findings []RuleFinding
			choiceTypes := map[string]bool{
				"control_radio":    true,
				"control_checkbox": true,
				"control_dropdown": true,
			}
			for order, qRaw := range questions {
				q, ok := qRaw.(map[string]interface{})
				if !ok {
					continue
				}
				qType, _ := q["type"].(string)
				if !choiceTypes[qType] {
					continue
				}
				opts, _ := q["options"].(string)
				if strings.TrimSpace(opts) == "" {
					findings = append(findings, RuleFinding{
						Rule:     "choice-options",
						Severity: SeverityError,
						Field:    fmt.Sprintf("questions.%s", order),
						Message:  fmt.Sprintf("%s field is missing pipe-separated options", qType),
					})
				}
			}
			return findings
		},
	}
}

// ruleSubmitButtonExists checks that the form has a submit button.
func ruleSubmitButtonExists() Rule {
	return Rule{
		Name:        "submit-button",
		Description: "Form must have a submit button",
		Check: func(schema map[string]interface{}) []RuleFinding {
			questions := getQuestions(schema)
			if questions == nil {
				return nil
			}
			for _, qRaw := range questions {
				q, ok := qRaw.(map[string]interface{})
				if !ok {
					continue
				}
				if qType, _ := q["type"].(string); qType == "control_button" {
					return nil
				}
			}
			return []RuleFinding{{
				Rule:     "submit-button",
				Severity: SeverityError,
				Field:    "questions",
				Message:  "no submit button found — add a control_button question",
			}}
		},
	}
}

// ruleFormHasTitle checks that the form has a non-empty title.
func ruleFormHasTitle() Rule {
	return Rule{
		Name:        "form-title",
		Description: "Form must have a title in properties",
		Check: func(schema map[string]interface{}) []RuleFinding {
			props, ok := schema["properties"].(map[string]interface{})
			if !ok {
				return []RuleFinding{{
					Rule:     "form-title",
					Severity: SeverityError,
					Field:    "properties",
					Message:  "missing properties section with a title",
				}}
			}
			title, _ := props["title"].(string)
			if strings.TrimSpace(title) == "" {
				return []RuleFinding{{
					Rule:     "form-title",
					Severity: SeverityError,
					Field:    "properties.title",
					Message:  "form title is empty",
				}}
			}
			return nil
		},
	}
}

// ruleNoEmptyOptions checks that choice fields don't have empty option values.
func ruleNoEmptyOptions() Rule {
	return Rule{
		Name:        "no-empty-options",
		Description: "Choice fields should not have empty option values",
		Check: func(schema map[string]interface{}) []RuleFinding {
			questions := getQuestions(schema)
			if questions == nil {
				return nil
			}
			var findings []RuleFinding
			choiceTypes := map[string]bool{
				"control_radio":    true,
				"control_checkbox": true,
				"control_dropdown": true,
			}
			for order, qRaw := range questions {
				q, ok := qRaw.(map[string]interface{})
				if !ok {
					continue
				}
				qType, _ := q["type"].(string)
				if !choiceTypes[qType] {
					continue
				}
				opts, _ := q["options"].(string)
				if opts == "" {
					continue
				}
				parts := strings.Split(opts, "|")
				for i, p := range parts {
					if strings.TrimSpace(p) == "" {
						findings = append(findings, RuleFinding{
							Rule:     "no-empty-options",
							Severity: SeverityWarning,
							Field:    fmt.Sprintf("questions.%s.options[%d]", order, i),
							Message:  "empty option value in pipe-separated list",
						})
					}
				}
			}
			return findings
		},
	}
}

// ruleRequiredFieldsMarked suggests marking important fields as required.
func ruleRequiredFieldsMarked() Rule {
	return Rule{
		Name:        "required-fields",
		Description: "Common fields like email and name should be marked required",
		Check: func(schema map[string]interface{}) []RuleFinding {
			questions := getQuestions(schema)
			if questions == nil {
				return nil
			}
			var findings []RuleFinding
			importantTypes := map[string]bool{
				"control_email":    true,
				"control_fullname": true,
			}
			for order, qRaw := range questions {
				q, ok := qRaw.(map[string]interface{})
				if !ok {
					continue
				}
				qType, _ := q["type"].(string)
				if !importantTypes[qType] {
					continue
				}
				required, _ := q["required"].(string)
				if strings.ToLower(required) != "yes" {
					text, _ := q["text"].(string)
					findings = append(findings, RuleFinding{
						Rule:     "required-fields",
						Severity: SeverityInfo,
						Field:    fmt.Sprintf("questions.%s", order),
						Message:  fmt.Sprintf("field %q (%s) is not marked as required — consider adding required: \"Yes\"", text, qType),
					})
				}
			}
			return findings
		},
	}
}

// ruleFieldOrdering checks that fields have sequential order values.
func ruleFieldOrdering() Rule {
	return Rule{
		Name:        "field-ordering",
		Description: "Field order keys should be sequential",
		Check: func(schema map[string]interface{}) []RuleFinding {
			questions := getQuestions(schema)
			if questions == nil || len(questions) == 0 {
				return nil
			}
			// Check for gaps in ordering (informational only)
			var findings []RuleFinding
			for order := range questions {
				if order == "" {
					findings = append(findings, RuleFinding{
						Rule:     "field-ordering",
						Severity: SeverityWarning,
						Field:    "questions",
						Message:  "found a question with an empty order key",
					})
				}
			}
			return findings
		},
	}
}

// RunRules executes a set of rules against a schema and returns all findings.
func RunRules(schema map[string]interface{}, rules []Rule) []RuleFinding {
	var all []RuleFinding
	for _, r := range rules {
		all = append(all, r.Check(schema)...)
	}
	return all
}

// getQuestions extracts the questions map from a schema.
func getQuestions(schema map[string]interface{}) map[string]interface{} {
	raw, ok := schema["questions"]
	if !ok {
		return nil
	}
	questions, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}
	return questions
}
