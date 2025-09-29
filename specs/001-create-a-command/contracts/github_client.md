# GitHub Client Interface Contract

## GitHubClient Interface

```go
type GitHubClient interface {
    // GetStarredRepositories fetches all starred repositories for a user
    // Returns paginated results with rate limit information
    GetStarredRepositories(ctx context.Context, username string, opts *StarredOptions) (*StarredResponse, error)
    
    // GetRateLimit returns current rate limit status
    GetRateLimit(ctx context.Context) (*RateLimitInfo, error)
    
    // ValidateUser checks if a GitHub username exists
    ValidateUser(ctx context.Context, username string) error
}

type StarredOptions struct {
    // Page pagination cursor (empty for first page)
    Cursor string
    
    // PerPage number of items per page (max 100)
    PerPage int
    
    // Sort order: "created", "updated", "pushed", "full_name"
    Sort string
    
    // Direction: "asc" or "desc"
    Direction string
}

type StarredResponse struct {
    Repositories []Repository
    PageInfo     PageInfo
    RateLimit    RateLimitInfo
}
```

## Contract Tests

### GetStarredRepositories Contract
- **Input**: Valid username, nil options → Should return first page with default pagination
- **Input**: Valid username, specific cursor → Should return next page of results
- **Input**: Invalid username → Should return specific error type (UserNotFoundError)
- **Input**: Valid username, PerPage > 100 → Should clamp to 100 or return validation error
- **Output**: Repository objects must have all required fields populated
- **Output**: PageInfo must correctly indicate if more pages available
- **Output**: RateLimit must contain current GitHub API rate limit status
- **Error Handling**: Network failures should return specific error types for retry logic
- **Error Handling**: Rate limit exceeded should return RateLimitError with retry information

### GetRateLimit Contract
- **Output**: Must return current rate limit status with Limit, Remaining, ResetTime
- **Error Handling**: API failures should return specific error types
- **Behavior**: Should work with both authenticated and unauthenticated clients

### ValidateUser Contract
- **Input**: Existing username → Should return nil error
- **Input**: Non-existent username → Should return UserNotFoundError
- **Input**: Invalid username format → Should return ValidationError
- **Error Handling**: API failures should return specific error types distinct from user validation