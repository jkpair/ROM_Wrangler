package organizer

import (
	"testing"

	"github.com/kurlmarx/romwrangler/internal/systems"
)

func TestBaseGameName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Super Mario 64 (USA)", "Super Mario 64"},
		{"Super Mario 64 (Japan)", "Super Mario 64"},
		{"Super Mario 64 (USA) (Rev 1)", "Super Mario 64"},
		{"Super Mario 64 (Europe) (En,Fr,De)", "Super Mario 64"},
		{"Legend of Zelda, The (USA) (v1.1)", "Legend of Zelda, The"},
		{"GoldenEye 007 (USA) (Rev A)", "GoldenEye 007"},
		{"Game (USA) (Disc 1)", "Game"},
		{"Game (Japan) (Disc 1)", "Game"},
		{"Game (Beta)", "Game"},
		{"Game (Proto) (USA)", "Game"},
		{"Game (2000) (USA)", "Game"},
		{"Game (NTSC)", "Game"},
		{"Game (PAL)", "Game"},
		// Subtitles in parentheses should NOT be stripped
		{"Resident Evil (Director's Cut) (USA)", "Resident Evil (Director's Cut)"},
		// Dump tags should be stripped
		{"Game (USA) [!]", "Game"},
		{"Game (Japan) [b1]", "Game"},
	}
	for _, tt := range tests {
		result := BaseGameName(tt.input)
		if result != tt.expected {
			t.Errorf("BaseGameName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestDetectVariants(t *testing.T) {
	scan := &ScanResult{
		Files: []ScannedFile{
			{Path: "/roms/n64/Super Mario 64 (USA).z64", System: systems.NintendoN64},
			{Path: "/roms/n64/Super Mario 64 (Japan).z64", System: systems.NintendoN64},
			{Path: "/roms/n64/Super Mario 64 (Europe).z64", System: systems.NintendoN64},
			{Path: "/roms/n64/GoldenEye 007 (USA).z64", System: systems.NintendoN64},
			{Path: "/roms/n64/Unique Game (USA).z64", System: systems.NintendoN64},
		},
		BySystem: map[systems.SystemID][]ScannedFile{},
	}

	groups := DetectVariants(scan)

	if len(groups) != 1 {
		t.Fatalf("expected 1 variant group, got %d", len(groups))
	}

	g := groups[0]
	if g.BaseName != "Super Mario 64" {
		t.Errorf("expected base name 'Super Mario 64', got %q", g.BaseName)
	}
	if len(g.Files) != 3 {
		t.Errorf("expected 3 files in group, got %d", len(g.Files))
	}
	// USA should sort first (region priority)
	if g.Files[0].Path != "/roms/n64/Super Mario 64 (USA).z64" {
		t.Errorf("expected USA variant first, got %s", g.Files[0].Path)
	}
}

func TestDetectVariants_NoGroups(t *testing.T) {
	scan := &ScanResult{
		Files: []ScannedFile{
			{Path: "/roms/n64/Game A (USA).z64", System: systems.NintendoN64},
			{Path: "/roms/n64/Game B (USA).z64", System: systems.NintendoN64},
		},
		BySystem: map[systems.SystemID][]ScannedFile{},
	}

	groups := DetectVariants(scan)
	if len(groups) != 0 {
		t.Errorf("expected 0 variant groups for unique games, got %d", len(groups))
	}
}

func TestDetectVariants_CrossSystem(t *testing.T) {
	// Same base name in different systems should NOT be grouped
	scan := &ScanResult{
		Files: []ScannedFile{
			{Path: "/roms/n64/Game (USA).z64", System: systems.NintendoN64},
			{Path: "/roms/nes/Game (USA).nes", System: systems.NintendoNES},
		},
		BySystem: map[systems.SystemID][]ScannedFile{},
	}

	groups := DetectVariants(scan)
	if len(groups) != 0 {
		t.Errorf("expected 0 groups for cross-system, got %d", len(groups))
	}
}

func TestDetectVariants_IgnoresMultiDisc(t *testing.T) {
	scan := &ScanResult{
		Files: []ScannedFile{
			{Path: "/roms/psx/Final Fantasy VII (USA) (Disc 1).chd", System: systems.SonyPSX},
			{Path: "/roms/psx/Final Fantasy VII (USA) (Disc 2).chd", System: systems.SonyPSX},
			{Path: "/roms/psx/Final Fantasy VII (USA) (Disc 3).chd", System: systems.SonyPSX},
			{Path: "/roms/psx/Game (USA) (Disc 1 of 2).chd", System: systems.SonyPSX},
			{Path: "/roms/psx/Game (USA) (Disc 2 of 2).chd", System: systems.SonyPSX},
			{Path: "/roms/psx/Game (USA) (Disk 1).chd", System: systems.SonyPSX},
			{Path: "/roms/psx/Game (USA) (Disk 2).chd", System: systems.SonyPSX},
		},
		BySystem: map[systems.SystemID][]ScannedFile{},
	}

	groups := DetectVariants(scan)
	if len(groups) != 0 {
		t.Errorf("expected 0 variant groups for multi-disc games, got %d", len(groups))
		for _, g := range groups {
			t.Logf("  group %q: %d files", g.BaseName, len(g.Files))
		}
	}
}

func TestRemoveFiles(t *testing.T) {
	scan := &ScanResult{
		Files: []ScannedFile{
			{Path: "/a.z64", System: systems.NintendoN64},
			{Path: "/b.z64", System: systems.NintendoN64},
			{Path: "/c.z64", System: systems.NintendoN64},
		},
		BySystem: map[systems.SystemID][]ScannedFile{
			systems.NintendoN64: {
				{Path: "/a.z64", System: systems.NintendoN64},
				{Path: "/b.z64", System: systems.NintendoN64},
				{Path: "/c.z64", System: systems.NintendoN64},
			},
		},
		Convertible: []ScannedFile{
			{Path: "/b.z64", System: systems.NintendoN64},
		},
	}

	scan.RemoveFiles([]string{"/b.z64"})

	if len(scan.Files) != 2 {
		t.Errorf("expected 2 files after remove, got %d", len(scan.Files))
	}
	if len(scan.BySystem[systems.NintendoN64]) != 2 {
		t.Errorf("expected 2 in BySystem after remove, got %d", len(scan.BySystem[systems.NintendoN64]))
	}
	if len(scan.Convertible) != 0 {
		t.Errorf("expected 0 convertible after remove, got %d", len(scan.Convertible))
	}
}
