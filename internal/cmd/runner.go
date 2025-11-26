package cmd

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/skevetter/bladerunner/internal/docker"
	gh "github.com/skevetter/bladerunner/internal/github"
	"github.com/skevetter/bladerunner/internal/runner"
	"github.com/spf13/cobra"
)

var (
	runnerOwner   string
	runnerRepo    string
	runnerName    string
	runnerGroup   string
	runnerLabels  []string
	runnerWorkDir string
)

var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "Manage GitHub Actions runners",
	Long:  `Manage self-hosted GitHub Actions runners.`,
}

var runnerRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new runner",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		token := githubToken
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		if token == "" {
			logrus.Fatal("GitHub token is required")
		}

		client := gh.NewClient(ctx, token)

		config := &runner.Config{
			Owner:       runnerOwner,
			Repo:        runnerRepo,
			RunnerName:  runnerName,
			RunnerGroup: runnerGroup,
			Labels:      runnerLabels,
			WorkDir:     runnerWorkDir,
			Token:       token,
		}

		mgr := runner.NewManager(client, config)

		if err := mgr.Register(ctx); err != nil {
			logrus.WithError(err).Fatal("Failed to register runner")
		}

		logrus.Info("Runner registered successfully")
	},
}

var runnerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the registered runner",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		token := githubToken
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		if token == "" {
			logrus.Fatal("GitHub token is required")
		}

		client := gh.NewClient(ctx, token)

		config := &runner.Config{
			Owner:       runnerOwner,
			Repo:        runnerRepo,
			RunnerName:  runnerName,
			RunnerGroup: runnerGroup,
			Labels:      runnerLabels,
			WorkDir:     runnerWorkDir,
			Token:       token,
		}

		mgr := runner.NewManager(client, config)

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			logrus.Info("Received shutdown signal, stopping runner...")
			cancel()
		}()

		if err := mgr.Run(ctx); err != nil {
			logrus.WithError(err).Fatal("Runner failed")
		}
	},
}

var runnerUnregisterCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Unregister a runner",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		token := githubToken
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		if token == "" {
			logrus.Fatal("GitHub token is required")
		}

		client := gh.NewClient(ctx, token)

		config := &runner.Config{
			Owner:      runnerOwner,
			Repo:       runnerRepo,
			RunnerName: runnerName,
			WorkDir:    runnerWorkDir,
			Token:      token,
		}

		mgr := runner.NewManager(client, config)

		if err := mgr.Unregister(ctx); err != nil {
			logrus.WithError(err).Fatal("Failed to unregister runner")
		}

		logrus.Info("Runner unregistered successfully")
	},
}

var runnerInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Download and install runner binary",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		token := githubToken
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		if token == "" {
			logrus.Fatal("GitHub token is required")
		}

		client := gh.NewClient(ctx, token)
		installer := runner.NewInstaller(client)

		if err := installer.Install(ctx, runnerOwner, runnerRepo, runnerWorkDir); err != nil {
			logrus.WithError(err).Fatal("Failed to install runner")
		}
	},
}

var runnerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start runner with full lifecycle (Register -> Run -> Unregister)",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		token := githubToken
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		if token == "" {
			// Try ACCESS_TOKEN (common in other runner images)
			token = os.Getenv("ACCESS_TOKEN")
		}

		if token == "" {
			logrus.Fatal("GitHub token is required (GITHUB_TOKEN or ACCESS_TOKEN)")
		}

		// Support env vars for other flags if not set
		if runnerOwner == "" {
			runnerOwner = os.Getenv("RUNNER_OWNER")
		}
		if runnerRepo == "" {
			runnerRepo = os.Getenv("RUNNER_REPO")
		}
		if runnerName == "" {
			runnerName = os.Getenv("RUNNER_NAME")
		}
		if runnerGroup == "" {
			runnerGroup = os.Getenv("RUNNER_GROUP")
		}
		if runnerWorkDir == "" {
			runnerWorkDir = os.Getenv("RUNNER_WORKDIR")
		}

		// Handle REPO_URL parsing if provided (e.g. https://github.com/owner/repo)
		if repoURL := os.Getenv("REPO_URL"); repoURL != "" && (runnerOwner == "" || runnerRepo == "") {
			// Simple parsing logic
			parts := strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")
			if len(parts) >= 2 {
				runnerOwner = parts[0]
				runnerRepo = parts[1]
			}
		}

		if runnerOwner == "" || runnerRepo == "" || runnerName == "" {
			logrus.Fatal("Owner, Repo, and Name are required")
		}

		client := gh.NewClient(ctx, token)

		config := &runner.Config{
			Owner:       runnerOwner,
			Repo:        runnerRepo,
			RunnerName:  runnerName,
			RunnerGroup: runnerGroup,
			Labels:      runnerLabels,
			WorkDir:     runnerWorkDir,
			Token:       token,
		}

		mgr := runner.NewManager(client, config)

		// 1. Register
		if err := mgr.Register(ctx); err != nil {
			logrus.WithError(err).Fatal("Failed to register runner")
		}

		// 2. Setup signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			s := <-sigChan
			logrus.Infof("Received signal %v, stopping runner...", s)
			cancel()
		}()

		// 3. Run
		if err := mgr.Run(ctx); err != nil {
			logrus.WithError(err).Error("Runner execution failed")
		}

		// 4. Unregister (on exit)
		// Create new context as the original might be cancelled
		unregCtx := context.Background()
		if err := mgr.Unregister(unregCtx); err != nil {
			logrus.WithError(err).Error("Failed to unregister runner")
		}
	},
}

