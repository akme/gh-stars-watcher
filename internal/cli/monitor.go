package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

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
	Use:   "monitor [username or usernames]",
	Short: "Monitor GitHub user(s) starred repositories for changes",
	Long: `Monitor one or more GitHub users' starred repositories and display only newly starred repositories since the last run.

For single users, the command works as before. For multiple users, provide a comma-separated list.
On the first run, this command establishes a baseline of currently starred repositories and shows no output.
Subsequent runs compare against the stored state and display only newly starred repositories.

By default, the tool uses unauthenticated GitHub API access (60 requests/hour). Use --auth to prompt for a token for higher rate limits (5000 requests/hour).

Examples:
  star-watcher monitor octocat
  star-watcher monitor octocat,github,torvalds --output json
  star-watcher monitor user1,user2 --verbose
  star-watcher monitor octocat --auth --verbose
  star-watcher monitor octocat --state-file ./custom-state.json`,
	Args: cobra.ExactArgs(1),
	RunE: runMonitor,
}

// parseUsernames parses the input string as either a single username or comma-separated usernames
func parseUsernames(input string) ([]string, error) {
	// Split by comma and trim whitespace
	rawUsernames := strings.Split(input, ",")
	usernames := make([]string, 0, len(rawUsernames))

	for _, username := range rawUsernames {
		username = strings.TrimSpace(username)
		if username == "" {
			continue // Skip empty strings
		}

		// Validate GitHub username format
		if !githubUsernamePattern.MatchString(username) {
			return nil, fmt.Errorf("invalid GitHub username format: %s\nUsername must contain only alphanumeric characters and hyphens, be 1-39 characters long, and not start or end with a hyphen", username)
		}

		usernames = append(usernames, username)
	}

	if len(usernames) == 0 {
		return nil, fmt.Errorf("no valid usernames provided")
	}

	return usernames, nil
}

func runMonitor(cmd *cobra.Command, args []string) error {
	usernames, err := parseUsernames(args[0])
	if err != nil {
		return err
	}

	if verbose {
		if len(usernames) == 1 {
			log.Printf("Starting monitor for user: %s", usernames[0])
		} else {
			log.Printf("Starting monitor for %d users: %s", len(usernames), strings.Join(usernames, ", "))
		}
		log.Printf("Output format: %s", output)
	}

	ctx := cmd.Context()

	// Handle single user (existing behavior)
	if len(usernames) == 1 {
		return runSingleUserMonitor(ctx, usernames[0])
	}

	// Handle multiple users
	return runMultiUserMonitor(ctx, usernames)
}

// runSingleUserMonitor handles monitoring for a single user (preserves existing behavior)
func runSingleUserMonitor(ctx context.Context, username string) error {
	if verbose {
		log.Printf("State file: %s", getStateFilePath(username))
	}

	// Create monitoring service with real implementations
	service, err := createMonitoringService()
	if err != nil {
		return fmt.Errorf("failed to create monitoring service: %w", err)
	}

	// Execute monitoring
	result, err := service.MonitorUser(ctx, username, getStateFilePath(username))
	if err != nil {
		if !quiet && output != "json" {
			fmt.Print("\r\033[K") // Clear the line completely before error
		}
		return fmt.Errorf("monitoring failed: %w", err)
	}

	if !quiet && output != "json" {
		fmt.Print("\r\033[K") // Clear the line completely before results
	}

	// Format and display results
	formatter := NewOutputFormatter(os.Stdout, output)
	return formatter.FormatMonitorResult(result)
}

// runMultiUserMonitor handles monitoring for multiple users with parallel processing
func runMultiUserMonitor(ctx context.Context, usernames []string) error {
	results := make(map[string]*monitor.MonitorResult)
	errors := make(map[string]error)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create monitoring service (shared for all users)
	service, err := createMonitoringService()
	if err != nil {
		return fmt.Errorf("failed to create monitoring service: %w", err)
	}

	// Process users in parallel
	for _, username := range usernames {
		wg.Add(1)
		go func(user string) {
			defer wg.Done()

			if verbose {
				log.Printf("Processing user: %s", user)
			}

			result, err := service.MonitorUser(ctx, user, getStateFilePath(user))

			mu.Lock()
			if err != nil {
				errors[user] = err
			} else {
				results[user] = result
			}
			mu.Unlock()
		}(username)
	}

	// Wait for all users to complete
	wg.Wait()

	if !quiet && output != "json" {
		fmt.Print("\r\033[K") // Clear the line completely before results
	}

	// Format and display results
	formatter := NewOutputFormatter(os.Stdout, output)
	return formatter.FormatMultiUserResults(results, errors)
}

// createMonitoringService creates a complete monitoring service with real implementations
func createMonitoringService() (*monitor.Service, error) {
	// Create GitHub client with empty token (will try to get from environment or keychain)
	githubClient := github.NewAPIClient("")

	// Create storage
	jsonStorage := storage.NewJSONStorage()

	// Create keychain authentication with the GitHub client as validator
	keychainAuth := auth.NewKeychainTokenManager(githubClient)

	// Check if we should use interactive prompts based on CLI flag and environment
	var tokenManager auth.TokenManager
	if authToken && os.Getenv("CI") == "" && isInteractiveTerminal() {
		// Interactive mode with explicit --auth flag: allow prompting
		tokenManager = auth.NewPromptTokenManager(keychainAuth)
	} else {
		// Default mode: only use existing tokens (keychain/environment), don't prompt
		tokenManager = keychainAuth
	}

	// Create monitoring service with configuration adjusted for verbosity
	cfg := config.DefaultConfig()

	// Adjust logging configuration based on CLI flags
	if quiet {
		cfg.Logging.LogLevel = "error"
		cfg.Logging.EnableAuditLog = false
		cfg.Logging.EnablePerformanceMetrics = false
		cfg.Logging.LogAPICallsSaved = false
	} else if verbose {
		cfg.Logging.LogLevel = "debug"
		cfg.Logging.EnableAuditLog = true
		cfg.Logging.EnablePerformanceMetrics = true
		cfg.Logging.LogAPICallsSaved = true
	} else {
		// Normal mode - less verbose than current default
		cfg.Logging.LogLevel = "warn"
		cfg.Logging.EnableAuditLog = false
		cfg.Logging.EnablePerformanceMetrics = false
		cfg.Logging.LogAPICallsSaved = false
	}

	service := monitor.NewService(githubClient, jsonStorage, tokenManager, cfg)

	// Set up progress callback only for non-JSON output to avoid polluting JSON
	if output != "json" && !quiet {
		if verbose {
			// Verbose mode: show all progress messages
			service.SetProgressCallback(func(message string) {
				// Clear the line and write the message
				fmt.Printf("\r\033[K%s", message)
			})
		} else {
			// Normal mode: only show essential progress messages
			service.SetProgressCallback(func(message string) {
				// Only show high-level progress, filter out technical details
				if isEssentialProgress(message) {
					// Clear the line and write the message
					fmt.Printf("\r\033[K%s", message)
				}
			})
		}
	}

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

// isEssentialProgress determines if a progress message should be shown in normal (non-verbose) mode
func isEssentialProgress(message string) bool {
	essentialPrefixes := []string{
		"Starting monitor",
		"Validating user",
		"Loading previous state",
		"Fetching starred repositories",
		"Analyzing repository changes",
		"Updating state",
		"Monitor complete",
		"Full sync completed",
		"Incremental fetch completed",
	}

	for _, prefix := range essentialPrefixes {
		if len(message) >= len(prefix) && message[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}
