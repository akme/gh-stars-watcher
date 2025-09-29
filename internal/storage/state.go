package storage

// StateStorage defines the interface for persisting and loading user state
type StateStorage interface {
	// SaveUserState persists user state to the specified file path
	// Should perform atomic writes to prevent corruption
	SaveUserState(filePath string, state *UserState) error

	// LoadUserState loads user state from the specified file path
	// Returns error if file doesn't exist or is corrupted
	LoadUserState(filePath string) (*UserState, error)
}

// StateFileNotFoundError represents an error when state file doesn't exist
type StateFileNotFoundError struct {
	FilePath string
}

func (e *StateFileNotFoundError) Error() string {
	return "state file not found: " + e.FilePath
}

// StateCorruptionError represents an error when state file is corrupted
type StateCorruptionError struct {
	FilePath string
	Cause    error
}

func (e *StateCorruptionError) Error() string {
	return "state file corrupted at " + e.FilePath + ": " + e.Cause.Error()
}
