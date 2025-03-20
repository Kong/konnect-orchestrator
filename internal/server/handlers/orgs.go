package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	services "github.com/Kong/konnect-orchestrator/internal/git/github"
)

// OrgHandler handles organization related requests
type OrgHandler struct {
	githubService *services.GitHubService
}

// NewOrgHandler creates a new OrgHandler
func NewOrgHandler(githubService *services.GitHubService) *OrgHandler {
	return &OrgHandler{
		githubService: githubService,
	}
}

// ListOrganizations lists organizations for the authenticated user
func (h *OrgHandler) ListOrganizations(c *gin.Context) {
	// Get the GitHub token from the context (set by the auth middleware)
	githubToken, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	// Get organizations
	orgs, err := h.githubService.GetUserOrganizations(
		c.Request.Context(),
		githubToken.(string),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get organizations: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"organizations": orgs,
	})
}
