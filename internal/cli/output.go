package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/akme/gh-stars-watcher/internal/monitor"
	"github.com/akme/gh-stars-watcher/internal/storage"
)

// OutputFormatter handles formatting of monitoring results
type OutputFormatter struct {
	writer io.Writer
	format string // "json", "text", "summary"
}

// NewOutputFormatter creates a new output formatter
func NewOutputFormatter(writer io.Writer, format string) *OutputFormatter {
	if writer == nil {
		writer = os.Stdout
	}
	if format == "" {
		format = "text"
	}
	return &OutputFormatter{
		writer: writer,
		format: format,
	}
}

// FormatMonitorResults formats monitoring results according to the configured format
func (f *OutputFormatter) FormatMonitorResults(result *monitor.ComparisonResult, username string) error {
	switch f.format {
	case "json":
		return f.formatJSON(result, username)
	case "summary":
		return f.formatSummary(result, username)
	default:
		return f.formatText(result, username)
	}
}

// FormatMonitorResult formats monitoring result from the service
func (f *OutputFormatter) FormatMonitorResult(result *monitor.MonitorResult) error {
	if f.format == "json" {
		encoder := json.NewEncoder(f.writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	// Text format
	if result.IsFirstRun {
		fmt.Fprintf(f.writer, "First run for %s - baseline established with %d starred repositories.\n",
			result.Username, result.TotalRepositories)
		fmt.Fprintf(f.writer, "Run again to detect newly starred repositories.\n")
		return nil
	}

	if len(result.NewRepositories) == 0 {
		fmt.Fprintf(f.writer, "No new starred repositories found for %s.\n", result.Username)
		fmt.Fprintf(f.writer, "Total repositories: %d\n", result.TotalRepositories)
		return nil
	}

	fmt.Fprintf(f.writer, "🌟 %s has starred %d new repositories!\n\n",
		result.Username, len(result.NewRepositories))

	// Sort by starred date (most recent first)
	sorted := make([]storage.Repository, len(result.NewRepositories))
	copy(sorted, result.NewRepositories)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StarredAt.After(sorted[j].StarredAt)
	})

	for _, repo := range sorted {
		f.formatRepository(repo, "added")
	}

	fmt.Fprintf(f.writer, "Total repositories: %d\n", result.TotalRepositories)
	if !result.PreviousCheck.IsZero() {
		fmt.Fprintf(f.writer, "Previous check: %s\n", result.PreviousCheck.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// formatJSON outputs results in JSON format
func (f *OutputFormatter) formatJSON(result *monitor.ComparisonResult, username string) error {
	output := struct {
		Username  string                    `json:"username"`
		Timestamp time.Time                 `json:"timestamp"`
		Summary   string                    `json:"summary"`
		Results   *monitor.ComparisonResult `json:"results"`
	}{
		Username:  username,
		Timestamp: time.Now(),
		Summary:   result.Summary(),
		Results:   result,
	}

	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatText outputs results in human-readable text format
func (f *OutputFormatter) formatText(result *monitor.ComparisonResult, username string) error {
	fmt.Fprintf(f.writer, "GitHub Stars Monitor Report for %s\n", username)
	fmt.Fprintf(f.writer, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	if len(result.Added) == 0 && len(result.Removed) == 0 && len(result.Updated) == 0 {
		fmt.Fprintf(f.writer, "No changes detected in starred repositories.\n")
		return nil
	}

	// New repositories
	if len(result.Added) > 0 {
		fmt.Fprintf(f.writer, "🌟 NEWLY STARRED REPOSITORIES (%d)\n", len(result.Added))
		fmt.Fprintf(f.writer, "%s\n", strings.Repeat("=", 50))
		for _, repo := range result.Added {
			f.formatRepository(repo, "added")
		}
		fmt.Fprintf(f.writer, "\n")
	}

	// Removed repositories
	if len(result.Removed) > 0 {
		fmt.Fprintf(f.writer, "💔 UNSTARRED REPOSITORIES (%d)\n", len(result.Removed))
		fmt.Fprintf(f.writer, "%s\n", strings.Repeat("=", 50))
		for _, repo := range result.Removed {
			f.formatRepository(repo, "removed")
		}
		fmt.Fprintf(f.writer, "\n")
	}

	// Updated repositories
	if len(result.Updated) > 0 {
		fmt.Fprintf(f.writer, "🔄 UPDATED REPOSITORIES (%d)\n", len(result.Updated))
		fmt.Fprintf(f.writer, "%s\n", strings.Repeat("=", 50))
		for _, update := range result.Updated {
			f.formatRepositoryUpdate(update)
		}
		fmt.Fprintf(f.writer, "\n")
	}

	fmt.Fprintf(f.writer, "Summary: %s\n", result.Summary())
	return nil
}

// formatSummary outputs a brief summary
func (f *OutputFormatter) formatSummary(result *monitor.ComparisonResult, username string) error {
	fmt.Fprintf(f.writer, "%s: %s\n", username, result.Summary())

	if len(result.Added) > 0 {
		fmt.Fprintf(f.writer, "New stars: ")
		names := make([]string, len(result.Added))
		for i, repo := range result.Added {
			names[i] = repo.FullName
		}
		fmt.Fprintf(f.writer, "%s\n", strings.Join(names, ", "))
	}

	return nil
}

// formatRepository formats a single repository
func (f *OutputFormatter) formatRepository(repo storage.Repository, action string) {
	icon := "⭐"
	if action == "removed" {
		icon = "💔"
	}

	fmt.Fprintf(f.writer, "%s %s\n", icon, repo.FullName)

	if repo.Description != "" {
		fmt.Fprintf(f.writer, "   %s\n", repo.Description)
	}

	fmt.Fprintf(f.writer, "   Language: %s | Stars: %d",
		f.formatLanguage(repo.Language), repo.StarCount)

	if action == "added" && !repo.StarredAt.IsZero() {
		fmt.Fprintf(f.writer, " | Starred: %s", repo.StarredAt.Format("2006-01-02"))
	}

	fmt.Fprintf(f.writer, "\n   %s\n\n", repo.URL)
}

// formatRepositoryUpdate formats a repository update
func (f *OutputFormatter) formatRepositoryUpdate(update monitor.RepositoryUpdate) {
	fmt.Fprintf(f.writer, "🔄 %s\n", update.Current.FullName)
	fmt.Fprintf(f.writer, "   Changes: %s\n", strings.Join(update.Changes, ", "))

	for _, change := range update.Changes {
		switch change {
		case "star_count":
			fmt.Fprintf(f.writer, "   Stars: %d → %d\n",
				update.Previous.StarCount, update.Current.StarCount)
		case "description":
			fmt.Fprintf(f.writer, "   Description updated\n")
		case "language":
			fmt.Fprintf(f.writer, "   Language: %s → %s\n",
				f.formatLanguage(update.Previous.Language),
				f.formatLanguage(update.Current.Language))
		case "updated_at":
			fmt.Fprintf(f.writer, "   Last updated: %s\n",
				update.Current.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Fprintf(f.writer, "   %s\n\n", update.Current.URL)
}

// formatLanguage formats programming language name
func (f *OutputFormatter) formatLanguage(language string) string {
	if language == "" {
		return "None"
	}
	return language
}

// FormatRepositoryList formats a simple list of repositories
func (f *OutputFormatter) FormatRepositoryList(repositories []storage.Repository, title string) error {
	if f.format == "json" {
		output := struct {
			Title        string               `json:"title"`
			Timestamp    time.Time            `json:"timestamp"`
			Count        int                  `json:"count"`
			Repositories []storage.Repository `json:"repositories"`
		}{
			Title:        title,
			Timestamp:    time.Now(),
			Count:        len(repositories),
			Repositories: repositories,
		}

		encoder := json.NewEncoder(f.writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	// Text format
	fmt.Fprintf(f.writer, "%s (%d repositories)\n", title, len(repositories))
	fmt.Fprintf(f.writer, "%s\n\n", strings.Repeat("=", len(title)+20))

	if len(repositories) == 0 {
		fmt.Fprintf(f.writer, "No repositories found.\n")
		return nil
	}

	// Sort by starred date (most recent first)
	sorted := make([]storage.Repository, len(repositories))
	copy(sorted, repositories)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StarredAt.After(sorted[j].StarredAt)
	})

	for _, repo := range sorted {
		f.formatRepository(repo, "listed")
	}

	return nil
}

// FormatError formats an error message
func (f *OutputFormatter) FormatError(err error, context string) error {
	if f.format == "json" {
		output := struct {
			Error     string    `json:"error"`
			Context   string    `json:"context"`
			Timestamp time.Time `json:"timestamp"`
		}{
			Error:     err.Error(),
			Context:   context,
			Timestamp: time.Now(),
		}

		encoder := json.NewEncoder(f.writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	// Text format
	fmt.Fprintf(f.writer, "Error %s: %v\n", context, err)
	return nil
}

// FormatStats formats statistics about repositories
func (f *OutputFormatter) FormatStats(repositories []storage.Repository) error {
	if len(repositories) == 0 {
		fmt.Fprintf(f.writer, "No repositories to analyze.\n")
		return nil
	}

	// Calculate statistics
	stats := f.calculateStats(repositories)

	if f.format == "json" {
		encoder := json.NewEncoder(f.writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(stats)
	}

	// Text format
	fmt.Fprintf(f.writer, "Repository Statistics\n")
	fmt.Fprintf(f.writer, "====================\n\n")
	fmt.Fprintf(f.writer, "Total repositories: %d\n", stats.Total)
	fmt.Fprintf(f.writer, "Total stars: %d\n", stats.TotalStars)
	fmt.Fprintf(f.writer, "Average stars per repository: %.1f\n", stats.AverageStars)
	fmt.Fprintf(f.writer, "Most starred: %s (%d stars)\n",
		stats.MostStarred.FullName, stats.MostStarred.StarCount)

	fmt.Fprintf(f.writer, "\nTop Languages:\n")
	for i, lang := range stats.TopLanguages {
		if i >= 5 { // Show top 5
			break
		}
		fmt.Fprintf(f.writer, "  %d. %s (%d repositories)\n",
			i+1, lang.Language, lang.Count)
	}

	return nil
}

// RepositoryStats contains statistics about repositories
type RepositoryStats struct {
	Total        int                `json:"total"`
	TotalStars   int                `json:"total_stars"`
	AverageStars float64            `json:"average_stars"`
	MostStarred  storage.Repository `json:"most_starred"`
	TopLanguages []LanguageStat     `json:"top_languages"`
}

// LanguageStat contains statistics for a programming language
type LanguageStat struct {
	Language string `json:"language"`
	Count    int    `json:"count"`
}

// calculateStats calculates statistics for a set of repositories
func (f *OutputFormatter) calculateStats(repositories []storage.Repository) RepositoryStats {
	stats := RepositoryStats{
		Total: len(repositories),
	}

	languageMap := make(map[string]int)
	maxStars := 0

	for _, repo := range repositories {
		stats.TotalStars += repo.StarCount

		if repo.StarCount > maxStars {
			maxStars = repo.StarCount
			stats.MostStarred = repo
		}

		language := repo.Language
		if language == "" {
			language = "Unknown"
		}
		languageMap[language]++
	}

	if stats.Total > 0 {
		stats.AverageStars = float64(stats.TotalStars) / float64(stats.Total)
	}

	// Convert language map to sorted slice
	for lang, count := range languageMap {
		stats.TopLanguages = append(stats.TopLanguages, LanguageStat{
			Language: lang,
			Count:    count,
		})
	}

	// Sort by count (descending)
	sort.Slice(stats.TopLanguages, func(i, j int) bool {
		return stats.TopLanguages[i].Count > stats.TopLanguages[j].Count
	})

	return stats
}
