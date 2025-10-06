// Package file provides utilities for file operations.
package file

import (
	"context"
	"os/exec"
)

// CopyFile copies a file from sourcePath to targetPath.
func CopyFile(ctx context.Context, sourcePath, targetPath string) error {
	return exec.CommandContext(ctx, "cp", "-a", "--", sourcePath, targetPath).Run()
}

// DeleteFile removes the specified file.
func DeleteFile(ctx context.Context, path string) error {
	return exec.CommandContext(ctx, "rm", "-f", "--", path).Run()
}
