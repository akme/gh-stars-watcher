package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// JSONStorage implements the StateStorage interface using JSON files
type JSONStorage struct{}

// NewJSONStorage creates a new JSON storage implementation
func NewJSONStorage() *JSONStorage {
	return &JSONStorage{}
}

// SaveUserState persists user state to the specified file path with atomic writes
func (j *JSONStorage) SaveUserState(filePath string, state *UserState) error {
	// Validate the state before saving
	if err := state.Validate(); err != nil {
		return fmt.Errorf("invalid user state: %v", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Create backup of existing file if it exists
	if _, err := os.Stat(filePath); err == nil {
		backupPath := filePath + ".bak"
		if err := copyFile(filePath, backupPath); err != nil {
			// Log warning but don't fail the save operation
			fmt.Fprintf(os.Stderr, "Warning: failed to create backup: %v\n", err)
		}
	}

	// Atomic write: write to temporary file first, then rename
	tempFile := filePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer func() {
		file.Close()
		// Clean up temp file if something goes wrong
		if _, err := os.Stat(tempFile); err == nil {
			os.Remove(tempFile)
		}
	}()

	// Write JSON with indentation for human readability
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(state); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}

	// Close file before rename
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	return nil
}

// LoadUserState loads user state from the specified file path
func (j *JSONStorage) LoadUserState(filePath string) (*UserState, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, &StateFileNotFoundError{FilePath: filePath}
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %v", err)
	}

	// Parse JSON
	var state UserState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, &StateCorruptionError{
			FilePath: filePath,
			Cause:    err,
		}
	}

	// Validate loaded state
	if err := state.Validate(); err != nil {
		return nil, &StateCorruptionError{
			FilePath: filePath,
			Cause:    fmt.Errorf("validation failed: %v", err),
		}
	}

	return &state, nil
}

// copyFile creates a copy of a file for backup purposes
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
