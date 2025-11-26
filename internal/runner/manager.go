package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	gh "github.com/skevetter/bladerunner/internal/github"
)

type Config struct {
	Owner         string
	Repo          string
	RunnerName    string
	RunnerGroup   string
	Labels        []string
	WorkDir       string
	Token         string
	RunnerVersion string
}

type Manager struct {
	client    *gh.Client
	installer *Installer
	config    *Config
	workDir   string
}

func NewManager(client *gh.Client, config *Config) *Manager {
	if config.WorkDir == "" {
		config.WorkDir = "/tmp/github-runner"
	}

	if config.RunnerVersion == "" {
		config.RunnerVersion = "latest"
	}

	return &Manager{
		client:    client,
		installer: NewInstaller(client),
		config:    config,
		workDir:   config.WorkDir,
	}
}

func (m *Manager) Register(ctx context.Context) error {
	logrus.Infof("Registering runner %s for %s/%s", m.config.RunnerName, m.config.Owner, m.config.Repo)

	// Check if runner files exist (should be baked in)
	configPath := filepath.Join(m.workDir, "config.sh")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logrus.Info("Runner files not found, attempting install...")
		if err := m.installer.Install(ctx, m.config.Owner, m.config.Repo, m.workDir); err != nil {
			return fmt.Errorf("failed to install runner: %w", err)
		}
	}

	// Check if already configured
	if _, err := os.Stat(filepath.Join(m.workDir, ".runner")); err == nil {
		logrus.Info("Runner already configured")
		return nil
	}

	regToken, err := m.client.CreateRegistrationToken(m.config.Owner, m.config.Repo)

	if err != nil {
		return fmt.Errorf("failed to create registration token: %w", err)
	}

	args := []string{
		"--url", fmt.Sprintf("https://github.com/%s/%s", m.config.Owner, m.config.Repo),
		"--token", regToken.GetToken(),
		"--name", m.config.RunnerName,
		"--work", "_work",
		"--unattended",
		"--replace",
	}

	if m.config.RunnerGroup != "" {
		args = append(args, "--runnergroup", m.config.RunnerGroup)
	}

	if len(m.config.Labels) > 0 {
		args = append(args, "--labels", strings.Join(m.config.Labels, ","))
	}

	cmd := exec.CommandContext(ctx, configPath, args...)
	cmd.Dir = m.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure runner: %w", err)
	}

	logrus.Info("Runner registered successfully")
	return nil
}

func (m *Manager) Unregister(ctx context.Context) error {
	logrus.Infof("Unregistering runner %s", m.config.RunnerName)

	configPath := filepath.Join(m.workDir, "config.sh")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logrus.Info("Runner not installed, skipping unregister")
		return nil
	}

	// Check if configured
	if _, err := os.Stat(filepath.Join(m.workDir, ".runner")); os.IsNotExist(err) {
		logrus.Info("Runner not configured, skipping unregister")
		return nil
	}

	cmd := exec.CommandContext(ctx, configPath, "remove", "--token", m.config.Token)
	cmd.Dir = m.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unregister runner: %w", err)
	}

	logrus.Info("Runner unregistered successfully")
	return nil
}

func (m *Manager) Run(ctx context.Context) error {
	logrus.Info("Starting runner...")

	runPath := filepath.Join(m.workDir, "run.sh")
	if _, err := os.Stat(runPath); os.IsNotExist(err) {
		return fmt.Errorf("runner executable not found at %s", runPath)
	}

	cmd := exec.CommandContext(ctx, runPath)
	cmd.Dir = m.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start runner: %w", err)
	}

	// Wait for context cancellation or process exit
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		logrus.Info("Context cancelled, stopping runner...")
		// Send SIGINT to allow graceful shutdown
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			logrus.Warnf("Failed to send SIGINT: %v", err)
			cmd.Process.Kill()
		}

		// Wait for exit
		select {
		case <-done:
		case <-ctx.Done(): // If it takes too long
			cmd.Process.Kill()
		}
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("runner exited with error: %w", err)
		}
		return nil
	}
}
