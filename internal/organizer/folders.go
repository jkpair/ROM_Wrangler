package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/kurlmarx/romwrangler/internal/systems"
)

// FolderStatus describes the state of a system folder.
type FolderStatus struct {
	System    systems.SystemID
	Folder    string
	FullPath  string
	Exists    bool
	FileCount int
}

// CheckFolders checks which ReplayOS system folders exist under baseDir.
func CheckFolders(baseDir string) []FolderStatus {
	var statuses []FolderStatus

	for id, folder := range systems.ReplayOSFolders {
		fullPath := filepath.Join(baseDir, folder)
		status := FolderStatus{
			System:   id,
			Folder:   folder,
			FullPath: fullPath,
		}

		info, err := os.Stat(fullPath)
		if err == nil && info.IsDir() {
			status.Exists = true
			entries, _ := os.ReadDir(fullPath)
			status.FileCount = len(entries)
		}

		statuses = append(statuses, status)
	}

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Folder < statuses[j].Folder
	})

	return statuses
}

// GenerateAllFolders creates all ReplayOS system folders under baseDir.
// Returns the number of folders created and any errors.
func GenerateAllFolders(baseDir string) (created int, errs []error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return 0, []error{fmt.Errorf("create base dir: %w", err)}
	}

	for _, folder := range systems.ReplayOSFolders {
		fullPath := filepath.Join(baseDir, folder)
		if _, err := os.Stat(fullPath); err == nil {
			continue // already exists
		}
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			errs = append(errs, fmt.Errorf("create %s: %w", folder, err))
			continue
		}
		created++
	}
	return created, errs
}
