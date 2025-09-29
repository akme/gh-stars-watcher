package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	keychainService = "gh-stars-watcher"
	keychainUser    = "github-token"
)

// KeychainTokenManager implements TokenManager using OS keychain and environment variables
type KeychainTokenManager struct {
	githubClient GitHubValidator // Interface for validating tokens
}

// GitHubValidator interface for validating GitHub tokens
type GitHubValidator interface {
	ValidateToken(ctx context.Context, token string) (bool, error)
}

// NewKeychainTokenManager creates a new keychain-based token manager
func NewKeychainTokenManager(validator GitHubValidator) *KeychainTokenManager {
	return &KeychainTokenManager{
		githubClient: validator,
	}
}

// GetToken retrieves a GitHub token from available sources in priority order:
// 1. GITHUB_TOKEN environment variable
// 2. OS keychain
// 3. Interactive prompt (not implemented in this function)
func (k *KeychainTokenManager) GetToken(ctx context.Context) (token string, source string, err error) {
	// First, try environment variable
	if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
		return envToken, "environment", nil
	}

	// Second, try OS keychain
	keychainToken, err := keyring.Get(keychainService, keychainUser)
	if err == nil && keychainToken != "" {
		return keychainToken, "keychain", nil
	}

	// No token found
	return "", "", &TokenNotFoundError{
		Message: "no GitHub token found in environment or keychain",
	}
}

// StoreToken stores a GitHub token securely in the OS keychain
func (k *KeychainTokenManager) StoreToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Validate token before storing
	if k.githubClient != nil {
		valid, err := k.githubClient.ValidateToken(ctx, token)
		if err != nil {
			return &TokenValidationError{
				Token: maskToken(token),
				Cause: err,
			}
		}
		if !valid {
			return &TokenValidationError{
				Token: maskToken(token),
				Cause: fmt.Errorf("token is not valid"),
			}
		}
	}

	// Store in keychain
	if err := keyring.Set(keychainService, keychainUser, token); err != nil {
		return fmt.Errorf("failed to store token in keychain: %v", err)
	}

	return nil
}

// RemoveToken removes stored GitHub token from the OS keychain
func (k *KeychainTokenManager) RemoveToken(ctx context.Context) error {
	if err := keyring.Delete(keychainService, keychainUser); err != nil {
		// Don't error if the token doesn't exist
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "cannot find") {
			return nil
		}
		return fmt.Errorf("failed to remove token from keychain: %v", err)
	}
	return nil
}

// ValidateToken checks if a GitHub token is valid by delegating to the GitHub client
func (k *KeychainTokenManager) ValidateToken(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("token cannot be empty")
	}

	if k.githubClient == nil {
		return false, fmt.Errorf("no GitHub client available for validation")
	}

	return k.githubClient.ValidateToken(ctx, token)
}

// maskToken masks a token for logging/error purposes
func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
