package formcode

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasChanges_Identical(t *testing.T) {
	a := map[string]interface{}{"title": "My Form", "id": "123"}
	b := map[string]interface{}{"title": "My Form", "id": "123"}
	assert.False(t, HasChanges(a, b))
}

func TestHasChanges_Different(t *testing.T) {
	a := map[string]interface{}{"title": "My Form"}
	b := map[string]interface{}{"title": "Other Form"}
	assert.True(t, HasChanges(a, b))
}

func TestHasChanges_Empty(t *testing.T) {
	assert.False(t, HasChanges(map[string]interface{}{}, map[string]interface{}{}))
}

func TestHasChanges_ExtraKey(t *testing.T) {
	a := map[string]interface{}{"title": "Form", "extra": "value"}
	b := map[string]interface{}{"title": "Form"}
	assert.True(t, HasChanges(a, b))
}

func TestComputeDiff_Identical(t *testing.T) {
	a := map[string]interface{}{"title": "Form", "id": "1"}
	b := map[string]interface{}{"title": "Form", "id": "1"}
	diff, err := ComputeDiff(a, b)
	require.NoError(t, err)
	assert.Empty(t, diff, "identical maps should produce empty diff")
}

func TestComputeDiff_Changed(t *testing.T) {
	remote := map[string]interface{}{"title": "Old Title"}
	local := map[string]interface{}{"title": "New Title"}
	diff, err := ComputeDiff(remote, local)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)
	assert.Contains(t, diff, "--- remote")
	assert.Contains(t, diff, "+++ local")
	assert.Contains(t, diff, "-")
	assert.Contains(t, diff, "+")
}

func TestComputeDiff_Header(t *testing.T) {
	remote := map[string]interface{}{"a": "1"}
	local := map[string]interface{}{"a": "2"}
	diff, err := ComputeDiff(remote, local)
	require.NoError(t, err)
	lines := strings.Split(diff, "\n")
	assert.Equal(t, "--- remote", lines[0])
	assert.Equal(t, "+++ local", lines[1])
}

func TestComputeDiff_AddedKey(t *testing.T) {
	remote := map[string]interface{}{"title": "Form"}
	local := map[string]interface{}{"title": "Form", "status": "ENABLED"}
	diff, err := ComputeDiff(remote, local)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)
}

func TestComputeDiff_RemovedKey(t *testing.T) {
	remote := map[string]interface{}{"title": "Form", "status": "ENABLED"}
	local := map[string]interface{}{"title": "Form"}
	diff, err := ComputeDiff(remote, local)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)
}

func TestComputeDiff_NestedMap(t *testing.T) {
	remote := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"text": "Name"},
		},
	}
	local := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"text": "Full Name"},
		},
	}
	diff, err := ComputeDiff(remote, local)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)
}
