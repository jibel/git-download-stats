# GitHub Download Statistics

A Go program that fetches and displays download statistics for release artifacts from GitHub projects.

## Features

- Fetch download statistics for all releases of a GitHub project
- Display detailed information about release artifacts
- Support for GitHub API authentication (higher rate limits)
- Show total downloads across all releases

## Installation

### Prerequisites

- Go 1.25 or later
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

## Usage

### Basic usage

```bash
./git-download-stats -owner <owner> -repo <repo>
```

### With GitHub token (for higher rate limits)

```bash
export GITHUB_TOKEN="your_github_token_here"
./git-download-stats -owner <owner> -repo <repo>
```

Or pass directly:

```bash
./git-download-stats -owner <owner> -repo <repo> -token "your_github_token_here"
```

### Detailed view

```bash
./git-download-stats -owner <owner> -repo <repo> -detailed
```

## Examples

### Get download stats for GitHub CLI

```bash
./git-download-stats -owner cli -repo cli
```

### Get detailed stats for GitHub CLI

```bash
./git-download-stats -owner cli -repo cli -detailed
```

### Using with make

```bash
make run ARGS="-owner cli -repo cli -detailed"
```

## Output

The program displays:
- Release name and tag
- Number of assets per release
- Total downloads per release
- Creation date

With the `-detailed` flag, it also shows:
- Individual asset names
- Download count per asset

## Example Output

```
Download Statistics for canonical/ubuntu-pro-for-wsl
Total Releases: 5 | Total Downloads: 3959
Last Updated: 2025-12-05 09:34:47 CET

RELEASE     TAG       ASSETS  DOWNLOADS  CREATED AT
---         ---       ---     ---        ---
1.0.1.0     1.0.1     2       2796       2025-12-01
0.9999.8.0  0.9999b8  2       156        2025-11-25
0.9999.7.0  0.9999b7  2       223        2025-11-17
0.9999.6.0  0.9999b6  2       774        2025-09-17
0.9999.5.0  0.9999b5  2       10         2025-09-17

✅ Statistics compiled successfully
```

## Options

- `-owner` (required): GitHub repository owner
- `-repo` (required): GitHub repository name
- `-token` (optional): GitHub API personal access token (defaults to `GITHUB_TOKEN` environment variable)
- `-detailed` (optional): Show detailed information including asset names

## API Rate Limits

- **Unauthenticated requests**: 60 requests per hour
- **Authenticated requests**: 5,000 requests per hour

For better performance with large repositories, set a GitHub personal access token.


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

## License

MIT License - feel free to use this project as you wish.
