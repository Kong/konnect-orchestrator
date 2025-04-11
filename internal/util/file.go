package util

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CopyFile(embedFS embed.FS, srcPath, dstPath string) error {
	// Open the embedded file
	srcFile, err := embedFS.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open embedded file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", filepath.Dir(dstPath), err)
	}

	// Create the destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	// Copy file contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy data to %s: %w", dstPath, err)
	}

	return nil
}

// copyEmbeddedFilesRecursive recursively copies files from an embedded FS to the target directory
func CopyEmbeddedFilesRecursive(embedFS embed.FS, srcDir, dstDir string) error {
	entries, err := embedFS.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", srcDir, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			// Ensure the destination directory exists
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dstPath, err)
			}
			// Recurse into subdirectory
			if err := CopyEmbeddedFilesRecursive(embedFS, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy individual file
			if err := CopyFile(embedFS, srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}
