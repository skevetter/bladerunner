package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bladerunner",
	Short: "GitHub Runner orchestration and management service",
	Long: `Bladerunner is a GitHub Runner orchestration and management service.
It provides a single (or scalable) self-hosted runner capable of processing
jobs for multiple repositories for personal accounts.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
}

func init() {}
