package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/akme/gh-stars-watcher/internal/auth"
	"github.com/akme/gh-stars-watcher/internal/github"
	"github.com/akme/gh-stars-watcher/internal/storage"
)

// Service provides the core monitoring functionality
type Service struct {
	githubClient github.GitHubClient
	storage      storage.StateStorage
	tokenManager auth.TokenManager
	progressFunc func(message string) // Optional progress callback
}

// NewService creates a new monitoring service
func NewService(githubClient github.GitHubClient, storage storage.StateStorage, tokenManager auth.TokenManager) *Service {
	return &Service{
		githubClient: githubClient,
		storage:      storage,
		tokenManager: tokenManager,
	}
}

// SetProgressCallback sets a callback function for progress updates
func (s *Service) SetProgressCallback(callback func(message string)) {
	s.progressFunc = callback
}

// MonitorUser monitors a GitHub user's starred repositories
func (s *Service) MonitorUser(ctx context.Context, username, stateFilePath string) (*MonitorResult, error) {
	s.progress("Starting monitor for user: " + username)

	// Try to get authentication token and create authenticated client if available
	if token, source, err := s.tokenManager.GetToken(ctx); err == nil && token != "" {
		s.progress("Using authentication from " + source)
		// Create new authenticated GitHub client
		s.githubClient = github.NewAPIClient(token)
	} else {
		s.progress("Using unauthenticated access (rate limits may apply)")
	}

	// Validate username
	s.progress("Validating user exists...")
	if err := s.githubClient.ValidateUser(ctx, username); err != nil {
		return nil, fmt.Errorf("user validation failed: %v", err)
	}

	// Load previous state
	s.progress("Loading previous state...")
	previousState, err := s.loadPreviousState(stateFilePath, username)
	if err != nil {
		return nil, fmt.Errorf("failed to load previous state: %v", err)
	}

	// Fetch current starred repositories
	s.progress("Fetching starred repositories...")
	currentRepos, rateLimit, err := s.fetchAllStarredRepos(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %v", err)
	}

	// Compare with previous state
	s.progress("Comparing with previous state...")
	newRepos := s.findNewRepositories(previousState.Repositories, currentRepos)

	// Update state
	s.progress("Updating state...")
	updatedState := &storage.UserState{
		Username:     username,
		LastCheck:    time.Now(),
		Repositories: currentRepos,
		TotalCount:   len(currentRepos),
		StateVersion: "1.0.0",
		CheckCount:   previousState.CheckCount + 1,
	}

	if err := s.storage.SaveUserState(stateFilePath, updatedState); err != nil {
		return nil, fmt.Errorf("failed to save state: %v", err)
	}

	s.progress("Monitor complete")

	return &MonitorResult{
		Username:          username,
		NewRepositories:   newRepos,
		TotalRepositories: len(currentRepos),
		PreviousCheck:     previousState.LastCheck,
		CurrentCheck:      updatedState.LastCheck,
		RateLimit:         *rateLimit,
		IsFirstRun:        previousState.CheckCount == 0,
	}, nil
}

// loadPreviousState loads previous state or creates new state for first run
func (s *Service) loadPreviousState(stateFilePath, username string) (*storage.UserState, error) {
	state, err := s.storage.LoadUserState(stateFilePath)
	if err != nil {
		// Handle file not found - first run
		if _, ok := err.(*storage.StateFileNotFoundError); ok {
			return storage.NewUserState(username), nil
		}
		// Handle corruption - rebuild state
		if _, ok := err.(*storage.StateCorruptionError); ok {
			log.Printf("Warning: State file corrupted, rebuilding from current state")
			return storage.NewUserState(username), nil
		}
		return nil, err
	}
	return state, nil
}

// fetchAllStarredRepos fetches all starred repositories with pagination
func (s *Service) fetchAllStarredRepos(ctx context.Context, username string) ([]storage.Repository, *github.RateLimitInfo, error) {
	var allRepos []storage.Repository
	var rateLimit *github.RateLimitInfo

	opts := &github.StarredOptions{
		PerPage:   100, // Maximum per page
		Sort:      "created",
		Direction: "desc", // Most recent first
	}

	for {
		response, err := s.githubClient.GetStarredRepositories(ctx, username, opts)
		if err != nil {
			return nil, nil, err
		}

		allRepos = append(allRepos, response.Repositories...)
		rateLimit = &response.RateLimit

		// Check if there are more pages
		if !response.PageInfo.HasNext {
			break
		}

		// Update cursor for next page
		opts.Cursor = response.PageInfo.NextCursor

		// Progress update for pagination
		s.progress(fmt.Sprintf("Fetched %d repositories...", len(allRepos)))
	}

	return allRepos, rateLimit, nil
}

// findNewRepositories compares current repositories with previous to find new ones
func (s *Service) findNewRepositories(previous, current []storage.Repository) []storage.Repository {
	// Create a map of previous repositories for fast lookup
	previousMap := make(map[string]bool)
	for _, repo := range previous {
		previousMap[repo.FullName] = true
	}

	// Find repositories in current but not in previous
	var newRepos []storage.Repository
	for _, repo := range current {
		if !previousMap[repo.FullName] {
			newRepos = append(newRepos, repo)
		}
	}

	return newRepos
}

// progress calls the progress callback if set
func (s *Service) progress(message string) {
	if s.progressFunc != nil {
		s.progressFunc(message)
	}
}

// MonitorResult contains the results of a monitoring operation
type MonitorResult struct {
	Username          string               `json:"username"`
	NewRepositories   []storage.Repository `json:"new_repositories"`
	TotalRepositories int                  `json:"total_repositories"`
	PreviousCheck     time.Time            `json:"previous_check"`
	CurrentCheck      time.Time            `json:"current_check"`
	RateLimit         github.RateLimitInfo `json:"rate_limit"`
	IsFirstRun        bool                 `json:"is_first_run"`
}
