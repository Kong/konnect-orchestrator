package github

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in the JWT token
type JWTClaims struct {
	UserID      int64  `json:"user_id"`
	Login       string `json:"login"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	Name        string `json:"name"`
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
func (c *JWTClaims) Valid() error {
	now := time.Now().Unix()

	if now > c.ExpiresAt {
		return errors.New("token is expired")
	}

	if c.IssuedAt > now {
		return errors.New("token issued in the future")
	}

	return nil
}

// GetExpirationTime implements the jwt.Claims interface
func (c *JWTClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	if c.ExpiresAt == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.ExpiresAt, 0)), nil
}

// GetIssuedAt implements the jwt.Claims interface
func (c *JWTClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	if c.IssuedAt == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.IssuedAt, 0)), nil
}

// GetNotBefore implements the jwt.Claims interface
func (c *JWTClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

// GetIssuer implements the jwt.Claims interface
func (c *JWTClaims) GetIssuer() (string, error) {
	return "", nil
}

// GetSubject implements the jwt.Claims interface
func (c *JWTClaims) GetSubject() (string, error) {
	return c.Login, nil
}

// GetAudience implements the jwt.Claims interface
func (c *JWTClaims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}
