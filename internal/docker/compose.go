package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/creack/pty"
)

const KoctlRunComposeFile = `services:
  koctl-api:
    image: ghcr.io/kong/koctl:latest
    ports:
      - "8080:8080"
    environment:
      - GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID}
      - GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET}
      - PLATFORM_REPO_URL=${PLATFORM_REPO_URL}
      - PLATFORM_REPO_GITHUB_TOKEN=${PLATFORM_REPO_GITHUB_TOKEN}
      - FRONTEND_URL=http://localhost:8081
      - GITHUB_REDIRECT_URI=http://localhost:8080/auth/github/callback
    command: ["run", "api"]
  koctl-ui:
    image: ghcr.io/kong/koctl-ui:latest
    ports:
      - "8081:8081"
    environment:
      - VITE_API_BASE_URL=http://koctl-api:8080
    depends_on:
      - koctl-api
`

type ComposeInstance struct {
	ProjectName string
	ComposeFile string
	Directory   string
	EnvVars     map[string]string
}

func ComposeUp(ctx context.Context, file, projectName string,
	envVars map[string]string, additionalArgs ...string,
) (*ComposeInstance, error) {
	tmpDir, err := os.MkdirTemp("", projectName+"-*")
	if err != nil {
		return nil, fmt.Errorf("ComposeUp: %w", err)
	}

	composeFile := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composeFile, []byte(file), 0o600); err != nil {
		return nil, fmt.Errorf("ComposeUp: %w", err)
	}

	instance := &ComposeInstance{
		ProjectName: projectName,
		ComposeFile: composeFile,
		Directory:   tmpDir,
		EnvVars:     envVars,
	}
	newArgs := []string{"up", "-d", "--pull=always", "--remove-orphans"}
	args := append(newArgs, additionalArgs...)

	if err := runComposeCmd(ctx, instance, args...); err != nil {
		return nil, fmt.Errorf("ComposeUp: %w", err)
	}

	return instance, nil
}

func ComposeDown(ctx context.Context, instance *ComposeInstance) error {
	defer os.RemoveAll(instance.Directory)

	if err := runComposeCmd(ctx, instance, "down", "--remove-orphans"); err != nil {
		return fmt.Errorf("ComposeUp: %w", err)
	}

	return nil
}

func runComposeCmd(ctx context.Context, instance *ComposeInstance, args ...string) error {
	fullArgs := append([]string{
		"--project-name",
		instance.ProjectName,
		"--file", "docker-compose.yaml",
	},
		args...)
	cmd := exec.CommandContext(ctx, "docker", append([]string{"compose"}, fullArgs...)...)
	cmd.Dir = instance.Directory
	env := os.Environ()
	for k, v := range instance.EnvVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	fmt.Printf("Running: docker %s\n", strings.Join(cmd.Args[1:], " "))

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("runCompose: %w", err)
	}

	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
	}()
	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("runCompose: %w", err)
	}

	fmt.Fprint(os.Stdout, "\r\n")
	fmt.Fprintln(os.Stdout, "To stop the project, run: docker compose --project-name koctl-run down")

	return nil
}
