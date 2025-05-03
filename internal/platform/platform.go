package platform

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Kong/konnect-orchestrator/internal/git"
	"github.com/Kong/konnect-orchestrator/internal/git/github"
	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/util"
	"github.com/go-git/go-git/v5/plumbing/transport"
	giturl "github.com/kubescape/go-git-url"
	"gopkg.in/yaml.v3"
)

func Init(platformGitCfg manifest.GitConfig, resourceFiles embed.FS, statusCh chan<- string, createNewRepo bool) error {
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

	platformRepoDir, err := git.Clone(platformGitCfg)
	if errors.Is(err, transport.ErrRepositoryNotFound) {
		if createNewRepo {
			if err = github.CreateRepo(context.Background(), gitURL.GetOwnerName(), gitURL.GetRepoName(), *platformGitCfg.GitHub); err != nil {
				return fmt.Errorf("failed to create platform repository: %w", err)
			}
			platformRepoDir, err = git.Clone(platformGitCfg)
		} else {
			return fmt.Errorf("failed to create platform repository: %w [Hint: re-run with --create to auto-create missing repo]", err)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to clone platform repository: %w", err)
	}
	statusCh <- fmt.Sprintf("✔ Cloned %s repository locally\n", gitURL.GetRepoName())

	branchName := "konnect-orchestrator-init"
	err = git.CheckoutBranch(platformRepoDir, branchName, platformGitCfg)
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
	statusCh <- "✔ Added default files to konnect directory\n"

	err = git.Add(platformRepoDir, ".")
	if err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	// 8. File PR
	// Detect changes to the repository
	isClean, err := git.IsClean(platformRepoDir)
	if err != nil {
		return fmt.Errorf("failed to check if platform repository is clean: %w", err)
	}
	var prURL string
	if !isClean {
		err = git.Commit(platformRepoDir, "Konnect Orchestrator initializing the platform repository", *platformGitCfg.Author)
		if err != nil {
			return fmt.Errorf("failed to commit changes: %w", err)
		}
		err = git.Push(platformRepoDir, platformGitCfg)
		if err != nil {
			return fmt.Errorf("failed to push changes: %w", err)
		}
		statusCh <- fmt.Sprintf("✔ Pushed changes to the %s repository %s branch\n", gitURL.GetRepoName(), branchName)

		pr, err := github.CreateOrUpdatePullRequest(
			context.Background(),
			gitURL.GetOwnerName(),
			gitURL.GetRepoName(),
			branchName,
			"[Konnect Orchestrator] - Init Platform",
			`The Konnect Orchestrator 'init' function was executed and filed this PR to initialize the Platform repository, 
			including GitHub Actions and default configuration files.
			
			Review and merge this PR to complete the Platform repository initialization (no actions will be performed except for the creation of the repository and github actions)`,
			*platformGitCfg.GitHub,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create or update pull request: %w", err)
		}
		prURL = *pr.HTMLURL
	}
	if prURL == "" {
		prURL = *platformGitCfg.Remote + "/pulls"
	}

	// 9. Write the provided GitHub auth token to the repository secrets API
	secretName := "KONNECT_ORCHESTRATOR_GITHUB_TOKEN" //nolint:gosec
	err = github.CreateRepoActionSecret(
		context.Background(),
		platformGitCfg.GitHub,
		gitURL.GetOwnerName(),
		gitURL.GetRepoName(),
		secretName,
		platformGitCfg.GitHub.Token)
	if err != nil {
		return fmt.Errorf("failed to create GitHub secret: %w", err)
	}
	statusCh <- fmt.Sprintf("✔ Added %s to %s repository secrets\n", secretName, gitURL.GetRepoName())

	statusCh <- fmt.Sprintf("✔ PR Filed: %s\n", prURL)
	statusCh <- "\tReview and Merge to complete platform repository initialization.\n\n"
	statusCh <- "Next add your Konnect Organization to the platform with\n"
	statusCh <- "\tkoctl add organization\n\n"
	statusCh <- "See the FAQ page for further questions on the Konnect Reference Platform:\n"
	statusCh <- "\thttps://deploy-preview-783--kongdeveloper.netlify.app/konnect-reference-platform/faq\n"

	return nil
}

func AddOrganization(
	platformGitCfg *manifest.GitConfig,
	orgName,
	konnectToken string,
	statusCh chan<- string,
) error {
	// TODO: Add an organization to the konnect/organizations.yaml file
	// Pre-requisites:
	// 		We need the platform gith configuration (url and token), either the platform file or args
	//		The repository must exist
	//		We prompt the user to provide the following:
	//			Organization name (we slugify for a YAML key)
	//			Konnect Token:
	if err := github.ValidateOrgName(orgName); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}

	defer close(statusCh)

	// 1. Clone the repository locally
	gitURL, err := giturl.NewGitURL(*platformGitCfg.Remote)
	if err != nil {
		return fmt.Errorf("failed to parse Git URL: %w", err)
	}

	platformRepoDir, err := git.Clone(*platformGitCfg)
	if err != nil {
		return fmt.Errorf("failed to clone platform repository: %w", err)
	}
	statusCh <- fmt.Sprintf("✔ Cloned %s repository locally\n", gitURL.GetRepoName())

	branchName := fmt.Sprintf("konnect-orchestrator-add-org-%s", orgName)
	err = git.CheckoutBranch(platformRepoDir, branchName, *platformGitCfg)
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	konnectPath := platformRepoDir + "/konnect"
	organizationsFilePath := konnectPath + "/organizations.yaml"
	konnectTokenEnvVarName := strings.ToUpper(orgName) + "_KONNECT_TOKEN"

	// 2. Load the organizations file into a manifest struct
	var man manifest.Orchestrator
	if err := util.ReadConfigFile(organizationsFilePath, &man); err != nil {
		return fmt.Errorf("failed to read organizations configuration: %w", err)
	}

	if man.Organizations == nil {
		man.Organizations = make(map[string]*manifest.Organization)
	}

	// 2a. Add the new organization to the organizations file
	man.Organizations[orgName] = &manifest.Organization{
		AccessToken: manifest.Secret{
			Value: konnectTokenEnvVarName,
			Type:  "env",
		},
		Environments: map[string]*manifest.Environment{
			"dev": {Type: "DEV", Region: "us"},
			"prd": {Type: "PROD", Region: "us"},
		},
	}

	// 2b. Write the updated organizations file back to the repository
	data, err := yaml.Marshal(&man)
	if err != nil {
		return fmt.Errorf("failed to marshal organizations configuration: %w", err)
	}

	file, err := os.Create(organizationsFilePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close() // Ensure that the file will be closed at the end

	// Write the YAML data to the file
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	statusCh <- fmt.Sprintf("✔ Added %s to organizations.yaml\n", gitURL.GetRepoName())

	// 3. Modify the workflow file to include
	//	    env:
	//			<ORGNAME>_KONNECT_TOKEN: {{ .secrets.<ORGNAME>_KONNECT_TOKEN }}
	// in the koctl apply step
	workflowFilePath := platformRepoDir + "/.github/workflows/konnect-koctl-apply.yaml"

	fd, err := os.ReadFile(workflowFilePath)
	if err != nil {
		return fmt.Errorf("failed to read workflow file: %w", err)
	}
	var root yaml.Node
	if err := yaml.Unmarshal(fd, &root); err != nil {
		return fmt.Errorf("failed to unmarshal workflow file: %w", err)
	}

	// Find the jobs > build > koctlSteps
	koctlSteps := findMapValuePath(root.Content[0],
		"jobs", "koctl-apply", "steps")
	if koctlSteps == nil || koctlSteps.Kind != yaml.SequenceNode {
		return fmt.Errorf("failed to find steps in workflow file")
	}

	applyStep := findStepByID(koctlSteps, "koctl-apply")
	if applyStep == nil {
		return fmt.Errorf("failed to find koctl-apply step in workflow file")
	}

	koctlApplyStepEnv := findMapValue(applyStep, "env")
	if koctlApplyStepEnv == nil {
		koctlApplyStepEnv = &yaml.Node{
			Kind:    yaml.MappingNode,
			Content: []*yaml.Node{},
		}
		applyStep.Content = append(applyStep.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: "env"}, koctlApplyStepEnv)
	}

	setEnvVariable(koctlApplyStepEnv, konnectTokenEnvVarName, fmt.Sprintf("${{secrets.%s}}", konnectTokenEnvVarName))

	out, err := yaml.Marshal(&root)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow file: %w", err)
	}
	err = os.WriteFile(workflowFilePath, out, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}
	statusCh <- fmt.Sprintf("✔ Added %s to koctl apply workflow\n", konnectTokenEnvVarName)

	// Detect changes to the repository
	isClean, err := git.IsClean(platformRepoDir)
	if err != nil {
		return fmt.Errorf("failed to check if platform repository is clean: %w", err)
	}
	var prURL string
	if !isClean {
		if err = git.Add(platformRepoDir, "."); err != nil {
			return fmt.Errorf("failed to add files to git: %w", err)
		}
		err = git.Commit(platformRepoDir, "Konnect Orchestrator initializing the platform repository", *platformGitCfg.Author)
		if err != nil {
			return fmt.Errorf("failed to commit changes: %w", err)
		}
		err = git.Push(platformRepoDir, *platformGitCfg)
		if err != nil {
			return fmt.Errorf("failed to push changes: %w", err)
		}

		pr, err := github.CreateOrUpdatePullRequest(
			context.Background(),
			gitURL.GetOwnerName(),
			gitURL.GetRepoName(),
			branchName,
			fmt.Sprintf("[Konnect Orchestrator] - Add %s Organization", orgName),
			`The Konnect Orchestrator Add Organization function was executed and 
			 filed this PR to add a new organization to the Platform repository`,
			*platformGitCfg.GitHub,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create or update pull request: %w", err)
		}
		prURL = *pr.HTMLURL
	}
	if prURL == "" {
		prURL = *platformGitCfg.Remote + "/pulls"
	}

	// 9. Write the provided GitHub auth token to the repository secrets API
	err = github.CreateRepoActionSecretFromString(
		context.Background(),
		platformGitCfg.GitHub,
		gitURL.GetOwnerName(),
		gitURL.GetRepoName(),
		konnectTokenEnvVarName,
		konnectToken)
	if err != nil {
		return fmt.Errorf("failed to create GitHub secret: %w", err)
	}
	statusCh <- fmt.Sprintf("✔ Added %s to %s repository secrets\n", konnectTokenEnvVarName, gitURL.GetRepoName())

	statusCh <- fmt.Sprintf("✔ PR Filed: %s\n", prURL)
	statusCh <- "\tReview and Merge to complete adding the organization to the platform repository.\n\n"
	statusCh <- "Next run the Reference Platform self service UI with:\n"
	statusCh <- "\tkoctl run\n\n"
	statusCh <- "See the FAQ page for further questions on the Konnect Reference Platform:\n"
	statusCh <- "\thttps://deploy-preview-783--kongdeveloper.netlify.app/konnect-reference-platform/faq\n"

	return nil
}

