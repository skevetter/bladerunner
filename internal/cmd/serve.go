package cmd

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Bladerunner server",
	Long:  `Start the Bladerunner server with PocketBase backend.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("Starting Bladerunner server...")

		app := pocketbase.New()

		app.OnServe().BindFunc(func(e *core.ServeEvent) error {
			logrus.Info("PocketBase server initialized")
			// Placeholder for custom initialization logic
			return nil
		})

		if err := app.Start(); err != nil {
			logrus.WithError(err).Fatal("Failed to start server")
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
