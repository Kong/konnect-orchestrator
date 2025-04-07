package github

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Kong/konnect-orchestrator/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
)

// AuthService handles authentication related operations
type AuthService struct {
	config      *config.Config
	oauthConfig *oauth2.Config
}

// NewAuthService creates a new AuthService
func NewAuthService(config *config.Config) *AuthService {
	// Create OAuth2 configuration for GitHub
	oauthConfig := &oauth2.Config{
		ClientID:     config.GitHubClientID,
		ClientSecret: config.GitHubClientSecret,
		RedirectURL:  config.GitHubRedirectURI,
		Scopes:       config.GitHubScopes,
		Endpoint:     githubOAuth.Endpoint,
	}

	return &AuthService{
		config:      config,
		oauthConfig: oauthConfig,
	}
}

// GetAuthorizationURL returns the GitHub authorization URL
func (s *AuthService) GetAuthorizationURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// ExchangeCodeForToken exchanges the authorization code for a token
func (s *AuthService) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.oauthConfig.Exchange(ctx, code)
}

// GetUserFromGitHub gets the user information from GitHub
func (s *AuthService) GetUserFromGitHub(ctx context.Context, token *oauth2.Token) (*github.User, error) {
	// Create a GitHub client with the token
	client := github.NewClient(s.oauthConfig.Client(ctx, token))

	// Get the authenticated user
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserEmail gets the primary email of the user from GitHub
func (s *AuthService) GetUserEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	// Create a GitHub client with the token
	client := github.NewClient(s.oauthConfig.Client(ctx, token))

	// Get emails
	emails, _, err := client.Users.ListEmails(ctx, nil)
	if err != nil {
		return "", err
	}

	// Find the primary email
	for _, email := range emails {
		if email.GetPrimary() && email.GetVerified() {
			return email.GetEmail(), nil
		}
	}

	return "", errors.New("no primary verified email found")
}

// GenerateJWT generates a JWT token for the user
func (s *AuthService) GenerateJWT(user *github.User, email, token string) (string, error) {
	// Create claims
	claims := NewJWTClaims(
		user.GetID(),
		user.GetLogin(),
		email,
		user.GetAvatarURL(),
		user.GetName(),
		token,
		s.config.JWTExpiration,
	)

	// Create token
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)

	// Sign the token
	tokenString, err := jwtToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(s.config.JWTSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithLeeway(5*time.Second), // Allow for clock skew
	)

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func CreateGitHubClient(ctx context.Context, token string) *github.Client {
	// Create a token source
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Create a GitHub client
	return github.NewClient(tc)
}
