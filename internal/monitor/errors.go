package monitor

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ErrorType represents different types of errors that can occur
type ErrorType int

const (
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeAuth
	ErrorTypeRateLimit
	ErrorTypeNetwork
	ErrorTypeUser
	ErrorTypeStorage
	ErrorTypeValidation
	ErrorTypeAPI
)

// MonitorError represents an error that occurred during monitoring
type MonitorError struct {
	Type      ErrorType
	Message   string
	Cause     error
	Timestamp time.Time
	Context   map[string]interface{}
}

// Error implements the error interface
func (e *MonitorError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *MonitorError) Unwrap() error {
	return e.Cause
}

// IsRetryable returns true if the error is retryable
func (e *MonitorError) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeNetwork, ErrorTypeRateLimit:
		return true
	case ErrorTypeAPI:
		// Some API errors are retryable (5xx)
		if httpErr, ok := e.Cause.(*HTTPError); ok {
			return httpErr.StatusCode >= 500
		}
		return false
	default:
		return false
	}
}

// GetRetryDelay returns the suggested delay before retrying
func (e *MonitorError) GetRetryDelay() time.Duration {
	switch e.Type {
	case ErrorTypeRateLimit:
		if resetTime, ok := e.Context["reset_time"].(time.Time); ok {
			delay := time.Until(resetTime)
			if delay > 0 {
				return delay
			}
		}
		return 1 * time.Hour // Default rate limit delay
	case ErrorTypeNetwork:
		return 30 * time.Second
	case ErrorTypeAPI:
		return 10 * time.Second
	default:
		return 5 * time.Second
	}
}

// HTTPError represents an HTTP-related error
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
	URL        string
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d %s: %s", e.StatusCode, e.Status, e.URL)
}

// ErrorHandler provides centralized error handling for monitoring operations
type ErrorHandler struct {
	maxRetries int
	baseDelay  time.Duration
}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		maxRetries: 3,
		baseDelay:  time.Second,
	}
}

// NewMonitorError creates a new monitor error
func (eh *ErrorHandler) NewMonitorError(errorType ErrorType, message string, cause error) *MonitorError {
	return &MonitorError{
		Type:      errorType,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// HandleError processes an error and returns an appropriate MonitorError
func (eh *ErrorHandler) HandleError(err error, context string) *MonitorError {
	if err == nil {
		return nil
	}

	// If it's already a MonitorError, return it
	if monitorErr, ok := err.(*MonitorError); ok {
		return monitorErr
	}

	// Classify the error
	errorType := eh.classifyError(err)
	message := fmt.Sprintf("Error %s", context)

	monitorErr := eh.NewMonitorError(errorType, message, err)
	eh.addErrorContext(monitorErr, err)

	return monitorErr
}

// classifyError determines the type of error
func (eh *ErrorHandler) classifyError(err error) ErrorType {
	errMsg := strings.ToLower(err.Error())

	// Check for specific error patterns
	switch {
	case strings.Contains(errMsg, "401") || strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "bad credentials") || strings.Contains(errMsg, "token"):
		return ErrorTypeAuth

	case strings.Contains(errMsg, "403") || strings.Contains(errMsg, "rate limit") ||
		strings.Contains(errMsg, "x-ratelimit"):
		return ErrorTypeRateLimit

	case strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "dns"):
		return ErrorTypeNetwork

	case strings.Contains(errMsg, "user") || strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "404"):
		return ErrorTypeUser

	case strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "format"):
		return ErrorTypeValidation

	case strings.Contains(errMsg, "storage") || strings.Contains(errMsg, "file") ||
		strings.Contains(errMsg, "permission"):
		return ErrorTypeStorage

	case strings.Contains(errMsg, "api") || strings.Contains(errMsg, "5") ||
		strings.Contains(errMsg, "server"):
		return ErrorTypeAPI

	default:
		return ErrorTypeUnknown
	}
}

// addErrorContext adds relevant context to the error
func (eh *ErrorHandler) addErrorContext(monitorErr *MonitorError, originalErr error) {
	// Add HTTP-specific context
	if httpErr, ok := originalErr.(*HTTPError); ok {
		monitorErr.Context["http_status"] = httpErr.StatusCode
		monitorErr.Context["http_url"] = httpErr.URL
		monitorErr.Context["http_body"] = httpErr.Body
	}

	// Add context-specific information
	switch monitorErr.Type {
	case ErrorTypeRateLimit:
		// Try to extract rate limit reset time from headers
		// This would be populated by the caller if available
		break
	case ErrorTypeAuth:
		monitorErr.Context["suggestion"] = "Check GitHub token permissions and validity"
	case ErrorTypeUser:
		monitorErr.Context["suggestion"] = "Verify the username exists and is accessible"
	case ErrorTypeNetwork:
		monitorErr.Context["suggestion"] = "Check network connection and GitHub availability"
	}
}

