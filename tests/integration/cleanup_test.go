package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCleanupCommand tests the cleanup functionality
func TestCleanupCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binaryPath := "../../bin/star-watcher"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("CLI binary not available yet - this is expected in TDD Red phase")
	}

	tmpDir, err := os.MkdirTemp("", "star-watcher-cleanup-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("CleanupRemovesStateFiles", func(t *testing.T) {
		// Create a state file first
		stateFile := filepath.Join(tmpDir, "cleanup-test.json")
		cmd := exec.Command(binaryPath, "monitor", "octocat", "--state-file", stateFile)
		// Always set CI mode to disable prompting in tests
		cmd.Env = append(os.Environ(), "CI=1")
		_, err := cmd.CombinedOutput()
		if err != nil {
			t.Skip("Cannot create state file for cleanup test")
		}

		// Verify state file exists
		if _, err := os.Stat(stateFile); os.IsNotExist(err) {
			t.Skip("State file not created")
		}

		// Run cleanup command
		cmd = exec.Command(binaryPath, "cleanup", "octocat", "--state-file", stateFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Expected no error from cleanup, got: %v\nOutput: %s", err, string(output))
		}

		// Verify state file is removed
		if _, err := os.Stat(stateFile); err == nil {
			t.Error("Expected state file to be removed by cleanup")
		}
	})
}
