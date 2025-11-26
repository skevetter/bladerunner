package github

import (
	"github.com/google/go-github/v66/github"
)

func (c *Client) ListRepositories() ([]*github.Repository, error) {
	opt := &github.RepositoryListByAuthenticatedUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := c.client.Repositories.ListByAuthenticatedUser(c.ctx, opt)
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func (c *Client) GetRepository(owner, repo string) (*github.Repository, error) {
	repository, _, err := c.client.Repositories.Get(c.ctx, owner, repo)
	return repository, err
}
