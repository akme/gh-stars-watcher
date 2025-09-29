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

// TimestampUpdate represents an audit record of timestamp changes for debugging and monitoring
type TimestampUpdate struct {
	Timestamp     time.Time `json:"timestamp"`       // When the update occurred
	OldValue      time.Time `json:"old_value"`       // Previous timestamp value
	NewValue      time.Time `json:"new_value"`       // New timestamp value
	UpdateType    string    `json:"update_type"`     // Type of update: "starred_at", "full_sync", "incremental"
	RepoCount     int       `json:"repo_count"`      // Number of repositories processed in this update
	APICallsSaved int       `json:"api_calls_saved"` // Number of API calls saved by this incremental fetch
	Reason        string    `json:"reason"`          // Reason for the update (e.g., "incremental_fetch", "full_sync", "first_run")
}

// Validate checks if the TimestampUpdate has valid field values
func (t *TimestampUpdate) Validate() error {
	// Timestamp must not be in the future
	if t.Timestamp.After(time.Now().Add(1 * time.Minute)) { // Allow 1 minute tolerance for clock skew
		return fmt.Errorf("timestamp update cannot be in the future: %v", t.Timestamp)
	}

	// UpdateType must be valid
	validTypes := map[string]bool{
		"starred_at":  true,
		"full_sync":   true,
		"incremental": true,
		"first_run":   true,
		"fallback":    true,
	}
	if !validTypes[t.UpdateType] {
		return fmt.Errorf("invalid update type: %s", t.UpdateType)
	}

	// RepoCount must be non-negative
	if t.RepoCount < 0 {
		return fmt.Errorf("repo count must be non-negative: %d", t.RepoCount)
	}

	// APICallsSaved must be non-negative
	if t.APICallsSaved < 0 {
		return fmt.Errorf("API calls saved must be non-negative: %d", t.APICallsSaved)
	}

	return nil
}

