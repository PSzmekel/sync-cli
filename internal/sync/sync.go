// Package sync provides synchronization functionalities.
package sync

import (
	"context"
	"log"
	"path/filepath"
	pkgDir "sync-cli/pkg/dir"
	pkgFile "sync-cli/pkg/file"
)

// SynchronizationDirectories synchronizes files from source to target directory.
func SynchronizationDirectories(ctx context.Context, source, target string, deleteMissing, deepSearch bool) error {
	results, errs := pkgDir.CompareDirs(source, target, deleteMissing, deepSearch)
	if len(errs) > 0 {
		for _, e := range errs {
			log.Printf("Error: %v\n", e)
		}
	} else if len(results.New) == 0 && len(results.Updated) == 0 && len(results.Deleted) == 0 {
		log.Println("No changes detected. Directories are already synchronized.")
		return nil
	}

	for _, file := range results.New {
		err := pkgFile.CopyFile(ctx, source, file, target, file)
		if err != nil {
			log.Printf("Failed to copy new file %s: %v\n", file, err)
			continue
		}
		log.Printf("New file: %s\n", file)
	}

	for _, file := range results.Updated {
		err := pkgFile.CopyFile(ctx, source, file, target, file)
		if err != nil {
			log.Printf("Failed to update file %s: %v\n", file, err)
			continue
		}
		log.Printf("Updated file: %s\n", file)
	}

	for _, file := range results.Deleted {
		err := pkgFile.DeleteFile(ctx, filepath.Join(target, file))
		if err != nil {
			log.Printf("Failed to delete file %s: %v\n", file, err)
			continue
		}
		log.Printf("Deleted file: %s\n", file)
	}

	return nil
}
