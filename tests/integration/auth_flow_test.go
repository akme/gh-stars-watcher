package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAuthenticationFlow tests various authentication methods
func TestAuthenticationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binaryPath := "../../bin/star-watcher"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("CLI binary not available yet - this is expected in TDD Red phase")
	}

	tmpDir, err := os.MkdirTemp("", "star-watcher-auth-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("UnauthenticatedAccess", func(t *testing.T) {
		// Clear any existing token
		os.Unsetenv("GITHUB_TOKEN")

		stateFile := filepath.Join(tmpDir, "unauth.json")
		cmd := exec.Command(binaryPath, "monitor", "octocat", "--state-file", stateFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Unauthenticated access may have rate limits: %v", err)
		}

		outputStr := string(output)
		// Should work with public repos, may have rate limiting warnings
		if strings.Contains(strings.ToLower(outputStr), "rate limit") {
			t.Log("Rate limiting encountered as expected for unauthenticated access")
		}
	})

	t.Run("EnvironmentTokenAuth", func(t *testing.T) {
		// This test requires a valid GitHub token to be meaningful
		// In real CI, this would use a test token
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			t.Skip("No GITHUB_TOKEN set - skipping authenticated test")
		}

		stateFile := filepath.Join(tmpDir, "auth.json")
		cmd := exec.Command(binaryPath, "monitor", "octocat", "--state-file", stateFile, "--verbose")
		// Pass the environment variable to the child process with CI mode
		cmd.Env = append(os.Environ(), "CI=1", "GITHUB_TOKEN="+token)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Expected no error with authentication, got: %v", err)
		}

		outputStr := string(output)
		// Should show higher rate limits or successful authenticated access
		if strings.Contains(strings.ToLower(outputStr), "authenticated") ||
			strings.Contains(outputStr, "5000") { // Higher rate limit
			t.Log("Authenticated access successful")
		}
	})
}
