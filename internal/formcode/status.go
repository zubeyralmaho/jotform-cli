package formcode

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// StatusReport represents a comparison between local and remote form schemas.
// This model complements the existing ComputeDiff and HasChanges functions by
// providing structured change detection for the `jotform status` command.
//
// Usage example:
//
//	report := StatusReport{
//	    FormID:         formID,
//	    FormName:       form.Title,
//	    LocalPath:      schemaPath,
//	    LocalModified:  fileInfo.ModTime(),
//	    RemoteModified: form.UpdatedAt,
//	    Changes:        detectChanges(remote, local),
//	    HasChanges:     HasChanges(remote, local),
//	}
//
// It provides a structured view of all differences for display to users.
type StatusReport struct {
	FormID         string    // Jotform form ID
	FormName       string    // Human-readable form name
	LocalPath      string    // Path to local schema file
	LocalModified  time.Time // Local file modification time
	RemoteModified time.Time // Remote form last update time
	Changes        []Change  // List of detected changes
	HasChanges     bool      // True if any changes exist
}

// Change represents a single modification between local and remote schemas.
// Changes are categorized by type and include the JSON path to the changed field.
type Change struct {
	Type        ChangeType  // Type of change (Added, Modified, Deleted)
	Path        string      // JSON path to changed field (e.g., "questions.3.text")
	Description string      // Human-readable change description
	OldValue    interface{} // Previous value (for Modified/Deleted changes)
	NewValue    interface{} // New value (for Added/Modified changes)
}

// ChangeType represents the type of change detected in a form schema.
type ChangeType int

const (
	// ChangeAdded indicates a new field was added to the schema
	ChangeAdded ChangeType = iota
	// ChangeModified indicates an existing field was changed
	ChangeModified
	// ChangeDeleted indicates a field was removed from the schema
	ChangeDeleted
)

// String returns the human-readable name of the change type.
func (ct ChangeType) String() string {
	switch ct {
	case ChangeAdded:
		return "Added"
	case ChangeModified:
		return "Modified"
	case ChangeDeleted:
		return "Deleted"
	default:
		return "Unknown"
	}
}

// ComputeStatus compares local and remote form schemas and generates a status report.
// It loads the local schema file, fetches the remote form from the API, and detects
// all structural differences between them.
//
// Parameters:
//   - client: API client for fetching remote form data
//   - formID: Jotform form ID to fetch
//   - schemaFile: Path to local schema file
//
// Returns a StatusReport with all detected changes, or an error if loading fails.
func ComputeStatus(client APIClient, formID, schemaFile string) (*StatusReport, error) {
	// Step 1: Load local schema file
	localSchema, err := ReadFile(schemaFile)
	if err != nil {
		return nil, fmt.Errorf("loading local schema: %w", err)
	}

	// Get local file modification time
	fileInfo, err := os.Stat(schemaFile)
	if err != nil {
		return nil, fmt.Errorf("reading file info: %w", err)
	}
	localModified := fileInfo.ModTime()

	// Step 2: Fetch remote form from API
	remoteForm, err := client.GetForm(formID)
	if err != nil {
		return nil, fmt.Errorf("fetching remote form: %w", err)
	}

	// Convert FormProperties to map for comparison
	remoteSchema, err := FormPropertiesToMap(remoteForm)
	if err != nil {
		return nil, fmt.Errorf("converting remote form: %w", err)
	}

	// Parse remote modification time (if available in properties)
	remoteModified := time.Now() // Default to now if not available
	if remoteSchema != nil {
		if props, ok := remoteSchema["properties"].(map[string]interface{}); ok {
			if updated, ok := props["updated_at"].(string); ok {
				// Try parsing common timestamp formats
				formats := []string{
					time.RFC3339,
					"2006-01-02 15:04:05",
					"2006-01-02T15:04:05",
				}
				for _, format := range formats {
					if t, err := time.Parse(format, updated); err == nil {
						remoteModified = t
						break
					}
				}
			}
		}
	}

	// Step 3: Compute structural diff
	changes := detectChanges(localSchema, remoteSchema)

	// Step 4: Build status report
	report := &StatusReport{
		FormID:         formID,
		FormName:       remoteForm.Title,
		LocalPath:      schemaFile,
		LocalModified:  localModified,
		RemoteModified: remoteModified,
		Changes:        changes,
		HasChanges:     len(changes) > 0,
	}

	return report, nil
}

// APIClient defines the interface for fetching form data from the Jotform API.
// This interface allows for easier testing with mock clients.
type APIClient interface {
	GetForm(id string) (*FormProperties, error)
}

// FormProperties represents the structure returned by the Jotform API.
// This matches the structure in internal/api/forms.go
type FormProperties struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Questions  map[string]interface{} `json:"questions"`
	Properties map[string]interface{} `json:"properties"`
}

