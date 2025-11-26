package github_test

import (
	"context"
	"os"
	"testing"

	gh "github.com/skevetter/bladerunner/internal/github"
)

func getToken(t *testing.T) string {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN not set, skipping integration test")
	}
	return token
}

func TestListRepositories(t *testing.T) {
	token := getToken(t)
	ctx := context.Background()
	client := gh.NewClient(ctx, token)

	repos, err := client.ListRepositories()
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}

	if len(repos) == 0 {
		t.Log("No repositories found (user may have no repos)")
	} else {
		t.Logf("Found %d repositories", len(repos))
		if len(repos) > 0 {
			t.Logf("Example repo: %s", repos[0].GetFullName())
		}
	}
}

func TestGetRepository(t *testing.T) {
	token := getToken(t)
	ctx := context.Background()
	client := gh.NewClient(ctx, token)

	repos, err := client.ListRepositories()
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}

	if len(repos) == 0 {
		t.Skip("No repositories to test GetRepository with")
	}

	testRepo := repos[0]
	owner := testRepo.GetOwner().GetLogin()
	name := testRepo.GetName()

	repo, err := client.GetRepository(owner, name)
	if err != nil {
		t.Fatalf("GetRepository failed: %v", err)
	}

	if repo.GetName() != name {
		t.Errorf("Expected repo name %s, got %s", name, repo.GetName())
	}

	t.Logf("Successfully retrieved repository: %s", repo.GetFullName())
}

func TestListWorkflows(t *testing.T) {
	token := getToken(t)
	ctx := context.Background()
	client := gh.NewClient(ctx, token)

	repos, err := client.ListRepositories()
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}

	if len(repos) == 0 {
		t.Skip("No repositories to test workflows with")
	}

	for _, repo := range repos {
		owner := repo.GetOwner().GetLogin()
		name := repo.GetName()

		workflows, err := client.ListWorkflows(owner, name)
		if err != nil {
			t.Logf("Failed to list workflows for %s: %v", repo.GetFullName(), err)
			continue
		}

		if workflows.GetTotalCount() > 0 {
			t.Logf("Found %d workflows in %s", workflows.GetTotalCount(), repo.GetFullName())
			for _, wf := range workflows.Workflows {
				t.Logf("  - Workflow: %s (ID: %d)", wf.GetName(), wf.GetID())
			}
			return // Success
		}
	}

	t.Log("No workflows found in any repository")
}

func TestListWorkflowRuns(t *testing.T) {
	token := getToken(t)
	ctx := context.Background()
	client := gh.NewClient(ctx, token)

	repos, err := client.ListRepositories()
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}

	if len(repos) == 0 {
		t.Skip("No repositories to test workflow runs with")
	}

	for _, repo := range repos {
		owner := repo.GetOwner().GetLogin()
		name := repo.GetName()

		workflows, err := client.ListWorkflows(owner, name)
		if err != nil {
			continue
		}

		if workflows.GetTotalCount() > 0 && len(workflows.Workflows) > 0 {
			workflowID := workflows.Workflows[0].GetID()
			runs, err := client.ListWorkflowRuns(owner, name, workflowID)
			if err != nil {
				t.Fatalf("ListWorkflowRuns failed: %v", err)
			}

			t.Logf("Found %d runs for workflow %s in %s",
				runs.GetTotalCount(),
				workflows.Workflows[0].GetName(),
				repo.GetFullName())
			return // Success
		}
	}

	t.Log("No workflow runs found in any repository")
}
