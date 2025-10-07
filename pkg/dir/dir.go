// Package dir provides utilities for directory comparison.
package dir

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
	deepSearch bool,
) (DiffResult, []error) {
	var (
		errors        []error
		diffResult    DiffResult
		sourceFileMap map[string]os.FileInfo
		targetFileMap map[string]os.FileInfo
		errs          []error
	)

	if deepSearch {
		sourceFileMap, targetFileMap, errs = listFilesDeep(source, target)
		if len(errs) > 0 {
			errors = append(errors, errs...)
			return diffResult, errors
		}
	} else {
		sourceFileMap, targetFileMap, errs = listFilesShallow(source, target)
		if len(errs) > 0 {
			errors = append(errors, errs...)
			return diffResult, errors
		}
	}

	// walk source files: classify new or updated
	for srcPath, srcFile := range sourceFileMap {
		tgtFile, exists := targetFileMap[srcPath]
		if !exists {
			diffResult.New = append(diffResult.New, srcPath)
			continue
		}

		if srcFile.Size() != tgtFile.Size() || srcFile.ModTime().After(tgtFile.ModTime()) {
			diffResult.Updated = append(diffResult.Updated, srcPath)
		}
	}

	if !deleteMissing {
		return diffResult, errors
	}

	// walk target files: classify deleted
	for tgtPath := range targetFileMap {
		_, exists := sourceFileMap[tgtPath]
		if !exists {
			diffResult.Deleted = append(diffResult.Deleted, tgtPath)
		}
	}

	return diffResult, nil
}

func listFilesShallow(source, target string) (map[string]os.FileInfo, map[string]os.FileInfo, []error) {
	errors := []error{}
	// read directories and create maps of file names to os.DirEntrys
	sourceFiles, err := os.ReadDir(source)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to read source directory: %v", err))
		return nil, nil, errors
	}

	sourceFileMap := make(map[string]os.FileInfo)
	for _, srcFile := range sourceFiles {
		if srcFile.IsDir() {
			// skip directories
			continue
		}
		sourceFileMap[srcFile.Name()], err = srcFile.Info()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get info for source file %s: %v", srcFile.Name(), err))
			continue
		}
	}

	targetFiles, err := os.ReadDir(target)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to read target directory: %v", err))
		return nil, nil, errors
	}

	targetFileMap := make(map[string]os.FileInfo)
	for _, tgtFile := range targetFiles {
		if tgtFile.IsDir() {
			// skip directories
			continue
		}
		targetFileMap[tgtFile.Name()], err = tgtFile.Info()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get info for target file %s: %v", tgtFile.Name(), err))
			continue
		}
	}
	return sourceFileMap, targetFileMap, nil
}

func listFilesDeep(source, target string) (map[string]os.FileInfo, map[string]os.FileInfo, []error) {
	errors := []error{}
	srcFilesPath := []string{}
	sourceFileMap := make(map[string]os.FileInfo)
	tgtFilesPath := []string{}
	targetFileMap := make(map[string]os.FileInfo)

	// walk source directory to get all file paths
	err := filepath.WalkDir(source, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			srcFilesPath = append(srcFilesPath, path)
		}
		return nil
	})
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to walk source directory: %v", err))
	}

	for _, filePath := range srcFilesPath {
		sf, err := os.Stat(filePath)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to stat source file %s: %v", filePath, err))
			continue
		}
		sourceFileMap[filePath[len(source)+1:]] = sf
	}

	// walk target directory to compare files
	err = filepath.WalkDir(target, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			tgtFilesPath = append(tgtFilesPath, path)
		}
		return nil
	})
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to walk target directory: %v", err))
	}

	for _, filePath := range tgtFilesPath {
		tf, err := os.Stat(filePath)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to stat target file %s: %v", filePath, err))
			continue
		}
		targetFileMap[filePath[len(target)+1:]] = tf
	}

	return sourceFileMap, targetFileMap, errors
}
