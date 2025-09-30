package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	verbose   bool
	quiet     bool
	stateFile string
	output    string
	authToken bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "star-watcher",
	Short: "Monitor GitHub user's starred repositories for changes",
	Long: `GitHub Stars Monitor CLI tracks changes in a user's starred repositories,
showing only newly starred repositories between runs.

The tool fetches current starred repos via GitHub API, persists state locally 
for comparison, and outputs changes with configurable formats and authentication methods.`,
	PersistentPreRun: setupLogging,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global persistent flags available to all commands
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output (detailed logging)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output (errors only)")
	rootCmd.PersistentFlags().StringVar(&stateFile, "state-file", "", "custom state file path (default: ~/.star-watcher/{username}.json)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "output format: text, json")
	rootCmd.PersistentFlags().BoolVarP(&authToken, "auth", "a", false, "prompt for GitHub token for authenticated requests (higher rate limits)")

	// Add subcommands
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(cleanupCmd)
}

// setupLogging configures logging based on verbosity flags
func setupLogging(cmd *cobra.Command, args []string) {
	if quiet && verbose {
		fmt.Fprintf(os.Stderr, "Warning: Both --quiet and --verbose specified. Using verbose mode.\n")
		quiet = false
	}

	if quiet {
		// Suppress all output except errors
		log.SetOutput(os.Stderr)
		log.SetFlags(0) // No timestamps in quiet mode
	} else if verbose {
		// Detailed logging with timestamps
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		// Normal logging - minimal
		log.SetOutput(os.Stderr)
		log.SetFlags(0)
	}
}

// getStateFilePath returns the state file path for a username
func getStateFilePath(username string) string {
	if stateFile != "" {
		return stateFile
	}

	// Default: ~/.star-watcher/{username}.json
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Sprintf(".star-watcher-%s.json", username)
	}

	stateDir := fmt.Sprintf("%s/.star-watcher", homeDir)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return fmt.Sprintf(".star-watcher-%s.json", username)
	}

	return fmt.Sprintf("%s/%s.json", stateDir, username)
}
