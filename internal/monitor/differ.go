package monitor

import (
	"fmt"
	"sort"
	"time"

	"github.com/akme/gh-stars-watcher/internal/storage"
)

// Differ provides repository comparison functionality
type Differ struct{}

// NewDiffer creates a new repository differ
func NewDiffer() *Differ {
	return &Differ{}
}

// CompareRepositories compares two sets of repositories and returns the differences
func (d *Differ) CompareRepositories(previous, current []storage.Repository) *ComparisonResult {
	// Create maps for efficient lookup
	prevMap := make(map[string]storage.Repository)
	currMap := make(map[string]storage.Repository)

	for _, repo := range previous {
		prevMap[repo.FullName] = repo
	}
	for _, repo := range current {
		currMap[repo.FullName] = repo
	}

	var added []storage.Repository
	var removed []storage.Repository
	var updated []RepositoryUpdate

	// Find added repositories (in current but not in previous)
	for _, repo := range current {
		if _, exists := prevMap[repo.FullName]; !exists {
			added = append(added, repo)
		}
	}

	// Find removed repositories (in previous but not in current)
	for _, repo := range previous {
		if _, exists := currMap[repo.FullName]; !exists {
			removed = append(removed, repo)
		}
	}

	// Find updated repositories (in both but with changes)
	for _, currRepo := range current {
		if prevRepo, exists := prevMap[currRepo.FullName]; exists {
			if d.hasRepositoryChanged(prevRepo, currRepo) {
				updated = append(updated, RepositoryUpdate{
					Previous: prevRepo,
					Current:  currRepo,
					Changes:  d.getRepositoryChanges(prevRepo, currRepo),
				})
			}
		}
	}

	// Sort results for consistent output
	sort.Slice(added, func(i, j int) bool {
		return added[i].StarredAt.After(added[j].StarredAt)
	})
	sort.Slice(removed, func(i, j int) bool {
		return removed[i].FullName < removed[j].FullName
	})
	sort.Slice(updated, func(i, j int) bool {
		return updated[i].Current.FullName < updated[j].Current.FullName
	})

	return &ComparisonResult{
		Added:   added,
		Removed: removed,
		Updated: updated,
	}
}

// hasRepositoryChanged checks if a repository has changed between two states
func (d *Differ) hasRepositoryChanged(prev, curr storage.Repository) bool {
	return prev.Description != curr.Description ||
		prev.StarCount != curr.StarCount ||
		!prev.UpdatedAt.Equal(curr.UpdatedAt) ||
		prev.Language != curr.Language ||
		prev.Private != curr.Private
}

// getRepositoryChanges identifies specific changes between two repository states
func (d *Differ) getRepositoryChanges(prev, curr storage.Repository) []string {
	var changes []string

	if prev.Description != curr.Description {
		changes = append(changes, "description")
	}
	if prev.StarCount != curr.StarCount {
		changes = append(changes, "star_count")
	}
	if !prev.UpdatedAt.Equal(curr.UpdatedAt) {
		changes = append(changes, "updated_at")
	}
	if prev.Language != curr.Language {
		changes = append(changes, "language")
	}
	if prev.Private != curr.Private {
		changes = append(changes, "private")
	}

	return changes
}

// FilterNewRepositories filters repositories to show only newly starred ones
func (d *Differ) FilterNewRepositories(repositories []storage.Repository, since time.Time) []storage.Repository {
	var newRepos []storage.Repository

	for _, repo := range repositories {
		if repo.StarredAt.After(since) {
			newRepos = append(newRepos, repo)
		}
	}

	// Sort by starred time (most recent first)
	sort.Slice(newRepos, func(i, j int) bool {
		return newRepos[i].StarredAt.After(newRepos[j].StarredAt)
	})

	return newRepos
}

// ComparisonResult contains the results of comparing two repository sets
type ComparisonResult struct {
	Added   []storage.Repository `json:"added"`
	Removed []storage.Repository `json:"removed"`
	Updated []RepositoryUpdate   `json:"updated"`
}

// RepositoryUpdate represents a repository that has been updated
type RepositoryUpdate struct {
	Previous storage.Repository `json:"previous"`
	Current  storage.Repository `json:"current"`
	Changes  []string           `json:"changes"`
}

// Summary returns a summary of the comparison results
func (r *ComparisonResult) Summary() string {
	return fmt.Sprintf("Added: %d, Removed: %d, Updated: %d",
		len(r.Added), len(r.Removed), len(r.Updated))
}
