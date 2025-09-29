package github

import (
	"fmt"
	"time"

	"github.com/akme/gh-stars-watcher/internal/storage"
)

// APIResponse represents structured data from GitHub API calls for rate limiting and pagination
type APIResponse struct {
	RateLimit       RateLimitInfo        `json:"rate_limit"`       // Current rate limit status
	PageInfo        PageInfo             `json:"page_info"`        // Pagination metadata
	Repositories    []storage.Repository `json:"repositories"`     // Repository data from current API call
	RequestDuration time.Duration        `json:"request_duration"` // Time taken for API request
	StatusCode      int                  `json:"status_code"`      // HTTP response status code
}

// RateLimitInfo contains GitHub API rate limit information
type RateLimitInfo struct {
	Limit     int       `json:"limit"`      // Maximum requests per hour
	Remaining int       `json:"remaining"`  // Requests remaining in current window
	ResetTime time.Time `json:"reset_time"` // When rate limit resets
	Used      int       `json:"used"`       // Requests used in current window
}

// PageInfo contains pagination metadata for GitHub API responses
type PageInfo struct {
	HasNext    bool   `json:"has_next"`    // Whether more pages are available
	NextCursor string `json:"next_cursor"` // Cursor for next page (GitHub pagination)
	TotalCount int    `json:"total_count"` // Total items across all pages
	PerPage    int    `json:"per_page"`    // Items per page
}

// Validate checks if the APIResponse has valid field values
func (a *APIResponse) Validate() error {
	// RateLimit values must be non-negative
	if a.RateLimit.Limit < 0 {
		return fmt.Errorf("rate limit must be non-negative: %d", a.RateLimit.Limit)
	}
	if a.RateLimit.Remaining < 0 {
		return fmt.Errorf("remaining requests must be non-negative: %d", a.RateLimit.Remaining)
	}
	if a.RateLimit.Used < 0 {
		return fmt.Errorf("used requests must be non-negative: %d", a.RateLimit.Used)
	}

	// ResetTime should be future timestamp during active monitoring
	// Note: We allow past timestamps for completed requests

	// StatusCode must be valid HTTP status code (100-599)
	if a.StatusCode < 100 || a.StatusCode >= 600 {
		return fmt.Errorf("invalid HTTP status code: %d", a.StatusCode)
	}

	// PageInfo validation
	if a.PageInfo.TotalCount < 0 {
		return fmt.Errorf("total count must be non-negative: %d", a.PageInfo.TotalCount)
	}
	if a.PageInfo.PerPage < 0 {
		return fmt.Errorf("per page must be non-negative: %d", a.PageInfo.PerPage)
	}

	// NextCursor validation when HasNext is true
	if a.PageInfo.HasNext && a.PageInfo.NextCursor == "" {
		return fmt.Errorf("next cursor cannot be empty when has_next is true")
	}

	// Validate all repositories
	for i, repo := range a.Repositories {
		if err := repo.Validate(); err != nil {
			return fmt.Errorf("invalid repository at index %d: %v", i, err)
		}
	}

	return nil
}

// StarredOptions contains options for fetching starred repositories
type StarredOptions struct {
	Cursor    string `json:"cursor"`    // Page pagination cursor (empty for first page)
	PerPage   int    `json:"per_page"`  // Number of items per page (max 100)
	Sort      string `json:"sort"`      // Sort order: "created", "updated", "pushed", "full_name"
	Direction string `json:"direction"` // Direction: "asc" or "desc"
}

// StarredResponse represents the response from GetStarredRepositories
type StarredResponse struct {
	Repositories []storage.Repository `json:"repositories"`
	PageInfo     PageInfo             `json:"page_info"`
	RateLimit    RateLimitInfo        `json:"rate_limit"`
}
