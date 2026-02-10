package converter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ConvertType determines which chdman command to use.
type ConvertType int

const (
	ConvertCD  ConvertType = iota // createcd (GDI, CUE+BIN)
	ConvertDVD                    // createdvd (ISO)
)

// ConvertResult holds the outcome of a single conversion.
type ConvertResult struct {
	InputPath  string
	OutputPath string
	Err        error
}

// FindChdman looks for the chdman binary in this order:
// 1. Explicit config path
// 2. PATH
// 3. Common install locations
func FindChdman(configPath string) (string, error) {
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	if path, err := exec.LookPath("chdman"); err == nil {
		return path, nil
	}

	commonPaths := []string{
		"/usr/bin/chdman",
		"/usr/local/bin/chdman",
		"/opt/mame/chdman",
	}
	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("chdman not found. Install MAME tools:\n  Arch/Manjaro: sudo pacman -S mame-tools\n  Ubuntu/Debian: sudo apt install mame-tools\n  Fedora: sudo dnf install mame-tools")
}

// DetectConvertType determines whether to use createcd or createdvd based on extension.
func DetectConvertType(inputPath string) ConvertType {
	ext := strings.ToLower(filepath.Ext(inputPath))
	switch ext {
	case ".gdi", ".cue":
		return ConvertCD
	default:
		return ConvertDVD
	}
}

// Convert runs chdman to convert a disc image to CHD format.
// The progressFn callback receives progress percentage (0-100).
func Convert(ctx context.Context, chdmanPath, inputPath, outputPath string, progressFn func(float64)) error {
	convType := DetectConvertType(inputPath)

	var cmd string
	switch convType {
	case ConvertCD:
		cmd = "createcd"
	case ConvertDVD:
		cmd = "createdvd"
	}

	args := []string{cmd, "-i", inputPath, "-o", outputPath}
	proc := exec.CommandContext(ctx, chdmanPath, args...)

	stderr, err := proc.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := proc.Start(); err != nil {
		return fmt.Errorf("failed to start chdman: %w", err)
	}

	// Parse progress from stderr
	parseProgress(stderr, progressFn)

	if err := proc.Wait(); err != nil {
		return fmt.Errorf("chdman failed: %w", err)
	}

	return nil
}

// OutputPath generates the CHD output path from an input path.
func OutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	return strings.TrimSuffix(inputPath, ext) + ".chd"
}

// IsConvertible returns true if the file extension can be converted to CHD.
func IsConvertible(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".gdi", ".cue", ".iso":
		return true
	default:
		return false
	}
}
