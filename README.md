# GitHub Stars Monitor

A command-line tool that tracks changes in a user's starred repositories, showing only newly starred repositories between runs.

## Features

- **Incremental Monitoring**: Only shows newly starred repositories since the last run
- **First Run Baseline**: Establishes a baseline on the first run without showing output
- **Multiple Output Formats**: Support for text (human-readable) and JSON formats
- **Secure Authentication**: Uses OS keychain for token storage with interactive fallback
- **Rate Limit Handling**: Intelligent rate limit detection and user-friendly error messages
- **Clean Architecture**: Well-structured codebase with interfaces and dependency injection
- **Comprehensive Error Handling**: User-friendly error messages with actionable guidance

## Installation

### Prerequisites

- Go 1.25 or later
- macOS (for keychain integration) or Linux

### Build from Source

```bash
git clone https://github.com/akme/gh-stars-watcher.git
cd gh-stars-watcher
make build
```

## Usage

### Basic Usage

Monitor a user's starred repositories:

```bash
./bin/star-watcher monitor octocat
```

### First Run

On the first run, the tool establishes a baseline:

```bash
$ ./bin/star-watcher monitor octocat
First run for octocat - baseline established with 3 starred repositories.
Run again to detect newly starred repositories.
```

### Subsequent Runs

Subsequent runs show only newly starred repositories:

```bash
$ ./bin/star-watcher monitor octocat
🌟 octocat has starred 2 new repositories!

⭐ octocat/Hello-World
   My first repository on GitHub!
   Language: None | Stars: 1
   Starred: 2024-01-15
   https://github.com/octocat/Hello-World

⭐ github/docs
   The open-source repo for docs.github.com
   Language: JavaScript | Stars: 2150
   Starred: 2024-01-16
   https://github.com/github/docs

Total repositories: 5
Previous check: 2024-01-14 10:30:45
```

### JSON Output

Get structured output in JSON format:

```bash
./bin/star-watcher monitor octocat --output json
```

```json
{
  "username": "octocat",
  "new_repositories": [
    {
      "full_name": "octocat/Hello-World",
      "description": "My first repository on GitHub!",
      "star_count": 1,
      "updated_at": "2024-01-15T10:30:00Z",
      "url": "https://github.com/octocat/Hello-World",
      "starred_at": "2024-01-15T12:00:00Z",
      "language": "",
      "private": false
    }
  ],
  "total_repositories": 5,
  "previous_check": "2024-01-14T10:30:45Z",
  "current_check": "2024-01-16T14:20:30Z",
  "rate_limit": {
    "limit": 60,
    "remaining": 45,
    "reset_time": "2024-01-16T15:00:00Z",
    "used": 15
  },
  "is_first_run": false
}
```

### Cleanup

Remove stored state for a user:

```bash
./bin/star-watcher cleanup octocat
```

Remove all stored states:

```bash
./bin/star-watcher cleanup --all
```

## Command Reference

### Global Flags

- `-o, --output string`: Output format: `text` (default) or `json`
- `-q, --quiet`: Quiet output (errors only)
- `-v, --verbose`: Verbose output (detailed logging)
- `--state-file string`: Custom state file path (default: `~/.star-watcher/{username}.json`)

### Monitor Command

```bash
star-watcher monitor [username] [flags]
```

Monitor a GitHub user's starred repositories for changes.

**Examples:**
```bash
star-watcher monitor octocat
star-watcher monitor octocat --output json
star-watcher monitor octocat --state-file ./custom-state.json
star-watcher monitor octocat --verbose
```

### Cleanup Command

```bash
star-watcher cleanup [username] [flags]
```

Remove stored state files for a specific user or all users.

**Flags:**
- `--all`: Remove all state files (use with caution)

**Examples:**
```bash
star-watcher cleanup octocat
star-watcher cleanup octocat --state-file ./custom-state.json
star-watcher cleanup --all
```

## Authentication

The tool supports multiple authentication methods:

1. **Environment Variable**: Set `GITHUB_TOKEN` environment variable
2. **OS Keychain**: Tokens are securely stored in the system keychain
3. **Interactive Prompt**: If no token is found, you'll be prompted to enter one

### Creating a GitHub Token

1. Go to [GitHub Settings > Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Add a note (e.g., "GitHub Stars Monitor")
4. Select scopes:
   - `public_repo` (for public repositories)
   - `repo` (if you want to monitor private starred repositories)
5. Click "Generate token"
6. Copy the token and provide it when prompted

## State Storage

The tool stores state in JSON files under `~/.star-watcher/`:

- `~/.star-watcher/{username}.json`: Contains the baseline of starred repositories
- Files include repository metadata, star counts, and timestamps
- State files are atomic-write protected to prevent corruption

## Rate Limiting

- **Unauthenticated**: 60 requests per hour
- **Authenticated**: 5000 requests per hour

The tool automatically handles rate limits and provides helpful error messages:

```bash
Error: GitHub API rate limit exceeded. Resets at: 2024-01-16T15:00:00Z
```

## Architecture

The project follows clean architecture principles:

```
cmd/star-watcher/          # CLI entry point
internal/
├── auth/                  # Authentication (keychain, prompts)
├── cli/                   # CLI commands and output formatting
├── github/                # GitHub API client
├── monitor/               # Core monitoring logic
└── storage/               # State persistence
tests/
├── contract/              # Interface contract tests
└── integration/           # End-to-end tests
```

## Development

### Building

```bash
go build -o star-watcher ./cmd/star-watcher
```

### Testing

```bash
go test ./...
```

### Code Style

The project follows standard Go conventions and uses:

- Clean architecture with interfaces
- Dependency injection
- Comprehensive error handling
- Test-driven development (TDD)

## Performance

- **Startup Time**: ~20ms for CLI operations
- **API Performance**: ~50 seconds for 3000 repositories (limited by GitHub API)
- **State File Size**: ~1MB for 3000 repositories
- **Memory Usage**: Minimal - processes repositories in batches

## Troubleshooting

### Rate Limit Exceeded

Wait for the rate limit to reset or authenticate with a GitHub token for higher limits.

### Invalid Username Format

Ensure the username follows GitHub's rules:
- 1-39 characters long
- Only alphanumeric characters and hyphens
- Cannot start or end with a hyphen

### State File Corruption

Remove the state file and re-run to establish a new baseline:

```bash
star-watcher cleanup username
star-watcher monitor username
```

### Permission Errors

Ensure the tool has permission to write to `~/.star-watcher/` directory.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following the existing code style
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [go-github](https://github.com/google/go-github) for GitHub API integration
- [Cobra](https://github.com/spf13/cobra) for CLI framework
- [go-keyring](https://github.com/zalando/go-keyring) for secure token storage
- `tests/` - Comprehensive test suite (contract, integration, unit)

## License

MIT License - see LICENSE file for details.