var runnerPoolCmd = &cobra.Command{
	Use:   "pool",
	Short: "Start runner pool manager",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		token := githubToken
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}
		if token == "" {
			logrus.Fatal("GitHub token is required")
		}

		if runnerOwner == "" || runnerRepo == "" {
			logrus.Fatal("Owner and Repo are required")
		}

		client := gh.NewClient(ctx, token)
		dockerClient, err := docker.NewClient()
		if err != nil {
			logrus.WithError(err).Fatal("Failed to create Docker client")
		}

		config := &runner.PoolConfig{
			Owner:      runnerOwner,
			Repo:       runnerRepo,
			Token:      token,
			MinRunners: runnerMin,
			MaxRunners: runnerMax,
			Image:      runnerImage,
		}

		pool := runner.NewPool(client, dockerClient, config)

		// Setup signal handling
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			logrus.Info("Received shutdown signal, stopping pool...")
			cancel()
		}()

		pool.Start(ctx)
	},
}

var (
	runnerMin   int
	runnerMax   int
	runnerImage string
)

func init() {
	rootCmd.AddCommand(runnerCmd)
	runnerCmd.AddCommand(runnerRegisterCmd)
	runnerCmd.AddCommand(runnerRunCmd)
	runnerCmd.AddCommand(runnerUnregisterCmd)
	runnerCmd.AddCommand(runnerInstallCmd)
	runnerCmd.AddCommand(runnerStartCmd)
	runnerCmd.AddCommand(runnerPoolCmd)

	flags := []*cobra.Command{runnerRegisterCmd, runnerRunCmd, runnerUnregisterCmd, runnerInstallCmd, runnerStartCmd, runnerPoolCmd}
	for _, cmd := range flags {
		cmd.Flags().StringVarP(&githubToken, "token", "t", "", "GitHub personal access token")
		cmd.Flags().StringVar(&runnerOwner, "owner", "", "Repository owner")
		cmd.Flags().StringVar(&runnerRepo, "repo", "", "Repository name")
		cmd.Flags().StringVar(&runnerWorkDir, "workdir", "/tmp/github-runner", "Runner work directory")
	}

	// Register specific flags
	runnerRegisterCmd.Flags().StringVar(&runnerName, "name", "", "Runner name")
	runnerRegisterCmd.Flags().StringVar(&runnerGroup, "group", "Default", "Runner group")
	runnerRegisterCmd.Flags().StringSliceVar(&runnerLabels, "labels", []string{}, "Runner labels")

	runnerRunCmd.Flags().StringVar(&runnerName, "name", "", "Runner name")
	runnerRunCmd.Flags().StringVar(&runnerGroup, "group", "Default", "Runner group")
	runnerRunCmd.Flags().StringSliceVar(&runnerLabels, "labels", []string{}, "Runner labels")

	runnerStartCmd.Flags().StringVar(&runnerName, "name", "", "Runner name")
	runnerStartCmd.Flags().StringVar(&runnerGroup, "group", "Default", "Runner group")
	runnerStartCmd.Flags().StringSliceVar(&runnerLabels, "labels", []string{}, "Runner labels")

	runnerUnregisterCmd.Flags().StringVar(&runnerName, "name", "", "Runner name")

	runnerPoolCmd.Flags().IntVar(&runnerMin, "min", 1, "Minimum number of runners")
	runnerPoolCmd.Flags().IntVar(&runnerMax, "max", 5, "Maximum number of runners")
	runnerPoolCmd.Flags().StringVar(&runnerImage, "image", "bladerunner-runner:latest", "Runner Docker image")

	// Required flags
	for _, cmd := range []*cobra.Command{runnerRegisterCmd, runnerRunCmd, runnerUnregisterCmd, runnerInstallCmd, runnerPoolCmd} {
		cmd.MarkFlagRequired("owner")
		cmd.MarkFlagRequired("repo")
	}
	runnerRegisterCmd.MarkFlagRequired("name")
	runnerRunCmd.MarkFlagRequired("name")
	runnerUnregisterCmd.MarkFlagRequired("name")
	// Start command flags are optional because they can be set via env vars
}
