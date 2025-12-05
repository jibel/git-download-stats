package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jibel/git-download-stats/internal"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root cobra command.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "git-download-stats",
		Short: "Fetch and store GitHub release download statistics",
		Long:  "Fetch and store GitHub release download statistics with SQLite persistence",
	}

	rootCmd.AddCommand(newFetchCmd())
	rootCmd.AddCommand(newShowCmd())
	rootCmd.AddCommand(newHistoryCmd())
	rootCmd.AddCommand(newCompareCmd())

	return rootCmd
}

func newFetchCmd() *cobra.Command {
	var ghOwner string
	var ghRepo string
	var ghToken string
	var detailedOutput bool
	var store bool
	var dbPath string

	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch GitHub release download statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			if ghOwner == "" || ghRepo == "" {
				return fmt.Errorf("owner and repo flags are required")
			}

			stats, err := internal.FetchReleaseStats(cmd.Context(), ghOwner, ghRepo, ghToken)
			if err != nil {
				return err
			}

			if len(stats.Releases) == 0 {
				log.Printf("No releases found for %s/%s\n", ghOwner, ghRepo)
				return nil
			}

			internal.DisplayStats(stats, detailedOutput)

			// Store in database if requested
			if store {
				db, err := internal.NewDatabase(dbPath)
				if err != nil {
					return fmt.Errorf("failed to connect to database: %w", err)
				}
				defer db.Close()

				if err := db.StoreStats(stats); err != nil {
					return fmt.Errorf("failed to store stats: %w", err)
				}
				dbFile := dbPath
				if dbFile == "" {
					dbFile = "github-stats.db"
				}
				log.Printf("\n✓ Statistics stored in %s\n", dbFile)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&ghOwner, "owner", "o", "", "GitHub repository owner (required)")
	cmd.Flags().StringVarP(&ghRepo, "repo", "r", "", "GitHub repository name (required)")
	cmd.Flags().StringVarP(&ghToken, "token", "t", os.Getenv("GITHUB_TOKEN"), "GitHub API token")
	cmd.Flags().BoolVarP(&detailedOutput, "detailed", "d", false, "Show detailed output with asset names")
	cmd.Flags().BoolVarP(&store, "store", "s", false, "Store statistics in database")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: github-stats.db)")

	return cmd
}

func newShowCmd() *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "show <owner> <repo>",
		Short: "Show latest stored statistics",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			owner := args[0]
			repo := args[1]

			db, err := internal.NewDatabase(dbPath)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			stats, err := db.GetLatestStats(owner, repo)
			if err != nil {
				return fmt.Errorf("failed to retrieve stats: %w", err)
			}

			if len(stats.Releases) == 0 {
				fmt.Printf("No statistics found for %s/%s\n", owner, repo)
				return nil
			}

			fmt.Printf("\nLatest Statistics for %s/%s\n", owner, repo)
			fmt.Printf("Fetched at: %s\n", stats.FetchedAt.Format("2006-01-02 15:04:05 MST"))
			internal.DisplayStats(stats, false)

			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: github-stats.db)")

	return cmd
}