// detectChanges compares two form schemas and returns a list of changes.
// It detects added, modified, and deleted fields using JSON path notation.
func detectChanges(local, remote map[string]interface{}) []Change {
	var changes []Change

	// Compare questions
	localQuestions := getQuestionsMap(local)
	remoteQuestions := getQuestionsMap(remote)

	// Find all question IDs in both schemas
	allQuestionIDs := make(map[string]bool)
	for qid := range localQuestions {
		allQuestionIDs[qid] = true
	}
	for qid := range remoteQuestions {
		allQuestionIDs[qid] = true
	}

	// Compare each question
	for qid := range allQuestionIDs {
		localQ, localExists := localQuestions[qid]
		remoteQ, remoteExists := remoteQuestions[qid]

		if localExists && !remoteExists {
			// Question added in local (not in remote)
			changes = append(changes, Change{
				Type:        ChangeAdded,
				Path:        fmt.Sprintf("questions.%s", qid),
				Description: fmt.Sprintf("Added: %s", getQuestionText(localQ)),
				NewValue:    localQ,
			})
		} else if !localExists && remoteExists {
			// Question deleted from local (exists in remote)
			changes = append(changes, Change{
				Type:        ChangeDeleted,
				Path:        fmt.Sprintf("questions.%s", qid),
				Description: fmt.Sprintf("Deleted: %s", getQuestionText(remoteQ)),
				OldValue:    remoteQ,
			})
		} else if localExists && remoteExists {
			// Question exists in both - check for modifications
			fieldChanges := compareQuestions(qid, localQ, remoteQ)
			changes = append(changes, fieldChanges...)
		}
	}

	// Compare top-level properties
	propChanges := compareProperties(local, remote)
	changes = append(changes, propChanges...)

	return changes
}

// getQuestionsMap extracts the questions map from a form schema.
// Returns an empty map if questions field doesn't exist or isn't a map.
func getQuestionsMap(schema map[string]interface{}) map[string]map[string]interface{} {
	questions := make(map[string]map[string]interface{})
	
	questionsField, ok := schema["questions"]
	if !ok {
		return questions
	}

	questionsMap, ok := questionsField.(map[string]interface{})
	if !ok {
		return questions
	}

	for qid, qdata := range questionsMap {
		if qmap, ok := qdata.(map[string]interface{}); ok {
			questions[qid] = qmap
		}
	}

	return questions
}

// getQuestionText extracts the text field from a question map.
// Returns a default string if text field doesn't exist.
func getQuestionText(question map[string]interface{}) string {
	if text, ok := question["text"].(string); ok {
		return text
	}
	if name, ok := question["name"].(string); ok {
		return name
	}
	return "field"
}

// compareQuestions compares two question objects and returns changes for modified fields.
func compareQuestions(qid string, local, remote map[string]interface{}) []Change {
	var changes []Change

	// Find all field names in both questions
	allFields := make(map[string]bool)
	for field := range local {
		allFields[field] = true
	}
	for field := range remote {
		allFields[field] = true
	}

	// Compare each field
	for field := range allFields {
		localVal, localExists := local[field]
		remoteVal, remoteExists := remote[field]

		if localExists && !remoteExists {
			// Field added in local
			changes = append(changes, Change{
				Type:        ChangeModified,
				Path:        fmt.Sprintf("questions.%s.%s", qid, field),
				Description: fmt.Sprintf("Added field: %s", field),
				NewValue:    localVal,
			})
		} else if !localExists && remoteExists {
			// Field deleted from local
			changes = append(changes, Change{
				Type:        ChangeModified,
				Path:        fmt.Sprintf("questions.%s.%s", qid, field),
				Description: fmt.Sprintf("Deleted field: %s", field),
				OldValue:    remoteVal,
			})
		} else if !deepEqual(localVal, remoteVal) {
			// Field modified
			changes = append(changes, Change{
				Type:        ChangeModified,
				Path:        fmt.Sprintf("questions.%s.%s", qid, field),
				Description: fmt.Sprintf("Modified: %s", field),
				OldValue:    remoteVal,
				NewValue:    localVal,
			})
		}
	}

	return changes
}

// compareProperties compares top-level properties between schemas.
func compareProperties(local, remote map[string]interface{}) []Change {
	var changes []Change

	// Compare properties field if it exists
	localProps, localHasProps := local["properties"].(map[string]interface{})
	remoteProps, remoteHasProps := remote["properties"].(map[string]interface{})

	if !localHasProps && !remoteHasProps {
		return changes
	}

	if localHasProps && !remoteHasProps {
		changes = append(changes, Change{
			Type:        ChangeAdded,
			Path:        "properties",
			Description: "Added properties section",
			NewValue:    localProps,
		})
		return changes
	}

	if !localHasProps && remoteHasProps {
		changes = append(changes, Change{
			Type:        ChangeDeleted,
			Path:        "properties",
			Description: "Deleted properties section",
			OldValue:    remoteProps,
		})
		return changes
	}

	// Both exist - compare individual properties
	allProps := make(map[string]bool)
	for prop := range localProps {
		allProps[prop] = true
	}
	for prop := range remoteProps {
		allProps[prop] = true
	}

	for prop := range allProps {
		localVal, localExists := localProps[prop]
		remoteVal, remoteExists := remoteProps[prop]

		if localExists && !remoteExists {
			changes = append(changes, Change{
				Type:        ChangeModified,
				Path:        fmt.Sprintf("properties.%s", prop),
				Description: fmt.Sprintf("Added property: %s", prop),
				NewValue:    localVal,
			})
		} else if !localExists && remoteExists {
			changes = append(changes, Change{
				Type:        ChangeModified,
				Path:        fmt.Sprintf("properties.%s", prop),
				Description: fmt.Sprintf("Deleted property: %s", prop),
				OldValue:    remoteVal,
			})
		} else if !deepEqual(localVal, remoteVal) {
			changes = append(changes, Change{
				Type:        ChangeModified,
				Path:        fmt.Sprintf("properties.%s", prop),
				Description: fmt.Sprintf("Modified property: %s", prop),
				OldValue:    remoteVal,
				NewValue:    localVal,
			})
		}
	}

	return changes
}

// deepEqual performs a deep comparison of two values.
// Uses JSON marshaling for consistent comparison of complex types.
func deepEqual(a, b interface{}) bool {
	aJSON, err1 := json.Marshal(a)
	bJSON, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(aJSON) == string(bJSON)
}
