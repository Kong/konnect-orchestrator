package github

import (
	"errors"
	"time"
)

// JWTClaims represents the claims in the JWT token
type JWTClaims struct {
	UserID    int64  `json:"user_id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"name"`
	// Store GitHub token securely (in production, consider encryption)
	GitHubToken string `json:"github_token"`
	// Standard JWT claims
	ExpiresAt int64 `json:"exp"`
	IssuedAt  int64 `json:"iat"`
}

// NewJWTClaims creates a new JWTClaims
func NewJWTClaims(userID int64, login, email, avatarURL, name, githubToken string, expiresIn time.Duration) JWTClaims {
	now := time.Now()
	return JWTClaims{
		UserID:      userID,
		Login:       login,
		Email:       email,
		AvatarURL:   avatarURL,
		Name:        name,
		GitHubToken: githubToken,
		IssuedAt:    now.Unix(),
		ExpiresAt:   now.Add(expiresIn).Unix(),
	}
}

// Valid implements the jwt.Claims interface.
// This is called during token verification to validate claims
func (c *JWTClaims) Valid() error {
	now := time.Now().Unix()

	// Check if the token is expired
	if now > c.ExpiresAt {
		return errors.New("token is expired")
	}

	// Check if the token was issued in the future (clock skew)
	if c.IssuedAt > now {
		return errors.New("token issued in the future")
	}

	// Token is valid
	return nil
}
