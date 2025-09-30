package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/akme/gh-stars-watcher/internal/auth"
	"github.com/akme/gh-stars-watcher/internal/config"
	"github.com/akme/gh-stars-watcher/internal/github"
	"github.com/akme/gh-stars-watcher/internal/storage"
)

// Configuration constants for re-star detection
const (
	// reStarThreshold is the minimum time difference to consider a repository
	// as re-starred rather than just updated metadata
	reStarThreshold = 10 * time.Minute
)

// Service provides the core monitoring functionality
type Service struct {
	githubClient github.GitHubClient
	storage      storage.StateStorage
	tokenManager auth.TokenManager
	progressFunc func(message string) // Optional progress callback
	config       *config.Config       // Configuration for incremental fetching
	retryManager *RetryManager        // Retry logic manager
	logger       *slog.Logger         // Structured logger
}

// NewService creates a new monitoring service
func NewService(githubClient github.GitHubClient, storage storage.StateStorage, tokenManager auth.TokenManager, cfg *config.Config) *Service {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Create structured logger based on configuration first
	logger := createLogger(cfg)

	// Validate configuration on creation
	if err := cfg.Validate(); err != nil {
		logger.Warn("Invalid configuration, using defaults", "error", err)
		cfg = config.DefaultConfig()
		// Recreate logger with validated config
		logger = createLogger(cfg)
	}

	retryManager := NewRetryManager(&cfg.Retry)

	return &Service{
		githubClient: githubClient,
		storage:      storage,
		tokenManager: tokenManager,
		config:       cfg,
		retryManager: retryManager,
		logger:       logger,
	}
}

// createLogger creates a structured logger based on configuration
func createLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	switch cfg.Logging.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Create appropriate handler based on format
	var handler slog.Handler
	writer := os.Stderr // Default to stderr
	if cfg.Logging.LogFormat == "json" {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	return slog.New(handler)
}

// SetProgressCallback sets a callback function for progress updates
func (s *Service) SetProgressCallback(callback func(message string)) {
	s.progressFunc = callback
	// Also configure retry manager to use the same progress function for logging
	if s.retryManager != nil {
		s.retryManager.SetLogger(func(format string, args ...interface{}) {
			msg := fmt.Sprintf(format, args...)
			if s.progressFunc != nil {
				s.progressFunc(msg)
			} else {
				s.logger.Info("Retry manager", "message", msg)
			}
		})
	}
}

// progress calls the progress callback if it's set
func (s *Service) progress(message string) {
	if s.progressFunc != nil {
		s.progressFunc(message)
	}
}

// logInfo logs an informational message with structured context
func (s *Service) logInfo(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Info(msg, args...)
	}
}

// logDebug logs a debug message with structured context
func (s *Service) logDebug(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Debug(msg, args...)
	}
}

// logError logs an error message with structured context
func (s *Service) logError(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Error(msg, args...)
	}
}

// logPerformanceMetrics logs performance metrics if enabled
func (s *Service) logPerformanceMetrics(msg string, args ...any) {
	if !s.config.Logging.EnablePerformanceMetrics {
		return
	}
	if s.logger != nil {
		s.logger.Info(msg, append(args, "type", "performance")...)
	}
}

