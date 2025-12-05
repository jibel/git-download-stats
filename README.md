# GitHub Download Statistics

A Go program that fetches, displays, and stores GitHub release download statistics in an SQLite database with full history tracking.

## Features

- 📊 Fetch download statistics for all releases of a GitHub project
- 💾 Store statistics in SQLite database with timestamps for historical analysis
- 📦 Display detailed information about release artifacts
- 🔐 Support for GitHub API authentication (higher rate limits)
- 📈 Track download trends over time
- 🔍 Query and compare statistics across different time periods
- 📑 Multiple commands for different use cases

## Installation

### Prerequisites

- Go 1.21 or later
- (Optional) GitHub personal access token for authenticated requests

### Build

```bash
make build
```

Or manually:

```bash
go mod download
go build -o git-download-stats
```

## Commands

### Fetch Command
Fetch GitHub release statistics and optionally store them in a database.

```bash
./git-download-stats fetch -o <owner> -r <repo> [OPTIONS]
```

**Options:**
- `-o, --owner` (required): GitHub repository owner
- `-r, --repo` (required): GitHub repository name
- `-t, --token`: GitHub API token (defaults to `GITHUB_TOKEN` env var)
- `-d, --detailed`: Show detailed output with asset names
- `-s, --store`: Store statistics in SQLite database
- `--db`: Custom database path (default: `github-stats.db`)

**Examples:**
```bash
# Fetch and display stats
./git-download-stats fetch -o cli -r cli

# Fetch with detailed output
./git-download-stats fetch -o cli -r cli -d

# Fetch and store in database
./git-download-stats fetch -o cli -r cli -s

# Fetch from custom database location
./git-download-stats fetch -o cli -r cli -s --db ./data/stats.db
```

### Show Command
Display the latest stored statistics for a repository.

```bash
./git-download-stats show <owner> <repo> [--db <path>]
```

**Examples:**
```bash
# Show latest stats for GitHub CLI
./git-download-stats show cli cli

# Show from custom database
./git-download-stats show cli cli --db ./data/stats.db
```

### History Command
Show historical snapshots of statistics over time.

```bash
./git-download-stats history <owner> <repo> [--limit <n>] [--db <path>]
```

**Options:**
- `--limit`: Number of historical snapshots to show (default: 10)
- `--db`: Custom database path

**Examples:**
```bash
# Show last 10 fetches
./git-download-stats history cli cli

# Show last 5 fetches
./git-download-stats history cli cli --limit 5
```

### Compare Command
Compare statistics between oldest and newest records within a time period.

```bash
./git-download-stats compare <owner> <repo> [--days <n>] [--db <path>]
```

**Options:**
- `--days`: Number of days to look back (default: 30)
- `--db`: Custom database path

**Examples:**
```bash
# Compare stats over last 30 days
./git-download-stats compare cli cli

# Compare over last 90 days
./git-download-stats compare cli cli --days 90
```

## Database Schema

The SQLite database stores statistics in two tables:

**stats table:**
- `id`: Primary key
- `owner`: Repository owner
- `repo`: Repository name
- `tag`: Release tag
- `release_name`: Release name
- `total_downloads`: Total downloads for the release
- `fetched_at`: Timestamp when data was fetched
- `created_at`: Release creation date

**assets table:**
- `id`: Primary key
- `stat_id`: Foreign key to stats
- `name`: Asset filename
- `download_count`: Number of downloads
- `size`: Asset file size in bytes
- `content_type`: MIME type

## Usage Examples

### Set up automated statistics collection

```bash
# Fetch and store stats once
./git-download-stats fetch -o hashicorp -r terraform -s

# Add to cron to run daily
0 0 * * * /path/to/git-download-stats fetch -o hashicorp -r terraform -s
```

### Track download trends

```bash
# View download history over last month
./git-download-stats history hashicorp terraform --limit 20

# Compare stats from different time periods
./git-download-stats compare hashicorp terraform --days 90
```

### Monitor multiple projects

```bash
# Fetch stats for multiple projects
./git-download-stats fetch -o cli -r cli -s
./git-download-stats fetch -o hashicorp -r terraform -s
./git-download-stats fetch -o golang -r go -s

# Check the stored statistics
./git-download-stats show cli cli
./git-download-stats show hashicorp terraform
./git-download-stats show golang go
```

## API Rate Limits

- **Unauthenticated requests**: 60 requests per hour
- **Authenticated requests**: 5,000 requests per hour

For better performance with large repositories, set a GitHub personal access token:

```bash
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

## Creating a GitHub Personal Access Token

1. Go to https://github.com/settings/tokens
2. Click "Generate new token"
3. Select `public_repo` scope (or `repo` for private repositories)
4. Copy the token and set it as environment variable:

```bash
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

## Output Examples

### Fetch Output
```
Download Statistics for cli/cli
Total Releases: 182 | Total Downloads: 68698450
Last Updated: 2025-12-05 10:53:51 CET

RELEASE            TAG      ASSETS  DOWNLOADS  CREATED AT
---                ---      ---     ---        ---
GitHub CLI 2.83.1  v2.83.1  22      571716     2025-11-13
GitHub CLI 2.83.0  v2.83.0  22      350999     2025-11-04
GitHub CLI 2.82.1  v2.82.1  22      516555     2025-10-22

✅ Statistics compiled successfully

✓ Statistics stored in github-stats.db
```

### History Output
```
📊 Statistics History for cli/cli (last 2 fetches)

[1] Fetched at: 2025-12-05 10:53:51 +0100 | Total Releases: 182 | Total Downloads: 68698450
    Top 3 releases:
      - GitHub CLI 2.83.1 (v2.83.1): 571716 downloads
      - GitHub CLI 2.3.0 (v2.3.0): 5085174 downloads
      - GitHub CLI 2.40.1 (v2.40.1): 3615787 downloads

[2] Fetched at: 2025-12-05 10:52:58 +0100 | Total Releases: 182 | Total Downloads: 68698421
    Top 3 releases:
      - GitHub CLI 2.83.1 (v2.83.1): 571716 downloads
      - GitHub CLI 2.3.0 (v2.3.0): 5085174 downloads
      - GitHub CLI 2.40.1 (v2.40.1): 3615787 downloads
```

### Compare Output
```
📈 Download Statistics Comparison for cli/cli
Period: Last 30 days
Oldest: 2025-11-05 | Newest: 2025-12-05

Total Downloads:
  Oldest: 67856234
  Newest: 68698450
  Growth: +842216 (+1.24%)

Top 5 releases by growth:
  1. GitHub CLI 2.83.1 (v2.83.1): +100000 (+21.27%)
  2. GitHub CLI 2.83.0 (v2.83.0): +45000 (+14.68%)
  3. GitHub CLI 2.82.1 (v2.82.1): +38000 (+7.86%)
```

## Development

### Running tests

```bash
make test
```

### Cleaning build artifacts

```bash
make clean
```

### Downloading dependencies

```bash
make deps
```

## Architecture

- **cmd/cmd.go**: Command-line interface using Cobra framework
- **internal/github.go**: GitHub API client for fetching release data
- **internal/database.go**: SQLite database operations and queries
- **internal/records.go**: Display formatting utilities

## License

MIT License - feel free to use this project as you wish.
