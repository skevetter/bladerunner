package github

import (
	"github.com/google/go-github/v66/github"
)

func (c *Client) ListWorkflows(owner, repo string) (*github.Workflows, error) {
	workflows, _, err := c.client.Actions.ListWorkflows(c.ctx, owner, repo, nil)
	return workflows, err
}

func (c *Client) GetWorkflow(owner, repo string, workflowID int64) (*github.Workflow, error) {
	workflow, _, err := c.client.Actions.GetWorkflowByID(c.ctx, owner, repo, workflowID)
	return workflow, err
}

func (c *Client) GetWorkflowByFileName(owner, repo, fileName string) (*github.Workflow, error) {
	workflow, _, err := c.client.Actions.GetWorkflowByFileName(c.ctx, owner, repo, fileName)
	return workflow, err
}

func (c *Client) ListWorkflowRuns(owner, repo string, workflowID int64) (*github.WorkflowRuns, error) {
	opts := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	runs, _, err := c.client.Actions.ListWorkflowRunsByID(c.ctx, owner, repo, workflowID, opts)
	return runs, err
}

func (c *Client) ListRepositoryWorkflowRuns(owner, repo string) (*github.WorkflowRuns, error) {
	opts := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	runs, _, err := c.client.Actions.ListWorkflowRunsByFileName(c.ctx, owner, repo, "", opts)
	return runs, err
}
