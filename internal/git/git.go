package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// getAuthMethod returns an ssh.AuthMethod based on the git config, or nil if no auth is specified
func getAuthMethod(gitConfig manifest.GitConfig) (ssh.AuthMethod, error) {
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

func GetRemoteFile(gitConfig manifest.GitConfig, branch, path string) ([]byte, error) {
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
func CloneInto(gitConfig manifest.GitConfig, dir string) error {

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
func Clone(gitConfig manifest.GitConfig) (string, error) {
	tempDir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return "", err
	}

	if err := CloneInto(gitConfig, tempDir); err != nil {
		return "", err
	}

	return tempDir, nil
}

func IsClean(dir string) (bool, error) {
	// Opens an already existing repository.
	r, err := git.PlainOpen(dir)
	if err != nil {
		return false, err
	}

	workTree, err := r.Worktree()
	if err != nil {
		return false, err
	}
	status, err := workTree.Status()
	if err != nil {
		return false, err
	}

	return status.IsClean(), nil
}
func Branch(dir string, branch string) error {
	r, err := git.PlainOpen(dir)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	branchRef := plumbing.NewBranchReferenceName(branch)
	branchCoOpts := git.CheckoutOptions{
		Branch: plumbing.ReferenceName(branchRef),
		Force:  true,
		Create: true,
	}
	err = w.Checkout(&branchCoOpts)
	if err != nil {
		return err
	}

	return nil
}

func Add(dir string, path string) error {
	r, err := git.PlainOpen(dir)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Add(path)
	if err != nil {
		return err
	}

	return nil
}

func Commit(dir string, message string, author manifest.Author) error {
	r, err := git.PlainOpen(dir)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  author.Name,
			Email: author.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	return nil
}
func Push(dir string, gitConfig manifest.GitConfig) error {
	r, err := git.PlainOpen(dir)
	if err != nil {
		return err
	}
	auth, err := getAuthMethod(gitConfig)
	if err != nil {
		return err
	}
	// push using default options
	err = r.Push(&git.PushOptions{
		Auth:       auth,
		Force:      true,
		RemoteName: "origin",
	})
	if err != nil {
		return err
	}

	return nil
}
