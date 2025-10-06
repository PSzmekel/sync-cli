// file/file_test.go
package file

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func haveCmd(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func requireUnixCoreutilsOrSkip(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("Tests require Unix-like coreutils: cp/mkdir/rm") // coreutils-dependent
	}
	// The implementation invokes "mkdir -p", "cp -a", "rm -f".
	// Check only the base commands; flags validation is not necessary here.
	for _, cmd := range []string{"mkdir", "cp", "rm"} {
		if !haveCmd(cmd) {
			t.Skipf("Skipping: required command %q not found", cmd)
		}
	}
}

func TestCopyFile_BasicCreatesDirsAndCopies(t *testing.T) {
	requireUnixCoreutilsOrSkip(t)

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source file with content.
	srcFile := "hello.txt"
	srcPath := filepath.Join(srcDir, srcFile)
	if err := os.WriteFile(srcPath, []byte("hello world"), 0o644); err != nil { //nolint: gosec
		t.Fatalf("write source: %v", err)
	}

	// Copy into nested directories under targetFile.
	targetFile := "nested/a/b/world.txt"
	ctx := context.Background()
	if err := CopyFile(ctx, srcDir, srcFile, dstDir, targetFile); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify content.
	got, err := os.ReadFile(filepath.Join(dstDir, targetFile)) //nolint: gosec
	if err != nil {
		t.Fatalf("read copied: %v", err)
	}
	if string(got) != "hello world" {
		t.Fatalf("content mismatch: got %q", string(got))
	}

	// Verify directories exist.
	if _, err := os.Stat(filepath.Join(dstDir, "nested", "a", "b")); err != nil {
		t.Fatalf("expected intermediate dirs to exist: %v", err)
	}
}

func TestCopyFile_ContextCanceled(t *testing.T) {
	requireUnixCoreutilsOrSkip(t)

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create a reasonably sized file; cancellation will be immediate anyway.
	srcFile := "big.bin"
	data := strings.Repeat("x", 5<<20)                                                        // 5 MiB
	if err := os.WriteFile(filepath.Join(srcDir, srcFile), []byte(data), 0o644); err != nil { //nolint: gosec
		t.Fatalf("write source: %v", err)
	}

	// Context canceled before invocation: CommandContext should kill the process promptly.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := CopyFile(ctx, srcDir, srcFile, dstDir, "out/xx.bin")
	if err == nil {
		t.Fatalf("expected error due to canceled context")
	}
	// Accept any non-nil error; depending on platform it may be context or process error.
}

func TestCopyFile_PermissionDeniedOnTargetParents(t *testing.T) {
	requireUnixCoreutilsOrSkip(t)

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcFile := "hello.txt"
	if err := os.WriteFile(filepath.Join(srcDir, srcFile), []byte("abc"), 0o644); err != nil { //nolint: gosec
		t.Fatalf("write source: %v", err)
	}

	// Create a non-writable dir under dstDir to force mkdir -p failure for deeper parents.
	block := filepath.Join(dstDir, "blocked")
	if err := os.Mkdir(block, 0o555); err != nil { //nolint: gosec
		t.Fatalf("mkdir blocked: %v", err)
	}

	// Attempt to create nested path under the read-only directory.
	targetFile := "blocked/x/y/z.bin"
	ctx := context.Background()
	err := CopyFile(ctx, srcDir, srcFile, dstDir, targetFile)
	if err == nil {
		t.Fatalf("expected error due to permission denial when creating parents")
	}
}

func TestDeleteFile_RemovesExisting(t *testing.T) {
	requireUnixCoreutilsOrSkip(t)

	dir := t.TempDir()
	p := filepath.Join(dir, "todelete.txt")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil { //nolint: gosec
		t.Fatalf("write file: %v", err)
	}
	ctx := context.Background()
	if err := DeleteFile(ctx, p); err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}
	if _, err := os.Stat(p); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected file to be gone, got err=%v", err)
	}
}

func TestDeleteFile_NonexistentOK(t *testing.T) {
	requireUnixCoreutilsOrSkip(t)

	ctx := context.Background()
	// Non-existing path; rm -f should not fail.
	if err := DeleteFile(ctx, filepath.Join(t.TempDir(), "missing.txt")); err != nil {
		t.Fatalf("expected nil error on non-existent file, got %v", err)
	}
}

func TestCopyFile_ContextTimeout(t *testing.T) {
	requireUnixCoreutilsOrSkip(t)

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create a larger file to increase the chance cp is running when timeout fires.
	srcFile := "large.bin"
	data := strings.Repeat("y", 20<<20)                                                       // 20 MiB
	if err := os.WriteFile(filepath.Join(srcDir, srcFile), []byte(data), 0o644); err != nil { //nolint: gosec
		t.Fatalf("write source: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err := CopyFile(ctx, srcDir, srcFile, dstDir, "deep/large_copied.bin")
	if err == nil {
		t.Fatalf("expected error due to context deadline")
	}
}
