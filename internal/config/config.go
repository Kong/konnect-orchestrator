package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	ServerPort  int
	ServerHost  string
	Environment string
	FrontendURL string

	// GitHub OAuth configuration
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURI  string
	GitHubScopes       []string

	// JWT configuration
	JWTSecret     string
	JWTExpiration time.Duration

	// Session configuration - added for CSRF protection
	SessionSecret string

	// Platform Repository configuration
	PlatformRepoOwnerName  string
	PlatformRepoOwnerEmail string
	PlatformRepoURL        string
	PlatformRepoGHToken    string

	KonnectToken string
	OrgName      string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if present
	_ = godotenv.Load()

	// Load server configuration
	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	// Set up default scopes for GitHub
	// These scopes will determine what your application can access
	defaultScopes := []string{"user:email", "read:user", "repo"}

	config := &Config{
		// Server configuration
		ServerPort:  serverPort,
		ServerHost:  getEnv("SERVER_HOST", "0.0.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),

		// GitHub OAuth configuration
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURI:  getEnv("GITHUB_REDIRECT_URI", "http://localhost:8080/auth/github/callback"),
		GitHubScopes:       defaultScopes,

		// JWT configuration
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiration: time.Duration(getEnvAsInt("JWT_EXPIRATION_HOURS", 24)) * time.Hour,

		// Session configuration - added for CSRF protection
		SessionSecret: getEnv("SESSION_SECRET", "session-secret-change-in-production"),

		// Platform Repository configuration
		PlatformRepoOwnerName:  getEnv("PLATFORM_REPO_OWNER_NAME", "Konnect Orchestrator"),
		PlatformRepoOwnerEmail: getEnv("PLATFORM_REPO_OWNER_EMAIL", "ko@konghq.com"),
		PlatformRepoURL:        getEnv("PLATFORM_REPO_URL", ""),
		PlatformRepoGHToken:    getEnv("PLATFORM_REPO_GITHUB_TOKEN", ""),

		KonnectToken: getEnv("KONNECT_TOKEN", ""),
		OrgName:      getEnv("ORG_NAME", ""),
	}

	return config, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
