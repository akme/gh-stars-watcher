package monitor

import (
	"fmt"
	"testing"
	"time"

	"github.com/akme/gh-stars-watcher/internal/config"
	"github.com/akme/gh-stars-watcher/internal/storage"
)

func TestReStarDetection(t *testing.T) {
	// Create a service with default config
	cfg := config.DefaultConfig()
	service := NewService(nil, nil, nil, cfg)

	// Create test repositories - simulate the scenario
	oldTime := time.Date(2025, 9, 29, 21, 12, 41, 0, time.UTC)
	newTime := time.Date(2025, 9, 30, 1, 30, 0, 0, time.UTC) // 4+ hours later (should be detected as new star)

	// Previous state - repo exists with old timestamp
	previousRepos := []storage.Repository{
		{
			FullName:    "test/repo1",
			StarredAt:   oldTime,
			Description: "Test repository 1",
			StarCount:   100,
		},
		{
			FullName:    "test/repo2",
			StarredAt:   oldTime.Add(-time.Hour), // Even older
			Description: "Test repository 2",
			StarCount:   200,
		},
	}

	// Current state - same repo but with newer timestamp (simulating re-star)
	currentRepos := []storage.Repository{
		{
			FullName:    "test/repo1",
			StarredAt:   newTime, // Much newer timestamp - should be detected as new star
			Description: "Test repository 1",
			StarCount:   100,
		},
		{
			FullName:    "test/repo2",
			StarredAt:   oldTime.Add(-time.Hour), // Same as before
			Description: "Test repository 2",
			StarCount:   200,
		},
		{
			FullName:    "test/repo3", // Truly new repo
			StarredAt:   newTime.Add(time.Minute),
			Description: "Test repository 3",
			StarCount:   50,
		},
	}

	// Test the change detection
	changes := service.findRepositoryChanges(previousRepos, currentRepos)

	fmt.Printf("Change detection results:\n")
	fmt.Printf("New stars: %d\n", len(changes.NewStars))
	for _, repo := range changes.NewStars {
		fmt.Printf("  - %s (starred at: %v)\n", repo.FullName, repo.StarredAt)
	}
	fmt.Printf("Re-stars: %d\n", len(changes.ReStars))
	for _, repo := range changes.ReStars {
		fmt.Printf("  - %s (starred at: %v)\n", repo.FullName, repo.StarredAt)
	}
	fmt.Printf("Updated: %d\n", len(changes.Updated))
	fmt.Printf("Unstars: %d\n", len(changes.Unstars))
	fmt.Printf("Total changes: %d\n", changes.TotalChanges)

	// Assertions
	if len(changes.NewStars) != 2 {
		t.Errorf("Expected 2 new stars (test/repo1 as re-star + test/repo3 as truly new), got %d", len(changes.NewStars))
	}

	// Check that test/repo1 was detected as a new star due to significant timestamp difference
	foundReStarAsNewStar := false
	foundNewRepo := false
	for _, repo := range changes.NewStars {
		if repo.FullName == "test/repo1" {
			foundReStarAsNewStar = true
		}
		if repo.FullName == "test/repo3" {
			foundNewRepo = true
		}
	}

	if !foundReStarAsNewStar {
		t.Error("Expected test/repo1 to be detected as new star due to significant timestamp difference")
	}
	if !foundNewRepo {
		t.Error("Expected test/repo3 to be detected as truly new star")
	}
}
