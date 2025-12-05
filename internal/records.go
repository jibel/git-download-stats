package internal

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

func DisplayStats(stats *ReleaseStats, detailed bool) {
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
