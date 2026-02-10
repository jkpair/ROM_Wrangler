package organizer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// PlanResult holds the outcome of executing a sort plan.
type PlanResult struct {
	FilesCopied  int
	FilesMoved   int
	M3UsWritten  int
	DirsCreated  int
	Errors       []error
}

// ExecutePlan carries out the sort plan by copying files and writing M3Us.
func ExecutePlan(plan *SortPlan, move bool, progressFn func(current, total int, filename string)) (*PlanResult, error) {
	result := &PlanResult{}
	total := len(plan.Files) + len(plan.M3Us)
	current := 0

	// Create directories
	for _, dir := range plan.DirsToCreate {
		if err := os.MkdirAll(dir, 0755); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("mkdir %s: %w", dir, err))
			continue
		}
		result.DirsCreated++
	}

	// Track source dirs for empty-dir cleanup when moving
	sourceDirs := make(map[string]bool)

	// Copy/move files
	for _, action := range plan.Files {
		current++
		if progressFn != nil {
			progressFn(current, total, filepath.Base(action.SourcePath))
		}

		// Skip no-op when source and dest are the same path
		absSrc, _ := filepath.Abs(action.SourcePath)
		absDst, _ := filepath.Abs(action.DestPath)
		if absSrc == absDst {
			continue
		}

		if move {
			sourceDirs[filepath.Dir(action.SourcePath)] = true
			if err := moveFile(action.SourcePath, action.DestPath); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("move %s: %w", action.SourcePath, err))
				continue
			}
			result.FilesMoved++
		} else {
			if err := copyFile(action.SourcePath, action.DestPath); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("copy %s: %w", action.SourcePath, err))
				continue
			}
			result.FilesCopied++
		}
	}

	// Write M3U files
	for _, m3u := range plan.M3Us {
		current++
		if progressFn != nil {
			progressFn(current, total, filepath.Base(m3u.Path))
		}

		dir := filepath.Dir(m3u.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("mkdir for m3u %s: %w", m3u.Path, err))
			continue
		}

		if err := os.WriteFile(m3u.Path, []byte(m3u.Content), 0644); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("write m3u %s: %w", m3u.Path, err))
			continue
		}
		result.M3UsWritten++
	}

	// Clean up empty source directories after moves
	if move {
		for dir := range sourceDirs {
			removeEmptyDirs(dir)
		}
	}

	return result, nil
}

// removeEmptyDirs removes a directory and its parents if they are empty.
// Stops at the first non-empty directory.
func removeEmptyDirs(dir string) {
	for {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			return
		}
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return dstFile.Sync()
}
