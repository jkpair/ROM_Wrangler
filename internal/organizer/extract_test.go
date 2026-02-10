package organizer

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kurlmarx/romwrangler/internal/systems"
)

func createTestZip(t *testing.T, zipPath string, files map[string]string) {
	t.Helper()
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		fw.Write([]byte(content))
	}
	w.Close()
}

// addBinaryToZip recreates a zip file, adding a binary entry to the existing entries.
func addBinaryToZip(t *testing.T, zipPath string, name string, data []byte) {
	t.Helper()

	// Read existing zip
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	type entry struct {
		name string
		data []byte
	}
	var entries []entry
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		d, _ := io.ReadAll(rc)
		rc.Close()
		entries = append(entries, entry{f.Name, d})
	}
	r.Close()

	// Rewrite zip with all entries plus new one
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for _, e := range entries {
		fw, _ := w.Create(e.name)
		fw.Write(e.data)
	}
	fw, _ := w.Create(name)
	fw.Write(data)
	w.Close()
}

func TestFindExtractable_DiscSystemOnly(t *testing.T) {
	dir := t.TempDir()

	// Disc-based system (Dreamcast) with a zip
	dcDir := filepath.Join(dir, "sega_dc")
	os.MkdirAll(dcDir, 0755)
	createTestZip(t, filepath.Join(dcDir, "game.zip"), map[string]string{
		"track01.bin": "data",
	})

	// Cartridge system (NES) with a zip — should NOT be found
	nesDir := filepath.Join(dir, "nes")
	os.MkdirAll(nesDir, 0755)
	createTestZip(t, filepath.Join(nesDir, "game.zip"), map[string]string{
		"game.nes": "data",
	})

	result := FindExtractable([]string{dir}, nil)

	if len(result) != 1 {
		t.Fatalf("expected 1 extractable, got %d", len(result))
	}
	if result[0].System != systems.SegaDC {
		t.Errorf("expected system %s, got %s", systems.SegaDC, result[0].System)
	}
}

func TestFindExtractable_AllArchiveFormats(t *testing.T) {
	dir := t.TempDir()

	dcDir := filepath.Join(dir, "sega_dc")
	os.MkdirAll(dcDir, 0755)
	// Create dummy files for each archive extension
	os.WriteFile(filepath.Join(dcDir, "game1.zip"), []byte("PK"), 0644)
	os.WriteFile(filepath.Join(dcDir, "game2.7z"), []byte("7z"), 0644)
	os.WriteFile(filepath.Join(dcDir, "game3.rar"), []byte("Rar"), 0644)
	os.WriteFile(filepath.Join(dcDir, "game4.bin.ecm"), []byte("ECM"), 0644)
	// Non-archive should not be found
	os.WriteFile(filepath.Join(dcDir, "game5.gdi"), []byte("data"), 0644)

	result := FindExtractable([]string{dir}, nil)

	if len(result) != 4 {
		t.Fatalf("expected 4 extractable archives, got %d", len(result))
	}
}

func TestFindExtractable_SkipsArchiveDir(t *testing.T) {
	dir := t.TempDir()

	archiveDir := filepath.Join(dir, "_archive", "sega_dc")
	os.MkdirAll(archiveDir, 0755)
	createTestZip(t, filepath.Join(archiveDir, "game.zip"), map[string]string{
		"track01.bin": "data",
	})

	result := FindExtractable([]string{dir}, nil)
	if len(result) != 0 {
		t.Errorf("expected 0 extractable from _archive, got %d", len(result))
	}
}

