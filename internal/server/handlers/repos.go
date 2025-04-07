package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	gh "github.com/Kong/konnect-orchestrator/internal/git/github"
)

// RepoHandler handles repository related requests
type RepoHandler struct {
	githubService *gh.GitHubService
}

// NewRepoHandler creates a new RepoHandler
func NewRepoHandler(githubService *gh.GitHubService) *RepoHandler {
	return &RepoHandler{
		githubService: githubService,
	}
}

// ListRepositories lists repositories for the authenticated user
func (h *RepoHandler) ListRepositories(c *gin.Context) {
	// Get the GitHub token from the context
	githubToken, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	// Get query parameters
	visibility := c.DefaultQuery("visibility", "all") // Can be "all", "public", "private"
	affiliation := c.DefaultQuery("affiliation", "owner")

	// Get repositories
	repos, err := h.githubService.GetRepositories(
		c.Request.Context(),
		githubToken.(string),
		visibility,
		affiliation,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get repositories: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repositories": repos,
	})
}

// ListOrganizationRepositories lists repositories for an organization
func (h *RepoHandler) ListOrganizationRepositories(c *gin.Context) {
	// Get the GitHub token from the context (set by the auth middleware)
	githubToken, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	// Get the organization from the URL
	org := c.Param("org")
	if org == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing organization parameter",
		})
		return
	}

	// Get query parameters
	visibility := c.DefaultQuery("visibility", "all") // Can be "all", "public", "private"

	// Get repositories
	repos, err := h.githubService.GetOrganizationRepositories(
		c.Request.Context(),
		githubToken.(string),
		org,
		visibility,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get organization repositories: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"organization": org,
		"repositories": repos,
	})
}

// GetRepositoryContent gets content from a repository
func (h *RepoHandler) GetRepositoryContent(c *gin.Context) {
	// Get the GitHub token from the context (set by the auth middleware)
	githubToken, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	// Get parameters from the URL
	owner := c.Param("owner")
	repo := c.Param("repo")
	path := c.Param("path")
	if path == "" {
		path = "/"
	}

	// Get query parameters
	ref := c.DefaultQuery("ref", "") // Branch or commit SHA

	// Check if the repository is accessible
	accessible, err := h.githubService.IsRepositoryAccessible(
		c.Request.Context(),
		githubToken.(string),
		owner,
		repo,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check repository access: " + err.Error(),
		})
		return
	}

	if !accessible {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Repository not accessible",
		})
		return
	}

	// Get content
	content, err := h.githubService.GetRepositoryContent(
		c.Request.Context(),
		githubToken.(string),
		owner,
		repo,
		path,
		ref,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get repository content: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, content)
}

// ListRepositoryBranches lists branches for a repository
func (h *RepoHandler) ListRepositoryBranches(c *gin.Context) {
	// Get the GitHub token from the context
	githubToken, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	// Get parameters from the URL
	owner := c.Param("owner")
	repo := c.Param("repo")

	// Check for missing parameters
	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing owner or repository parameter",
		})
		return
	}

	// Check if the repository is accessible
	accessible, err := h.githubService.IsRepositoryAccessible(
		c.Request.Context(),
		githubToken.(string),
		owner,
		repo,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check repository access: " + err.Error(),
		})
		return
	}

	if !accessible {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Repository not accessible",
		})
		return
	}

	// Get branches
	branches, err := h.githubService.GetRepositoryBranches(
		c.Request.Context(),
		githubToken.(string),
		owner,
		repo,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get repository branches: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"branches": branches,
	})
}

// ListEnterpriseOrganizations lists organizations for a GitHub Enterprise instance
func (h *RepoHandler) ListEnterpriseOrganizations(c *gin.Context) {
	// Get the GitHub token from the context (set by the auth middleware)
	githubToken, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	// Get the enterprise server from the URL
	server := c.Param("server")
	if server == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing enterprise server parameter",
		})
		return
	}

	// Get organizations
	orgs, err := h.githubService.GetEnterpriseOrganizations(
		c.Request.Context(),
		githubToken.(string),
		server,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get enterprise organizations: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"server":        server,
		"organizations": orgs,
	})
}

// ListUserRepositories lists repositories for a specific user
func (h *RepoHandler) ListUserRepositories(c *gin.Context) {
	// Get the GitHub token from the context
	githubToken, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	// Get the username from the URL
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing username parameter",
		})
		return
	}

	visibility := c.DefaultQuery("visibility", "all")

	// Get repositories
	repos, err := h.githubService.GetUserRepositories(
		c.Request.Context(),
		githubToken.(string),
		username,
		visibility,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user repositories: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":     username,
		"repositories": repos,
	})
}
