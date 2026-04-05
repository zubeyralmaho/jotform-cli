package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaContract_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, jotformSchemaContract)
}

func TestSchemaContract_ContainsFieldTypes(t *testing.T) {
	assert.Contains(t, jotformSchemaContract, "control_textbox")
	assert.Contains(t, jotformSchemaContract, "control_textarea")
	assert.Contains(t, jotformSchemaContract, "control_email")
	assert.Contains(t, jotformSchemaContract, "control_phone")
	assert.Contains(t, jotformSchemaContract, "control_radio")
	assert.Contains(t, jotformSchemaContract, "control_checkbox")
	assert.Contains(t, jotformSchemaContract, "control_dropdown")
	assert.Contains(t, jotformSchemaContract, "control_button")
	assert.Contains(t, jotformSchemaContract, "control_head")
}

func TestSchemaContract_ContainsStructure(t *testing.T) {
	assert.Contains(t, jotformSchemaContract, "questions")
	assert.Contains(t, jotformSchemaContract, "properties")
	assert.Contains(t, jotformSchemaContract, "type")
	assert.Contains(t, jotformSchemaContract, "text")
	assert.Contains(t, jotformSchemaContract, "order")
}

func TestSchemaContract_ContainsInstructions(t *testing.T) {
	assert.Contains(t, jotformSchemaContract, "JSON")
	assert.Contains(t, jotformSchemaContract, "options")
	assert.Contains(t, jotformSchemaContract, "pipe-separated")
}
