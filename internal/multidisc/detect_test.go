package multidisc

import (
	"testing"
)

func TestDetectSets(t *testing.T) {
	files := []string{
		"/roms/Final Fantasy VII (USA) (Disc 1).cue",
		"/roms/Final Fantasy VII (USA) (Disc 2).cue",
		"/roms/Final Fantasy VII (USA) (Disc 3).cue",
		"/roms/Crash Bandicoot (USA).cue",
		"/roms/Metal Gear Solid (USA) (Disc 1).cue",
		"/roms/Metal Gear Solid (USA) (Disc 2).cue",
	}

	sets, standalone := DetectSets(files)

	if len(sets) != 2 {
		t.Fatalf("expected 2 multi-disc sets, got %d", len(sets))
	}

	if len(standalone) != 1 {
		t.Fatalf("expected 1 standalone file, got %d", len(standalone))
	}

	// Find FF7 set
	var ff7Set *MultiDiscSet
	for i, s := range sets {
		if s.BaseName == "Final Fantasy VII (USA)" {
			ff7Set = &sets[i]
			break
		}
	}
	if ff7Set == nil {
		t.Fatal("expected to find Final Fantasy VII set")
	}
	if len(ff7Set.Files) != 3 {
		t.Errorf("expected 3 discs in FF7, got %d", len(ff7Set.Files))
	}

	// Verify disc ordering
	for i, disc := range ff7Set.Files {
		if disc.DiscNum != i+1 {
			t.Errorf("disc %d has number %d", i, disc.DiscNum)
		}
	}
}

func TestDetectSets_CDPattern(t *testing.T) {
	files := []string{
		"/roms/Game (USA) (CD1).bin",
		"/roms/Game (USA) (CD2).bin",
	}

	sets, standalone := DetectSets(files)
	if len(sets) != 1 {
		t.Fatalf("expected 1 set, got %d", len(sets))
	}
	if len(standalone) != 0 {
		t.Fatalf("expected 0 standalone, got %d", len(standalone))
	}
}

func TestDetectSets_SingleDiscWithPattern(t *testing.T) {
	// A single file with disc pattern should be standalone
	files := []string{
		"/roms/Game (Disc 1).cue",
	}

	sets, standalone := DetectSets(files)
	if len(sets) != 0 {
		t.Errorf("expected 0 sets, got %d", len(sets))
	}
	if len(standalone) != 1 {
		t.Errorf("expected 1 standalone, got %d", len(standalone))
	}
}

func TestDetectSets_DiscOfPattern(t *testing.T) {
	files := []string{
		"/roms/Resident Evil - Code Veronica v1.000 (2000)(Capcom)(NTSC)(US)(Disc 1 of 2).chd",
		"/roms/Resident Evil - Code Veronica v1.000 (2000)(Capcom)(NTSC)(US)(Disc 2 of 2).chd",
	}

	sets, standalone := DetectSets(files)
	if len(sets) != 1 {
		t.Fatalf("expected 1 set, got %d", len(sets))
	}
	if len(standalone) != 0 {
		t.Fatalf("expected 0 standalone, got %d", len(standalone))
	}
	if len(sets[0].Files) != 2 {
		t.Errorf("expected 2 discs, got %d", len(sets[0].Files))
	}
	if sets[0].Files[0].DiscNum != 1 || sets[0].Files[1].DiscNum != 2 {
		t.Errorf("disc numbers wrong: %d, %d", sets[0].Files[0].DiscNum, sets[0].Files[1].DiscNum)
	}
}

func TestExtractDiscNumber(t *testing.T) {
	tests := []struct {
		filename string
		expected int
	}{
		{"Game (Disc 1)", 1},
		{"Game (Disc 2)", 2},
		{"Game (Disc 1 of 2)", 1},
		{"Game (Disc 2 of 2)", 2},
		{"Game (CD1)", 1},
		{"Game (CD 3)", 3},
		{"Game (Disk 1)", 1},
		{"Game_d1", 1},
		{"Game_d2", 2},
		{"Game", 0},
	}
	for _, tt := range tests {
		result := ExtractDiscNumber(tt.filename)
		if result != tt.expected {
			t.Errorf("ExtractDiscNumber(%q) = %d, want %d", tt.filename, result, tt.expected)
		}
	}
}

func TestHasDiscPattern(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"Game (Disc 1)", true},
		{"Game (Disc 1 of 2)", true},
		{"Game (CD2)", true},
		{"Game (Disk 1)", true},
		{"Game_d1", true},
		{"Game", false},
		{"Game (USA)", false},
	}
	for _, tt := range tests {
		result := HasDiscPattern(tt.filename)
		if result != tt.expected {
			t.Errorf("HasDiscPattern(%q) = %v, want %v", tt.filename, result, tt.expected)
		}
	}
}

func TestGenerateM3U(t *testing.T) {
	set := MultiDiscSet{
		BaseName: "Final Fantasy VII (USA)",
		Files: []DiscFile{
			{Path: "/roms/Final Fantasy VII (USA) (Disc 1).chd", DiscNum: 1},
			{Path: "/roms/Final Fantasy VII (USA) (Disc 2).chd", DiscNum: 2},
			{Path: "/roms/Final Fantasy VII (USA) (Disc 3).chd", DiscNum: 3},
		},
	}

	content := GenerateM3U(set, "", false)
	expected := "Final Fantasy VII (USA) (Disc 1).chd\nFinal Fantasy VII (USA) (Disc 2).chd\nFinal Fantasy VII (USA) (Disc 3).chd\n"
	if content != expected {
		t.Errorf("M3U content:\ngot:  %q\nwant: %q", content, expected)
	}
}

func TestGenerateM3U_WithExtension(t *testing.T) {
	set := MultiDiscSet{
		BaseName: "Game",
		Files: []DiscFile{
			{Path: "/roms/Game (Disc 1).cue", DiscNum: 1},
			{Path: "/roms/Game (Disc 2).cue", DiscNum: 2},
		},
	}

	content := GenerateM3U(set, ".chd", false)
	expected := "Game (Disc 1).chd\nGame (Disc 2).chd\n"
	if content != expected {
		t.Errorf("M3U content:\ngot:  %q\nwant: %q", content, expected)
	}
}
