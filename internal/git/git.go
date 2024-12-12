package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// getAuthMethod returns an ssh.AuthMethod based on the git config, or nil if no auth is specified
func getAuthMethod(gitConfig manifest.Git) (ssh.AuthMethod, error) {
	// Return nil if no auth configured
	if gitConfig.Auth.Type == "" {
		return nil, nil
	}

	// Only support SSH auth for now
	if gitConfig.Auth.Type != "ssh" {
		return nil, errors.New("not supported")
	}

	// Get the SSH key content based on the key type
	switch gitConfig.Auth.SSH.Key.Type {
	case "file":
		expandedPath := os.ExpandEnv(gitConfig.Auth.SSH.Key.Value)
		// Create SSH key authentication
		publicKeys, err := ssh.NewPublicKeysFromFile("git", expandedPath, "" /*pwd*/)
		if err != nil {
			return nil, err
		}
		return publicKeys, nil
	case "env":
		//keyContent = os.Getenv(platform.Git.Auth.SSH.Key.Value)
		return nil, errors.New("not supported")
	case "literal":
		//keyContent = platform.Git.Auth.SSH.Key.Value
		return nil, errors.New("not supported")
	default:
		return nil, errors.New("not supported")
	}
}

func GetRemoteFile(gitConfig manifest.Git, branch, path string) ([]byte, error) {
	auth, err := getAuthMethod(gitConfig)
	if err != nil {
		return nil, err
	}

	tempDir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return nil, err
	}

	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:           gitConfig.Remote,
		Auth:          auth,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(tempDir, path))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

// CloneInto clones a git repository into the specified directory
func CloneInto(gitConfig manifest.Git, dir string) error {

	auth, err := getAuthMethod(gitConfig)
	if err != nil {
		return err
	}

	// Clone the repository
	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:  gitConfig.Remote,
		Auth: auth,
	})
	return err
}

// Clone clones a git repository into a temporary directory and returns the directory path
func Clone(gitConfig manifest.Git) (string, error) {
	tempDir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return "", err
	}

	if err := CloneInto(gitConfig, tempDir); err != nil {
		return "", err
	}

	return tempDir, nil
}
