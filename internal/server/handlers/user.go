package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	services "github.com/Kong/konnect-orchestrator/internal/git/github"
)

// UserHandler handles user related requests
type UserHandler struct {
	githubService *services.GitHubService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(githubService *services.GitHubService) *UserHandler {
	return &UserHandler{
		githubService: githubService,
	}
}

// GetProfile gets the user profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get the GitHub token from the context (set by the auth middleware)
	githubToken, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	// Get the user profile
	profile, err := h.githubService.GetUserProfile(
		c.Request.Context(),
		githubToken.(string),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user profile: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}
