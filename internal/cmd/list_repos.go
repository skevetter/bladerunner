package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	gh "github.com/skevetter/bladerunner/internal/github"
	"github.com/spf13/cobra"
)

var (
	githubToken string
)

var listReposCmd = &cobra.Command{
	Use:   "list-repos",
	Short: "List GitHub repositories",
	Long:  `List all repositories for the authenticated user.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Get token from flag or environment
		token := githubToken
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		if token == "" {
			logrus.Fatal("GitHub token is required. Set GITHUB_TOKEN env var or use --token flag")
		}

		client := gh.NewClient(ctx, token)

		repos, err := client.ListRepositories()
		if err != nil {
			logrus.WithError(err).Fatal("Failed to list repositories")
		}

		logrus.Infof("Found %d repositories:", len(repos))
		for _, repo := range repos {
			private := ""
			if repo.GetPrivate() {
				private = " (private)"
			}
			fmt.Printf("  - %s%s\n", repo.GetFullName(), private)
			if repo.Description != nil && *repo.Description != "" {
				fmt.Printf("    %s\n", *repo.Description)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listReposCmd)
	listReposCmd.Flags().StringVarP(&githubToken, "token", "t", "", "GitHub personal access token")
}
