package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// getAuthMethod returns an ssh.AuthMethod based on the git config, or nil if no auth is specified
func getAuthMethod(gitConfig manifest.GitConfig) (transport.AuthMethod, error) {
	// Return nil if no auth configured
	if gitConfig.Auth.Type == "" {
		return nil, nil
	}

	// Only support SSH auth for now
	if gitConfig.Auth.Type == "ssh" {
		key, err := util.ResolveSecretValue(gitConfig.Auth.SSH.Key)
		if err != nil {
			return nil, err
		}
		publicKeys, err := ssh.NewPublicKeys("git", []byte(key), "")
		if err != nil {
			return nil, err
		}
		return publicKeys, nil
	} else if gitConfig.Auth.Type == "token" {
		key, err := util.ResolveSecretValue(gitConfig.Auth.Token)
		if err != nil {
			return nil, err
		}
		bashAuth := &http.BasicAuth{
			Username: "x-access-token",
			Password: key,
		}
		return bashAuth, nil
	} else {
		return nil, errors.New("unsupported auth type: " + gitConfig.Auth.Type)
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

func CheckoutBranch(dir string, branch string) error {
	r, err := git.PlainOpen(dir)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	// Try to fetch the remote branch first
	err = r.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", branch, branch)),
		},
		Force: true,
	})
	// Only return error if it's not "already up to date" or "no matching ref spec"
	if err != nil && err != git.NoErrAlreadyUpToDate && !errors.Is(err, git.NoMatchingRefSpecError{}) {
		return fmt.Errorf("failed to fetch remote branch: %w", err)
	}

	// Check if remote branch exists
	remoteBranch := plumbing.NewRemoteReferenceName("origin", branch)
	remoteRef, err := r.Reference(remoteBranch, true)

	branchRef := plumbing.NewBranchReferenceName(branch)
	if err == nil {
		// Remote branch exists, create local branch from remote
		_, err = r.Reference(branchRef, true)
		if err == plumbing.ErrReferenceNotFound {
			// Create new local branch tracking remote branch
			headRef := plumbing.NewHashReference(branchRef, remoteRef.Hash())
			err = r.Storer.SetReference(headRef)
			if err != nil {
				return fmt.Errorf("failed to create local branch: %w", err)
			}
		}

		// Checkout the branch
		err = w.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
			Force:  true,
		})
		if err != nil {
			return fmt.Errorf("failed to checkout branch: %w", err)
		}
	} else {
		// Checkout the branch (will create if doesn't exist)
		err = w.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
			Force:  true,
			Create: true,
		})
		if err != nil {
			return fmt.Errorf("failed to checkout branch: %w", err)
		}

		// If remote branch exists, reset to its state
		if remoteRef != nil {
			err = w.Reset(&git.ResetOptions{
				Commit: remoteRef.Hash(),
				Mode:   git.HardReset,
			})
			if err != nil {
				return fmt.Errorf("failed to reset to remote branch: %w", err)
			}
		}
	}

	return nil
}
