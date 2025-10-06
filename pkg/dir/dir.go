// Package dir provides utilities for directory comparison.
package dir

import (
	"fmt"
	"os"
)

// DiffResult holds the results of directory comparison.
type DiffResult struct {
	New     []string
	Updated []string
	Deleted []string
}

// CompareDirs compares two directories and returns the differences.
// source: path to the source directory
// target: path to the target directory
// deleteMissing: if true, files present in target but missing in source are marked for deletion
// returns a DiffResult containing lists of new, updated, and deleted files, along with any errors encountered.
func CompareDirs(
	source, target string,
	deleteMissing bool,
) (DiffResult, []error) {
	errors := []error{}
	diffResult := DiffResult{}
	// read directories and create maps of file names to os.DirEntrys
	sourceFiles, err := os.ReadDir(source)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to read source directory: %v", err))
		return diffResult, errors
	}

	sourceFileMap := make(map[string]os.DirEntry)
	for _, srcFile := range sourceFiles {
		sourceFileMap[srcFile.Name()] = srcFile
	}

	targetFiles, err := os.ReadDir(target)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to read target directory: %v", err))
		return diffResult, errors
	}

	targetFileMap := make(map[string]os.DirEntry)
	for _, tgtFile := range targetFiles {
		targetFileMap[tgtFile.Name()] = tgtFile
	}

	// walk source files: classify new or updated
	for _, srcFile := range sourceFiles {
		tgtFile, exists := targetFileMap[srcFile.Name()]
		if !exists {
			diffResult.New = append(diffResult.New, srcFile.Name())
			continue
		}
		// both exist, check if updated
		srcInfo, err := srcFile.Info()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get info for source file %s: %v", srcFile.Name(), err))
			continue
		}
		tgtInfo, err := tgtFile.Info()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get info for target file %s: %v", tgtFile.Name(), err))
			continue
		}

		if srcInfo.Size() != tgtInfo.Size() || srcInfo.ModTime().After(tgtInfo.ModTime()) {
			diffResult.Updated = append(diffResult.Updated, srcFile.Name())
		}
	}

	if !deleteMissing {
		return diffResult, errors
	}

	// walk target files: classify deleted
	for _, tgtFile := range targetFiles {
		_, exists := sourceFileMap[tgtFile.Name()]
		if !exists {
			diffResult.Deleted = append(diffResult.Deleted, tgtFile.Name())
		}
	}

	return diffResult, nil
}
