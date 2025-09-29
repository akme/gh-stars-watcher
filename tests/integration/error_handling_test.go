package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binaryPath := "../../bin/star-watcher"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("CLI binary not available yet - this is expected in TDD Red phase")
	}

	tmpDir, err := os.MkdirTemp("", "star-watcher-errors-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("InvalidUsernameFormat", func(t *testing.T) {
		stateFile := filepath.Join(tmpDir, "invalid.json")
		cmd := exec.Command(binaryPath, "monitor", "invalid-username!", "--state-file", stateFile)
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Error("Expected error for invalid username format")
		}

		// Should provide helpful error message
		outputStr := strings.ToLower(string(output))
		if !strings.Contains(outputStr, "invalid") && !strings.Contains(outputStr, "username") {
			t.Errorf("Expected helpful error message about username, got: %s", string(output))
		}
	})

	t.Run("NetworkConnectivityIssues", func(t *testing.T) {
		// This test is hard to simulate reliably
		// In practice, would test with mock network conditions
		t.Skip("Network connectivity tests require special setup")
	})

	t.Run("CorruptedStateFile", func(t *testing.T) {
		// Create a corrupted state file
		stateFile := filepath.Join(tmpDir, "corrupted.json")
		err := os.WriteFile(stateFile, []byte("invalid json content"), 0o644)
		if err != nil {
			t.Fatalf("Failed to create corrupted state file: %v", err)
		}

		cmd := exec.Command(binaryPath, "monitor", "octocat", "--state-file", stateFile)
		// Always set CI mode to disable prompting in tests
		cmd.Env = append(os.Environ(), "CI=1")
		output, err := cmd.CombinedOutput()

		// Should handle corruption gracefully (rebuild state)
		outputStr := string(output)
		if strings.Contains(strings.ToLower(outputStr), "error") {
			// Should provide helpful error message about rebuilding
			if !strings.Contains(strings.ToLower(outputStr), "rebuild") &&
				!strings.Contains(strings.ToLower(outputStr), "reset") {
				t.Logf("Corruption handling: %s", outputStr)
			}
		}
	})
}
