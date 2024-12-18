package github

import (
	"context"
	"fmt"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/util"
	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

func CreateOrUpdatePullRequest(ctx context.Context,
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

	// First, check if there's an existing PR for this branch
	existingPRs, _, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		Head:  branch,
		Base:  "main",
		State: "open",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	if len(existingPRs) > 0 {
		// Update existing PR
		pr := existingPRs[0]
		if pr.Title != nil && *pr.Title != title || pr.Body != nil && *pr.Body != body {
			pr, _, err = client.PullRequests.Edit(ctx, owner, repo, pr.GetNumber(), &github.PullRequest{
				Title: github.String(title),
				Body:  github.String(body),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update pull request: %w", err)
			}
		}
		return pr, nil
	}

	// Create new PR if none exists
	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(branch),
		Base:                github.String("main"),
		Body:                github.String(body),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return pr, nil
}
