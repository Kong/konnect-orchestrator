package github

import (
	"context"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/util"
	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

func CreatePullRequest(ctx context.Context,
	owner, repo, branch, title, body string,
	githubConfig manifest.GitHubConfig) (*github.PullRequest, error) {
	// Create GitHub client with token
	token, err := util.ResolveSecretValue(githubConfig.Token)
	if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Create the pull request
	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(branch),
		Base:                github.String("main"), // or your default branch
		Body:                github.String(body),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return nil, err
	}

	return pr, nil
}
