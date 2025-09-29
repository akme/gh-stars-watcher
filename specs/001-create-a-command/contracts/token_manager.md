# Authentication Token Interface Contract

## TokenManager Interface

```go
type TokenManager interface {
    // GetToken retrieves GitHub token from configured sources
    // Returns TokenNotFoundError if no token available from any source
    GetToken() (string, error)
    
    // SetToken stores token securely in preferred storage (keychain)
    SetToken(token string) error
    
    // RemoveToken removes stored token from all sources
    RemoveToken() error
    
    // ValidateToken checks if token is valid with GitHub API
    ValidateToken(token string) (*TokenInfo, error)
    
    // IsTokenRequired returns true if operations require authentication
    IsTokenRequired(operation string) bool
}

type TokenInfo struct {
    Username   string   `json:"username"`
    Scopes     []string `json:"scopes"`
    ExpiresAt  *time.Time `json:"expires_at,omitempty"`
    TokenType  string   `json:"token_type"`
    RateLimit  RateLimitInfo `json:"rate_limit"`
}

// Token discovery priority:
// 1. Command line flag --token
// 2. Environment variable GITHUB_TOKEN
// 3. OS keychain storage
// 4. Config file ~/.star-watcher/config.json
// 5. Interactive prompt (when terminal is available)
```

## Contract Tests

### GetToken Contract
- **Behavior**: Should check sources in priority order (env var → keychain → config → prompt)
- **Output**: Valid GitHub token string when found
- **Error**: TokenNotFoundError when no token available from any source
- **Error**: Should not prompt interactively when running in non-interactive environment
- **Security**: Should not log token values in any error messages or debugging output

### SetToken Contract
- **Input**: Valid GitHub token → Should store in OS keychain (preferred) or config file fallback
- **Input**: Invalid token format → Should return ValidationError before storage
- **Behavior**: Should encrypt/secure token according to platform capabilities
- **Error Handling**: Should gracefully fallback to config file if keychain unavailable

### RemoveToken Contract
- **Behavior**: Should remove token from all possible storage locations
- **Output**: Should return nil error even if token wasn't found (idempotent)
- **Security**: Should securely overwrite/clear token data where possible

### ValidateToken Contract
- **Input**: Valid GitHub token → Should return TokenInfo with user details and scopes
- **Input**: Invalid/expired token → Should return InvalidTokenError
- **Input**: Token with insufficient permissions → Should return InsufficientScopesError
- **Output**: TokenInfo must include username, scopes, and rate limit information
- **Behavior**: Should make actual GitHub API call to validate token

### IsTokenRequired Contract
- **Input**: "public_repos" → Should return false (public data accessible without auth)
- **Input**: "private_repos" → Should return true (requires authentication)
- **Input**: "user_info" → Should return true (requires authentication)
- **Behavior**: Should help users understand when authentication is needed

## Error Types

```go
type TokenNotFoundError struct {
    Message string
    Sources []string // Sources that were checked
}

type InvalidTokenError struct {
    Message string
    Reason  string // "expired", "invalid_format", "revoked"
}

type InsufficientScopesError struct {
    Message       string
    RequiredScope string
    AvailableScopes []string
}
```

## Security Requirements

- Token values must never appear in logs, error messages, or stdout
- Keychain integration should use platform-specific secure storage
- Config file token storage should have restricted file permissions (600)
- Interactive prompts should mask token input (no echo)
- Memory containing tokens should be cleared after use when possible