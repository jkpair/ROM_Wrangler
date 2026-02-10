package organizer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/systems"
)

// archiveExtensions lists file extensions treated as extractable archives.
var archiveExtensions = map[string]bool{
	".zip": true,
	".7z":  true,
	".rar": true,
	".ecm": true,
}

// ExtractableFile represents an archive found in a disc-based system folder.
type ExtractableFile struct {
	Path   string
	System systems.SystemID
}

// ExtractResult holds the outcome of extracting archives.
type ExtractResult struct {
	Extracted    int
	FilesCreated int
	Errors       []error
}

// Find7z looks for the 7z binary. Returns empty string if not found.
func Find7z() string {
	for _, name := range []string{"7z", "7zz", "7za"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

// FindUnecm looks for an unecm binary (unecm or ecm-uncompress).
// Returns empty string if not found. No longer required since ECM
// decompression is now handled natively in Go, but kept for reference.
func FindUnecm() string {
	for _, name := range []string{"unecm", "ecm-uncompress"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

// FindExtractable walks source directories and finds archive files (.zip,
// .7z, .rar) in any recognized system subdirectory.
func FindExtractable(dirs []string, aliases map[string]string) []ExtractableFile {
	var result []ExtractableFile

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == "_archive" {
				continue
			}

			systemID, ok := config.ResolveAlias(entry.Name(), aliases)
			if !ok {
				continue
			}

			if _, ok := systems.AllSystems[systemID]; !ok {
				continue
			}

			subPath := filepath.Join(dir, entry.Name())
			filepath.Walk(subPath, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return nil
				}
				ext := strings.ToLower(filepath.Ext(path))
				if archiveExtensions[ext] {
					result = append(result, ExtractableFile{
						Path:   path,
						System: systemID,
					})
				}
				return nil
			})
		}
	}

	return result
}

// ExtractArchives extracts all provided archive files into per-game subfolders.
func ExtractArchives(files []ExtractableFile, progressFn func(current, total int, filename string)) *ExtractResult {
	result := &ExtractResult{}
	total := len(files)
	sevenZPath := Find7z()

	for i, f := range files {
		if progressFn != nil {
			progressFn(i+1, total, filepath.Base(f.Path))
		}

		count, err := extractSingleArchive(f.Path, sevenZPath, "")
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", filepath.Base(f.Path), err))
			continue
		}
		result.Extracted++
		result.FilesCreated += count
	}

	return result
}

// ExtractAll performs iterative extraction: extracts archives, then re-scans
// for newly-exposed extractable files (e.g. .ecm inside a .rar), repeating
// until no new work is found. Returns the combined result and the full list
// of original archive files to be archived later (.zip, .rar, .7z only —
// intermediate .ecm files are cleaned up after decompression).
func ExtractAll(dirs []string, aliases map[string]string, progressFn func(current, total int, filename string)) (*ExtractResult, []ExtractableFile) {
	totalResult := &ExtractResult{}
	var allProcessed []ExtractableFile
	processed := make(map[string]bool)

	for {
		extractable := FindExtractable(dirs, aliases)

		// Filter out already-processed files
		var newFiles []ExtractableFile
		for _, f := range extractable {
			if !processed[f.Path] {
				newFiles = append(newFiles, f)
				processed[f.Path] = true
			}
		}

		if len(newFiles) == 0 {
			break
		}

		result := ExtractArchives(newFiles, progressFn)
		totalResult.Extracted += result.Extracted
		totalResult.FilesCreated += result.FilesCreated
		totalResult.Errors = append(totalResult.Errors, result.Errors...)

		// Only add real archives to the processed list for later archiving.
		// ECM files are intermediates (exposed by extracting a .rar/.zip)
		// and get cleaned up after decompression.
		for _, f := range newFiles {
			ext := strings.ToLower(filepath.Ext(f.Path))
			if ext != ".ecm" {
				allProcessed = append(allProcessed, f)
			}
		}
	}

	return totalResult, allProcessed
}

// extractSingleArchive dispatches to the appropriate extractor based on
// file extension. Uses Go stdlib for .zip and .ecm, 7z CLI for .7z and .rar.
func extractSingleArchive(archivePath, sevenZPath, _ string) (int, error) {
	ext := strings.ToLower(filepath.Ext(archivePath))

	switch ext {
	case ".zip":
		return extractSingleZip(archivePath)
	case ".7z", ".rar":
		if sevenZPath == "" {
			return 0, fmt.Errorf("7z not found; install p7zip to extract %s files", ext)
		}
		return extractWith7z(archivePath, sevenZPath)
	case ".ecm":
		return decompressEcm(archivePath, "")
	default:
		return 0, fmt.Errorf("unsupported archive format: %s", ext)
	}
}

// extractWith7z uses the 7z CLI to extract an archive. It checks if the
// archive has a single root folder; if not, extracts into a subfolder
// named after the archive.
func extractWith7z(archivePath, sevenZPath string) (int, error) {
	parentDir := filepath.Dir(archivePath)

	// List archive contents to determine structure
	listCmd := exec.Command(sevenZPath, "l", "-slt", archivePath)
	listOut, err := listCmd.Output()
	if err != nil {
		return 0, fmt.Errorf("7z list: %w", err)
	}

	entries := parse7zList(string(listOut))
	destDir := compute7zExtractDir(archivePath, parentDir, entries)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return 0, fmt.Errorf("mkdir: %w", err)
	}

	// Extract
	cmd := exec.Command(sevenZPath, "x", "-y", "-o"+destDir, archivePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return 0, fmt.Errorf("7z extract: %w\n%s", err, string(out))
	}

	// Count extracted files
	count := 0
	for _, e := range entries {
		if !e.isDir {
			count++
		}
	}

	return count, nil
}

