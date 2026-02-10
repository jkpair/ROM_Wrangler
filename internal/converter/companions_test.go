package converter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompanionFiles_GDI(t *testing.T) {
	dir := t.TempDir()

	gdiContent := `3
1 0 4 2048 track01.bin 0
2 300 0 2352 track02.raw 0
3 45000 4 2048 track03.bin 0
`
	gdiPath := filepath.Join(dir, "game.gdi")
	os.WriteFile(gdiPath, []byte(gdiContent), 0644)

	// Create the track files so they exist
	for _, name := range []string{"track01.bin", "track02.raw", "track03.bin"} {
		os.WriteFile(filepath.Join(dir, name), []byte("data"), 0644)
	}

	files, err := CompanionFiles(gdiPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 4 {
		t.Fatalf("expected 4 files (gdi + 3 tracks), got %d: %v", len(files), files)
	}

	// First file should be the GDI itself
	if filepath.Base(files[0]) != "game.gdi" {
		t.Errorf("expected first file to be game.gdi, got %s", files[0])
	}

	// Check track files are included
	expected := map[string]bool{
		"game.gdi":    true,
		"track01.bin": true,
		"track02.raw": true,
		"track03.bin": true,
	}
	for _, f := range files {
		name := filepath.Base(f)
		if !expected[name] {
			t.Errorf("unexpected file: %s", name)
		}
		delete(expected, name)
	}
	for name := range expected {
		t.Errorf("missing expected file: %s", name)
	}
}

func TestCompanionFiles_CUE(t *testing.T) {
	dir := t.TempDir()

	cueContent := `FILE "Track 1.bin" BINARY
  TRACK 01 MODE1/2352
    INDEX 01 00:00:00
FILE "Track 2.bin" BINARY
  TRACK 02 AUDIO
    INDEX 01 00:00:00
`
	cuePath := filepath.Join(dir, "game.cue")
	os.WriteFile(cuePath, []byte(cueContent), 0644)

	files, err := CompanionFiles(cuePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 files (cue + 2 tracks), got %d: %v", len(files), files)
	}

	if filepath.Base(files[0]) != "game.cue" {
		t.Errorf("expected first file to be game.cue, got %s", files[0])
	}

	expected := map[string]bool{
		"game.cue":    true,
		"Track 1.bin": true,
		"Track 2.bin": true,
	}
	for _, f := range files {
		name := filepath.Base(f)
		if !expected[name] {
			t.Errorf("unexpected file: %s", name)
		}
		delete(expected, name)
	}
	for name := range expected {
		t.Errorf("missing expected file: %s", name)
	}
}

func TestCompanionFiles_ISO(t *testing.T) {
	dir := t.TempDir()
	isoPath := filepath.Join(dir, "game.iso")
	os.WriteFile(isoPath, []byte("data"), 0644)

	files, err := CompanionFiles(isoPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if filepath.Base(files[0]) != "game.iso" {
		t.Errorf("expected game.iso, got %s", files[0])
	}
}

func TestCompanionFiles_GDI_DuplicateTracks(t *testing.T) {
	dir := t.TempDir()

	// GDI with duplicate track filename
	gdiContent := `2
1 0 4 2048 track01.bin 0
2 300 0 2352 track01.bin 0
`
	gdiPath := filepath.Join(dir, "game.gdi")
	os.WriteFile(gdiPath, []byte(gdiContent), 0644)

	files, err := CompanionFiles(gdiPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should deduplicate: gdi + 1 unique track
	if len(files) != 2 {
		t.Fatalf("expected 2 files (deduplicated), got %d: %v", len(files), files)
	}
}