// RetryWithBackoff executes an operation with retry logic
func (eh *ErrorHandler) RetryWithBackoff(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= eh.maxRetries; attempt++ {
		// Execute the operation
		err := operation()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Handle the error
		monitorErr := eh.HandleError(err, "during retry operation")

		// Check if we should retry
		if attempt == eh.maxRetries || !monitorErr.IsRetryable() {
			break
		}

		// Calculate delay
		delay := eh.calculateBackoffDelay(attempt, monitorErr)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// calculateBackoffDelay calculates the delay for the next retry attempt
func (eh *ErrorHandler) calculateBackoffDelay(attempt int, err *MonitorError) time.Duration {
	// Use error-specific delay if available
	if err.IsRetryable() {
		if delay := err.GetRetryDelay(); delay > 0 {
			return delay
		}
	}

	// Exponential backoff with jitter
	delay := eh.baseDelay * time.Duration(1<<uint(attempt))

	// Add jitter (±25%)
	jitter := delay / 4
	jitterOffset := time.Duration(float64(jitter) * 2 * float64(time.Now().UnixNano()%1000) / 1000.0)
	delay = delay - jitter + jitterOffset

	// Cap at 5 minutes
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}

	return delay
}

// FormatUserFriendlyError formats an error for user display
func (eh *ErrorHandler) FormatUserFriendlyError(err error) string {
	monitorErr := eh.HandleError(err, "")

	switch monitorErr.Type {
	case ErrorTypeAuth:
		return "Authentication failed. Please check your GitHub token and permissions."
	case ErrorTypeRateLimit:
		resetTime := "unknown"
		if rt, ok := monitorErr.Context["reset_time"].(time.Time); ok {
			resetTime = rt.Format("15:04:05")
		}
		return fmt.Sprintf("Rate limit exceeded. Try again after %s.", resetTime)
	case ErrorTypeNetwork:
		return "Network error. Please check your internet connection and try again."
	case ErrorTypeUser:
		return "User not found or not accessible. Please verify the username."
	case ErrorTypeStorage:
		return "Storage error. Please check file permissions and available disk space."
	case ErrorTypeValidation:
		return "Invalid input. Please check your parameters and try again."
	case ErrorTypeAPI:
		return "GitHub API error. The service may be temporarily unavailable."
	default:
		return fmt.Sprintf("An unexpected error occurred: %v", err)
	}
}

// IsTemporaryError checks if an error is temporary and might resolve on its own
func IsTemporaryError(err error) bool {
	if monitorErr, ok := err.(*MonitorError); ok {
		return monitorErr.IsRetryable()
	}

	// Check for common temporary error patterns
	errMsg := strings.ToLower(err.Error())
	temporaryPatterns := []string{
		"timeout",
		"connection refused",
		"network",
		"temporary",
		"rate limit",
		"server error",
		"503",
		"502",
		"500",
	}

	for _, pattern := range temporaryPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// CombineErrors combines multiple errors into a single error
func CombineErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		return errors[0]
	}

	var messages []string
	for _, err := range errors {
		if err != nil {
			messages = append(messages, err.Error())
		}
	}

	if len(messages) == 0 {
		return nil
	}

	return fmt.Errorf("multiple errors occurred: %s", strings.Join(messages, "; "))
}

// RecoverableError represents an error from which recovery might be possible
type RecoverableError struct {
	Err          error
	RecoveryTips []string
	RetryAfter   *time.Duration
}

// Error implements the error interface
func (r *RecoverableError) Error() string {
	return r.Err.Error()
}

// Unwrap returns the underlying error
func (r *RecoverableError) Unwrap() error {
	return r.Err
}

// NewRecoverableError creates a new recoverable error
func NewRecoverableError(err error, tips []string, retryAfter *time.Duration) *RecoverableError {
	return &RecoverableError{
		Err:          err,
		RecoveryTips: tips,
		RetryAfter:   retryAfter,
	}
}

// IsRecoverable checks if an error is recoverable
func IsRecoverable(err error) bool {
	_, ok := err.(*RecoverableError)
	return ok
}
