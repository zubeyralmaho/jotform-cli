package watch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckpoint_HasSeen_Empty(t *testing.T) {
	cp := &Checkpoint{}
	assert.False(t, cp.HasSeen("100", "2024-01-01"))
}

func TestCheckpoint_HasSeen_ByDate(t *testing.T) {
	cp := &Checkpoint{
		LastSeenID:    "100",
		LastCreatedAt: "2024-06-15",
	}

	assert.True(t, cp.HasSeen("99", "2024-06-14"))
	assert.True(t, cp.HasSeen("100", "2024-06-15"))
	assert.False(t, cp.HasSeen("101", "2024-06-16"))
}

func TestCheckpoint_HasSeen_ByID_Fallback(t *testing.T) {
	cp := &Checkpoint{
		LastSeenID: "100",
	}

	// When createdAt is empty but LastCreatedAt is also empty,
	// it falls through to ID comparison (string comparison)
	assert.True(t, cp.HasSeen("100", ""))
	assert.False(t, cp.HasSeen("101", ""))
	// "99" < "100" is false in string comparison ("9" > "1")
	assert.False(t, cp.HasSeen("99", ""))
}

func TestCheckpoint_Update(t *testing.T) {
	cp := &Checkpoint{}
	cp.Update("200", "2024-07-01")

	assert.Equal(t, "200", cp.LastSeenID)
	assert.Equal(t, "2024-07-01", cp.LastCreatedAt)
}

func TestCheckpoint_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	cpPath := filepath.Join(dir, "watch-test123.cursor")

	cp := &Checkpoint{
		FormID:        "test123",
		LastSeenID:    "500",
		LastCreatedAt: "2024-08-01",
		filePath:      cpPath,
	}

	err := cp.Save()
	require.NoError(t, err)

	// Verify file was created
	data, err := os.ReadFile(cpPath)
	require.NoError(t, err)

	var loaded Checkpoint
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)
	assert.Equal(t, "test123", loaded.FormID)
	assert.Equal(t, "500", loaded.LastSeenID)
	assert.NotEmpty(t, loaded.UpdatedAt)
}

func TestLoad_NonExistentFile(t *testing.T) {
	// Use a form ID that won't have a checkpoint file
	cp, err := Load("nonexistent-form-id-999999")
	require.NoError(t, err)
	require.NotNil(t, cp)
	assert.Equal(t, "nonexistent-form-id-999999", cp.FormID)
	assert.Empty(t, cp.LastSeenID)
}

func TestCheckpoint_SaveCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	cpPath := filepath.Join(dir, "watch-test.cursor")

	cp := &Checkpoint{
		FormID:   "test",
		filePath: cpPath,
	}

	err := cp.Save()
	require.NoError(t, err)

	_, err = os.Stat(cpPath)
	require.NoError(t, err)
}
