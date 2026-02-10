package multidisc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateM3U creates M3U playlist content for a multi-disc set.
// If useSubdir is true, paths are relative to a subdirectory named after the game.
func GenerateM3U(set MultiDiscSet, ext string, useSubdir bool) string {
	var lines []string
	for _, disc := range set.Files {
		filename := filepath.Base(disc.Path)
		// If the files will be converted to CHD, update the extension
		if ext != "" {
			oldExt := filepath.Ext(filename)
			filename = strings.TrimSuffix(filename, oldExt) + ext
		}
		if useSubdir {
			filename = filepath.Join(set.BaseName, filename)
		}
		lines = append(lines, filename)
	}
	return strings.Join(lines, "\n") + "\n"
}

// WriteM3U writes an M3U file to disk.
func WriteM3U(dir string, set MultiDiscSet, ext string, useSubdir bool) (string, error) {
	content := GenerateM3U(set, ext, useSubdir)
	m3uPath := filepath.Join(dir, set.BaseName+".m3u")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(m3uPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write M3U: %w", err)
	}

	return m3uPath, nil
}
