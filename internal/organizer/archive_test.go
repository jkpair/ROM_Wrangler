package organizer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kurlmarx/romwrangler/internal/converter"
)

func TestBuildArchivePlan(t *testing.T) {
	dir := t.TempDir()
	sourceDir := filepath.Join(dir, "roms", "sega_dc")
	archiveDir := filepath.Join(dir, "roms", "_archive")
	os.MkdirAll(sourceDir, 0755)

	// Create a GDI with track files
	gdiContent := `2
1 0 4 2048 track01.bin 0
2 300 0 2352 track02.raw 0
`
	gdiPath := filepath.Join(sourceDir, "game.gdi")
	os.WriteFile(gdiPath, []byte(gdiContent), 0644)
	os.WriteFile(filepath.Join(sourceDir, "track01.bin"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(sourceDir, "track02.raw"), []byte("data"), 0644)

	results := []converter.ConvertResult{
		{InputPath: gdiPath, OutputPath: filepath.Join(sourceDir, "game.chd"), Err: nil},
	}

	plan := BuildArchivePlan([]string{filepath.Join(dir, "roms")}, results, archiveDir)

	if len(plan.Actions) != 3 {
		t.Fatalf("expected 3 archive actions (gdi + 2 tracks), got %d", len(plan.Actions))
	}

	// Verify archive paths preserve folder structure
	for _, action := range plan.Actions {
		relToArchive, _ := filepath.Rel(archiveDir, action.ArchivePath)
		if !startsWith(relToArchive, "sega_dc") {
			t.Errorf("expected archive path under sega_dc, got %s", relToArchive)
		}
	}
}

func TestBuildArchivePlan_SkipsFailedConversions(t *testing.T) {
	dir := t.TempDir()
	isoPath := filepath.Join(dir, "game.iso")
	os.WriteFile(isoPath, []byte("data"), 0644)

	results := []converter.ConvertResult{
		{InputPath: isoPath, OutputPath: "", Err: os.ErrNotExist},
	}

	plan := BuildArchivePlan([]string{dir}, results, filepath.Join(dir, "_archive"))

	if len(plan.Actions) != 0 {
		t.Errorf("expected 0 actions for failed conversion, got %d", len(plan.Actions))
	}
}

func TestExecuteArchive(t *testing.T) {
	dir := t.TempDir()
	sourceDir := filepath.Join(dir, "source")
	archiveDir := filepath.Join(dir, "archive")
	os.MkdirAll(sourceDir, 0755)

	// Create source files
	file1 := filepath.Join(sourceDir, "game.iso")
	os.WriteFile(file1, []byte("iso data"), 0644)

	plan := &ArchivePlan{
		ArchiveRoot: archiveDir,
		Actions: []ArchiveAction{
			{SourcePath: file1, ArchivePath: filepath.Join(archiveDir, "game.iso")},
		},
	}

	result := ExecuteArchive(plan, nil)

	if result.FilesMoved != 1 {
		t.Errorf("expected 1 file moved, got %d", result.FilesMoved)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(result.Errors), result.Errors)
	}

	// Source should be gone
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Error("expected source file to be removed after archive")
	}

	// Archive should exist
	archived := filepath.Join(archiveDir, "game.iso")
	if _, err := os.Stat(archived); err != nil {
		t.Errorf("expected archived file to exist: %v", err)
	}

	// Verify content
	data, _ := os.ReadFile(archived)
	if string(data) != "iso data" {
		t.Errorf("expected 'iso data', got %q", string(data))
	}
}

func startsWith(path, prefix string) bool {
	return len(path) >= len(prefix) && path[:len(prefix)] == prefix
}
