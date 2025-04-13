package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/Kong/konnect-orchestrator/internal/git"
	gh "github.com/Kong/konnect-orchestrator/internal/git/github"
	"github.com/Kong/konnect-orchestrator/internal/manifest"
	giturl "github.com/kubescape/go-git-url"
	"gopkg.in/yaml.v3"
)

// RepoHandler handles repository related requests
type PlatformHandler struct {
	githubService     *gh.GitHubService
	platformGitConfig manifest.GitConfig
	teamsFilePath     string
	orgsFilePath      string
}

// NewRepoHandler creates a new RepoHandler
func NewPlatformHandler(githubService *gh.GitHubService, platformGitConfig manifest.GitConfig,
	teamsFilePath, orgsFilePath string) *PlatformHandler {
	return &PlatformHandler{
		githubService:     githubService,
		platformGitConfig: platformGitConfig,
		teamsFilePath:     teamsFilePath,
		orgsFilePath:      orgsFilePath,
	}
}

func (h *PlatformHandler) GetRepositoryPullRequests(c *gin.Context) {
	// Get the GitHub token from the context
	githubToken, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	gitURL, err := giturl.NewGitURL(*h.platformGitConfig.Remote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get path parameters
	owner := gitURL.GetOwnerName()
	repo := gitURL.GetRepoName()

	// Get query parameters with defaults
	state := c.DefaultQuery("state", "all")
	sort := c.DefaultQuery("sort", "created")
	direction := c.DefaultQuery("direction", "desc")

	// Validate state parameter
	validStates := map[string]bool{"open": true, "closed": true, "all": true}
	if !validStates[state] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter. Must be 'open', 'closed', or 'all'"})
		return
	}

	// Call the service to get pull requests
	pullRequests, err := h.githubService.GetRepositoryPullRequests(c.Request.Context(), githubToken.(string), owner, repo, state, sort, direction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Construct the response
	response := gh.PullRequestResponse{
		PullRequests: pullRequests,
	}

	c.JSON(http.StatusOK, response)
}

func (h *PlatformHandler) GetExistingServices(c *gin.Context) {
	auth, err := git.GetAuthMethod(h.platformGitConfig)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Failed to get authentication method: " + err.Error(),
		})
		return
	}

	// Setup in-memory filesystem
	fs := memfs.New()

	// Clone repository in memory
	r, err := goGit.Clone(memory.NewStorage(), fs, &goGit.CloneOptions{
		URL:           *h.platformGitConfig.Remote,
		SingleBranch:  true,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", "main")), // Assuming main branch
		Auth:          auth,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clone repository: " + err.Error(),
		})
		return
	}

	// Get the worktree
	w, err := r.Worktree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get worktree: " + err.Error(),
		})
		return
	}

	file, err := w.Filesystem.Open(h.teamsFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open config file: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read config file: " + err.Error(),
		})
		return
	}

	// Parse yaml
	var orchestrator manifest.Orchestrator
	if err := yaml.Unmarshal(content, &orchestrator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse config file: " + err.Error(),
		})
		return
	}

	// Create response structure for frontend
	type ServiceResponse struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		SpecPath    string `json:"specPath,omitempty"`
		ProdBranch  string `json:"prodBranch"`
		DevBranch   string `json:"devBranch"`
		Git         struct {
			Repo string `json:"repo"`
		} `json:"git"`
		Team string `json:"team"`
	}

	// Convert orchestrator services to response format
	services := []ServiceResponse{}
	for teamName, team := range orchestrator.Teams {
		for serviceKey, service := range team.Services {
			if service == nil || service.Name == nil {
				continue // Skip invalid services
			}

			// Get repo path from the remote URL or use the key as fallback
			repoPath := serviceKey
			if service.Git != nil && service.Git.Remote != nil {
				// Extract repo path from full URL
				remoteURL := *service.Git.Remote
				if strings.Contains(remoteURL, "github.com") {
					parts := strings.Split(remoteURL, "github.com/")
					if len(parts) > 1 {
						repoPath = strings.TrimSuffix(parts[1], ".git")
					}
				}
			}

			// Create service response
			serviceResp := ServiceResponse{
				Name:       *service.Name,
				Team:       teamName,
				ProdBranch: service.ProdBranch,
				DevBranch:  service.DevBranch,
				Git: struct {
					Repo string `json:"repo"`
				}{
					Repo: repoPath,
				},
			}

			// Add optional fields if available
			if service.Description != nil {
				serviceResp.Description = *service.Description
			}
			if service.SpecPath != "" {
				serviceResp.SpecPath = service.SpecPath
			}

			services = append(services, serviceResp)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
	})
}

