# State Storage Interface Contract

## StateStorage Interface

```go
type StateStorage interface {
    // Load retrieves the stored state for a user
    // Returns StateNotFoundError if no state exists for user
    Load(username string) (*UserState, error)
    
    // Save persists the user state atomically
    // Creates backup of previous state before overwriting
    Save(username string, state *UserState) error
    
    // Exists checks if state file exists for user
    Exists(username string) bool
    
    // Delete removes state file for user
    Delete(username string) error
    
    // List returns all usernames with stored states
    List() ([]string, error)
    
    // Cleanup removes state files older than specified duration
    Cleanup(olderThan time.Duration) ([]string, error)
}

type UserState struct {
    Username      string       `json:"username"`
    LastCheck     time.Time    `json:"last_check"`
    Repositories  []Repository `json:"repositories"`
    TotalCount    int          `json:"total_count"`
    StateVersion  string       `json:"state_version"`
    CheckCount    int          `json:"check_count"`
}
```

## Contract Tests

### Load Contract
- **Input**: Existing username → Should return complete UserState with all fields
- **Input**: Non-existent username → Should return StateNotFoundError (not generic error)
- **Input**: Username with corrupted state file → Should return CorruptedStateError with recovery suggestion
- **Output**: Returned UserState must pass validation (valid username, non-future timestamps)
- **Behavior**: Should handle concurrent reads safely

### Save Contract
- **Input**: Valid UserState → Should persist atomically with backup creation
- **Input**: UserState with invalid data → Should return ValidationError before persistence
- **Behavior**: Should create parent directories if they don't exist
- **Behavior**: Failed saves should not corrupt existing state (atomic operation)
- **Behavior**: Should handle concurrent writes safely
- **Post-condition**: Saved state should be immediately loadable via Load method

### Exists Contract
- **Input**: Username with existing state → Should return true
- **Input**: Username without state → Should return false
- **Behavior**: Should be fast (no file parsing, just existence check)

### Delete Contract
- **Input**: Existing state file → Should remove file and return nil error
- **Input**: Non-existent state file → Should return nil error (idempotent)
- **Behavior**: Should also remove backup files if present

### List Contract
- **Output**: Should return all usernames with valid state files
- **Output**: Should exclude corrupted or temporary files
- **Behavior**: Should handle directory access errors gracefully

### Cleanup Contract
- **Input**: Duration parameter → Should remove files older than specified time
- **Output**: Should return list of removed filenames
- **Behavior**: Should preserve files modified within specified duration
- **Behavior**: Should handle file access errors gracefully (partial cleanup is acceptable)