func newHistoryCmd() *cobra.Command {
	var dbPath string
	var limit int

	cmd := &cobra.Command{
		Use:   "history <owner> <repo>",
		Short: "Show statistics history",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			owner := args[0]
			repo := args[1]

			db, err := internal.NewDatabase(dbPath)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			allStats, err := db.GetStatsHistory(owner, repo, limit)
			if err != nil {
				return fmt.Errorf("failed to retrieve history: %w", err)
			}

			if len(allStats) == 0 {
				fmt.Printf("No history found for %s/%s\n", owner, repo)
				return nil
			}

			fmt.Printf("\n📊 Statistics History for %s/%s (last %d fetches)\n\n", owner, repo, len(allStats))

			for i, stats := range allStats {
				fmt.Printf("[%d] Fetched at: %s | Total Releases: %d | Total Downloads: %d\n",
					i+1,
					stats.FetchedAt.Format("2006-01-02 15:04:05 MST"),
					len(stats.Releases),
					stats.TotalDownloads,
				)

				if len(stats.Releases) > 0 {
					fmt.Printf("    Top 3 releases:\n")
					count := 3
					if len(stats.Releases) < 3 {
						count = len(stats.Releases)
					}
					for j := 0; j < count; j++ {
						rel := stats.Releases[j]
						fmt.Printf("      - %s (%s): %d downloads\n", rel.Name, rel.Tag, rel.TotalDownloads)
					}
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: github-stats.db)")
	cmd.Flags().IntVar(&limit, "limit", 10, "Number of historical snapshots to show")

	return cmd
}

func newCompareCmd() *cobra.Command {
	var dbPath string
	var days int

	cmd := &cobra.Command{
		Use:   "compare <owner> <repo>",
		Short: "Compare statistics across time",
		Long:  "Compare download statistics between the oldest and newest records",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			owner := args[0]
			repo := args[1]

			db, err := internal.NewDatabase(dbPath)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			endTime := time.Now()
			startTime := endTime.AddDate(0, 0, -days)

			allStats, err := db.GetStatsBetween(owner, repo, startTime, endTime)
			if err != nil {
				return fmt.Errorf("failed to retrieve stats: %w", err)
			}

			if len(allStats) < 2 {
				fmt.Printf("Need at least 2 data points to compare (found %d)\n", len(allStats))
				return nil
			}

			oldest := allStats[len(allStats)-1]
			newest := allStats[0]

			fmt.Printf("\n📈 Download Statistics Comparison for %s/%s\n", owner, repo)
			fmt.Printf("Period: Last %d days\n", days)
			fmt.Printf("Oldest: %s | Newest: %s\n\n", oldest.FetchedAt.Format("2006-01-02"), newest.FetchedAt.Format("2006-01-02"))

			totalGrowth := newest.TotalDownloads - oldest.TotalDownloads
			var percentGrowth float64
			if oldest.TotalDownloads > 0 {
				percentGrowth = (float64(totalGrowth) / float64(oldest.TotalDownloads)) * 100
			}

			fmt.Printf("Total Downloads:\n")
			fmt.Printf("  Oldest: %d\n", oldest.TotalDownloads)
			fmt.Printf("  Newest: %d\n", newest.TotalDownloads)
			fmt.Printf("  Growth: %+d (%+.2f%%)\n\n", totalGrowth, percentGrowth)

			// Compare top releases
			fmt.Printf("Top 5 releases by growth:\n")
			type relComparison struct {
				name   string
				tag    string
				oldDL  int
				newDL  int
				growth int
			}

			comparisons := make([]relComparison, 0)
			for _, newRel := range newest.Releases {
				for _, oldRel := range oldest.Releases {
					if oldRel.Tag == newRel.Tag {
						growth := newRel.TotalDownloads - oldRel.TotalDownloads
						comparisons = append(comparisons, relComparison{
							name:   newRel.Name,
							tag:    newRel.Tag,
							oldDL:  oldRel.TotalDownloads,
							newDL:  newRel.TotalDownloads,
							growth: growth,
						})
						break
					}
				}
			}

			// Sort by growth (simple bubble sort for small list)
			for i := 0; i < len(comparisons); i++ {
				for j := i + 1; j < len(comparisons); j++ {
					if comparisons[j].growth > comparisons[i].growth {
						comparisons[i], comparisons[j] = comparisons[j], comparisons[i]
					}
				}
			}

			count := 5
			if len(comparisons) < 5 {
				count = len(comparisons)
			}
			for i := 0; i < count; i++ {
				c := comparisons[i]
				pct := 0.0
				if c.oldDL > 0 {
					pct = (float64(c.growth) / float64(c.oldDL)) * 100
				}
				fmt.Printf("  %d. %s (%s): %+d (%+.2f%%)\n", i+1, c.name, c.tag, c.growth, pct)
			}

			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: github-stats.db)")
	cmd.Flags().IntVar(&days, "days", 30, "Number of days to look back")

	return cmd
}

// Execute runs the command and handles errors.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}
