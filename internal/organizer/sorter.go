package organizer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kurlmarx/romwrangler/internal/devices"
	"github.com/kurlmarx/romwrangler/internal/multidisc"
	"github.com/kurlmarx/romwrangler/internal/systems"
)

// FileAction describes what to do with a single file.
type FileAction struct {
	SourcePath string
	DestPath   string
	System     systems.SystemID
}

// M3UAction describes an M3U file to create.
type M3UAction struct {
	Path    string
	Content string
	System  systems.SystemID
}

// SortPlan describes all operations needed to organize ROMs.
type SortPlan struct {
	Files     []FileAction
	M3Us      []M3UAction
	DirsToCreate []string
}

// BuildSortPlan creates a plan to organize scanned files for a device.
func BuildSortPlan(scanResult *ScanResult, device devices.Device, outputDir string, cleanNames bool) *SortPlan {
	plan := &SortPlan{}
	dirsNeeded := make(map[string]bool)

	for systemID, files := range scanResult.BySystem {
		folder, ok := device.FolderForSystem(systemID)
		if !ok {
			continue
		}

		destDir := filepath.Join(outputDir, folder)
		dirsNeeded[destDir] = true

		// Deduplicate: if a .chd exists, skip the pre-conversion source
		// file with the same base name (.cue, .gdi, .bin, .iso).
		chdSet := make(map[string]bool)
		for _, f := range files {
			if strings.ToLower(filepath.Ext(f.Path)) == ".chd" {
				base := strings.TrimSuffix(f.Path, filepath.Ext(f.Path))
				chdSet[base] = true
			}
		}

		var filePaths []string
		for _, f := range files {
			ext := strings.ToLower(filepath.Ext(f.Path))
			if ext != ".chd" {
				base := strings.TrimSuffix(f.Path, filepath.Ext(f.Path))
				if chdSet[base] {
					continue // skip, CHD version exists
				}
			}
			filePaths = append(filePaths, f.Path)
		}

		// Detect multi-disc sets
		sets, standalone := multidisc.DetectSets(filePaths)

		// Handle standalone files
		for _, path := range standalone {
			name := filepath.Base(path)
			if cleanNames {
				ext := filepath.Ext(name)
				nameNoExt := name[:len(name)-len(ext)]
				name = CleanFilename(nameNoExt) + ext
			}

			plan.Files = append(plan.Files, FileAction{
				SourcePath: path,
				DestPath:   filepath.Join(destDir, name),
				System:     systemID,
			})
		}

		// Handle multi-disc sets
		for _, set := range sets {
			for _, disc := range set.Files {
				name := filepath.Base(disc.Path)
				if cleanNames {
					ext := filepath.Ext(name)
					nameNoExt := name[:len(name)-len(ext)]
					name = CleanFilename(nameNoExt) + ext
				}

				plan.Files = append(plan.Files, FileAction{
					SourcePath: disc.Path,
					DestPath:   filepath.Join(destDir, name),
					System:     systemID,
				})
			}

			// Generate M3U (skip if already exists)
			m3uPath := filepath.Join(destDir, set.BaseName+".m3u")
			if _, err := os.Stat(m3uPath); err != nil {
				m3uContent := multidisc.GenerateM3U(set, "", false)
				plan.M3Us = append(plan.M3Us, M3UAction{
					Path:    m3uPath,
					Content: m3uContent,
					System:  systemID,
				})
			}
		}
	}

	for dir := range dirsNeeded {
		plan.DirsToCreate = append(plan.DirsToCreate, dir)
	}

	return plan
}