func TestExtractSingleZip_SingleRootFolder(t *testing.T) {
	dir := t.TempDir()

	// Zip with a single root folder containing files
	zipPath := filepath.Join(dir, "game.zip")
	createTestZip(t, zipPath, map[string]string{
		"MyGame/track01.bin": "track1data",
		"MyGame/track02.bin": "track2data",
		"MyGame/disc.gdi":    "gdidata",
	})

	count, err := extractSingleZip(zipPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 files extracted, got %d", count)
	}

	// Files should be in dir/MyGame/ (no double-nesting)
	if _, err := os.Stat(filepath.Join(dir, "MyGame", "track01.bin")); err != nil {
		t.Error("expected MyGame/track01.bin to exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "MyGame", "disc.gdi")); err != nil {
		t.Error("expected MyGame/disc.gdi to exist")
	}
}

func TestExtractSingleZip_LooseFiles(t *testing.T) {
	dir := t.TempDir()

	// Zip with loose files (no common root folder)
	zipPath := filepath.Join(dir, "game.zip")
	createTestZip(t, zipPath, map[string]string{
		"track01.bin": "track1data",
		"track02.bin": "track2data",
		"disc.gdi":    "gdidata",
	})

	count, err := extractSingleZip(zipPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 files extracted, got %d", count)
	}

	// Files should be wrapped in dir/game/ subfolder
	if _, err := os.Stat(filepath.Join(dir, "game", "track01.bin")); err != nil {
		t.Error("expected game/track01.bin to exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "game", "disc.gdi")); err != nil {
		t.Error("expected game/disc.gdi to exist")
	}
}

func TestSafeJoin_ZipSlipPrevention(t *testing.T) {
	base := "/safe/dir"

	// Normal path should work
	_, err := safeJoin(base, "file.bin")
	if err != nil {
		t.Errorf("normal path should succeed: %v", err)
	}

	// Nested path should work
	_, err = safeJoin(base, "subdir/file.bin")
	if err != nil {
		t.Errorf("nested path should succeed: %v", err)
	}

	// Zip-slip attempt should fail
	_, err = safeJoin(base, "../../../etc/passwd")
	if err == nil {
		t.Error("zip-slip path should fail")
	}
}

func TestDeleteArchiveDir(t *testing.T) {
	dir := t.TempDir()

	archiveDir := filepath.Join(dir, "_archive")
	os.MkdirAll(filepath.Join(archiveDir, "sega_dc"), 0755)
	os.WriteFile(filepath.Join(archiveDir, "sega_dc", "game.zip"), []byte("data"), 0644)

	if err := DeleteArchiveDir(archiveDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(archiveDir); !os.IsNotExist(err) {
		t.Error("expected archive dir to be deleted")
	}
}

func buildTestEcm(data []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString("ECM\x00")
	// Type 0 chunk (raw copy), count = len(data)
	// num = count - 1, encoded as: first byte = (num << 2) | type
	num := uint32(len(data) - 1)
	first := byte((num&0x1F)<<2) | 0 // type 0
	remaining := num >> 5
	if remaining > 0 {
		first |= 0x80 // continuation bit
	}
	buf.WriteByte(first)
	for remaining > 0 {
		b := byte(remaining & 0x7F)
		remaining >>= 7
		if remaining > 0 {
			b |= 0x80
		}
		buf.WriteByte(b)
	}
	buf.Write(data)
	// End marker
	buf.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	return buf.Bytes()
}

func TestExtractAll_IteratesOverNewlyExposed(t *testing.T) {
	dir := t.TempDir()

	// Create a PSX folder with a zip containing a .cue and a valid .ecm
	psxDir := filepath.Join(dir, "sony_psx")
	os.MkdirAll(psxDir, 0755)

	// Build a valid ECM file containing some raw sector data
	ecmContent := buildTestEcm(bytes.Repeat([]byte{0xAB}, 100))

	// Zip contains a .cue and a .bin.ecm file (valid ECM format)
	createTestZip(t, filepath.Join(psxDir, "game.zip"), map[string]string{
		"game/game.cue": "FILE game.bin BINARY",
	})
	// Write the ECM file into the zip using binary content
	addBinaryToZip(t, filepath.Join(psxDir, "game.zip"), "game/game.bin.ecm", ecmContent)

	result, processed := ExtractAll([]string{dir}, nil, nil)

	// Round 1: extracts the zip (finds game.zip)
	// Round 2: finds the .ecm exposed by zip extraction, decompresses with native Go
	if result.Extracted != 2 {
		t.Errorf("expected 2 extracted (zip + ecm), got %d", result.Extracted)
		for _, e := range result.Errors {
			t.Logf("  error: %v", e)
		}
	}

	// The zip should be in the processed list (for later archiving)
	foundZip := false
	for _, p := range processed {
		if filepath.Ext(p.Path) == ".zip" {
			foundZip = true
		}
	}
	if !foundZip {
		t.Error("expected zip in processed list")
	}

	// The .ecm should NOT be in the processed list — it's intermediate
	for _, p := range processed {
		if filepath.Ext(p.Path) == ".ecm" {
			t.Error("ecm files should not be in processed list (they are intermediates)")
		}
	}

	// The decompressed .bin file should exist
	binPath := filepath.Join(psxDir, "game", "game.bin")
	if _, err := os.Stat(binPath); err != nil {
		t.Errorf("expected decompressed game.bin at %s", binPath)
	}

	// The .ecm file should be deleted after decompression
	ecmPath := filepath.Join(psxDir, "game", "game.bin.ecm")
	if _, err := os.Stat(ecmPath); !os.IsNotExist(err) {
		t.Error("expected .ecm file to be deleted after decompression")
	}
}

func TestArchiveExtractedZips(t *testing.T) {
	dir := t.TempDir()

	// Create a zip in a disc system folder
	dcDir := filepath.Join(dir, "sega_dc")
	os.MkdirAll(dcDir, 0755)
	zipPath := filepath.Join(dcDir, "game.zip")
	os.WriteFile(zipPath, []byte("zipdata"), 0644)

	archiveDir := filepath.Join(dir, "_archive")
	files := []ExtractableFile{{Path: zipPath, System: systems.SegaDC}}

	result := ArchiveExtractedZips(files, []string{dir}, archiveDir)

	if result.FilesMoved != 1 {
		t.Errorf("expected 1 file moved, got %d", result.FilesMoved)
	}
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}

	// Original should be gone
	if _, err := os.Stat(zipPath); !os.IsNotExist(err) {
		t.Error("expected original zip to be moved")
	}

	// Should exist in archive
	archivedPath := filepath.Join(archiveDir, "sega_dc", "game.zip")
	if _, err := os.Stat(archivedPath); err != nil {
		t.Errorf("expected archived zip at %s", archivedPath)
	}
}
