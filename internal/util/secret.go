package util

import (
	"fmt"
	"os"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
)

func ResolveSecretValue(secret manifest.Secret) (string, error) {
	switch secret.Type {
	case "file":
		expandedPath := os.ExpandEnv(secret.Value)
		content, err := os.ReadFile(expandedPath)
		if err != nil {
			return "", fmt.Errorf("failed to read token file: %w", err)
		}
		return string(content), nil
	case "env":
		value := os.Getenv(secret.Value)
		if value == "" {
			return "", fmt.Errorf("environment variable %s not set", secret.Value)
		}
		return value, nil
	case "literal":
		return secret.Value, nil
	default:
		return "", fmt.Errorf("unsupported token type: %s", secret.Type)
	}
}
