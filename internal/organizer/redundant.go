package organizer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/converter"
)

// FindSupersededDiscImages walks source directories and returns paths of
// disc image files (.cue/.gdi and their companion .bin tracks) where a
// .chd conversion already exists. Skips _archive directories.
func FindSupersededDiscImages(dirs []string, aliases map[string]string) []string {
	seen := make(map[string]bool)
	var redundant []string

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == "_archive" {
				continue
			}

			if _, ok := config.ResolveAlias(entry.Name(), aliases); !ok {
				continue
			}

			subPath := filepath.Join(dir, entry.Name())
			filepath.Walk(subPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() && info.Name() == "_archive" {
					return filepath.SkipDir
				}
				if info.IsDir() {
					return nil
				}
				if !converter.IsConvertible(path) {
					return nil
				}

				chdPath := converter.OutputPath(path)
				if _, err := os.Stat(chdPath); err == nil {
					// CHD exists — mark disc image and companions as redundant
					companions, cerr := converter.CompanionFiles(path)
					if cerr != nil {
						if !seen[path] {
							seen[path] = true
							redundant = append(redundant, path)
						}
						return nil
					}
					for _, c := range companions {
						if !seen[c] {
							seen[c] = true
							redundant = append(redundant, c)
						}
					}
				}
				return nil
			})
		}
	}

	return redundant
}

// FindExtractedArchives walks source directories and returns paths of
// archive files (.zip/.7z/.rar) that appear to have already been extracted
// (a folder with the same base name exists alongside the archive).
// Skips _archive directories.
func FindExtractedArchives(dirs []string, aliases map[string]string) []string {
	archiveExts := map[string]bool{".zip": true, ".7z": true, ".rar": true}
	var redundant []string

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == "_archive" {
				continue
			}

			if _, ok := config.ResolveAlias(entry.Name(), aliases); !ok {
				continue
			}

			subPath := filepath.Join(dir, entry.Name())
			filepath.Walk(subPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() && info.Name() == "_archive" {
					return filepath.SkipDir
				}
				if info.IsDir() {
					return nil
				}

				ext := strings.ToLower(filepath.Ext(path))
				if !archiveExts[ext] {
					return nil
				}

				// Check if a folder with the archive's base name exists alongside
				baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
				extractedDir := filepath.Join(filepath.Dir(path), baseName)
				if fi, ferr := os.Stat(extractedDir); ferr == nil && fi.IsDir() {
					redundant = append(redundant, path)
				}

				return nil
			})
		}
	}

	return redundant
}