// Looks up nested keys like jobs > build > steps
func findMapValuePath(root *yaml.Node, keys ...string) *yaml.Node {
	node := root
	for _, key := range keys {
		node = findMapValue(node, key)
		if node == nil {
			return nil
		}
	}
	return node
}

// Finds a value in a map node
func findMapValue(mapNode *yaml.Node, key string) *yaml.Node {
	if mapNode.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(mapNode.Content); i += 2 {
		k := mapNode.Content[i]
		v := mapNode.Content[i+1]
		if k.Value == key {
			return v
		}
	}
	return nil
}

// findStepByID returns the step node with a given id
func findStepByID(steps *yaml.Node, targetID string) *yaml.Node {
	if steps.Kind != yaml.SequenceNode {
		return nil
	}

	for _, step := range steps.Content {
		if step.Kind != yaml.MappingNode {
			continue
		}
		for i := 0; i < len(step.Content); i += 2 {
			k := step.Content[i]
			v := step.Content[i+1]
			if k.Value == "id" && v.Value == targetID {
				return step
			}
		}
	}
	return nil
}

// Sets or updates a key in a MappingNode
func setEnvVariable(env *yaml.Node, key, value string) {
	for i := 0; i < len(env.Content); i += 2 {
		k := env.Content[i]
		v := env.Content[i+1]
		if k.Value == key {
			v.Value = value
			return
		}
	}
	// Not found, append it
	env.Content = append(env.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value},
	)
}
