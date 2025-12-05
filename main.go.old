package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"
)

func main() {
	owner := flag.String("owner", "", "GitHub repository owner (required)")
	repo := flag.String("repo", "", "GitHub repository name (required)")
	token := flag.String("token", os.Getenv("GITHUB_TOKEN"), "GitHub API token (optional, defaults to GITHUB_TOKEN env var)")
	detailed := flag.Bool("detailed", false, "Show detailed information including asset names and sizes")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage: git-download-stats [options]

Fetches and displays download statistics for release artifacts from a GitHub project.

Options:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	if *owner == "" || *repo == "" {
		fmt.Fprintf(os.Stderr, "Error: -owner and -repo flags are required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	stats, err := FetchReleaseStats(context.Background(), *owner, *repo, *token)
	if err != nil {
		log.Fatalf("Failed to fetch release stats: %v", err)
	}

	if len(stats.Releases) == 0 {
		fmt.Printf("No releases found for %s/%s\n", *owner, *repo)
		return
	}

	displayStats(stats, *detailed)
}

func displayStats(stats *ReleaseStats, detailed bool) {
	fmt.Printf("\nDownload Statistics for %s/%s\n", stats.Owner, stats.Repo)
	fmt.Printf("Total Releases: %d | Total Downloads: %d\n", len(stats.Releases), stats.TotalDownloads)
	fmt.Printf("Last Updated: %s\n\n", time.Now().Format("2006-01-02 15:04:05 MST"))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if detailed {
		fmt.Fprintln(w, "RELEASE\tTAG\tASSETS\tTOTAL DOWNLOADS\tCREATED AT")
		fmt.Fprintln(w, "---\t---\t---\t---\t---")
		for _, rel := range stats.Releases {
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
				rel.Name,
				rel.Tag,
				len(rel.Assets),
				rel.TotalDownloads,
				rel.CreatedAt.Format("2006-01-02"),
			)
			for _, asset := range rel.Assets {
				fmt.Fprintf(w, "\t→ %s\t\t%s\t%d downloads\n",
					"",
					asset.Name,
					asset.DownloadCount,
				)
			}
		}
	} else {
		fmt.Fprintln(w, "RELEASE\tTAG\tASSETS\tDOWNLOADS\tCREATED AT")
		fmt.Fprintln(w, "---\t---\t---\t---\t---")
		for _, rel := range stats.Releases {
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
				rel.Name,
				rel.Tag,
				len(rel.Assets),
				rel.TotalDownloads,
				rel.CreatedAt.Format("2006-01-02"),
			)
		}
	}

	w.Flush()

	fmt.Printf("\n✅ Statistics compiled successfully\n\n")
}
