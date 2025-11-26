package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skevetter/bladerunner/internal/docker"
	gh "github.com/skevetter/bladerunner/internal/github"
)

type PoolConfig struct {
	MinRunners    int
	MaxRunners    int
	Image         string
	Owner         string
	Repo          string
	Token         string
	ScaleInterval time.Duration
}

type RunnerInstance struct {
	Name        string
	ContainerID string
	CreatedAt   time.Time
}

type Pool struct {
	client       *gh.Client
	dockerClient *docker.Client
	config       *PoolConfig
	runners      map[string]*RunnerInstance
}

func NewPool(client *gh.Client, dockerClient *docker.Client, config *PoolConfig) *Pool {
	if config.ScaleInterval == 0 {
		config.ScaleInterval = 30 * time.Second
	}
	return &Pool{
		client:       client,
		dockerClient: dockerClient,
		config:       config,
		runners:      make(map[string]*RunnerInstance),
	}
}

func (p *Pool) Start(ctx context.Context) {
	ticker := time.NewTicker(p.config.ScaleInterval)
	defer ticker.Stop()

	logrus.Infof("Starting runner pool for %s/%s (Min: %d, Max: %d)", p.config.Owner, p.config.Repo, p.config.MinRunners, p.config.MaxRunners)

	for {
		select {
		case <-ctx.Done():
			p.Stop(context.Background())
			return
		case <-ticker.C:
			if err := p.reconcile(ctx); err != nil {
				logrus.WithError(err).Error("Failed to reconcile runner pool")
			}
		}
	}
}

func (p *Pool) Stop(ctx context.Context) {
	logrus.Info("Stopping runner pool...")
	for name, runner := range p.runners {
		if err := p.dockerClient.StopContainer(ctx, runner.ContainerID); err != nil {
			logrus.Warnf("Failed to stop container %s (%s): %v", name, runner.ContainerID, err)
		}
	}
}

func (p *Pool) reconcile(ctx context.Context) error {
	// 1. Get current queued jobs
	workflowRuns, err := p.client.ListRepositoryWorkflowRuns(p.config.Owner, p.config.Repo)
	if err != nil {
		return fmt.Errorf("failed to list workflow runs: %w", err)
	}

	queuedJobs := 0
	for _, run := range workflowRuns.WorkflowRuns {
		if run.GetStatus() == "queued" || run.GetStatus() == "in_progress" {
			queuedJobs++
		}
	}

	logrus.Debugf("Current queued/in_progress jobs: %d", queuedJobs)

	// 2. Determine desired runner count
	// Simple logic: 1 runner per job, clamped by Min/Max
	desiredRunners := min(max(queuedJobs, p.config.MinRunners), p.config.MaxRunners)

	currentRunners := len(p.runners)
	logrus.Infof("Reconciling: Current=%d, Desired=%d", currentRunners, desiredRunners)

	if currentRunners < desiredRunners {
		// Scale up
		toAdd := desiredRunners - currentRunners
		for range toAdd {
			if err := p.addRunner(ctx); err != nil {
				logrus.WithError(err).Error("Failed to add runner")
			}
		}
	} else if currentRunners > desiredRunners {
		// Scale down
		// Note: Ideally we should only remove idle runners.
		// For now, we'll just remove random ones (which might kill active jobs - risky!)
		// TODO: Check runner status before removing
		toRemove := currentRunners - desiredRunners
		for range toRemove {
			p.removeRunner(ctx)
		}
	}

	return nil
}

func (p *Pool) addRunner(ctx context.Context) error {
	name := fmt.Sprintf("runner-%s-%d", p.config.Repo, time.Now().UnixNano())

	env := []string{
		"GITHUB_TOKEN=" + p.config.Token,
		"RUNNER_OWNER=" + p.config.Owner,
		"RUNNER_REPO=" + p.config.Repo,
		"RUNNER_NAME=" + name,
		"RUNNER_LABELS=docker,pool",
	}

	id, err := p.dockerClient.StartRunnerContainer(ctx, p.config.Image, env, name)
	if err != nil {
		return err
	}

	p.runners[name] = &RunnerInstance{
		Name:        name,
		ContainerID: id,
		CreatedAt:   time.Now(),
	}
	logrus.Infof("Started runner %s (%s)", name, id)
	return nil
}

func (p *Pool) removeRunner(ctx context.Context) {
	// Pick a random runner to remove
	for name, runner := range p.runners {
		if err := p.dockerClient.StopContainer(ctx, runner.ContainerID); err != nil {
			logrus.WithError(err).Errorf("Failed to stop runner %s", name)
		} else {
			logrus.Infof("Stopped runner %s", name)
			delete(p.runners, name)
		}
		return // Remove one at a time
	}
}
