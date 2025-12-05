package internal

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// captureOutput redirects stdout to a buffer for the duration of f and returns its contents.
func captureOutput(t *testing.T, f func()) string {
	t.Helper()
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	f()

	_ = w.Close()
	os.Stdout = origStdout
	<-done

	return buf.String()
}

func sampleStats() *ReleaseStats {
	return &ReleaseStats{
		Owner: "owner",
		Repo:  "repo",
		Releases: []Release{
			{
				Name:           "Release One",
				Tag:            "v1.0.0",
				Assets:         []Asset{{Name: "asset1.tar.gz", DownloadCount: 5}},
				TotalDownloads: 5,
				CreatedAt:      time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				Name:           "Release Two",
				Tag:            "v1.1.0",
				Assets:         []Asset{{Name: "asset2.tar.gz", DownloadCount: 10}},
				TotalDownloads: 10,
				CreatedAt:      time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		TotalDownloads: 15,
	}
}

func TestDisplayStatsSummary(t *testing.T) {
	stats := sampleStats()

	out := captureOutput(t, func() {
		DisplayStats(stats, false)
	})

	checks := []string{
		"Download Statistics for owner/repo",
		"Total Releases: 2 | Total Downloads: 15",
		"RELEASE      TAG     ASSETS  DOWNLOADS  CREATED AT",
		"Release One  v1.0.0  1       5          2024-05-01",
		"Release Two  v1.1.0  1       10         2024-06-01",
		"Statistics compiled successfully",
	}

	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Fatalf("expected output to contain %q, got:\n%s", c, out)
		}
	}
}

func TestDisplayStatsDetailed(t *testing.T) {
	stats := sampleStats()

	out := captureOutput(t, func() {
		DisplayStats(stats, true)
	})

	checks := []string{
		"RELEASE      TAG     ASSETS  TOTAL DOWNLOADS  CREATED AT",
		"Release One  v1.0.0  1       5                2024-05-01",
		"Release Two  v1.1.0  1       10               2024-06-01",
		"asset1.tar.gz    5 downloads",
		"asset2.tar.gz    10 downloads",
	}

	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Fatalf("expected output to contain %q, got:\n%s", c, out)
		}
	}
}
