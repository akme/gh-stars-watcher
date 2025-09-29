package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup [username]",
	Short: "Remove stored state files for a user",
	Long: `Remove stored state files for a specific user or all users.

This command permanently deletes the stored baseline state, which means the next
monitor run will establish a new baseline from the current starred repositories.

Examples:
  star-watcher cleanup octocat              # Remove state for specific user
  star-watcher cleanup octocat --state-file ./custom-state.json
  star-watcher cleanup --all               # Remove all state files (use with caution)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCleanup,
}

var cleanupAll bool

func init() {
	cleanupCmd.Flags().BoolVar(&cleanupAll, "all", false, "remove all state files (use with caution)")
}

func runCleanup(cmd *cobra.Command, args []string) error {
	if cleanupAll {
		return cleanupAllStateFiles()
	}

	if len(args) == 0 {
		return fmt.Errorf("username required unless --all flag is specified")
	}

	username := args[0]

	// Validate GitHub username format
	if !githubUsernamePattern.MatchString(username) {
		return fmt.Errorf("invalid GitHub username format: %s", username)
	}

	return cleanupUserStateFile(username)
}

func cleanupUserStateFile(username string) error {
	statePath := getStateFilePath(username)

	if verbose {
		log.Printf("Cleaning up state file: %s", statePath)
	}

	// Check if state file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		if !quiet {
			fmt.Printf("No state file found for user: %s\n", username)
		}
		return nil
	}

	// Remove the state file
	if err := os.Remove(statePath); err != nil {
		return fmt.Errorf("failed to remove state file %s: %v", statePath, err)
	}

	// Also remove backup file if it exists
	backupPath := statePath + ".bak"
	if _, err := os.Stat(backupPath); err == nil {
		if err := os.Remove(backupPath); err != nil {
			log.Printf("Warning: failed to remove backup file %s: %v", backupPath, err)
		}
	}

	if !quiet {
		fmt.Printf("Cleaned up state for user: %s\n", username)
	}

	return nil
}

func cleanupAllStateFiles() error {
	if verbose {
		log.Printf("Cleaning up all state files...")
	}

	// Get the default state directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	stateDir := filepath.Join(homeDir, ".star-watcher")

	// Check if state directory exists
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		if !quiet {
			fmt.Println("No state directory found.")
		}
		return nil
	}

	// Read directory contents
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return fmt.Errorf("failed to read state directory: %v", err)
	}

	removedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if strings.HasSuffix(filename, ".json") || strings.HasSuffix(filename, ".bak") {
			filePath := filepath.Join(stateDir, filename)
			if err := os.Remove(filePath); err != nil {
				log.Printf("Warning: failed to remove %s: %v", filePath, err)
			} else {
				removedCount++
				if verbose {
					log.Printf("Removed: %s", filePath)
				}
			}
		}
	}

	// Try to remove the directory if it's empty
	if err := os.Remove(stateDir); err != nil {
		if verbose {
			log.Printf("State directory not empty or could not be removed: %v", err)
		}
	}

	if !quiet {
		fmt.Printf("Cleaned up %d state files.\n", removedCount)
	}

	return nil
}