// MonitorUser monitors a GitHub user's starred repositories with enhanced incremental capabilities
func (s *Service) MonitorUser(ctx context.Context, username, stateFilePath string) (*MonitorResult, error) {
	startTime := time.Now()
	s.logPerformanceMetrics("Starting monitor", "username", username)
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
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// Load previous state
	s.progress("Loading previous state...")
	previousState, err := s.loadPreviousState(stateFilePath, username)
	if err != nil {
		return nil, fmt.Errorf("failed to load previous state: %w", err)
	}

	// Fetch current starred repositories using incremental approach
	s.progress("Fetching starred repositories...")
	currentRepos, rateLimit, apiCallsSaved, isFullSync, err := s.fetchStarredReposWithFallback(ctx, username, previousState)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}

	// Compare with previous state and detect all types of changes
	s.progress("Analyzing repository changes...")
	changes := s.findRepositoryChanges(previousState.Repositories, currentRepos)

	// Update state with incremental fetch information
	s.progress("Updating state...")
	updatedState := &storage.UserState{
		Username:     username,
		LastCheck:    time.Now(),
		Repositories: currentRepos,
		TotalCount:   len(currentRepos),
		StateVersion: "1.0.0",
		CheckCount:   previousState.CheckCount + 1,

		// Copy incremental fetch settings from previous state
		LastStarredAt:      previousState.LastStarredAt,
		LastFullSyncAt:     previousState.LastFullSyncAt,
		IncrementalEnabled: previousState.IncrementalEnabled,
		FullSyncInterval:   previousState.FullSyncInterval,
		LastIncrementalAt:  previousState.LastIncrementalAt,
		APICallsSaved:      previousState.APICallsSaved,
	}

	// Update timestamps based on fetch type and results
	if isFullSync {
		updatedState.UpdateFullSyncTimestamp(len(currentRepos), "scheduled_full_sync")
		s.progress("Full sync completed")
	} else {
		updatedState.UpdateIncrementalTimestamp(len(changes.NewStars), apiCallsSaved, "incremental_fetch")
		s.progress("Incremental fetch completed")
	}

	// Update the most recent starred_at timestamp
	if len(currentRepos) > 0 {
		mostRecent := updatedState.GetMostRecentStarredAt()
		if mostRecent.After(updatedState.LastStarredAt) {
			s.logDebug("Updating last starred timestamp",
				"from", updatedState.LastStarredAt,
				"to", mostRecent,
				"new_stars", len(changes.NewStars),
				"api_calls_saved", apiCallsSaved)
			updatedState.UpdateLastStarredAt(mostRecent, len(changes.NewStars), apiCallsSaved, "repository_update")
		}
	}

	if err := s.storage.SaveUserState(stateFilePath, updatedState); err != nil {
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	s.progress("Monitor complete")

	// Log performance metrics
	duration := time.Since(startTime)
	s.logPerformanceMetrics("Monitor completed", "username", username, "duration", duration)
	if s.config.Logging.LogAPICallsSaved && apiCallsSaved > 0 {
		s.logDebug("API calls saved through incremental fetching", "api_calls_saved", apiCallsSaved)
	}

	var rateLimitInfo github.RateLimitInfo
	if rateLimit != nil {
		rateLimitInfo = *rateLimit
	}

	return &MonitorResult{
		Username:           username,
		Changes:            changes,
		TotalRepositories:  len(currentRepos),
		PreviousCheck:      previousState.LastCheck,
		CurrentCheck:       updatedState.LastCheck,
		RateLimit:          rateLimitInfo,
		IsFirstRun:         previousState.CheckCount == 0,
		IsFullSync:         isFullSync,
		APICallsSaved:      apiCallsSaved,
		IncrementalEnabled: updatedState.IncrementalEnabled,
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
			s.logger.Warn("State file corrupted, rebuilding from current state")
			return storage.NewUserState(username), nil
		}
		return nil, err
	}

	// Migrate old state files that don't have incremental fields set
	s.migrateStateToIncrementalFields(state)

	return state, nil
}

