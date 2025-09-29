package cli

import (
	"fmt"
	"regexp"
)

// githubUsernamePattern validates GitHub usernames according to GitHub's rules
// GitHub usernames can:
// - be 1-39 characters long
// - contain alphanumeric characters and hyphens
// - not start or end with a hyphen
var githubUsernameValidationPattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$`)

// ValidateGitHubUsername checks if a username follows GitHub's username rules
func ValidateGitHubUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if len(username) > 39 {
		return fmt.Errorf("username too long: %d characters (maximum 39)", len(username))
	}

	if !githubUsernameValidationPattern.MatchString(username) {
		return fmt.Errorf("invalid GitHub username format: %s\nUsername must contain only alphanumeric characters and hyphens, be 1-39 characters long, and not start or end with a hyphen", username)
	}

	return nil
}

// SanitizeGitHubUsername performs basic sanitization on a username
func SanitizeGitHubUsername(username string) string {
	// Remove leading and trailing whitespace
	username = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(username, "")

	// Convert to lowercase (GitHub usernames are case-insensitive)
	return username
}
