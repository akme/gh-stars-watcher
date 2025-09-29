package cli

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/akme/gh-stars-watcher/internal/auth"
	"github.com/akme/gh-stars-watcher/internal/config"
	"github.com/akme/gh-stars-watcher/internal/github"
	"github.com/akme/gh-stars-watcher/internal/monitor"
	"github.com/akme/gh-stars-watcher/internal/storage"
	"github.com/spf13/cobra"
)

// githubUsernamePattern validates GitHub usernames
var githubUsernamePattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$`)

// monitorCmd represents the monitor command
var monitorCmd = &cobra.Command{
	Use:   "monitor [username]",
	Short: "Monitor a GitHub user's starred repositories for changes",
	Long: `Monitor a GitHub user's starred repositories and display only newly starred repositories since the last run.

On the first run, this command establishes a baseline of currently starred repositories and shows no output.
Subsequent runs compare against the stored state and display only newly starred repositories.

Examples:
  star-watcher monitor octocat
  star-watcher monitor octocat --output json
  star-watcher monitor octocat --state-file ./custom-state.json
  star-watcher monitor octocat --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: runMonitor,
}

func runMonitor(cmd *cobra.Command, args []string) error {
	username := args[0]

	// Validate GitHub username format
	if !githubUsernamePattern.MatchString(username) {
		return fmt.Errorf("invalid GitHub username format: %s\nUsername must contain only alphanumeric characters and hyphens, be 1-39 characters long, and not start or end with a hyphen", username)
	}

	if verbose {
		log.Printf("Starting monitor for user: %s", username)
		log.Printf("Output format: %s", output)
		log.Printf("State file: %s", getStateFilePath(username))
	}

	// Create monitoring service with real implementations
	ctx := cmd.Context()
	service, err := createMonitoringService()
	if err != nil {
		return fmt.Errorf("failed to create monitoring service: %w", err)
	}

	// Execute monitoring
	result, err := service.MonitorUser(ctx, username, getStateFilePath(username))
	if err != nil {
		if !quiet {
			fmt.Println() // New line after progress
		}
		return fmt.Errorf("monitoring failed: %w", err)
	}

	if !quiet {
		fmt.Println() // New line after progress
	}

	// Format and display results
	formatter := NewOutputFormatter(os.Stdout, output)
	return formatter.FormatMonitorResult(result)
}

// createMonitoringService creates a complete monitoring service with real implementations
func createMonitoringService() (*monitor.Service, error) {
	// Create GitHub client with empty token (will try to get from environment or keychain)
	githubClient := github.NewAPIClient("")

	// Create storage
	jsonStorage := storage.NewJSONStorage()

	// Create keychain authentication with the GitHub client as validator
	keychainAuth := auth.NewKeychainTokenManager(githubClient)

	// Check if we should use interactive prompts (not in tests or CI)
	var tokenManager auth.TokenManager
	if os.Getenv("CI") != "" || !isInteractiveTerminal() {
		// Non-interactive mode: don't prompt for tokens
		tokenManager = keychainAuth
	} else {
		// Interactive mode: allow prompting
		tokenManager = auth.NewPromptTokenManager(keychainAuth)
	}

	// Create monitoring service with default configuration
	cfg := config.DefaultConfig()
	service := monitor.NewService(githubClient, jsonStorage, tokenManager, cfg)

	return service, nil
}

// isInteractiveTerminal checks if we're running in an interactive terminal
func isInteractiveTerminal() bool {
	// Check if stdout is a terminal and stdin is available
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	// Check if stdin is a pipe/redirect (non-interactive)
	return (stat.Mode() & os.ModeCharDevice) != 0
}
