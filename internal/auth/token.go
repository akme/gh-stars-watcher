package auth

import (
	"context"
)

// TokenManager defines the interface for managing GitHub authentication tokens
type TokenManager interface {
	// GetToken retrieves a GitHub token from available sources
	// Returns token, source, and error. Source indicates where token was found.
	GetToken(ctx context.Context) (token string, source string, err error)

	// StoreToken stores a GitHub token securely (typically in OS keychain)
	StoreToken(ctx context.Context, token string) error

	// RemoveToken removes stored GitHub token from secure storage
	RemoveToken(ctx context.Context) error

	// ValidateToken checks if a GitHub token is valid by making a test API call
	ValidateToken(ctx context.Context, token string) (valid bool, err error)
}

// TokenNotFoundError represents an error when no token is available
type TokenNotFoundError struct {
	Message string
}

func (e *TokenNotFoundError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "no GitHub token found"
}

// TokenValidationError represents an error during token validation
type TokenValidationError struct {
	Token string
	Cause error
}

func (e *TokenValidationError) Error() string {
	return "token validation failed: " + e.Cause.Error()
}
