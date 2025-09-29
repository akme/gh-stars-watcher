# Quickstart: GitHub Stars Monitor CLI

This guide validates the complete user workflow from installation to monitoring newly starred repositories.

## Prerequisites
- Go 1.25+ installed
- GitHub account with some starred repositories
- Terminal access

## Installation

```bash
# Install from source
git clone https://github.com/akme/gh-stars-watcher.git
cd gh-stars-watcher
go install ./cmd/star-watcher

# Verify installation
star-watcher --version
```

## First Run (Baseline Setup)

```bash
# Monitor a GitHub user (replace with actual username)
star-watcher --user octocat

# Expected output for first run:
# ✓ Fetching starred repositories for octocat...
# ✓ Found 42 starred repositories
# ✓ Saved initial state to ~/.star-watcher/octocat.json
# 
# First run complete! No new repositories to display.
# Future runs will show newly starred repositories since this baseline.
```

## Subsequent Monitoring Runs

```bash
# Run after user has starred new repositories
star-watcher --user octocat

# Expected output when new repositories found:
# ✓ Fetching starred repositories for octocat...
# ✓ Found 45 starred repositories (3 new since last check)
# 
# 📦 New starred repositories:
# 
# microsoft/vscode
# │ Description: Visual Studio Code
# │ Stars: 158,234 | Updated: 2025-09-29T10:30:00Z
# │ URL: https://github.com/microsoft/vscode
# 
# golang/go
# │ Description: The Go programming language
# │ Stars: 118,567 | Updated: 2025-09-29T09:15:00Z  
# │ URL: https://github.com/golang/go
# 
# kubernetes/kubernetes
# │ Description: Production-Grade Container Scheduling and Management
# │ Stars: 102,345 | Updated: 2025-09-29T08:45:00Z
# │ URL: https://github.com/kubernetes/kubernetes
# 
# ✓ State updated successfully
```

## Authentication Setup (Optional but Recommended)

```bash
# Set up GitHub token for higher rate limits and private repo access
export GITHUB_TOKEN="ghp_your_personal_access_token_here"

# Or use interactive setup
star-watcher --user octocat --setup-auth

# Expected interactive prompt:
# GitHub token not found. Would you like to set one up? (y/N): y
# Enter your GitHub personal access token: [hidden input]
# ✓ Token validated successfully for user: your-username
# ✓ Token stored securely in OS keychain
```

## Output Format Options

```bash
# JSON output for programmatic use
star-watcher --user octocat --output json

# Expected JSON output:
# {
#   "user": "octocat",
#   "timestamp": "2025-09-29T12:00:00Z",
#   "new_repositories": [
#     {
#       "full_name": "microsoft/vscode",
#       "description": "Visual Studio Code",
#       "star_count": 158234,
#       "updated_at": "2025-09-29T10:30:00Z",
#       "url": "https://github.com/microsoft/vscode",
#       "starred_at": "2025-09-29T11:45:00Z"
#     }
#   ],
#   "total_new": 1,
#   "total_starred": 45
# }

# Quiet mode for scripts
star-watcher --user octocat --output json --quiet > new-stars.json
```

## State Management

```bash
# Use custom state file location
star-watcher --user octocat --state-file ./custom-state.json

# List all monitored users
star-watcher cleanup --list

# Expected output:
# Monitored users:
# - octocat (last check: 2025-09-29T12:00:00Z, 45 repositories)
# - github (last check: 2025-09-28T15:30:00Z, 123 repositories)

# Clean up old state files
star-watcher cleanup --older-than 30d

# Expected output:
# ✓ Removed 2 state files older than 30 days:
#   - old-user-1.json (last modified: 2025-08-15)
#   - old-user-2.json (last modified: 2025-08-20)
```

## Automated Monitoring Setup

```bash
# Add to crontab for daily monitoring
crontab -e

# Add this line to check daily at 9 AM and save results
0 9 * * * /usr/local/bin/star-watcher --user octocat --output json --quiet >> ~/.star-watcher/daily-log.json

# Or create a monitoring script
cat > monitor-stars.sh << 'EOF'
#!/bin/bash
USERS=("octocat" "github" "microsoft")
for user in "${USERS[@]}"; do
    echo "Checking $user..."
    star-watcher --user "$user" --output json --quiet > "/tmp/stars-$user-$(date +%Y%m%d).json"
done
EOF

chmod +x monitor-stars.sh
```

## Error Scenarios Validation

```bash
# Test with non-existent user
star-watcher --user nonexistentuser12345

# Expected error output:
# ❌ Error: GitHub user 'nonexistentuser12345' not found
# 
# Please check the username and try again.
# You can verify the user exists at: https://github.com/nonexistentuser12345

# Test without network connectivity (disconnect network)
star-watcher --user octocat

# Expected error output:
# ❌ Error: Unable to connect to GitHub API
# 
# Please check your internet connection and try again.
# If the problem persists, GitHub may be experiencing issues.
# Check status at: https://www.githubstatus.com

# Test with rate limiting (make many rapid requests)
for i in {1..100}; do star-watcher --user octocat; done

# Expected rate limit handling:
# ⏳ Rate limit reached. Waiting 3 seconds before retry...
# ⏳ Rate limit reached. Waiting 6 seconds before retry...
# ✓ Continuing after rate limit reset
```

## Validation Checklist

- [ ] Installation completes without errors
- [ ] First run creates baseline state file
- [ ] Subsequent runs detect newly starred repositories
- [ ] Authentication setup works via environment variable and interactive prompt
- [ ] JSON output format is valid and programmatically parseable
- [ ] Custom state file locations work correctly
- [ ] Cleanup command removes old state files
- [ ] Automated monitoring via cron works reliably
- [ ] Error messages are actionable and user-friendly
- [ ] Rate limiting is handled gracefully with progress indication
- [ ] Performance meets constitutional requirements (<500ms cached, <2s fresh)

## Success Criteria

This quickstart validates constitutional compliance:
- **Code Quality**: CLI follows Go idioms and provides clear help text
- **Testing Excellence**: All user scenarios can be automated as integration tests
- **User Experience**: Consistent flag patterns, actionable errors, progress indicators
- **Performance**: Startup and operation times meet specified requirements