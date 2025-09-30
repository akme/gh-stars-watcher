package monitor

import (
	"testing"
	"time"

	"github.com/akme/gh-stars-watcher/internal/config"
	"github.com/akme/gh-stars-watcher/internal/storage"
)

func TestService_findRepositoryChanges_ReStarDetection(t *testing.T) {
	cfg := config.DefaultConfig()
	service := NewService(nil, nil, nil, cfg)

	baseTime := time.Date(2025, 9, 29, 21, 12, 41, 0, time.UTC)

	tests := []struct {
		name         string
		previous     []storage.Repository
		current      []storage.Repository
		wantNewStars []string
		wantReStars  []string
	}{
		{
			name: "DetectReStarAsNewStar",
			previous: []storage.Repository{
				{
					FullName:    "test/repo1",
					StarredAt:   baseTime,
					Description: "Test repository 1",
					StarCount:   100,
				},
			},
			current: []storage.Repository{
				{
					FullName:    "test/repo1",
					StarredAt:   baseTime.Add(4 * time.Hour), // Much newer - re-starred
					Description: "Test repository 1",
					StarCount:   100,
				},
			},
			wantNewStars: []string{"test/repo1"},
			wantReStars:  []string{},
		},
		{
			name: "DetectMinorUpdateAsReStar",
			previous: []storage.Repository{
				{
					FullName:    "test/repo2",
					StarredAt:   baseTime,
					Description: "Test repository 2",
					StarCount:   200,
				},
			},
			current: []storage.Repository{
				{
					FullName:    "test/repo2",
					StarredAt:   baseTime.Add(5 * time.Minute), // Small diff - minor update
					Description: "Test repository 2",
					StarCount:   200,
				},
			},
			wantNewStars: []string{},
			wantReStars:  []string{"test/repo2"},
		},
		{
			name: "DetectTrulyNewRepo",
			previous: []storage.Repository{
				{
					FullName:    "test/existing",
					StarredAt:   baseTime,
					Description: "Existing repo",
					StarCount:   50,
				},
			},
			current: []storage.Repository{
				{
					FullName:    "test/existing",
					StarredAt:   baseTime,
					Description: "Existing repo",
					StarCount:   50,
				},
				{
					FullName:    "test/new-repo",
					StarredAt:   baseTime.Add(time.Hour),
					Description: "New repository",
					StarCount:   10,
				},
			},
			wantNewStars: []string{"test/new-repo"},
			wantReStars:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := service.findRepositoryChanges(tt.previous, tt.current)

			// Check new stars
			gotNewStars := make([]string, len(changes.NewStars))
			for i, repo := range changes.NewStars {
				gotNewStars[i] = repo.FullName
			}
			if !slicesEqual(gotNewStars, tt.wantNewStars) {
				t.Errorf("NewStars = %v, want %v", gotNewStars, tt.wantNewStars)
			}

			// Check re-stars
			gotReStars := make([]string, len(changes.ReStars))
			for i, repo := range changes.ReStars {
				gotReStars[i] = repo.FullName
			}
			if !slicesEqual(gotReStars, tt.wantReStars) {
				t.Errorf("ReStars = %v, want %v", gotReStars, tt.wantReStars)
			}
		})
	}
}

// slicesEqual compares two string slices for equality
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
