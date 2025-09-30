package monitor

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/akme/gh-stars-watcher/internal/config"
)

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err         error
	IsRateLimit bool
	RetryAfter  time.Duration
	IsTemporary bool
}

func (r *RetryableError) Error() string {
	return r.Err.Error()
}

func (r *RetryableError) Unwrap() error {
	return r.Err
}

// RetryManager handles retry logic with exponential backoff
type RetryManager struct {
	config *config.RetryConfig
	logger func(format string, args ...interface{})
}

// NewRetryManager creates a new retry manager with the given configuration
func NewRetryManager(cfg *config.RetryConfig) *RetryManager {
	return &RetryManager{
		config: cfg,
		logger: log.Printf, // Default logger
	}
}

// SetLogger sets a custom logger function
func (r *RetryManager) SetLogger(logger func(format string, args ...interface{})) {
	r.logger = logger
}

// ExecuteWithRetry executes a function with retry logic
func (r *RetryManager) ExecuteWithRetry(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Check context before each attempt
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := operation()
		if err == nil {
			// Success
			if attempt > 0 {
				r.logger("Operation succeeded after %d retry attempts", attempt)
			}
			return nil
		}

		lastErr = err

		// Check if this is the last attempt
		if attempt == r.config.MaxRetries {
			r.logger("Operation failed after %d attempts: %v", attempt+1, err)
			break
		}

		// Determine if error is retryable
		retryableErr, isRetryable := err.(*RetryableError)
		if !isRetryable {
			// For non-retryable errors, convert them to retryable if they seem temporary
			retryableErr = &RetryableError{
				Err:         err,
				IsTemporary: r.isTemporaryError(err),
			}
		}

		// Don't retry non-retryable errors
		if !retryableErr.IsTemporary && !retryableErr.IsRateLimit {
			r.logger("Non-retryable error encountered: %v", err)
			return err
		}

		// Handle rate limit specifically
		if retryableErr.IsRateLimit && !r.config.RetryOnRateLimit {
			r.logger("Rate limit encountered but retry disabled: %v", err)
			return err
		}

		// Calculate delay
		delay := r.calculateDelay(attempt, retryableErr)

		r.logger("Operation failed (attempt %d/%d), retrying after %v: %v",
			attempt+1, r.config.MaxRetries+1, delay, err)

		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("operation failed after %d attempts, last error: %w", r.config.MaxRetries+1, lastErr)
}

// calculateDelay calculates the delay for the next retry attempt
func (r *RetryManager) calculateDelay(attempt int, retryableErr *RetryableError) time.Duration {
	// For rate limits, use the specified retry after time
	if retryableErr.IsRateLimit && retryableErr.RetryAfter > 0 {
		return retryableErr.RetryAfter + r.config.RateLimitBuffer
	}

	// Exponential backoff
	delay := time.Duration(float64(r.config.InitialDelay) * math.Pow(r.config.BackoffMultiplier, float64(attempt)))

	// Cap at max delay
	if delay > r.config.MaxDelay {
		delay = r.config.MaxDelay
	}

	return delay
}

// isTemporaryError determines if an error is likely temporary and retryable
func (r *RetryManager) isTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Common temporary error patterns
	temporaryPatterns := []string{
		"connection reset",
		"connection refused",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"no such host",
		"i/o timeout",
		"context deadline exceeded",
		"server error",
		"internal server error",
		"bad gateway",
		"service unavailable",
		"gateway timeout",
	}

	for _, pattern := range temporaryPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// WrapRetryableError wraps an error as retryable
func WrapRetryableError(err error, isRateLimit bool, retryAfter time.Duration) error {
	if err == nil {
		return nil
	}

	return &RetryableError{
		Err:         err,
		IsRateLimit: isRateLimit,
		RetryAfter:  retryAfter,
		IsTemporary: true,
	}
}

// WrapNonRetryableError wraps an error as non-retryable
func WrapNonRetryableError(err error) error {
	if err == nil {
		return nil
	}

	return &RetryableError{
		Err:         err,
		IsTemporary: false,
	}
}