func (h *PlatformHandler) AddServiceRepo(c *gin.Context) {
	auth, err := git.GetAuthMethod(h.platformGitConfig)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	// Parse the request body
	var repoInfo gh.Repository
	if err := c.ShouldBindJSON(&repoInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid repository data: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if repoInfo.Name == "" || repoInfo.Owner.Login == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Repository name and owner are required",
		})
		return
	}

	baseBranch := "main"

	// Create unique branch name with timestamp
	newBranchName := fmt.Sprintf("add-service-%s", repoInfo.Name)

	// Setup in-memory filesystem
	fs := memfs.New()

	// Clone repository in memory
	r, err := goGit.Clone(memory.NewStorage(), fs, &goGit.CloneOptions{
		URL:           *h.platformGitConfig.Remote,
		SingleBranch:  true,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", baseBranch)),
		Auth:          auth,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error cloning repository": err.Error()})
		return
	}

	// Get the worktree
	w, err := r.Worktree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error getting worktree": err.Error()})
		return
	}

	// Create and checkout new branch
	headRef, err := r.Head()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error getting HEAD": err.Error()})
		return
	}

	branchRefName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", newBranchName))
	ref := plumbing.NewHashReference(branchRefName, headRef.Hash())

	err = r.Storer.SetReference(ref)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error creating branch": err.Error()})
		return
	}

	err = w.Checkout(&goGit.CheckoutOptions{
		Branch: branchRefName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error checking out branch": err.Error()})
		return
	}

	// Read existing file content (if any)
	var oldContent []byte
	file, err := w.Filesystem.Open(h.teamsFilePath)
	if err == nil {
		// File exists, read its content
		oldContent, err = io.ReadAll(file)
		file.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error reading file": err.Error()})
			return
		}
	}

	var man manifest.Orchestrator

	if err := yaml.Unmarshal(oldContent, &man); err != nil {
		// Fall back to JSON if YAML fails
		if jsonErr := json.Unmarshal(oldContent, &man); jsonErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"failed to parse file as YAML or JSON": err.Error()})
			return
		}
	}

	if man.Teams == nil {
		man.Teams = make(map[string]*manifest.Team)
		man.Teams[repoInfo.Team] = &manifest.Team{
			Services:    make(map[string]*manifest.Service),
			Users:       []string{},
			Description: &repoInfo.Team,
		}
	}

	var prodBranch = "main"
	var devBranch = "dev"

	if repoInfo.ProdBranch != "" {
		prodBranch = repoInfo.ProdBranch
	}

	if repoInfo.DevBranch != "" {
		devBranch = repoInfo.DevBranch
	}

	newService := manifest.Service{
		Name: &repoInfo.Name,
		Git: &manifest.GitConfig{
			Remote: &repoInfo.CloneURL,
		},
		Description: &repoInfo.Description,
		SpecPath:    "openapi.yaml",
		ProdBranch:  prodBranch,
		DevBranch:   devBranch,
	}

	man.Teams[repoInfo.Team].Services[repoInfo.FullName] = &newService

	file, err = w.Filesystem.Create(h.teamsFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error creating file": err.Error()})
		return
	}

	// Use the encoder directly on your struct
	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(man); err != nil { // Encode the struct directly
		fmt.Printf("Error encoding YAML: %v\n", err)
		file.Close()
		return
	}

	// Close the encoder when done
	if err := encoder.Close(); err != nil {
		fmt.Printf("Error closing encoder: %v\n", err)
		file.Close()
		return
	}
	file.Close()

	_, err = w.Add(h.teamsFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error staging file": err.Error()})

		return
	}

	_, err = w.Commit("message", &goGit.CommitOptions{
		Author: &object.Signature{
			Name:  *h.platformGitConfig.Author.Name,
			Email: *h.platformGitConfig.Author.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error committing changes": err.Error()})
		return
	}

	// Push the branch
	err = r.Push(&goGit.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec(fmt.Sprintf("%s:%s", branchRefName, branchRefName))},
		Auth:       auth,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error pushing changes": err.Error()})

		return
	}

	gitURL, err := giturl.NewGitURL(*h.platformGitConfig.Remote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get path parameters
	owner := gitURL.GetOwnerName()
	repo := gitURL.GetRepoName()

	_, err = gh.CreateOrUpdatePullRequest(
		c.Request.Context(),
		owner,
		repo,
		newBranchName,
		fmt.Sprintf("[Konnect Orchestrator] - Add Service: %s", repoInfo.Name), "Adding service manifest", *h.platformGitConfig.GitHub)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error creating pull request": err.Error()})
		return
	}

	// RESPONSE

	c.JSON(http.StatusOK, gin.H{
		"message": "Repository registered successfully",
		"repo": gin.H{
			"id":             repoInfo.ID,
			"name":           repoInfo.Name,
			"full_name":      repoInfo.FullName,
			"owner":          repoInfo.Owner,
			"url":            repoInfo.CloneURL,
			"is_private":     repoInfo.Private,
			"default_branch": repoInfo.DefaultBranch,
		},
	})
}
