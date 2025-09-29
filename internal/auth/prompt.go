package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// PromptTokenManager handles interactive token prompts
type PromptTokenManager struct {
	tokenManager TokenManager // Delegate for actual token management
}

// NewPromptTokenManager creates a new interactive prompt token manager
func NewPromptTokenManager(tokenManager TokenManager) *PromptTokenManager {
	return &PromptTokenManager{
		tokenManager: tokenManager,
	}
}

// PromptForToken interactively prompts the user for a GitHub token
func (p *PromptTokenManager) PromptForToken(ctx context.Context) (string, error) {
	fmt.Print("GitHub token not found. Please enter your GitHub personal access token.\n")
	fmt.Print("You can create one at: https://github.com/settings/tokens\n")
	fmt.Print("Required scopes: public_repo (or repo for private repositories)\n\n")

	fmt.Print("GitHub Token: ")

	// Try to read password without echo (hidden input)
	token, err := readPassword()
	if err != nil {
		// Fall back to regular input if terminal doesn't support hidden input
		fmt.Print("\nWarning: Token will be visible. Press Ctrl+C to cancel.\n")
		fmt.Print("GitHub Token: ")
		reader := bufio.NewReader(os.Stdin)
		token, err = reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read token: %v", err)
		}
	}

	fmt.Println() // New line after hidden input

	// Clean up the token
	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}

	// Validate the token
	if p.tokenManager != nil {
		valid, err := p.tokenManager.ValidateToken(ctx, token)
		if err != nil {
			return "", fmt.Errorf("failed to validate token: %v", err)
		}
		if !valid {
			return "", fmt.Errorf("invalid GitHub token")
		}

		// Ask if user wants to store the token
		fmt.Print("Would you like to store this token securely for future use? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err == nil {
			response = strings.ToLower(strings.TrimSpace(response))
			if response == "y" || response == "yes" {
				if err := p.tokenManager.StoreToken(ctx, token); err != nil {
					fmt.Printf("Warning: Failed to store token: %v\n", err)
				} else {
					fmt.Println("Token stored securely.")
				}
			}
		}
	}

	return token, nil
}

// GetTokenWithPrompt tries to get a token from storage, prompting if not found
func (p *PromptTokenManager) GetTokenWithPrompt(ctx context.Context) (string, string, error) {
	// First try to get existing token
	if p.tokenManager != nil {
		token, source, err := p.tokenManager.GetToken(ctx)
		if err == nil && token != "" {
			return token, source, nil
		}
	}

	// No token found, prompt for one
	token, err := p.PromptForToken(ctx)
	if err != nil {
		return "", "", err
	}

	return token, "prompt", nil
}

// readPassword reads a password from stdin without echoing it to the terminal
func readPassword() (string, error) {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(syscall.Stdin)) {
		// Not a terminal, fall back to regular input
		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		return strings.TrimSpace(password), err
	}

	// Read password without echo
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	return string(bytePassword), nil
}

// GetToken tries to get token from the underlying token manager first, then prompts if needed
func (p *PromptTokenManager) GetToken(ctx context.Context) (token string, source string, err error) {
	// First try to get token from the underlying token manager (e.g., keychain)
	token, source, err = p.tokenManager.GetToken(ctx)
	if err == nil && token != "" {
		return token, source, nil
	}

	// If no token found, prompt the user
	token, err = p.PromptForToken(ctx)
	if err != nil {
		return "", "", err
	}

	// Store the token for future use
	if storeErr := p.tokenManager.StoreToken(ctx, token); storeErr != nil {
		// Log the error but don't fail - we have a valid token
		fmt.Printf("Warning: Failed to store token in keychain: %v\n", storeErr)
	}

	return token, "user_prompt", nil
}

// StoreToken delegates to the underlying token manager
func (p *PromptTokenManager) StoreToken(ctx context.Context, token string) error {
	return p.tokenManager.StoreToken(ctx, token)
}

// RemoveToken delegates to the underlying token manager
func (p *PromptTokenManager) RemoveToken(ctx context.Context) error {
	return p.tokenManager.RemoveToken(ctx)
}

// ValidateToken delegates to the underlying token manager
func (p *PromptTokenManager) ValidateToken(ctx context.Context, token string) (bool, error) {
	return p.tokenManager.ValidateToken(ctx, token)
}
