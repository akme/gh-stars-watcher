package storage

import (
	"fmt"
	"net/url"
	"regexp"
	"time"
)

// Repository represents a starred GitHub repository with all metadata needed for comparison and display
type Repository struct {
	FullName    string    `json:"full_name"`   // Owner/repo format (e.g., "microsoft/vscode")
	Description string    `json:"description"` // Repository description (nullable)
	StarCount   int       `json:"star_count"`  // Current number of stars
	UpdatedAt   time.Time `json:"updated_at"`  // Last repository update timestamp
	URL         string    `json:"url"`         // Repository URL for browser access
	StarredAt   time.Time `json:"starred_at"`  // When user starred this repository
	Language    string    `json:"language"`    // Primary programming language (optional)
	Private     bool      `json:"private"`     // Whether repository is private
}

// githubRepoNamePattern validates GitHub repository full names
var githubRepoNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`)

// Validate checks if the Repository has valid field values
func (r *Repository) Validate() error {
	// FullName must match GitHub pattern
	if !githubRepoNamePattern.MatchString(r.FullName) {
		return fmt.Errorf("invalid repository full name format: %s", r.FullName)
	}

	// URL must be valid HTTPS GitHub repository URL
	if r.URL != "" {
		parsedURL, err := url.Parse(r.URL)
		if err != nil {
			return fmt.Errorf("invalid repository URL: %v", err)
		}
		if parsedURL.Scheme != "https" {
			return fmt.Errorf("repository URL must use HTTPS: %s", r.URL)
		}
		if parsedURL.Host != "github.com" {
			return fmt.Errorf("repository URL must be on github.com: %s", r.URL)
		}
	}

	// StarCount must be non-negative
	if r.StarCount < 0 {
		return fmt.Errorf("star count must be non-negative: %d", r.StarCount)
	}

	// StarredAt must not be future timestamp
	if r.StarredAt.After(time.Now()) {
		return fmt.Errorf("starred timestamp cannot be in the future: %v", r.StarredAt)
	}

	return nil
}

// String returns a human-readable representation of the repository
func (r *Repository) String() string {
	return fmt.Sprintf("%s (%d stars) - %s", r.FullName, r.StarCount, r.Description)
}

// UserState represents the persisted state for a GitHub user's monitoring session
type UserState struct {
	Username     string       `json:"username"`      // GitHub username being monitored
	LastCheck    time.Time    `json:"last_check"`    // Timestamp of last successful check
	Repositories []Repository `json:"repositories"`  // Previously seen starred repositories
	TotalCount   int          `json:"total_count"`   // Total repositories at last check (for pagination validation)
	StateVersion string       `json:"state_version"` // Schema version for backward compatibility
	CheckCount   int          `json:"check_count"`   // Number of successful checks performed
}

// githubUsernamePattern validates GitHub usernames
var githubUsernamePattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$`)

// semanticVersionPattern validates semantic version strings
var semanticVersionPattern = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// Validate checks if the UserState has valid field values
func (u *UserState) Validate() error {
	// Username must match GitHub username pattern
	if !githubUsernamePattern.MatchString(u.Username) {
		return fmt.Errorf("invalid GitHub username format: %s", u.Username)
	}

	// LastCheck must not be future timestamp
	if u.LastCheck.After(time.Now()) {
		return fmt.Errorf("last check timestamp cannot be in the future: %v", u.LastCheck)
	}

	// StateVersion must follow semantic versioning pattern
	if u.StateVersion != "" && !semanticVersionPattern.MatchString(u.StateVersion) {
		return fmt.Errorf("invalid semantic version format: %s", u.StateVersion)
	}

	// TotalCount must be non-negative
	if u.TotalCount < 0 {
		return fmt.Errorf("total count must be non-negative: %d", u.TotalCount)
	}

	// CheckCount must be non-negative
	if u.CheckCount < 0 {
		return fmt.Errorf("check count must be non-negative: %d", u.CheckCount)
	}

	// Validate all repositories
	for i, repo := range u.Repositories {
		if err := repo.Validate(); err != nil {
			return fmt.Errorf("invalid repository at index %d: %v", i, err)
		}
	}

	return nil
}

// NewUserState creates a new UserState with default values
func NewUserState(username string) *UserState {
	return &UserState{
		Username:     username,
		LastCheck:    time.Time{}, // Zero time for first run
		Repositories: make([]Repository, 0),
		TotalCount:   0,
		StateVersion: "1.0.0",
		CheckCount:   0,
	}
}
