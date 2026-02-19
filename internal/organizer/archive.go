package organizer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kurlmarx/romwrangler/internal/converter"
)

// ArchiveAction describes moving a single file to the archive.
type ArchiveAction struct {
	SourcePath  string
	ArchivePath string
}

// ArchivePlan describes all files to archive after conversion.
type ArchivePlan struct {
	Actions     []ArchiveAction
	ArchiveRoot string
}

// ArchiveResult holds the outcome of executing an archive plan.
type ArchiveResult struct {
	FilesMoved int
	Errors     []error
}

// BuildArchivePlan creates a plan to move original disc image files (and their
// companions) into an archive directory, preserving the folder structure
// relative to the source roots.
func BuildArchivePlan(sourceRoots []string, convertResults []converter.ConvertResult, archiveDir string) *ArchivePlan {
	plan := &ArchivePlan{ArchiveRoot: archiveDir}

	for _, cr := range convertResults {
		if cr.Err != nil {
			continue
		}

		companions, err := converter.CompanionFiles(cr.InputPath)
		if err != nil {
			companions = []string{cr.InputPath}
		}

		for _, filePath := range companions {
			relPath := computeRelativePath(sourceRoots, filePath)
			archivePath := filepath.Join(archiveDir, relPath)

			plan.Actions = append(plan.Actions, ArchiveAction{
				SourcePath:  filePath,
				ArchivePath: archivePath,
			})
		}
	}

	return plan
}

// computeRelativePath finds which source root a file belongs to and returns
// its path relative to that root. If no root matches, returns the base filename.
func computeRelativePath(sourceRoots []string, filePath string) string {
	for _, root := range sourceRoots {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		absFile, err := filepath.Abs(filePath)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absFile, absRoot+string(filepath.Separator)) {
			rel, err := filepath.Rel(absRoot, absFile)
			if err == nil {
				return rel
			}
		}
	}
	return filepath.Base(filePath)
}

// ExecuteArchive moves files according to the archive plan. Uses os.Rename for
// same-filesystem moves, falling back to copy+delete for cross-device.
func ExecuteArchive(plan *ArchivePlan, progressFn func(current, total int, filename string)) *ArchiveResult {
	result := &ArchiveResult{}
	total := len(plan.Actions)

	for i, action := range plan.Actions {
		if progressFn != nil {
			progressFn(i+1, total, filepath.Base(action.SourcePath))
		}

		if err := os.MkdirAll(filepath.Dir(action.ArchivePath), 0755); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("mkdir %s: %w", action.ArchivePath, err))
			continue
		}

		if err := moveFile(action.SourcePath, action.ArchivePath); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("move %s: %w", action.SourcePath, err))
			continue
		}
		result.FilesMoved++
	}

	return result
}

// ArchiveFilteredFiles moves dedup-filtered variant files to the archive
// directory, preserving their relative path from the source roots.
func ArchiveFilteredFiles(paths []string, sourceRoots []string, archiveDir string) *ArchiveResult {
	result := &ArchiveResult{}

	for _, filePath := range paths {
		relPath := computeRelativePath(sourceRoots, filePath)
		archivePath := filepath.Join(archiveDir, relPath)

		if err := os.MkdirAll(filepath.Dir(archivePath), 0755); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("mkdir %s: %w", archivePath, err))
			continue
		}

		if err := moveFile(filePath, archivePath); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("move %s: %w", filePath, err))
			continue
		}
		result.FilesMoved++
	}

	return result
}

// ArchiveUnsupported moves unresolved and unsupported files to the archive
// directory, preserving their relative path from the source roots.
func ArchiveUnsupported(unresolved, unsupported []string, sourceRoots []string, archiveDir string) *ArchiveResult {
	result := &ArchiveResult{}
	all := make([]string, 0, len(unresolved)+len(unsupported))
	all = append(all, unresolved...)
	all = append(all, unsupported...)

	for _, filePath := range all {
		relPath := computeRelativePath(sourceRoots, filePath)
		archivePath := filepath.Join(archiveDir, relPath)

		if err := os.MkdirAll(filepath.Dir(archivePath), 0755); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("mkdir %s: %w", archivePath, err))
			continue
		}

		if err := moveFile(filePath, archivePath); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("move %s: %w", filePath, err))
			continue
		}
		result.FilesMoved++
	}

	return result
}

// moveFile attempts os.Rename first, falling back to copy+delete for
// cross-device moves.
func moveFile(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Fall back to copy + delete
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	if err := dstFile.Sync(); err != nil {
		return err
	}

	srcFile.Close()
	return os.Remove(src)
}
