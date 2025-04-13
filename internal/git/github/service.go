package github

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/util"
	"github.com/google/go-github/v60/github"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
)

type GitHubService struct {
	authService *AuthService
}

// NewGitHubService creates a new GitHubService
func NewGitHubService(authService *AuthService) *GitHubService {
	return &GitHubService{
		authService: authService,
	}
}

func CreateOrUpdatePullRequest(ctx context.Context,
	owner, repo, branch, title, body string,
	githubConfig manifest.GitHubConfig,
) (*github.PullRequest, error) {
	// Create GitHub client with token
	token, err := util.ResolveSecretValue(*githubConfig.Token)
	if err != nil {
		return nil, err
	}

	client := CreateGitHubClient(ctx, token)

	// First, check if there's an existing PR for this branch
	existingPRs, _, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		Head:  fmt.Sprintf("%s:%s", owner, branch),
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

// GetUserProfile gets the user profile from GitHub
func (s *GitHubService) GetUserProfile(ctx context.Context, token string) (UserProfile, error) {
	client := s.createClient(ctx, token)

	// Get user information
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return UserProfile{}, err
	}

	// Get user email
	email, err := s.getUserPrimaryEmail(ctx, client)
	if err != nil {
		// Non-critical error, proceed without email
		email = ""
	}

	// Map to our model
	return UserProfile{
		ID:        user.GetID(),
		Login:     user.GetLogin(),
		Name:      user.GetName(),
		Email:     email,
		AvatarURL: user.GetAvatarURL(),
		HTMLURL:   user.GetHTMLURL(),
		Bio:       user.GetBio(),
		Company:   user.GetCompany(),
		Location:  user.GetLocation(),
	}, nil
}

// getUserPrimaryEmail gets the primary email of the user
func (s *GitHubService) getUserPrimaryEmail(ctx context.Context, client *github.Client) (string, error) {
	emails, _, err := client.Users.ListEmails(ctx, nil)
	if err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.GetPrimary() && email.GetVerified() {
			return email.GetEmail(), nil
		}
	}

	return "", errors.New("no primary verified email found")
}

