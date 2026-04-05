package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltin_Count(t *testing.T) {
	templates := Builtin()
	assert.Len(t, templates, 8)
}

func TestBuiltin_AllHaveRequiredFields(t *testing.T) {
	for _, tmpl := range Builtin() {
		t.Run(tmpl.Name, func(t *testing.T) {
			assert.NotEmpty(t, tmpl.Name)
			assert.NotEmpty(t, tmpl.Description)
			assert.NotEmpty(t, tmpl.Category)
			assert.NotNil(t, tmpl.Schema)
			assert.NotNil(t, tmpl.Schema["questions"])
			assert.NotNil(t, tmpl.Schema["properties"])
		})
	}
}

func TestBuiltin_UniqueNames(t *testing.T) {
	seen := map[string]bool{}
	for _, tmpl := range Builtin() {
		assert.False(t, seen[tmpl.Name], "duplicate template name: %s", tmpl.Name)
		seen[tmpl.Name] = true
	}
}

func TestBuiltin_AllHaveSubmitButton(t *testing.T) {
	for _, tmpl := range Builtin() {
		t.Run(tmpl.Name, func(t *testing.T) {
			questions, ok := tmpl.Schema["questions"].(map[string]interface{})
			require.True(t, ok)

			hasButton := false
			for _, q := range questions {
				qMap, ok := q.(map[string]interface{})
				if ok && qMap["type"] == "control_button" {
					hasButton = true
					break
				}
			}
			assert.True(t, hasButton, "template %s has no submit button", tmpl.Name)
		})
	}
}

func TestGet_Found(t *testing.T) {
	tmpl := Get("contact")
	require.NotNil(t, tmpl)
	assert.Equal(t, "contact", tmpl.Name)
	assert.Equal(t, "General", tmpl.Category)
}

func TestGet_NotFound(t *testing.T) {
	tmpl := Get("nonexistent")
	assert.Nil(t, tmpl)
}

func TestGet_AllTemplates(t *testing.T) {
	names := []string{"contact", "feedback", "rsvp", "order", "registration", "survey", "bug-report", "job-application"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			tmpl := Get(name)
			require.NotNil(t, tmpl, "template %s not found", name)
			assert.Equal(t, name, tmpl.Name)
		})
	}
}

func TestBuiltin_AllHaveHeader(t *testing.T) {
	for _, tmpl := range Builtin() {
		t.Run(tmpl.Name, func(t *testing.T) {
			questions, ok := tmpl.Schema["questions"].(map[string]interface{})
			require.True(t, ok)

			q1, ok := questions["1"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, "control_head", q1["type"])
		})
	}
}
