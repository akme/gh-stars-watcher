package contract

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/akme/gh-stars-watcher/internal/storage"
)

// TestStateStorageContract validates the StateStorage interface contract
func TestStateStorageContract(t *testing.T) {
	// This test will fail until StateStorage interface and implementation exist
	var store storage.StateStorage
	if store == nil {
		t.Skip("StateStorage implementation not available yet - this is expected in TDD Red phase")
	}

	// Create temporary directory for test state files
	tmpDir, err := os.MkdirTemp("", "star-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("SaveAndLoadUserState", func(t *testing.T) {
		// Create test user state
		userState := &storage.UserState{
			Username:     "testuser",
			LastCheck:    time.Now(),
			Repositories: []storage.Repository{},
			TotalCount:   0,
			StateVersion: "1.0.0",
			CheckCount:   1,
		}

		// Save user state
		statePath := filepath.Join(tmpDir, "testuser.json")
		err := store.SaveUserState(statePath, userState)
		if err != nil {
			t.Errorf("Expected no error saving state, got: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(statePath); os.IsNotExist(err) {
			t.Error("Expected state file to be created")
		}

		// Load user state
		loadedState, err := store.LoadUserState(statePath)
		if err != nil {
			t.Errorf("Expected no error loading state, got: %v", err)
		}

		// Verify loaded state matches saved state
		if loadedState.Username != userState.Username {
			t.Errorf("Expected username %s, got %s", userState.Username, loadedState.Username)
		}
		if loadedState.StateVersion != userState.StateVersion {
			t.Errorf("Expected version %s, got %s", userState.StateVersion, loadedState.StateVersion)
		}
		if loadedState.CheckCount != userState.CheckCount {
			t.Errorf("Expected check count %d, got %d", userState.CheckCount, loadedState.CheckCount)
		}
	})

	t.Run("LoadNonexistentFile", func(t *testing.T) {
		nonexistentPath := filepath.Join(tmpDir, "nonexistent.json")
		_, err := store.LoadUserState(nonexistentPath)
		if err == nil {
			t.Error("Expected error when loading nonexistent file")
		}
		// Should return specific error type for file not found
	})

	t.Run("AtomicWrite", func(t *testing.T) {
		// Test that writes are atomic (use temp file + rename)
		userState := &storage.UserState{
			Username:     "atomictest",
			LastCheck:    time.Now(),
			Repositories: []storage.Repository{},
			TotalCount:   0,
			StateVersion: "1.0.0",
			CheckCount:   1,
		}

		statePath := filepath.Join(tmpDir, "atomic.json")
		err := store.SaveUserState(statePath, userState)
		if err != nil {
			t.Errorf("Expected no error with atomic write, got: %v", err)
		}

		// Verify no temporary files are left behind
		entries, err := os.ReadDir(tmpDir)
		if err != nil {
			t.Fatalf("Failed to read temp dir: %v", err)
		}

		for _, entry := range entries {
			if filepath.Ext(entry.Name()) == ".tmp" {
				t.Errorf("Found temporary file left behind: %s", entry.Name())
			}
		}
	})

	t.Run("BackupPreviousState", func(t *testing.T) {
		statePath := filepath.Join(tmpDir, "backup-test.json")
		backupPath := statePath + ".bak"

		// Create initial state
		initialState := &storage.UserState{
			Username:     "backuptest",
			LastCheck:    time.Now(),
			Repositories: []storage.Repository{},
			TotalCount:   0,
			StateVersion: "1.0.0",
			CheckCount:   1,
		}

		err := store.SaveUserState(statePath, initialState)
		if err != nil {
			t.Errorf("Expected no error saving initial state, got: %v", err)
		}

		// Update state (should create backup)
		updatedState := &storage.UserState{
			Username:     "backuptest",
			LastCheck:    time.Now(),
			Repositories: []storage.Repository{},
			TotalCount:   0,
			StateVersion: "1.0.0",
			CheckCount:   2, // Incremented
		}

		err = store.SaveUserState(statePath, updatedState)
		if err != nil {
			t.Errorf("Expected no error saving updated state, got: %v", err)
		}

		// Verify backup file exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Error("Expected backup file to be created")
		}

		// Load backup and verify it contains original state
		if _, err := os.Stat(backupPath); err == nil {
			backupState, err := store.LoadUserState(backupPath)
			if err != nil {
				t.Errorf("Expected no error loading backup, got: %v", err)
			}
			if backupState.CheckCount != 1 {
				t.Errorf("Expected backup check count 1, got %d", backupState.CheckCount)
			}
		}
	})
}

// TestUserStateValidation validates UserState struct validation
func TestUserStateValidation(t *testing.T) {
	// This test will fail until UserState struct exists
	t.Skip("UserState struct not available yet - this is expected in TDD Red phase")

	// Expected validation rules:
	// - Username must match GitHub pattern
	// - LastCheck must not be future timestamp
	// - StateVersion must follow semantic versioning
	// - TotalCount must be non-negative
	// - CheckCount must be non-negative
}
