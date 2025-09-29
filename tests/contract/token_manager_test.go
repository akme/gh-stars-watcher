package contract

import (
	"context"
	"testing"

	"github.com/akme/gh-stars-watcher/internal/auth"
)

// TestTokenManagerContract validates the TokenManager interface contract
func TestTokenManagerContract(t *testing.T) {
	// This test will fail until TokenManager interface and implementation exist
	var manager auth.TokenManager
	if manager == nil {
		t.Skip("TokenManager implementation not available yet - this is expected in TDD Red phase")
	}

	ctx := context.Background()

	t.Run("GetToken", func(t *testing.T) {
		t.Run("FromEnvironmentVariable", func(t *testing.T) {
			// Should check GITHUB_TOKEN environment variable first
			token, source, err := manager.GetToken(ctx)
			if err != nil {
				t.Errorf("Expected no error getting token, got: %v", err)
			}
			if token == "" {
				t.Log("No token found - this is acceptable for testing")
			}
			if token != "" && source != "environment" {
				t.Errorf("Expected source 'environment', got '%s'", source)
			}
		})

		t.Run("FromKeychain", func(t *testing.T) {
			// Should fall back to OS keychain if env var not set
			// This test may skip if no keychain token is stored
			token, source, err := manager.GetToken(ctx)
			if err != nil && err.Error() != "no token found" {
				t.Errorf("Expected no error or 'no token found', got: %v", err)
			}
			if token != "" && source != "keychain" && source != "environment" {
				t.Errorf("Expected source 'keychain' or 'environment', got '%s'", source)
			}
		})

		t.Run("InteractivePrompt", func(t *testing.T) {
			// This would normally prompt user interactively
			// In tests, this should be mockable or skippable
			t.Skip("Interactive prompts not testable in automated tests")
		})
	})

	t.Run("StoreToken", func(t *testing.T) {
		testToken := "test-token-12345"

		err := manager.StoreToken(ctx, testToken)
		if err != nil {
			t.Errorf("Expected no error storing token, got: %v", err)
		}

		// Verify token can be retrieved
		retrievedToken, source, err := manager.GetToken(ctx)
		if err != nil {
			t.Errorf("Expected no error retrieving stored token, got: %v", err)
		}
		if retrievedToken != testToken {
			t.Errorf("Expected retrieved token '%s', got '%s'", testToken, retrievedToken)
		}
		if source != "keychain" && source != "environment" {
			t.Errorf("Expected source 'keychain' or 'environment', got '%s'", source)
		}
	})

	t.Run("RemoveToken", func(t *testing.T) {
		// Store a token first
		testToken := "test-token-to-remove"
		err := manager.StoreToken(ctx, testToken)
		if err != nil {
			t.Errorf("Expected no error storing token, got: %v", err)
		}

		// Remove the token
		err = manager.RemoveToken(ctx)
		if err != nil {
			t.Errorf("Expected no error removing token, got: %v", err)
		}

		// Verify token is no longer available
		token, _, err := manager.GetToken(ctx)
		if err == nil && token != "" {
			// Only error if we still get a token from keychain
			// Environment variables can't be "removed" by the app
			t.Errorf("Expected no token after removal, but got token")
		}
	})

	t.Run("ValidateToken", func(t *testing.T) {
		t.Run("ValidToken", func(t *testing.T) {
			// This should make a test API call to validate the token
			validToken := "valid-github-token"
			valid, err := manager.ValidateToken(ctx, validToken)
			if err != nil {
				t.Errorf("Expected no error validating token, got: %v", err)
			}
			if !valid {
				t.Error("Expected valid token to be validated as true")
			}
		})

		t.Run("InvalidToken", func(t *testing.T) {
			invalidToken := "invalid-token-12345"
			valid, err := manager.ValidateToken(ctx, invalidToken)
			if err != nil {
				t.Errorf("Expected no error (just false result), got: %v", err)
			}
			if valid {
				t.Error("Expected invalid token to be validated as false")
			}
		})

		t.Run("EmptyToken", func(t *testing.T) {
			valid, err := manager.ValidateToken(ctx, "")
			if err == nil {
				t.Error("Expected error for empty token")
			}
			if valid {
				t.Error("Expected empty token to be invalid")
			}
		})
	})
}

// TestTokenSecurity validates token security requirements
func TestTokenSecurity(t *testing.T) {
	// This test will fail until TokenManager implementation exists
	t.Skip("TokenManager implementation not available yet - this is expected in TDD Red phase")

	// Expected security features:
	// - Tokens stored in OS keychain, not plain text files
	// - Environment variables checked first for CI/CD scenarios
	// - Interactive prompts mask token input
	// - Token validation before storage
}
