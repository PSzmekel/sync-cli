// Package file provides utilities for file operations.
package file

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CopyFile copies a file from sourcePath to targetPath.
func CopyFile(ctx context.Context, sourcePath, sourceFile, targetPath, targetFile string) error {
	// split targetFile by "/" and create necessary directories
	targetDir := strings.Split(targetFile, "/")

	for i := 0; i < len(targetDir)-1; i++ {
		dir := strings.Join(targetDir[:i+1], "/")

		// check if dir exists and create if not
		dirPath := filepath.Join(targetPath, dir)

		stat, err := os.Stat(dirPath)
		if err != nil && stat == nil {
			err := exec.CommandContext(ctx, "mkdir", "-p", "--", dirPath).Run() //nolint: gosec
			if err != nil {
				return err
			}
		}
	}

	return exec.CommandContext( //nolint: gosec // uses external command to copy metadata
		ctx,
		"cp",
		"-a",
		"--",
		filepath.Join(sourcePath, sourceFile),
		filepath.Join(targetPath, targetFile),
	).
		Run()
}

// DeleteFile removes the specified file.
func DeleteFile(ctx context.Context, path string) error {
	return exec.CommandContext(ctx, "rm", "-f", "--", path).Run()
}
