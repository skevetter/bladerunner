package github

import (
	"github.com/google/go-github/v66/github"
)

// ListRunners lists all self-hosted runners for a repository
func (c *Client) ListRunners(owner, repo string) (*github.Runners, error) {
	opts := &github.ListRunnersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	runners, _, err := c.client.Actions.ListRunners(c.ctx, owner, repo, opts)
	return runners, err
}

// GetRunner gets a specific runner by ID
func (c *Client) GetRunner(owner, repo string, runnerID int64) (*github.Runner, error) {
	runner, _, err := c.client.Actions.GetRunner(c.ctx, owner, repo, runnerID)
	return runner, err
}

// CreateRegistrationToken creates a registration token for adding a self-hosted runner
func (c *Client) CreateRegistrationToken(owner, repo string) (*github.RegistrationToken, error) {
	token, _, err := c.client.Actions.CreateRegistrationToken(c.ctx, owner, repo)
	return token, err
}

// RemoveRunner removes a self-hosted runner from a repository
func (c *Client) RemoveRunner(owner, repo string, runnerID int64) error {
	_, err := c.client.Actions.RemoveRunner(c.ctx, owner, repo, runnerID)
	return err
}

// ListRunnerApplicationDownloads lists runner application downloads
func (c *Client) ListRunnerApplicationDownloads(owner, repo string) ([]*github.RunnerApplicationDownload, error) {
	downloads, _, err := c.client.Actions.ListRunnerApplicationDownloads(c.ctx, owner, repo)
	return downloads, err
}
