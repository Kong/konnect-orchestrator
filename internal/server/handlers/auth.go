package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/Kong/konnect-orchestrator/internal/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	gh "github.com/Kong/konnect-orchestrator/internal/git/github"
)

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *gh.AuthService, config *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		config:      config,
	}
}

// And update the struct to include the config
type AuthHandler struct {
	authService *gh.AuthService
	config      *config.Config
}

// Login initiates the GitHub OAuth flow
func (h *AuthHandler) Login(c *gin.Context) {
	// Generate a random state
	state, err := generateRandomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate state",
		})
		return
	}

	// Store the state in a cookie
	c.SetCookie("oauth_state", state, int(5*time.Minute.Seconds()), "/", "", true, true)
	c.SetSameSite(http.SameSiteStrictMode)

	// Get the authorization URL
	authURL := h.authService.GetAuthorizationURL(state)

	// Redirect to GitHub
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// Callback handles the GitHub OAuth callback
func (h *AuthHandler) Callback(c *gin.Context) {
	// Get the state from the URL
	state := c.Query("state")
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing state parameter",
		})
		return
	}

	// Get the state from the cookie
	savedState, err := c.Cookie("oauth_state")
	if err != nil || state != savedState {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid state parameter",
		})
		return
	}

	// Get the code from the URL
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing code parameter",
		})
		return
	}

	// Exchange the code for a token
	token, err := h.authService.ExchangeCodeForToken(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to exchange code for token",
		})
		return
	}

	// Get the user from GitHub
	user, err := h.authService.GetUserFromGitHub(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user from GitHub",
		})
		return
	}

	// Get the user's email
	email, err := h.authService.GetUserEmail(c.Request.Context(), token)
	if err != nil {
		// Non-critical error, proceed without email
		email = ""
	}

	// Generate a JWT token
	jwtToken, err := h.authService.GenerateJWT(user, email, token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate JWT token",
		})
		return
	}

	// Set the token in a cookie for web clients
	c.SetCookie("auth_token", jwtToken, int(24*time.Hour.Seconds()), "/", "", true, true)
	c.SetSameSite(http.SameSiteStrictMode)

	// Get the frontend URL from config (you'll need to add this to your config)
	frontendURL := h.config.FrontendURL // Add this to your config structure

	// Redirect to the Vue.js frontend's callback route with the token
	c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/auth/success?token="+jwtToken)
}

// Success handles the successful authentication
func (h *AuthHandler) Success(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing token parameter",
		})
		return
	}

	// Generate a CSRF token
	csrfToken, err := generateRandomToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate CSRF token",
		})
		return
	}

	// Store the CSRF token in session
	session := sessions.Default(c)
	session.Set("csrf_token", csrfToken)
	session.Save()

	// For successful authentication, return the token and CSRF token
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Authentication successful",
		"token":      token,
		"csrf_token": csrfToken,
	})
}

// Logout handles the logout request
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear the auth token cookie
	c.SetCookie("auth_token", "", -1, "/", "", true, true)

	// Clear the session
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// RefreshToken refreshes the JWT token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get the user from the context (set by the auth middleware)
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	// Get the GitHub token from the context
	githubToken, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GitHub token not found",
		})
		return
	}

	// Create a GitHub client
	client := gh.CreateGitHubClient(c.Request.Context(), githubToken.(string))

	// Get the user from GitHub
	user, _, err := client.Users.Get(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user from GitHub",
		})
		return
	}

	// Generate a new JWT token
	jwtToken, err := h.authService.GenerateJWT(
		user,
		claims.(*gh.JWTClaims).Email,
		githubToken.(string),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate JWT token",
		})
		return
	}

	// Generate a new CSRF token
	csrfToken, err := generateRandomToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate CSRF token",
		})
		return
	}

	// Update the CSRF token in session
	session := sessions.Default(c)
	session.Set("csrf_token", csrfToken)
	session.Save()

	// Set the token in a cookie for web clients
	c.SetCookie("auth_token", jwtToken, int(24*time.Hour.Seconds()), "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"token":      jwtToken,
		"csrf_token": csrfToken,
	})
}

// generateRandomState generates a random state for OAuth
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// generateRandomToken generates a random token for CSRF
func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
