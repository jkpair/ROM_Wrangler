package organizer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/converter"
	"github.com/kurlmarx/romwrangler/internal/systems"
)

// skipFolders contains directory names that should be skipped during scanning.
// These include the archive folder and ReplayOS special folders that contain
// system files (.lr, .sh, .rec) rather than ROMs.
var skipFolders = map[string]bool{
	"_archive":   true,
	"_extra":     true,
	"_recent":    true,
	"_favorites": true,
	"_autostart": true,
}

// ScannedFile represents a discovered file in the source directories.
type ScannedFile struct {
	Path     string
	System   systems.SystemID
	Resolved bool // true if system was resolved via alias
}

// ScanResult holds all discovered files grouped by system.
type ScanResult struct {
	Files       []ScannedFile
	BySystem    map[systems.SystemID][]ScannedFile
	Convertible []ScannedFile // disc images that can be converted to CHD
	Unresolved  []string      // files in dirs that couldn't be matched to a system
	Unsupported []string      // files in system dirs with unsupported extensions
	Errors      []error
}

// Scan walks source directories and builds a file inventory.
// Subdirectory names are resolved to systems via aliases.
func Scan(dirs []string, aliases map[string]string) *ScanResult {
	result := &ScanResult{
		BySystem: make(map[systems.SystemID][]ScannedFile),
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			result.Errors = append(result.Errors, err)
			continue
		}

		for _, entry := range entries {
			if skipFolders[entry.Name()] {
				continue // skip archive and special directories
			}
			subPath := filepath.Join(dir, entry.Name())

			if entry.IsDir() {
				// Try to resolve directory name to a system
				systemID, ok := config.ResolveAlias(entry.Name(), aliases)
				if !ok {
					// Walk dir but mark files as unresolved
					filepath.Walk(subPath, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return nil
						}
						if info.IsDir() && skipFolders[info.Name()] {
							return filepath.SkipDir
						}
						if info.IsDir() {
							return nil
						}
						result.Unresolved = append(result.Unresolved, path)
						return nil
					})
					continue
				}

				// Walk the system directory
				filepath.Walk(subPath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return nil
					}
					if info.IsDir() && skipFolders[info.Name()] {
						return filepath.SkipDir
					}
					if info.IsDir() {
						return nil
					}

					ext := strings.ToLower(filepath.Ext(path))
					if !systems.IsValidFormat(systemID, ext) {
						result.Unsupported = append(result.Unsupported, path)
						return nil
					}

					sf := ScannedFile{
						Path:     path,
						System:   systemID,
						Resolved: true,
					}
					result.Files = append(result.Files, sf)
					result.BySystem[systemID] = append(result.BySystem[systemID], sf)
					return nil
				})
			} else {
				// Files at root level are unresolved
				result.Unresolved = append(result.Unresolved, subPath)
			}
		}
	}

	// Populate Convertible: disc-based system files that can be converted to CHD.
	// Skip files whose .chd output already exists (already converted).
	for _, f := range result.Files {
		if !converter.IsConvertible(f.Path) {
			continue
		}
		info, ok := systems.AllSystems[f.System]
		if !ok || !info.IsDiscBased {
			continue
		}
		if _, err := os.Stat(converter.OutputPath(f.Path)); err == nil {
			continue // .chd already exists
		}
		result.Convertible = append(result.Convertible, f)
	}

	return result
}

// UpdateForConversions updates the scan result after CHD conversions complete.
// For each successful conversion, the source file entry is replaced with the
// .chd output, and companion track files are removed from Files/BySystem.
func (sr *ScanResult) UpdateForConversions(results []converter.ConvertResult) {
	// Build lookup of source paths to their CHD outputs
	converted := make(map[string]string) // inputPath -> outputPath
	for _, cr := range results {
		if cr.Err != nil {
			continue
		}
		converted[cr.InputPath] = cr.OutputPath
	}

	// Gather all companion files that should be removed (track files embedded in CHD)
	removeSet := make(map[string]bool)
	for inputPath := range converted {
		companions, err := converter.CompanionFiles(inputPath)
		if err != nil {
			continue
		}
		for _, c := range companions {
			removeSet[c] = true
		}
	}

	// Rebuild Files: replace converted entries with .chd, remove companions
	var newFiles []ScannedFile
	for _, f := range sr.Files {
		if outputPath, ok := converted[f.Path]; ok {
			// Replace with CHD
			newFiles = append(newFiles, ScannedFile{
				Path:     outputPath,
				System:   f.System,
				Resolved: f.Resolved,
			})
		} else if removeSet[f.Path] {
			// Skip companion track files
			continue
		} else {
			newFiles = append(newFiles, f)
		}
	}
	sr.Files = newFiles

	// Rebuild BySystem
	sr.BySystem = make(map[systems.SystemID][]ScannedFile)
	for _, f := range sr.Files {
		sr.BySystem[f.System] = append(sr.BySystem[f.System], f)
	}

	// Clear convertible list
	sr.Convertible = nil
}

// RemoveFailedConversions removes files associated with failed conversions
// from the scan result. This prevents the sorter from flattening unconverted
// disc images whose track files would collide at the destination.
func (sr *ScanResult) RemoveFailedConversions(results []converter.ConvertResult) {
	// Collect input paths of failed conversions
	var failedPaths []string
	for _, cr := range results {
		if cr.Err != nil {
			failedPaths = append(failedPaths, cr.InputPath)
		}
	}

	if len(failedPaths) == 0 {
		return
	}

	// Gather all files to remove: the disc image + its companion tracks
	removeSet := make(map[string]bool)
	for _, inputPath := range failedPaths {
		companions, err := converter.CompanionFiles(inputPath)
		if err != nil {
			removeSet[inputPath] = true
			continue
		}
		for _, c := range companions {
			removeSet[c] = true
		}
	}

	// Rebuild Files excluding removed paths
	var newFiles []ScannedFile
	for _, f := range sr.Files {
		if !removeSet[f.Path] {
			newFiles = append(newFiles, f)
		}
	}
	sr.Files = newFiles

	// Rebuild BySystem
	sr.BySystem = make(map[systems.SystemID][]ScannedFile)
	for _, f := range sr.Files {
		sr.BySystem[f.System] = append(sr.BySystem[f.System], f)
	}
}
