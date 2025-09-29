package config

import (
	"fmt"
	"time"
)

// IncrementalConfig contains configuration options for incremental fetching
type IncrementalConfig struct {
	// Enable or disable incremental fetching globally
	Enabled bool `json:"enabled" yaml:"enabled"`

	// FullSyncInterval specifies hours between full synchronizations
	// 0 means never perform full sync after initial fetch
	// 24 means perform full sync every 24 hours
	FullSyncInterval int `json:"full_sync_interval" yaml:"full_sync_interval"`

	// FallbackOnError determines if full sync should be used when incremental fails
	FallbackOnError bool `json:"fallback_on_error" yaml:"fallback_on_error"`

	// MaxIncrementalPages limits the number of pages to fetch incrementally
	// This prevents runaway fetching for users with many new stars
	MaxIncrementalPages int `json:"max_incremental_pages" yaml:"max_incremental_pages"`

	// DetectUnstars enables detection of unstarred repositories
	// Requires periodic full sync to work properly
	DetectUnstars bool `json:"detect_unstars" yaml:"detect_unstars"`

	// DetectReStars enables detection of re-starred repositories
	DetectReStars bool `json:"detect_re_stars" yaml:"detect_re_stars"`

	// TimestampTolerance allows for small timestamp differences to handle clock skew
	TimestampTolerance time.Duration `json:"timestamp_tolerance" yaml:"timestamp_tolerance"`
}

// RetryConfig contains configuration for retry logic and error handling
type RetryConfig struct {
	// MaxRetries specifies the maximum number of retry attempts
	MaxRetries int `json:"max_retries" yaml:"max_retries"`

	// InitialDelay is the initial delay between retries
	InitialDelay time.Duration `json:"initial_delay" yaml:"initial_delay"`

	// MaxDelay is the maximum delay between retries (for exponential backoff)
	MaxDelay time.Duration `json:"max_delay" yaml:"max_delay"`

	// BackoffMultiplier is the multiplier for exponential backoff
	BackoffMultiplier float64 `json:"backoff_multiplier" yaml:"backoff_multiplier"`

	// RetryOnRateLimit enables automatic retry when rate limit is hit
	RetryOnRateLimit bool `json:"retry_on_rate_limit" yaml:"retry_on_rate_limit"`

	// RateLimitBuffer adds buffer time when waiting for rate limit reset
	RateLimitBuffer time.Duration `json:"rate_limit_buffer" yaml:"rate_limit_buffer"`
}

// LoggingConfig contains configuration for monitoring and debugging
type LoggingConfig struct {
	// LogLevel controls the verbosity of logging (error, warn, info, debug)
	LogLevel string `json:"log_level" yaml:"log_level"`

	// LogFormat specifies the log format (json, text)
	LogFormat string `json:"log_format" yaml:"log_format"`

	// EnableAuditLog enables audit logging for timestamp updates
	EnableAuditLog bool `json:"enable_audit_log" yaml:"enable_audit_log"`

	// EnablePerformanceMetrics enables performance metric collection
	EnablePerformanceMetrics bool `json:"enable_performance_metrics" yaml:"enable_performance_metrics"`

	// LogAPICallsSaved enables logging of API calls saved by incremental fetching
	LogAPICallsSaved bool `json:"log_api_calls_saved" yaml:"log_api_calls_saved"`
}

// Config contains all configuration options for the star watcher
type Config struct {
	Incremental IncrementalConfig `json:"incremental" yaml:"incremental"`
	Retry       RetryConfig       `json:"retry" yaml:"retry"`
	Logging     LoggingConfig     `json:"logging" yaml:"logging"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Incremental: IncrementalConfig{
			Enabled:             true,
			FullSyncInterval:    24, // Full sync every 24 hours
			FallbackOnError:     true,
			MaxIncrementalPages: 10, // Limit to 10 pages (1000 repos) per incremental fetch
			DetectUnstars:       true,
			DetectReStars:       true,
			TimestampTolerance:  1 * time.Minute, // 1 minute tolerance for clock skew
		},
		Retry: RetryConfig{
			MaxRetries:        3,
			InitialDelay:      1 * time.Second,
			MaxDelay:          30 * time.Second,
			BackoffMultiplier: 2.0,
			RetryOnRateLimit:  true,
			RateLimitBuffer:   30 * time.Second, // Extra 30 seconds buffer
		},
		Logging: LoggingConfig{
			LogLevel:                 "info",
			LogFormat:                "text",
			EnableAuditLog:           true,
			EnablePerformanceMetrics: true,
			LogAPICallsSaved:         true,
		},
	}
}

// Validate checks if the configuration values are valid
func (c *Config) Validate() error {
	// Validate incremental config
	if c.Incremental.FullSyncInterval < 0 {
		return fmt.Errorf("full_sync_interval must be non-negative")
	}

	if c.Incremental.MaxIncrementalPages <= 0 {
		c.Incremental.MaxIncrementalPages = 10 // Set default
	}

	if c.Incremental.TimestampTolerance < 0 {
		c.Incremental.TimestampTolerance = 1 * time.Minute // Set default
	}

	// Validate retry config
	if c.Retry.MaxRetries < 0 {
		c.Retry.MaxRetries = 3 // Set to default
	}

	if c.Retry.InitialDelay <= 0 {
		c.Retry.InitialDelay = 1 * time.Second
	}

	if c.Retry.MaxDelay <= 0 {
		c.Retry.MaxDelay = 30 * time.Second
	}

	if c.Retry.BackoffMultiplier <= 1.0 {
		c.Retry.BackoffMultiplier = 2.0
	}

	// Validate logging config
	validLogLevels := map[string]bool{
		"error": true,
		"warn":  true,
		"info":  true,
		"debug": true,
	}

	if !validLogLevels[c.Logging.LogLevel] {
		c.Logging.LogLevel = "info"
	}

	validLogFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if !validLogFormats[c.Logging.LogFormat] {
		c.Logging.LogFormat = "text"
	}

	return nil
}
