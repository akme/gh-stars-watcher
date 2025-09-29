package github

import (
	"context"
)

// GitHubClient defines the interface for interacting with the GitHub API
type GitHubClient interface {
	// GetStarredRepositories fetches all starred repositories for a user
	// Returns paginated results with rate limit information
	GetStarredRepositories(ctx context.Context, username string, opts *StarredOptions) (*StarredResponse, error)

	// GetRateLimit returns current rate limit status
	GetRateLimit(ctx context.Context) (*RateLimitInfo, error)

	// ValidateUser checks if a GitHub username exists
	ValidateUser(ctx context.Context, username string) error
}

// UserNotFoundError represents an error when a GitHub user doesn't exist
type UserNotFoundError struct {
	Username string
}

func (e *UserNotFoundError) Error() string {
	return "GitHub user not found: " + e.Username
}

// RateLimitError represents an error when API rate limit is exceeded
type RateLimitError struct {
	ResetTime string
	Limit     int
	Used      int
}

func (e *RateLimitError) Error() string {
	return "GitHub API rate limit exceeded. Resets at: " + e.ResetTime
}
