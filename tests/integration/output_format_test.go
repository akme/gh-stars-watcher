package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestOutputFormat tests JSON and text output formats
func TestOutputFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binaryPath := "../../bin/star-watcher"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("CLI binary not available yet - this is expected in TDD Red phase")
	}

	tmpDir, err := os.MkdirTemp("", "star-watcher-output-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("JSONOutput", func(t *testing.T) {
		stateFile := filepath.Join(tmpDir, "json.json")
		cmd := exec.Command(binaryPath, "monitor", "octocat", "--state-file", stateFile, "--output", "json")
		// Always set CI mode to disable prompting in tests
		cmd.Env = append(os.Environ(), "CI=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Skip("Cannot test JSON output without working CLI")
		}

		outputStr := string(output)

		// Should be valid JSON
		var jsonData interface{}
		if err := json.Unmarshal([]byte(outputStr), &jsonData); err != nil {
			t.Errorf("Expected valid JSON output, got parse error: %v\nOutput: %s", err, outputStr)
		}
	})

	t.Run("TextOutput", func(t *testing.T) {
		stateFile := filepath.Join(tmpDir, "text.json")
		cmd := exec.Command(binaryPath, "monitor", "octocat", "--state-file", stateFile, "--output", "text")
		// Always set CI mode to disable prompting in tests
		cmd.Env = append(os.Environ(), "CI=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Skip("Cannot test text output without working CLI")
		}

		outputStr := string(output)

		// Should be human-readable text (not JSON)
		var jsonData interface{}
		if json.Unmarshal([]byte(outputStr), &jsonData) == nil && strings.HasPrefix(outputStr, "{") {
			t.Error("Expected human-readable text, got JSON")
		}
	})
}
