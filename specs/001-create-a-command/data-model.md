# Data Model: GitHub Stars Monitor CLI

## Core Entities

### Repository
Represents a starred GitHub repository with all metadata needed for comparison and display.

**Fields**:
- `FullName` (string): Owner/repo format (e.g., "microsoft/vscode")
- `Description` (string): Repository description (nullable)
- `StarCount` (int): Current number of stars
- `UpdatedAt` (time.Time): Last repository update timestamp
- `URL` (string): Repository URL for browser access
- `StarredAt` (time.Time): When user starred this repository (from API)
- `Language` (string): Primary programming language (optional)
- `Private` (bool): Whether repository is private

**Validation Rules**:
- FullName must match pattern `^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`
- URL must be valid HTTPS GitHub repository URL
- StarCount must be non-negative
- StarredAt must not be future timestamp

**Relationships**:
- Many repositories belong to one UserState
- No direct relationships to other entities

### UserState
Represents the persisted state for a GitHub user's monitoring session.

**Fields**:
- `Username` (string): GitHub username being monitored
- `LastCheck` (time.Time): Timestamp of last successful check
- `Repositories` ([]Repository): Previously seen starred repositories
- `TotalCount` (int): Total repositories at last check (for pagination validation)
- `StateVersion` (string): Schema version for backward compatibility
- `CheckCount` (int): Number of successful checks performed

**Validation Rules**:
- Username must match GitHub username pattern `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$`
- LastCheck must not be future timestamp
- StateVersion must follow semantic versioning pattern
- TotalCount must be non-negative
- CheckCount must be non-negative

**State Transitions**:
- New → FirstRun: Initialize with current starred repositories, no "new" items reported
- FirstRun → Monitoring: Subsequent runs compare against stored state
- Monitoring → Monitoring: Update with new repositories and refresh metadata
- Any → Corrupted: Invalid file format triggers rebuild from current state

### APIResponse
Represents structured data from GitHub API calls for rate limiting and pagination.

**Fields**:
- `RateLimit` (RateLimitInfo): Current rate limit status
- `PageInfo` (PageInfo): Pagination metadata
- `Repositories` ([]Repository): Repository data from current API call
- `RequestDuration` (time.Duration): Time taken for API request
- `StatusCode` (int): HTTP response status code

**Sub-entities**:

#### RateLimitInfo
- `Limit` (int): Maximum requests per hour
- `Remaining` (int): Requests remaining in current window
- `ResetTime` (time.Time): When rate limit resets
- `Used` (int): Requests used in current window

#### PageInfo
- `HasNext` (bool): Whether more pages are available
- `NextCursor` (string): Cursor for next page (GitHub pagination)
- `TotalCount` (int): Total items across all pages
- `PerPage` (int): Items per page

**Validation Rules**:
- RateLimit values must be non-negative
- ResetTime must be future timestamp during active monitoring
- StatusCode must be valid HTTP status code (100-599)
- NextCursor must be valid base64 string when HasNext is true

## Data Flow

### Repository Comparison Logic
1. Load previous UserState from local storage
2. Fetch current starred repositories via GitHub API
3. Compare current repositories against UserState.Repositories by FullName
4. Identify new repositories: present in current but not in previous
5. Update UserState with current repositories and new LastCheck timestamp
6. Persist updated UserState atomically

### State Persistence Strategy
- **File Format**: JSON with indentation for human readability
- **File Location**: `~/.star-watcher/{username}.json` (configurable via --state-file)
- **Atomic Updates**: Write to temporary file, then rename to prevent corruption
- **Backup Strategy**: Keep previous state as `.bak` file for recovery
- **Schema Evolution**: StateVersion field enables backward-compatible updates

### Error Recovery
- **Corrupted State**: Rebuild from current GitHub API state with warning
- **API Failures**: Preserve existing state, provide actionable error messages
- **Network Issues**: Implement exponential backoff with clear status reporting
- **Rate Limiting**: Wait with progress indication, respect GitHub headers

## Performance Considerations

### Memory Usage
- Stream large repository lists rather than loading all into memory
- Use pagination to limit memory footprint per API call
- Implement repository comparison using maps for O(n) lookup performance

### Storage Efficiency
- Compress state files for users with thousands of starred repositories
- Implement cleanup logic to remove old state files (configurable retention)
- Use efficient JSON marshaling with struct tags for optimal serialization

### API Efficiency
- Cache repository metadata to minimize API calls
- Use conditional requests (If-Modified-Since headers) when supported
- Implement intelligent pagination sizing based on total repository count