// GetRepositories gets repositories for the user
func (s *GitHubService) GetRepositories(ctx context.Context, token, visibility, affiliation string) ([]Repository, error) {
	client := s.createClient(ctx, token)

	// Set up options for listing repositories
	opts := &github.RepositoryListOptions{
		Visibility:  visibility,  // Can be "all", "public", or "private"
		Affiliation: affiliation, // Can be "owner", "collaborator", "organization_member", or a combination
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []Repository
	for {
		repos, resp, err := client.Repositories.List(ctx, "", opts)
		if err != nil {
			return nil, err
		}

		// Map GitHub repos to our model
		for _, repo := range repos {
			allRepos = append(allRepos, mapRepository(repo))
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// GetOrganizationRepositories gets repositories for an organization
func (s *GitHubService) GetOrganizationRepositories(ctx context.Context, token, org, visibility string) ([]Repository, error) {
	client := s.createClient(ctx, token)

	// Set up options for listing repositories
	opts := &github.RepositoryListByOrgOptions{
		Type:        visibility, // Can be "all", "public", or "private"
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, err
		}

		// Map GitHub repos to our model
		for _, repo := range repos {
			allRepos = append(allRepos, mapRepository(repo))
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

func validatePath(path string) (string, error) {
	// Remove any ".." path traversal attempts
	if strings.Contains(path, "..") {
		return "", errors.New("invalid path: contains path traversal sequences")
	}

	// Ensure path starts with a forward slash
	if !strings.HasPrefix(path, "/") && path != "" {
		path = "/" + path
	}

	// Additional path validation
	validPath := regexp.MustCompile(`^[\/a-zA-Z0-9_\-\.]+$`)
	if path != "" && !validPath.MatchString(path) {
		return "", errors.New("invalid path: contains unauthorized characters")
	}

	return path, nil
}

// GetRepositoryContent gets content from a repository
func (s *GitHubService) GetRepositoryContent(ctx context.Context, token, owner, repo, path string, ref string) (interface{}, error) {
	// Validate and sanitize path
	sanitizedPath, err := validatePath(path)
	if err != nil {
		return nil, err
	}

	client := s.createClient(ctx, token)
	// Continue with sanitizedPath instead of path
	fileContent, directoryContent, _, err := client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		sanitizedPath,
		&github.RepositoryContentGetOptions{Ref: ref},
	)
	if err != nil {
		return nil, err
	}

	// Check if it's a file or directory
	if fileContent != nil {
		// It's a file// Correct approach
		encodeddContent, err := fileContent.GetContent()
		if err != nil {
			return nil, err // Handle the error first
		}

		// Now pass only the content to the decode function
		content, err := decodeContent(encodeddContent)
		if err != nil {
			return nil, err
		}

		return FileContent{
			Name:        fileContent.GetName(),
			Path:        fileContent.GetPath(),
			SHA:         fileContent.GetSHA(),
			Size:        fileContent.GetSize(),
			Type:        "file",
			Content:     content,
			Encoding:    fileContent.GetEncoding(),
			DownloadURL: fileContent.GetDownloadURL(),
			URL:         fileContent.GetURL(),
			HTMLURL:     fileContent.GetHTMLURL(),
		}, nil
	}

	// It's a directory
	var dirContent DirectoryContent
	for _, c := range directoryContent {
		dirContent = append(dirContent, FileContent{
			Name:        c.GetName(),
			Path:        c.GetPath(),
			SHA:         c.GetSHA(),
			Size:        c.GetSize(),
			Type:        c.GetType(),
			DownloadURL: c.GetDownloadURL(),
			URL:         c.GetURL(),
			HTMLURL:     c.GetHTMLURL(),
		})
	}

	return dirContent, nil
}

// GetRepositoryBranches gets branches for a repository
func (s *GitHubService) GetRepositoryBranches(ctx context.Context, token, owner, repo string) ([]Branch, error) {
	client := s.createClient(ctx, token)

	// First get the repository to find the default branch
	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	defaultBranch := repository.GetDefaultBranch()

	// Set up options for listing branches
	opts := &github.BranchListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allBranches []Branch
	for {
		branches, resp, err := client.Repositories.ListBranches(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list branches: %w", err)
		}

		// Map GitHub branches to our model
		for _, branch := range branches {
			allBranches = append(allBranches, Branch{
				Name:      branch.GetName(),
				CommitSHA: branch.GetCommit().GetSHA(),
				Protected: branch.GetProtected(),
				IsDefault: branch.GetName() == defaultBranch,
			})
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allBranches, nil
}

// GetEnterpriseOrganizations gets organizations for a GitHub Enterprise instance
func (s *GitHubService) GetEnterpriseOrganizations(ctx context.Context, token, enterpriseServer string) ([]*github.Organization, error) {
	// For GitHub Enterprise, we need to create a custom client with the enterprise server URL
	if !strings.HasPrefix(enterpriseServer, "https://") {
		enterpriseServer = "https://" + enterpriseServer
	}

	// Make sure the URL ends with a trailing slash
	if !strings.HasSuffix(enterpriseServer, "/") {
		enterpriseServer = enterpriseServer + "/"
	}

	// Create a token source
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Construct the API and upload URLs for GitHub Enterprise
	baseAPIURL := enterpriseServer + "api/v3/"
	uploadURL := enterpriseServer + "api/uploads/"

	// Create the client using the new builder pattern approach
	client, err := github.NewClient(tc).WithEnterpriseURLs(baseAPIURL, uploadURL)
	if err != nil {
		return nil, err
	}

	// List organizations
	opts := &github.ListOptions{PerPage: 100}
	var allOrgs []*github.Organization

	for {
		orgs, resp, err := client.Organizations.List(ctx, "", opts)
		if err != nil {
			return nil, err
		}

		allOrgs = append(allOrgs, orgs...)

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allOrgs, nil
}

// IsRepositoryAccessible checks if a repository is accessible
func (s *GitHubService) IsRepositoryAccessible(ctx context.Context, token, owner, repo string) (bool, error) {
	client := s.createClient(ctx, token)

	// Try to get the repository
	_, resp, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// createClient creates a GitHub client with the token
func (s *GitHubService) createClient(ctx context.Context, token string) *github.Client {
	// Create a token source
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Create a GitHub client
	return github.NewClient(tc)
}

// mapRepository maps a GitHub repository to our model
func mapRepository(repo *github.Repository) Repository {
	return Repository{
		ID:          repo.GetID(),
		Name:        repo.GetName(),
		FullName:    repo.GetFullName(),
		Description: repo.GetDescription(),
		Private:     repo.GetPrivate(),
		HTMLURL:     repo.GetHTMLURL(),
		CloneURL:    repo.GetCloneURL(),
		SSHURL:      repo.GetSSHURL(),
		Owner: struct {
			Login     string `json:"login"`
			ID        int64  `json:"id"`
			AvatarURL string `json:"avatar_url"`
		}{
			Login:     repo.GetOwner().GetLogin(),
			ID:        repo.GetOwner().GetID(),
			AvatarURL: repo.GetOwner().GetAvatarURL(),
		},
		DefaultBranch: repo.GetDefaultBranch(),
		CreatedAt:     repo.GetCreatedAt().Format(time.RFC3339),
		UpdatedAt:     repo.GetUpdatedAt().Format(time.RFC3339),
		IsEnterprise:  IsEnterpriseURL(repo.GetHTMLURL()),
	}
}

// GetUserOrganizations gets organizations for the authenticated user
func (s *GitHubService) GetUserOrganizations(ctx context.Context, token string) ([]Organization, error) {
	client := s.createClient(ctx, token)

	// Set up options for listing organizations
	opts := &github.ListOptions{PerPage: 100}

	var allOrgs []Organization
	for {
		orgs, resp, err := client.Organizations.List(ctx, "", opts)
		if err != nil {
			return nil, err
		}

		// Map GitHub organizations to our model
		for _, org := range orgs {
			allOrgs = append(allOrgs, Organization{
				ID:        org.GetID(),
				Login:     org.GetLogin(),
				Name:      org.GetName(),
				AvatarURL: org.GetAvatarURL(),
				HTMLURL:   org.GetHTMLURL(),
			})
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allOrgs, nil
}

// GetUserRepositories gets repositories for a specific user
func (s *GitHubService) GetUserRepositories(ctx context.Context, token, username, visibility string) ([]Repository, error) {
	client := s.createClient(ctx, token)

	// Set up options for listing repositories
	opts := &github.RepositoryListByUserOptions{
		Type:        visibility, // Include all repositories (public, private)
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []Repository
	for {
		repos, resp, err := client.Repositories.ListByUser(ctx, username, opts)
		if err != nil {
			return nil, err
		}

		// Map GitHub repos to our model
		for _, repo := range repos {
			allRepos = append(allRepos, mapRepository(repo))
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

func naclEncrypt(recipientPublicKey string, content string) string {
	// taken from here: https://jefflinse.io/posts/encrypting-github-secrets-using-go/
	// currently ignoring errors here
	b, _ := base64.StdEncoding.DecodeString(recipientPublicKey)
	recipientKey := new([32]byte)
	copy(recipientKey[:], b)
	pubKey, privKey, _ := box.GenerateKey(rand.Reader)
	nonceHash, _ := blake2b.New(24, nil)
	nonceHash.Write(pubKey[:])
	nonceHash.Write(recipientKey[:])
	nonce := new([24]byte)
	copy(nonce[:], nonceHash.Sum(nil))
	out := box.Seal(pubKey[:], []byte(content), nonce, recipientKey, privKey)
	return base64.StdEncoding.EncodeToString(out)
}

func CreateRepoActionSecretFromString(ctx context.Context, githubConfig *manifest.GitHubConfig, owner, repo, secretName, secretValue string) error {
	token, err := util.ResolveSecretValue(*githubConfig.Token)
	if err != nil {
		return err
	}

	client := CreateGitHubClient(ctx, token)

	repoPubKey, _, err := client.Actions.GetRepoPublicKey(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	encryptedSecretValue := naclEncrypt(repoPubKey.GetKey(), secretValue)

	// Create a new secret
	secret := &github.EncryptedSecret{
		Name:           secretName,
		KeyID:          repoPubKey.GetKeyID(),
		EncryptedValue: encryptedSecretValue,
	}

	_, err = client.Actions.CreateOrUpdateRepoSecret(ctx, owner, repo, secret)
	if err != nil {
		return fmt.Errorf("failed to create or update secret: %w", err)
	}

	return nil
}

func CreateRepoActionSecret(ctx context.Context, githubConfig *manifest.GitHubConfig, owner, repo, secretName string, secretValue *manifest.Secret) error {
	s, err := util.ResolveSecretValue(*secretValue)
	if err != nil {
		return err
	}
	return CreateRepoActionSecretFromString(ctx, githubConfig, owner, repo, secretName, s)
}

func (s *GitHubService) GetActionsSecrets(ctx context.Context, token, owner, repo string) ([]Secrets, error) {
	client := s.createClient(ctx, token)
	opts := &github.ListOptions{PerPage: 100}
	_, _, err := client.Actions.ListRepoSecrets(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	return nil, nil
}

// IsEnterpriseURL checks if a URL is for a GitHub Enterprise instance
func IsEnterpriseURL(url string) bool {
	return !strings.Contains(url, "github.com")
}

// decodeContent decodes base64 encoded content
func decodeContent(content string) (string, error) {
	// GitHub API returns content with newlines, which need to be removed for base64 decoding
	content = strings.ReplaceAll(content, "\n", "")
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", fmt.Errorf("failed to decode content: %w", err)
	}
	return string(decoded), nil
}

// GetRepositoryPullRequests fetches pull requests for a repository from GitHub
func (s *GitHubService) GetRepositoryPullRequests(ctx context.Context, token, owner, repo string, state, sort, direction string) ([]PullRequest, error) {
	// Get the GitHub client
	client := s.createClient(ctx, token)

	// Set up the options for the GitHub API call
	opts := &github.PullRequestListOptions{
		State:     state,
		Sort:      sort,
		Direction: direction,
		ListOptions: github.ListOptions{
			PerPage: 100, // Adjust based on your needs
		},
	}

	// Call GitHub API to get pull requests
	pullRequests, _, err := client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pull requests: %w", err)
	}

	// Convert GitHub pull requests to our model
	result := make([]PullRequest, 0, len(pullRequests))

	for _, pr := range pullRequests {
		pullRequest := PullRequest{
			ID:        *pr.ID,
			Number:    *pr.Number,
			Title:     *pr.Title,
			State:     *pr.State,
			HTMLURL:   *pr.HTMLURL,
			CreatedAt: *pr.CreatedAt,
			UpdatedAt: *pr.UpdatedAt,
		}

		if pr.Body != nil {
			pullRequest.Body = *pr.Body
		}

		if pr.ClosedAt != nil {
			closedAt := *pr.ClosedAt
			pullRequest.ClosedAt = &closedAt
		}

		if pr.MergedAt != nil {
			mergedAt := *pr.MergedAt
			pullRequest.MergedAt = &mergedAt
		}

		if pr.User != nil {
			pullRequest.User = PullRequestUser{
				ID:        *pr.User.ID,
				Login:     *pr.User.Login,
				AvatarURL: *pr.User.AvatarURL,
				HTMLURL:   *pr.User.HTMLURL,
			}
		}

		result = append(result, pullRequest)
	}

	return result, nil
}