// migrateStateToIncrementalFields ensures old state files have proper incremental defaults
func (s *Service) migrateStateToIncrementalFields(state *storage.UserState) {
	migrated := false

	// If IncrementalEnabled is false and FullSyncInterval is 0, this is likely an old state file
	if !state.IncrementalEnabled && state.FullSyncInterval == 0 {
		state.IncrementalEnabled = true
		state.FullSyncInterval = 24 // Default 24 hours
		migrated = true
		s.logInfo("Migrated state file to enable incremental fetching", "interval", "24h")
	}

	if migrated {
		s.logInfo("State file migration completed", "username", state.Username)
	}
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
		var response *github.StarredResponse
		err := s.retryManager.ExecuteWithRetry(ctx, func() error {
			var err error
			response, err = s.githubClient.GetStarredRepositories(ctx, username, opts)
			if err != nil {
				// Check if this is a rate limit error
				if isRateLimitError(err) {
					retryAfter := extractRetryAfter(err)
					return WrapRetryableError(err, true, retryAfter)
				}
				// For other errors, let retry manager decide if retryable
				return err
			}
			return nil
		})
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

// fetchStarredReposIncremental fetches starred repositories incrementally using previous state
func (s *Service) fetchStarredReposIncremental(ctx context.Context, username string, previousState *storage.UserState) ([]storage.Repository, *github.RateLimitInfo, int, error) {
	s.progress("Starting incremental fetch...")
	s.logDebug("Incremental fetch starting", "username", username, "from_timestamp", previousState.LastStarredAt)

	var allRepos []storage.Repository
	var rateLimit *github.RateLimitInfo
	var apiCallsSaved int = 0

	// Use sort=created, direction=desc to get most recently starred repos first
	opts := &github.StarredOptions{
		PerPage:   100,       // Maximum per page
		Sort:      "created", // Sort by starred_at timestamp
		Direction: "desc",    // Most recent first
	}

	// Track the most recent starred_at we've seen
	mostRecentStarredAt := previousState.LastStarredAt
	foundNewRepos := false
	pagesProcessed := 0

	for {
		// Check max pages limit from configuration
		if pagesProcessed >= s.config.Incremental.MaxIncrementalPages {
			s.progress(fmt.Sprintf("Reached maximum incremental pages limit (%d), stopping", s.config.Incremental.MaxIncrementalPages))
			break
		}
		var response *github.StarredResponse
		err := s.retryManager.ExecuteWithRetry(ctx, func() error {
			s.logDebug("Fetching starred repositories", "username", username, "page", pagesProcessed+1, "type", "incremental")
			var err error
			response, err = s.githubClient.GetStarredRepositories(ctx, username, opts)
			if err != nil {
				s.logError("GitHub API call failed during incremental fetch", "username", username, "error", err)
				// Check if this is a rate limit error
				if isRateLimitError(err) {
					retryAfter := extractRetryAfter(err)
					s.logInfo("Rate limit hit during incremental fetch", "username", username, "retry_after", retryAfter)
					return WrapRetryableError(err, true, retryAfter)
				}
				// For other errors, let retry manager decide if retryable
				return err
			}
			s.logDebug("Successfully fetched repositories from GitHub API", "count", len(response.Repositories), "type", "incremental")
			return nil
		})
		if err != nil {
			return nil, nil, 0, err
		}

		rateLimit = &response.RateLimit
		pagesProcessed++

		// Process repositories in this page
		var newReposInPage []storage.Repository

		for _, repo := range response.Repositories {
			// Check if we've reached repositories we've seen before, with timestamp tolerance
			timeDiff := repo.StarredAt.Sub(mostRecentStarredAt)
			if !mostRecentStarredAt.IsZero() && timeDiff <= s.config.Incremental.TimestampTolerance {
				s.logDebug("Reached previously seen timestamp (within tolerance)", "timestamp", repo.StarredAt, "tolerance", s.config.Incremental.TimestampTolerance)
				break
			}

			newReposInPage = append(newReposInPage, repo)
			foundNewRepos = true

			// Update the most recent timestamp
			if repo.StarredAt.After(mostRecentStarredAt) {
				mostRecentStarredAt = repo.StarredAt
			}
		}

		allRepos = append(allRepos, newReposInPage...)

		// If we didn't find any new repos in this page, we can stop
		if len(newReposInPage) == 0 {
			s.logDebug("No new repositories found in this page, stopping incremental fetch")
			break
		}

		// If the page wasn't full, we've reached the end
		if len(response.Repositories) < opts.PerPage {
			s.progress("Reached end of starred repositories")
			break
		}

		// Check if there are more pages
		if !response.PageInfo.HasNext {
			break
		}

		// Update cursor for next page
		opts.Cursor = response.PageInfo.NextCursor

		// Estimate API calls saved (very rough estimate)
		if foundNewRepos && len(allRepos) < previousState.TotalCount {
			// We stopped early, so we saved some API calls
			remainingRepos := previousState.TotalCount - len(allRepos)
			estimatedPagesSkipped := remainingRepos / opts.PerPage
			apiCallsSaved += estimatedPagesSkipped
		}

		// Progress update for pagination (debug level for detailed info)
		s.logDebug("Incremental fetch progress", "new_repositories_found", len(allRepos))
	}

	s.progress(fmt.Sprintf("Incremental fetch complete: %d new repositories, estimated %d API calls saved", len(allRepos), apiCallsSaved))
	return allRepos, rateLimit, apiCallsSaved, nil
}

// fetchStarredReposWithFallback attempts incremental fetch first, falls back to full fetch if needed
func (s *Service) fetchStarredReposWithFallback(ctx context.Context, username string, previousState *storage.UserState) ([]storage.Repository, *github.RateLimitInfo, int, bool, error) {
	isFullSync := false
	apiCallsSaved := 0

	// Determine fetch strategy based on configuration
	if s.config.Incremental.Enabled && previousState.ShouldUseIncremental() && !previousState.ShouldPerformFullSync() {
		s.progress("Attempting incremental fetch...")
		s.logInfo("Using incremental fetch", "username", username)

		// Try incremental fetch
		newRepos, rateLimit, saved, err := s.fetchStarredReposIncremental(ctx, username, previousState)
		if err != nil {
			if s.config.Incremental.FallbackOnError {
				s.progress(fmt.Sprintf("Incremental fetch failed: %v, falling back to full sync", err))
			} else {
				s.progress(fmt.Sprintf("Incremental fetch failed: %v, fallback disabled", err))
				return nil, nil, 0, false, fmt.Errorf("incremental fetch failed and fallback disabled: %w", err)
			}
		} else {
			// Merge new repos with existing repos for change detection
			mergedRepos := s.mergeRepositories(previousState.Repositories, newRepos)
			return mergedRepos, rateLimit, saved, isFullSync, nil
		}
	}

	// Fallback to full sync
	s.progress("Performing full sync...")
	s.logInfo("Using full sync", "username", username)
	isFullSync = true
	allRepos, rateLimit, err := s.fetchAllStarredRepos(ctx, username)
	return allRepos, rateLimit, apiCallsSaved, isFullSync, err
}

// mergeRepositories merges new repositories with existing ones, handling duplicates
func (s *Service) mergeRepositories(existing []storage.Repository, newRepos []storage.Repository) []storage.Repository {
	repoMap := make(map[string]storage.Repository, len(existing)+len(newRepos))

	for _, repo := range existing {
		repoMap[repo.FullName] = repo
	}

	for _, repo := range newRepos {
		repoMap[repo.FullName] = repo
	}

	merged := make([]storage.Repository, 0, len(repoMap))
	for _, repo := range repoMap {
		merged = append(merged, repo)
	}

	sort.Slice(merged, func(i, j int) bool {
		if merged[i].StarredAt.Equal(merged[j].StarredAt) {
			return merged[i].FullName < merged[j].FullName
		}
		return merged[i].StarredAt.After(merged[j].StarredAt)
	})

	return merged
}

// RepositoryChanges represents the changes between two repository states
type RepositoryChanges struct {
	NewStars     []storage.Repository `json:"new_stars"`     // Newly starred repositories
	Unstars      []storage.Repository `json:"unstars"`       // Unstarred repositories
	ReStars      []storage.Repository `json:"re_stars"`      // Re-starred repositories (starred, unstarred, then starred again)
	Updated      []storage.Repository `json:"updated"`       // Repositories with updated metadata
	TotalChanges int                  `json:"total_changes"` // Total number of changes detected
}

// findRepositoryChanges compares current repositories with previous to find all types of changes
func (s *Service) findRepositoryChanges(previous, current []storage.Repository) *RepositoryChanges {
	changes := &RepositoryChanges{
		NewStars: make([]storage.Repository, 0),
		Unstars:  make([]storage.Repository, 0),
		ReStars:  make([]storage.Repository, 0),
		Updated:  make([]storage.Repository, 0),
	}

	// Create maps for efficient lookup
	previousMap := make(map[string]storage.Repository)
	currentMap := make(map[string]storage.Repository)

	for _, repo := range previous {
		previousMap[repo.FullName] = repo
	}

	for _, repo := range current {
		currentMap[repo.FullName] = repo
	}

	// Find new stars (in current but not in previous)
	for _, currentRepo := range current {
		if _, exists := previousMap[currentRepo.FullName]; !exists {
			changes.NewStars = append(changes.NewStars, currentRepo)
		} else {
			// Check for updates (same repo but different metadata)
			prevRepo := previousMap[currentRepo.FullName]
			if s.hasRepositoryChanged(prevRepo, currentRepo) {
				changes.Updated = append(changes.Updated, currentRepo)
			}

			// Check for re-stars (starred_at timestamp changed) - only if enabled in config
			if s.config.Incremental.DetectReStars {
				if !currentRepo.StarredAt.Equal(prevRepo.StarredAt) && currentRepo.StarredAt.After(prevRepo.StarredAt) {
					// Calculate time difference to detect significant re-starring
					timeDiff := currentRepo.StarredAt.Sub(prevRepo.StarredAt)

					// If the time difference is substantial, treat as new star
					// This indicates the repo was likely unstarred and re-starred
					if timeDiff > reStarThreshold {
						s.logDebug("Detected re-starred repository as new star", "repository", currentRepo.FullName, "time_diff", timeDiff)
						changes.NewStars = append(changes.NewStars, currentRepo)
					} else {
						// Small time difference, just track as re-star
						changes.ReStars = append(changes.ReStars, currentRepo)
					}
				}
			}
		}
	}

	// Find unstars (in previous but not in current) - only if enabled in config
	if s.config.Incremental.DetectUnstars {
		for _, prevRepo := range previous {
			if _, exists := currentMap[prevRepo.FullName]; !exists {
				changes.Unstars = append(changes.Unstars, prevRepo)
			}
		}
	}

	changes.TotalChanges = len(changes.NewStars) + len(changes.Unstars) + len(changes.ReStars) + len(changes.Updated)
	return changes
}

// isRateLimitError checks if an error is related to rate limiting

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	rateLimitPatterns := []string{
		"rate limit",
		"api rate limit exceeded",
		"403 forbidden",
		"secondary rate limit",
		"abuse detection",
	}

	for _, pattern := range rateLimitPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// extractRetryAfter attempts to extract retry-after duration from error message
func extractRetryAfter(err error) time.Duration {
	// For now, return a default duration
	// In a real implementation, you'd parse the error message or response headers
	return 60 * time.Second // Default 1 minute wait for rate limits
}

// hasRepositoryChanged checks if repository metadata has changed
func (s *Service) hasRepositoryChanged(prev, current storage.Repository) bool {
	// Check significant metadata changes
	return prev.Description != current.Description ||
		prev.StarCount != current.StarCount ||
		prev.Language != current.Language ||
		prev.Private != current.Private ||
		!prev.UpdatedAt.Equal(current.UpdatedAt)
}

// MonitorResult contains comprehensive results including incremental fetch information
type MonitorResult struct {
	PreviousCheck      time.Time            `json:"previous_check"`
	CurrentCheck       time.Time            `json:"current_check"`
	RateLimit          github.RateLimitInfo `json:"rate_limit"`
	Username           string               `json:"username"`
	Changes            *RepositoryChanges   `json:"changes"` // Detailed change analysis
	TotalRepositories  int                  `json:"total_repositories"`
	APICallsSaved      int                  `json:"api_calls_saved"` // Estimated API calls saved by incremental fetch
	IsFirstRun         bool                 `json:"is_first_run"`
	IsFullSync         bool                 `json:"is_full_sync"`        // Whether a full sync was performed
	IncrementalEnabled bool                 `json:"incremental_enabled"` // Whether incremental fetching is enabled
}
