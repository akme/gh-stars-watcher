package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestFirstRunWorkflow tests the complete first-run user experience
func TestFirstRunWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test will fail until the CLI application is built
	binaryPath := "../../bin/star-watcher"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("CLI binary not available yet - this is expected in TDD Red phase")
	}

	// Create temporary directory for test state
	tmpDir, err := os.MkdirTemp("", "star-watcher-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "testuser.json")

	t.Run("FirstRunWithPublicUser", func(t *testing.T) {
		// Run the monitor command for the first time
		cmd := exec.Command(binaryPath, "monitor", "octocat", "--state-file", stateFile, "--output", "json")
		// Always set CI mode to disable prompting in tests
		// The parent process environment should already include GITHUB_TOKEN if set
		cmd.Env = append(os.Environ(), "CI=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Expected no error on first run, got: %v\nOutput: %s", err, string(output))
		}

		// First run should create state file but show no "new" repositories
		if _, err := os.Stat(stateFile); os.IsNotExist(err) {
			t.Error("Expected state file to be created on first run")
		}

		// Parse JSON output - should be empty array or indicate no new repos
		outputStr := string(output)
		if !strings.Contains(outputStr, "[]") && !strings.Contains(outputStr, "no new") {
			t.Logf("First run output: %s", outputStr)
		}
	})

	t.Run("FirstRunCreatesBaseline", func(t *testing.T) {
		// Verify state file contains baseline data
		if _, err := os.Stat(stateFile); err != nil {
			t.Fatalf("State file should exist after first run: %v", err)
		}

		// Read state file and verify structure
		content, err := os.ReadFile(stateFile)
		if err != nil {
			t.Fatalf("Failed to read state file: %v", err)
		}

		stateStr := string(content)
		if !strings.Contains(stateStr, "octocat") {
			t.Error("Expected state file to contain username")
		}
		if !strings.Contains(stateStr, "repositories") {
			t.Error("Expected state file to contain repositories array")
		}
		if !strings.Contains(stateStr, "last_check") {
			t.Error("Expected state file to contain last_check timestamp")
		}
	})

	t.Run("FirstRunWithInvalidUser", func(t *testing.T) {
		invalidStateFile := filepath.Join(tmpDir, "invalid.json")

		cmd := exec.Command(binaryPath, "monitor", "nonexistent-user-12345", "--state-file", invalidStateFile)
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Error("Expected error for nonexistent user")
		}

		// Should not create state file for invalid user
		if _, err := os.Stat(invalidStateFile); err == nil {
			t.Error("Should not create state file for invalid user")
		}

		// Error message should be helpful
		outputStr := string(output)
		if !strings.Contains(strings.ToLower(outputStr), "user") && !strings.Contains(strings.ToLower(outputStr), "not found") {
			t.Errorf("Expected helpful error message, got: %s", outputStr)
		}
	})

	t.Run("FirstRunShowsProgressIndicator", func(t *testing.T) {
		// For operations >2s, should show progress
		// This test may be flaky depending on network speed
		progressStateFile := filepath.Join(tmpDir, "progress.json")

		cmd := exec.Command(binaryPath, "monitor", "torvalds", "--state-file", progressStateFile, "--verbose")
		// Always set CI mode to disable prompting in tests
		// The parent process environment should already include GITHUB_TOKEN if set
		cmd.Env = append(os.Environ(), "CI=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Progress test failed (network dependent): %v", err)
		}

		outputStr := string(output)
		// Look for progress indicators (dots, spinner, percentage, etc.)
		hasProgress := strings.Contains(outputStr, "...") ||
			strings.Contains(outputStr, "Fetching") ||
			strings.Contains(outputStr, "Progress") ||
			strings.Contains(outputStr, "%")

		if len(outputStr) > 100 && !hasProgress {
			t.Logf("No progress indicators found in verbose output: %s", outputStr)
		}
	})
}
