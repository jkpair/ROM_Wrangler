package organizer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kurlmarx/romwrangler/internal/systems"
)

func TestScan(t *testing.T) {
	// Create a temp source directory structure
	dir := t.TempDir()

	// Create system subdirectories with ROM files
	nesDir := filepath.Join(dir, "nes")
	os.MkdirAll(nesDir, 0755)
	os.WriteFile(filepath.Join(nesDir, "game1.nes"), []byte("NES ROM"), 0644)
	os.WriteFile(filepath.Join(nesDir, "game2.nes"), []byte("NES ROM"), 0644)

	genesisDir := filepath.Join(dir, "genesis")
	os.MkdirAll(genesisDir, 0755)
	os.WriteFile(filepath.Join(genesisDir, "sonic.md"), []byte("MD ROM"), 0644)

	// Unknown dir
	unknownDir := filepath.Join(dir, "unknownsystem")
	os.MkdirAll(unknownDir, 0755)
	os.WriteFile(filepath.Join(unknownDir, "file.bin"), []byte("data"), 0644)

	result := Scan([]string{dir}, nil)

	// Check NES files
	nesFiles := result.BySystem[systems.NintendoNES]
	if len(nesFiles) != 2 {
		t.Errorf("expected 2 NES files, got %d", len(nesFiles))
	}

	// Check Genesis/MD files
	mdFiles := result.BySystem[systems.SegaMD]
	if len(mdFiles) != 1 {
		t.Errorf("expected 1 MD file, got %d", len(mdFiles))
	}

	// Check unresolved
	if len(result.Unresolved) != 1 {
		t.Errorf("expected 1 unresolved file, got %d", len(result.Unresolved))
	}

	// Total
	if len(result.Files) != 3 {
		t.Errorf("expected 3 total files, got %d", len(result.Files))
	}
}