type archiveEntry struct {
	path  string
	isDir bool
}

// parse7zList parses the output of `7z l -slt` to get entry paths.
func parse7zList(output string) []archiveEntry {
	var entries []archiveEntry
	var currentPath string
	var currentIsDir bool

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Path = ") {
			if currentPath != "" {
				entries = append(entries, archiveEntry{path: currentPath, isDir: currentIsDir})
			}
			currentPath = strings.TrimPrefix(line, "Path = ")
			currentIsDir = false
		} else if line == "Folder = +" {
			currentIsDir = true
		}
	}
	if currentPath != "" {
		entries = append(entries, archiveEntry{path: currentPath, isDir: currentIsDir})
	}

	// Skip the first entry which is the archive name itself
	if len(entries) > 0 {
		entries = entries[1:]
	}

	return entries
}

// compute7zExtractDir determines the extraction directory for a 7z/rar archive.
// Same logic as zip: if all entries share a single root dir, extract to parent;
// otherwise create a wrapper subfolder.
func compute7zExtractDir(archivePath, parentDir string, entries []archiveEntry) string {
	if len(entries) == 0 {
		return parentDir
	}

	var commonRoot string
	for _, e := range entries {
		parts := strings.SplitN(filepath.ToSlash(e.path), "/", 2)
		if len(parts) == 0 {
			continue
		}
		root := parts[0]
		if commonRoot == "" {
			commonRoot = root
		} else if root != commonRoot {
			base := strings.TrimSuffix(filepath.Base(archivePath), filepath.Ext(archivePath))
			return filepath.Join(parentDir, base)
		}
	}

	// All entries share one root — extract directly
	return parentDir
}

// decompressEcm decompresses a .ecm file using the native Go implementation.
// The output file is the input path with the .ecm extension stripped
// (e.g. game.bin.ecm → game.bin). After successful decompression, any .cue
// files referencing the .ecm are patched, and the source .ecm is deleted.
func decompressEcm(ecmPath, _ string) (int, error) {
	outputPath, err := decompressEcmNative(ecmPath)
	if err != nil {
		return 0, fmt.Errorf("ecm decompress %s: %w", filepath.Base(ecmPath), err)
	}
	// Fix .cue files that reference .bin.ecm instead of .bin
	fixCueEcmReferences(ecmPath, outputPath)
	// Remove the intermediate .ecm file — the .bin is what matters.
	// The original archive (.rar/.zip) is preserved for archiving.
	os.Remove(ecmPath)
	return 1, nil
}

// extractSingleZip extracts a zip file into the same directory as the zip.
// If all entries share a single root folder, they are extracted directly
// (preserving that folder). Otherwise, a subfolder named after the zip
// (without extension) is created to contain the files.
func extractSingleZip(zipPath string) (int, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	destDir := computeExtractDir(zipPath, r.File)

	count := 0
	for _, f := range r.File {
		target, err := safeJoin(destDir, f.Name)
		if err != nil {
			return count, err
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return count, fmt.Errorf("mkdir: %w", err)
		}

		if err := extractFile(f, target); err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

// computeExtractDir determines the extraction directory. If all zip entries
// share a single root directory, extract to the zip's parent dir (the root
// dir in the zip serves as the subfolder). Otherwise, create a subfolder
// named after the zip file.
func computeExtractDir(zipPath string, files []*zip.File) string {
	parentDir := filepath.Dir(zipPath)

	if len(files) == 0 {
		return parentDir
	}

	// Check if all entries share a single root directory
	var commonRoot string
	for _, f := range files {
		parts := strings.SplitN(filepath.ToSlash(f.Name), "/", 2)
		if len(parts) == 0 {
			continue
		}
		root := parts[0]
		if commonRoot == "" {
			commonRoot = root
		} else if root != commonRoot {
			// Multiple roots — need a wrapper subfolder
			base := strings.TrimSuffix(filepath.Base(zipPath), filepath.Ext(zipPath))
			return filepath.Join(parentDir, base)
		}
	}

	// All entries share one root — extract directly, the root dir acts as subfolder
	return parentDir
}

// safeJoin joins a base directory with a relative path, preventing zip-slip
// attacks by verifying the result stays within the base.
func safeJoin(base, rel string) (string, error) {
	// Clean the relative path to remove ../ sequences
	cleaned := filepath.FromSlash(rel)
	target := filepath.Join(base, cleaned)

	// Resolve to absolute for comparison
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", fmt.Errorf("abs base: %w", err)
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("abs target: %w", err)
	}

	if !strings.HasPrefix(absTarget, absBase+string(filepath.Separator)) && absTarget != absBase {
		return "", fmt.Errorf("zip-slip detected: %s escapes %s", rel, base)
	}

	return target, nil
}

func extractFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open entry %s: %w", f.Name, err)
	}
	defer rc.Close()

	out, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("create %s: %w", target, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, rc); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}

	return nil
}

// ArchiveExtractedZips moves extracted archive files to the archive directory,
// preserving their relative path from the source roots.
func ArchiveExtractedZips(files []ExtractableFile, sourceRoots []string, archiveDir string) *ArchiveResult {
	result := &ArchiveResult{}

	for _, f := range files {
		relPath := computeRelativePath(sourceRoots, f.Path)
		archivePath := filepath.Join(archiveDir, relPath)

		if err := os.MkdirAll(filepath.Dir(archivePath), 0755); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("mkdir %s: %w", archivePath, err))
			continue
		}

		if err := moveFile(f.Path, archivePath); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("move %s: %w", f.Path, err))
			continue
		}
		result.FilesMoved++
	}

	return result
}

// DeleteArchiveDir removes the archive directory and all its contents.
func DeleteArchiveDir(archiveDir string) error {
	return os.RemoveAll(archiveDir)
}
