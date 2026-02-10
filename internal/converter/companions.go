package converter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CompanionFiles returns all files associated with a disc image, including
// the disc image itself. For GDI/CUE files this includes referenced track
// files. For ISO files, returns just the ISO path. All returned paths are absolute.
func CompanionFiles(discImagePath string) ([]string, error) {
	absPath, err := filepath.Abs(discImagePath)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(absPath))
	switch ext {
	case ".gdi":
		return parseGDI(absPath)
	case ".cue":
		return parseCUE(absPath)
	default:
		return []string{absPath}, nil
	}
}

// parseGDI parses a GDI file and returns all referenced track files plus the
// GDI file itself. GDI format: first line is track count, subsequent lines
// have track info where field index 4 is the filename.
func parseGDI(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open GDI: %w", err)
	}
	defer f.Close()

	dir := filepath.Dir(path)
	files := []string{path}
	seen := map[string]bool{path: true}

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum == 1 {
			continue // skip track count line
		}

		fields := strings.Fields(scanner.Text())
		if len(fields) < 5 {
			continue
		}

		trackFile := filepath.Join(dir, fields[4])
		if !seen[trackFile] {
			seen[trackFile] = true
			files = append(files, trackFile)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read GDI: %w", err)
	}

	return files, nil
}

// cueFileRe matches FILE "filename" BINARY/WAVE lines in CUE sheets.
var cueFileRe = regexp.MustCompile(`(?i)^\s*FILE\s+"([^"]+)"`)

// parseCUE parses a CUE sheet and returns all referenced files plus the
// CUE file itself.
func parseCUE(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open CUE: %w", err)
	}
	defer f.Close()

	dir := filepath.Dir(path)
	files := []string{path}
	seen := map[string]bool{path: true}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		matches := cueFileRe.FindStringSubmatch(scanner.Text())
		if len(matches) < 2 {
			continue
		}

		trackFile := filepath.Join(dir, matches[1])
		if !seen[trackFile] {
			seen[trackFile] = true
			files = append(files, trackFile)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read CUE: %w", err)
	}

	return files, nil
}
