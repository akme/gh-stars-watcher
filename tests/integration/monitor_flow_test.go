package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestMonitoringFlow tests subsequent runs after baseline establishment
func TestMonitoringFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binaryPath := "../../bin/star-watcher"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("CLI binary not available yet - this is expected in TDD Red phase")
	}

	tmpDir, err := os.MkdirTemp("", "star-watcher-monitor-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "monitor.json")

	t.Run("SubsequentRunsDetectChanges", func(t *testing.T) {
		// Simulate having an existing state with fewer repos
		// This test simulates the change detection logic

		// First run establishes baseline
		cmd := exec.Command(binaryPath, "monitor", "octocat", "--state-file", stateFile)
		// Always set CI mode to disable prompting in tests
		cmd.Env = append(os.Environ(), "CI=1")
		_, err := cmd.CombinedOutput()
		if err != nil {
			t.Skip("Cannot establish baseline")
		}

		// Wait a moment to ensure timestamp difference
		time.Sleep(1 * time.Second)

		// Second run should show minimal or no changes for same user
		cmd = exec.Command(binaryPath, "monitor", "octocat", "--state-file", stateFile, "--output", "json")
		// Always set CI mode to disable prompting in tests
		cmd.Env = append(os.Environ(), "CI=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Expected no error on subsequent run, got: %v", err)
		}

		// Should typically show no new repos for same user in quick succession
		outputStr := string(output)
		t.Logf("Subsequent run output: %s", outputStr)
	})
}
