package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	gh "github.com/Kong/konnect-orchestrator/internal/git/github"
)

// RepoHandler handles repository related requests
type PlatformHandler struct {
	githubService *gh.GitHubService
}

// NewRepoHandler creates a new RepoHandler
func NewPlatformHandler(githubService *gh.GitHubService) *PlatformHandler {
	return &PlatformHandler{
		githubService: githubService,
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
	// Get path parameters
	owner := "johnharris85"
	repo := "testing-koctl"

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
