package watch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Checkpoint persists the last seen submission state for durable watch.
type Checkpoint struct {
	FormID        string `json:"form_id"`
	LastSeenID    string `json:"last_seen_id"`
	LastCreatedAt string `json:"last_created_at"`
	UpdatedAt     string `json:"updated_at"`

	filePath string
}

// Load reads a checkpoint file for the given form ID.
// Returns a zero-value Checkpoint if the file doesn't exist (first run).
func Load(formID string) (*Checkpoint, error) {
	cp := &Checkpoint{
		FormID:   formID,
		filePath: checkpointPath(formID),
	}

	data, err := os.ReadFile(cp.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return cp, nil
		}
		return nil, fmt.Errorf("reading checkpoint: %w", err)
	}

	if err := json.Unmarshal(data, cp); err != nil {
		return nil, fmt.Errorf("parsing checkpoint: %w", err)
	}
	cp.filePath = checkpointPath(formID)
	return cp, nil
}

// Save writes the checkpoint to disk.
func (c *Checkpoint) Save() error {
	c.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	dir := filepath.Dir(c.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.filePath, data, 0644)
}

// HasSeen returns true if we've already processed this submission ID.
func (c *Checkpoint) HasSeen(submissionID, createdAt string) bool {
	if c.LastSeenID == "" {
		return false
	}
	// If the submission was created before or at our checkpoint, we've likely seen it.
	// We compare created_at strings (ISO format sorts lexicographically).
	if createdAt != "" && c.LastCreatedAt != "" {
		return createdAt <= c.LastCreatedAt
	}
	// Fallback: compare IDs (numeric, higher = newer)
	return submissionID <= c.LastSeenID
}

// Update advances the checkpoint to track a new submission.
func (c *Checkpoint) Update(submissionID, createdAt string) {
	c.LastSeenID = submissionID
	c.LastCreatedAt = createdAt
}

func checkpointPath(formID string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".jotform", fmt.Sprintf("watch-%s.cursor", formID))
}
