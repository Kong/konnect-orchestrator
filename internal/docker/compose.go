package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ComposeInstance struct {
	ProjectName string
	ComposeFile string
	Directory   string
}

func ComposeUp(ctx context.Context, file, projectName string, additionalArgs ...string) (*ComposeInstance, error) {
	tmpDir, err := os.MkdirTemp("", projectName+"-*")
	if err != nil {
		return nil, fmt.Errorf("ComposeUp: %w", err)
	}

	composeFile := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composeFile, []byte(file), 0644); err != nil {
		return nil, fmt.Errorf("ComposeUp: %w", err)
	}

	instance := &ComposeInstance{
		ProjectName: projectName,
		ComposeFile: composeFile,
		Directory:   tmpDir,
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
	fullArgs := append([]string{"--project-name", instance.ProjectName, "--file", "docker-compose.yaml"}, args...)
	cmd := exec.CommandContext(ctx, "docker", append([]string{"compose"}, fullArgs...)...)
	cmd.Dir = instance.Directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running: docker %s\n", strings.Join(cmd.Args[1:], " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("runCompose: %w", err)
	}
	return nil
}
