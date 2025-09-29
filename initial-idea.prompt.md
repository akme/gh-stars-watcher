Create a command-line application in Go that monitors GitHub user's starred repositories and displays changes between runs. The application should:

1. Accept a GitHub username as a command-line argument
2. Fetch the user's currently starred repositories using the GitHub API
3. Store the repository data in a local file for comparison between runs
4. Compare current starred repositories with previously stored data
5. Output only newly starred repositories since the last run
6. Include for each repository:
   - Full repository name (owner/repo)
   - Description
   - Star count
   - Last update timestamp
   - Repository URL

Technical requirements:

- Use GitHub REST API v3 or GraphQL API v4
- Implement proper error handling and rate limit awareness
- Store data in a structured format (JSON/YAML)
- Support both authenticated and unauthenticated API access
- Follow Go best practices and standard project layout
- Include appropriate logging for debugging and monitoring
- Exit with proper status codes based on execution success/failure

Example usage:

```bash
star-watcher --user username [--token GITHUB_TOKEN] [--output json|text] [--state-file path/to/state.json]
```

The application should be suitable for automated execution via cron jobs or scheduled tasks, with machine-readable output options for integration with other tools.
