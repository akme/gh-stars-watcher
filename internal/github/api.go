package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/akme/gh-stars-watcher/internal/storage"
	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

// APIClient implements the GitHubClient interface using go-github
type APIClient struct {
	client *github.Client
}

// NewAPIClient creates a new GitHub API client
func NewAPIClient(token string) *APIClient {
	var client *github.Client

	if token != "" {
		// Authenticated client
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		client = github.NewClient(tc)
	} else {
		// Unauthenticated client
		client = github.NewClient(nil)
	}

	return &APIClient{
		client: client,
	}
}

// GetStarredRepositories fetches all starred repositories for a user
func (a *APIClient) GetStarredRepositories(ctx context.Context, username string, opts *StarredOptions) (*StarredResponse, error) {
	// Set default options
	if opts == nil {
		opts = &StarredOptions{}
	}

	// GitHub API options
	listOpts := &github.ActivityListStarredOptions{
		Sort:      opts.Sort,
		Direction: opts.Direction,
		ListOptions: github.ListOptions{
			PerPage: opts.PerPage,
		},
	}

	// Set defaults
	if listOpts.Sort == "" {
		listOpts.Sort = "created"
	}
	if listOpts.Direction == "" {
		listOpts.Direction = "desc"
	}
	if listOpts.PerPage == 0 {
		listOpts.PerPage = 30
	}
	if listOpts.PerPage > 100 {
		listOpts.PerPage = 100 // GitHub API maximum
	}

	// Handle pagination cursor
	if opts.Cursor != "" {
		// Parse cursor as page number for simplicity
		if page, err := strconv.Atoi(opts.Cursor); err == nil && page > 0 {
			listOpts.Page = page
		}
	}

	// Make API call
	starred, resp, err := a.client.Activity.ListStarred(ctx, username, listOpts)
	if err != nil {
		// Handle specific GitHub API errors
		if strings.Contains(err.Error(), "404") {
			return nil, &UserNotFoundError{Username: username}
		}
		if strings.Contains(err.Error(), "403") && strings.Contains(err.Error(), "rate limit") {
			return nil, &RateLimitError{
				ResetTime: time.Now().Add(time.Hour).Format(time.RFC3339),
				Limit:     5000,
				Used:      5000,
			}
		}
		return nil, fmt.Errorf("GitHub API error: %v", err)
	}

	// Convert GitHub repositories to our Repository model
	repositories := make([]storage.Repository, len(starred))
	for i, star := range starred {
		repo := star.GetRepository()
		repositories[i] = storage.Repository{
			FullName:    repo.GetFullName(),
			Description: repo.GetDescription(),
			StarCount:   repo.GetStargazersCount(),
			UpdatedAt:   repo.GetUpdatedAt().Time,
			URL:         repo.GetHTMLURL(),
			StarredAt:   star.GetStarredAt().Time,
			Language:    repo.GetLanguage(),
			Private:     repo.GetPrivate(),
		}
	}

	// Build response
	response := &StarredResponse{
		Repositories: repositories,
		PageInfo: PageInfo{
			HasNext:    resp.NextPage > 0,
			NextCursor: "",
			TotalCount: len(repositories), // GitHub doesn't provide total count easily
			PerPage:    listOpts.PerPage,
		},
		RateLimit: RateLimitInfo{
			Limit:     resp.Rate.Limit,
			Remaining: resp.Rate.Remaining,
			ResetTime: resp.Rate.Reset.Time,
			Used:      resp.Rate.Limit - resp.Rate.Remaining,
		},
	}

	// Set next cursor if there are more pages
	if resp.NextPage > 0 {
		response.PageInfo.NextCursor = strconv.Itoa(resp.NextPage)
	}

	return response, nil
}

// GetRateLimit returns current rate limit status
func (a *APIClient) GetRateLimit(ctx context.Context) (*RateLimitInfo, error) {
	rateLimits, _, err := a.client.RateLimits(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limits: %v", err)
	}

	core := rateLimits.GetCore()
	return &RateLimitInfo{
		Limit:     core.Limit,
		Remaining: core.Remaining,
		ResetTime: core.Reset.Time,
		Used:      core.Limit - core.Remaining,
	}, nil
}

// ValidateUser checks if a GitHub username exists
func (a *APIClient) ValidateUser(ctx context.Context, username string) error {
	_, _, err := a.client.Users.Get(ctx, username)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return &UserNotFoundError{Username: username}
		}
		return fmt.Errorf("failed to validate user: %v", err)
	}
	return nil
}

// ValidateToken validates a GitHub personal access token
func (a *APIClient) ValidateToken(ctx context.Context, token string) (bool, error) {
	// Create a temporary client with the token
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	tempClient := github.NewClient(tc)

	// Try to get the authenticated user to validate the token
	_, _, err := tempClient.Users.Get(ctx, "")
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			return false, nil // Token is invalid but no error occurred
		}
		return false, fmt.Errorf("failed to validate token: %v", err)
	}
	return true, nil
}
