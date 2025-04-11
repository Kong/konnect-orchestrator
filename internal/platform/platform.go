package platform

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/Kong/konnect-orchestrator/internal/git"
	"github.com/Kong/konnect-orchestrator/internal/git/github"
	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/util"
	giturl "github.com/kubescape/go-git-url"
)

func Init(platformGitCfg *manifest.GitConfig, resourceFiles embed.FS) error {
	// TODO: Initialize the platform repository with the following steps:
	// Pre-requisites:
	//		The repository must exist
	//		We prompt the user to provide the following:
	//			git URL
	//			git author (name and email)
	//			GitHub token

	// 1. Clone the repository locally
	gitURL, err := giturl.NewGitURL(*platformGitCfg.Remote)
	if err != nil {
		return fmt.Errorf("failed to parse Git URL: %w", err)
	}

	platformRepoDir, err := git.Clone(*platformGitCfg)
	if err != nil {
		return fmt.Errorf("failed to clone platform repository: %w", err)
	}

	branchName := "konnect-orchestrator-init"
	err = git.CheckoutBranch(platformRepoDir, branchName, *platformGitCfg)
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	// 2. Create the konnect/ directory structure
	konnectPath := platformRepoDir + "/konnect"
	if err := os.MkdirAll(konnectPath, 0o755); err != nil {
		return fmt.Errorf("failed to create konnect directory: %w", err)
	}

	// 3. Create the .github/workflow diretory structure
	if err := os.MkdirAll(".github/workflows", 0o755); err != nil {
		return fmt.Errorf("failed to create .github/workflows directory: %w", err)
	}

	// 4. Create the default konnect/platform.yaml file
	//    a. Write the provided auth token as an env var to the github auth section
	// 5. Create the default konnect/teams.yaml file
	// 6. Create the default konnect/organizations.yaml file (including environments)
	// 7. Create the Workflow files in .github/workflow
	err = util.CopyEmbeddedFilesRecursive(resourceFiles, "resources/platform", platformRepoDir)
	if err != nil {
		return fmt.Errorf("failed to copy default konnect/ files: %w", err)
	}

	// 8. File PR
	// Detect changes to the repository
	isClean, err := git.IsClean(platformRepoDir)
	if err != nil {
		return fmt.Errorf("failed to check if platform repository is clean: %w", err)
	}
	if !isClean {
		err = git.Commit(platformRepoDir, "Konnect Orchestrator initializing the platform repository", *platformGitCfg.Author)
		if err != nil {
			return fmt.Errorf("failed to commit changes: %w", err)
		}
		err = git.Push(platformRepoDir, *platformGitCfg)
		if err != nil {
			return fmt.Errorf("failed to push changes: %w", err)
		}

		_, err = github.CreateOrUpdatePullRequest(
			context.Background(),
			gitURL.GetOwnerName(),
			gitURL.GetRepoName(),
			branchName,
			"[Konnect Orchestrator] - Init Platform",
			`The Konnect Orchestrator 'init' function was executed and filed this PR to initialize the Platform repository, 
			including GitHub Actions and default configuration files.`,
			*platformGitCfg.GitHub)
		if err != nil {
			return fmt.Errorf("failed to create or update pull request: %w", err)
		}
	}

	// 9. Write the provided GitHub auth token to the repository secrets API
	err = github.CreateRepoActionSecret(
		context.Background(),
		platformGitCfg.GitHub,
		gitURL.GetOwnerName(),
		gitURL.GetRepoName(),
		"KONNECT_ORCHESTRATOR_GITHUB_TOKEN",
		platformGitCfg.GitHub.Token)
	if err != nil {
		return fmt.Errorf("failed to create GitHub secret: %w", err)
	}

	return nil
}
