package github

import (
	"context"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

func NewClient(ctx context.Context, token string) *Client {
	var httpClient *github.Client

	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		httpClient = github.NewClient(tc)
	} else {
		httpClient = github.NewClient(nil)
	}

	return &Client{
		client: httpClient,
		ctx:    ctx,
	}
}