// UserState represents the persisted state for a GitHub user's monitoring session
type UserState struct {
	Username     string       `json:"username"`      // GitHub username being monitored
	LastCheck    time.Time    `json:"last_check"`    // Timestamp of last successful check
	Repositories []Repository `json:"repositories"`  // Previously seen starred repositories
	TotalCount   int          `json:"total_count"`   // Total repositories at last check (for pagination validation)
	StateVersion string       `json:"state_version"` // Schema version for backward compatibility
	CheckCount   int          `json:"check_count"`   // Number of successful checks performed

	// Incremental fetching fields
	LastStarredAt      time.Time `json:"last_starred_at"`     // Most recent starred_at timestamp from previous fetch
	LastFullSyncAt     time.Time `json:"last_full_sync_at"`   // Timestamp of last complete repository fetch
	IncrementalEnabled bool      `json:"incremental_enabled"` // Whether incremental fetching is enabled
	FullSyncInterval   int       `json:"full_sync_interval"`  // Hours between full syncs (0 = disabled)

	// Audit and monitoring fields
	LastIncrementalAt time.Time         `json:"last_incremental_at"` // Timestamp of last incremental fetch
	APICallsSaved     int               `json:"api_calls_saved"`     // Cumulative API calls saved by incremental fetching
	TimestampUpdates  []TimestampUpdate `json:"timestamp_updates"`   // Audit log of timestamp changes
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

	// Validate incremental fetching fields
	if u.FullSyncInterval < 0 {
		return fmt.Errorf("full sync interval must be non-negative: %d", u.FullSyncInterval)
	}

	if u.APICallsSaved < 0 {
		return fmt.Errorf("API calls saved must be non-negative: %d", u.APICallsSaved)
	}

	// LastStarredAt should not be in the future
	if u.LastStarredAt.After(time.Now().Add(1 * time.Minute)) { // Allow 1 minute tolerance
		return fmt.Errorf("last starred timestamp cannot be in the future: %v", u.LastStarredAt)
	}

	// LastFullSyncAt should not be in the future
	if u.LastFullSyncAt.After(time.Now().Add(1 * time.Minute)) {
		return fmt.Errorf("last full sync timestamp cannot be in the future: %v", u.LastFullSyncAt)
	}

	// LastIncrementalAt should not be in the future
	if u.LastIncrementalAt.After(time.Now().Add(1 * time.Minute)) {
		return fmt.Errorf("last incremental timestamp cannot be in the future: %v", u.LastIncrementalAt)
	}

	// Validate all timestamp updates
	for i, update := range u.TimestampUpdates {
		if err := update.Validate(); err != nil {
			return fmt.Errorf("invalid timestamp update at index %d: %v", i, err)
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

		// Incremental fetching defaults
		LastStarredAt:      time.Time{}, // Zero time for first run
		LastFullSyncAt:     time.Time{}, // Zero time for first run
		IncrementalEnabled: true,        // Enable incremental fetching by default
		FullSyncInterval:   24,          // Full sync every 24 hours by default

		// Audit and monitoring defaults
		LastIncrementalAt: time.Time{},                // Zero time for first run
		APICallsSaved:     0,                          // No calls saved initially
		TimestampUpdates:  make([]TimestampUpdate, 0), // Empty audit log
	}
}

// ShouldUseIncremental determines if incremental fetching should be used
func (u *UserState) ShouldUseIncremental() bool {
	// Must be enabled and have a previous starred timestamp
	return u.IncrementalEnabled && !u.LastStarredAt.IsZero()
}

// ShouldPerformFullSync determines if a full sync is needed
func (u *UserState) ShouldPerformFullSync() bool {
	// Force full sync if never performed or interval exceeded
	if u.LastFullSyncAt.IsZero() || u.FullSyncInterval == 0 {
		return true
	}

	// Check if interval has passed
	nextFullSync := u.LastFullSyncAt.Add(time.Duration(u.FullSyncInterval) * time.Hour)
	return time.Now().After(nextFullSync)
}

// UpdateLastStarredAt updates the last starred timestamp and logs the change
func (u *UserState) UpdateLastStarredAt(newTimestamp time.Time, repoCount int, apiCallsSaved int, reason string) {
	oldValue := u.LastStarredAt
	u.LastStarredAt = newTimestamp

	// Add audit record
	update := TimestampUpdate{
		Timestamp:     time.Now(),
		OldValue:      oldValue,
		NewValue:      newTimestamp,
		UpdateType:    "starred_at",
		RepoCount:     repoCount,
		APICallsSaved: apiCallsSaved,
		Reason:        reason,
	}

	u.TimestampUpdates = append(u.TimestampUpdates, update)
	u.APICallsSaved += apiCallsSaved // Accumulate total savings

	// Keep only last 100 audit records to prevent unbounded growth
	if len(u.TimestampUpdates) > 100 {
		u.TimestampUpdates = u.TimestampUpdates[len(u.TimestampUpdates)-100:]
	}
}

// UpdateFullSyncTimestamp updates the last full sync timestamp
func (u *UserState) UpdateFullSyncTimestamp(repoCount int, reason string) {
	oldValue := u.LastFullSyncAt
	u.LastFullSyncAt = time.Now()

	// Add audit record
	update := TimestampUpdate{
		Timestamp:     time.Now(),
		OldValue:      oldValue,
		NewValue:      u.LastFullSyncAt,
		UpdateType:    "full_sync",
		RepoCount:     repoCount,
		APICallsSaved: 0, // Full sync doesn't save API calls
		Reason:        reason,
	}

	u.TimestampUpdates = append(u.TimestampUpdates, update)

	// Keep only last 100 audit records
	if len(u.TimestampUpdates) > 100 {
		u.TimestampUpdates = u.TimestampUpdates[len(u.TimestampUpdates)-100:]
	}
}

// UpdateIncrementalTimestamp updates the last incremental fetch timestamp
func (u *UserState) UpdateIncrementalTimestamp(repoCount int, apiCallsSaved int, reason string) {
	oldValue := u.LastIncrementalAt
	u.LastIncrementalAt = time.Now()

	// Add audit record
	update := TimestampUpdate{
		Timestamp:     time.Now(),
		OldValue:      oldValue,
		NewValue:      u.LastIncrementalAt,
		UpdateType:    "incremental",
		RepoCount:     repoCount,
		APICallsSaved: apiCallsSaved,
		Reason:        reason,
	}

	u.TimestampUpdates = append(u.TimestampUpdates, update)
	u.APICallsSaved += apiCallsSaved // Accumulate total savings

	// Keep only last 100 audit records
	if len(u.TimestampUpdates) > 100 {
		u.TimestampUpdates = u.TimestampUpdates[len(u.TimestampUpdates)-100:]
	}
}

// GetMostRecentStarredAt finds the most recent starred_at timestamp from current repositories
func (u *UserState) GetMostRecentStarredAt() time.Time {
	var mostRecent time.Time

	for _, repo := range u.Repositories {
		if repo.StarredAt.After(mostRecent) {
			mostRecent = repo.StarredAt
		}
	}

	return mostRecent
}
