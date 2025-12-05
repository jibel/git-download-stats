package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v56/github"
)

type Asset struct {
	Name          string
	DownloadCount int
	Size          int64
	ContentType   string
}

type Release struct {
	Name           string
	Tag            string
	Assets         []Asset
	TotalDownloads int
	CreatedAt      time.Time
	PublishedAt    time.Time
	IsPrerelease   bool
	IsDraft        bool
}

type ReleaseStats struct {
	Owner          string
	Repo           string
	TotalDownloads int
	Releases       []Release
	FetchedAt      time.Time
}

// FetchReleaseStats fetches all releases and their asset download statistics from GitHub
func FetchReleaseStats(ctx context.Context, owner, repo, token string) (*ReleaseStats, error) {
	client := github.NewClient(nil)

	// If token is provided, create an authenticated client for higher rate limits
	if token != "" {
		client = github.NewClient(nil).WithAuthToken(token)
	}

	stats := &ReleaseStats{
		Owner:     owner,
		Repo:      repo,
		Releases:  make([]Release, 0),
		FetchedAt: time.Now(),
	}

	// Fetch all releases (paginated)
	opt := &github.ListOptions{PerPage: 100}
	for {
		releases, resp, err := client.Repositories.ListReleases(ctx, owner, repo, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list releases: %w", err)
		}

		for _, ghRelease := range releases {
			rel := Release{
				Tag:          ghRelease.GetTagName(),
				CreatedAt:    ghRelease.GetCreatedAt().Time,
				PublishedAt:  ghRelease.GetPublishedAt().Time,
				IsPrerelease: ghRelease.GetPrerelease(),
				IsDraft:      ghRelease.GetDraft(),
				Assets:       make([]Asset, 0),
			}

			// Use release name if available, otherwise use tag
			if ghRelease.GetName() != "" {
				rel.Name = ghRelease.GetName()
			} else {
				rel.Name = ghRelease.GetTagName()
			}

			// Process assets
			for _, ghAsset := range ghRelease.Assets {
				asset := Asset{
					Name:          ghAsset.GetName(),
					DownloadCount: ghAsset.GetDownloadCount(),
					Size:          int64(ghAsset.GetSize()),
					ContentType:   ghAsset.GetContentType(),
				}
				rel.Assets = append(rel.Assets, asset)
				rel.TotalDownloads += asset.DownloadCount
			}

			stats.Releases = append(stats.Releases, rel)
			stats.TotalDownloads += rel.TotalDownloads
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return stats, nil
}
