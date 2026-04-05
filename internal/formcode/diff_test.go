package formcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeDiff_Identical(t *testing.T) {
	m := map[string]interface{}{"title": "Same"}
	diff, err := ComputeDiff(m, m)
	require.NoError(t, err)
	assert.Empty(t, diff)
}

func TestComputeDiff_Different(t *testing.T) {
	remote := map[string]interface{}{"title": "Remote"}
	local := map[string]interface{}{"title": "Local"}

	diff, err := ComputeDiff(remote, local)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)
	assert.Contains(t, diff, "--- remote")
	assert.Contains(t, diff, "+++ local")
	assert.Contains(t, diff, "-")
	assert.Contains(t, diff, "+")
}

func TestComputeDiff_AddedField(t *testing.T) {
	remote := map[string]interface{}{"title": "Form"}
	local := map[string]interface{}{"title": "Form", "extra": "field"}

	diff, err := ComputeDiff(remote, local)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)
}

func TestHasChanges_True(t *testing.T) {
	remote := map[string]interface{}{"title": "A"}
	local := map[string]interface{}{"title": "B"}
	assert.True(t, HasChanges(remote, local))
}

func TestHasChanges_False(t *testing.T) {
	m := map[string]interface{}{"title": "Same"}
	assert.False(t, HasChanges(m, m))
}

func TestHasChanges_DifferentKeys(t *testing.T) {
	a := map[string]interface{}{"a": 1}
	b := map[string]interface{}{"b": 1}
	assert.True(t, HasChanges(a, b))
}

func TestComputeDiff_NestedChanges(t *testing.T) {
	remote := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"type": "control_head", "text": "Old"},
		},
	}
	local := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"type": "control_head", "text": "New"},
		},
	}

	diff, err := ComputeDiff(remote, local)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)
}
