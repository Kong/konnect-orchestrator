package middleware

import (
	"net/http"
	"strings"
	"time"

	models "github.com/Kong/konnect-orchestrator/internal/git/github"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v60/github"

	services "github.com/Kong/konnect-orchestrator/internal/git/github"
)

// AuthMiddleware is a middleware for authentication
type AuthMiddleware struct {
	authService *services.AuthService
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth middleware checks if the user is authenticated
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header format must be Bearer {token}",
			})
			return
		}

		// Validate the token
		tokenString := parts[1]
		claims, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// Set the claims in the context
		c.Set("user", claims)
		c.Set("github_token", claims.GitHubToken)

		c.Next()
	}
}

// RefreshToken refreshes the token if it's about to expire
func (m *AuthMiddleware) RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request first
		c.Next()

		// Check if we have a user in the context
		user, exists := c.Get("user")
		if !exists {
			return
		}

		// Check if the token is about to expire
		claims := user.(*models.JWTClaims)
		if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) < 24*time.Hour {
			// Token is about to expire, generate a new one
			newToken, err := m.authService.GenerateJWT(
				&github.User{
					ID:        github.Int64(claims.UserID),
					Login:     github.String(claims.Login),
					AvatarURL: github.String(claims.AvatarURL),
					Name:      github.String(claims.Name),
				},
				claims.Email,
				claims.GitHubToken,
			)
			if err == nil {
				// Set the new token in the response header
				c.Header("X-Refresh-Token", newToken)
			}
		}
	}
}

// ErrorHandler handles errors in a standardized way
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Return the first error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": c.Errors.Last().Error(),
			})
		}
	}
}
