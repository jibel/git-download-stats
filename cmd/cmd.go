package cmd

import (
	"log"
	"os"

	"github.com/jibel/git-download-stats/internal"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root cobra command.
func NewRootCmd() *cobra.Command {
	var ghOwner string
	var ghRepo string
	var ghToken string
	var detailedOutput bool

	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch GitHub release download statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := internal.FetchReleaseStats(cmd.Context(), ghOwner, ghRepo, ghToken)
			if err != nil {
				return err
			}

			if len(stats.Releases) == 0 {
				log.Printf("No releases found for %s/%s\n", ghOwner, ghRepo)
				return nil
			}

			internal.DisplayStats(stats, detailedOutput)
			return nil
		},
	}

	cmd.Flags().StringVarP(&ghOwner, "owner", "o", "", "GitHub repository owner")
	cmd.Flags().StringVarP(&ghRepo, "repo", "r", "", "GitHub repository name")
	cmd.Flags().StringVarP(&ghToken, "token", "t", os.Getenv("GITHUB_TOKEN"), "GitHub API token (default from GITHUB_TOKEN env)")
	cmd.Flags().BoolVarP(&detailedOutput, "detailed", "d", false, "Show detailed download statistics including asset names and sizes")

	//cmd.MarkFlagRequired("project")

	return cmd
}

// Execute runs the command and handles errors.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}
