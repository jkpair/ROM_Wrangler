package organizer

import (
	"testing"

	"github.com/kurlmarx/romwrangler/internal/systems"
)

func TestDetectSystemByExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantSys  systems.SystemID
		wantOK   bool
	}{
		{"NES rom", "Super Mario Bros.nes", systems.NintendoNES, true},
		{"SNES rom", "Zelda.sfc", systems.NintendoSNES, true},
		{"SNES rom smc", "DKC.smc", systems.NintendoSNES, true},
		{"GBA rom", "Pokemon.gba", systems.NintendoGBA, true},
		{"GBC rom", "Pokemon Crystal.gbc", systems.NintendoGBC, true},
		{"GB rom", "Tetris.gb", systems.NintendoGB, true},
		{"N64 rom z64", "Mario64.z64", systems.NintendoN64, true},
		{"N64 rom n64", "Mario64.n64", systems.NintendoN64, true},
		{"NDS rom", "Pokemon.nds", systems.NintendoNDS, true},
		{"Genesis md", "Sonic.md", systems.SegaMD, true},
		{"Genesis smd", "Sonic.smd", systems.SegaMD, true},
		{"Genesis gen", "Sonic.gen", systems.SegaMD, true},
		{"32X", "Knuckles.32x", systems.Sega32X, true},
		{"Game Gear", "Sonic.gg", systems.SegaGG, true},
		{"Master System", "Alex Kidd.sms", systems.SegaMS, true},
		{"Atari 2600", "Pitfall.a26", systems.Atari2600, true},
		{"Atari Lynx", "APBC.lnx", systems.AtariLynx, true},
		{"PC Engine", "Bonk.pce", systems.NECPCE, true},
		{"ScummVM", "monkey.scummvm", systems.ScummVM, true},
		{"C64 d64", "game.d64", systems.CommodoreC64, true},
		{"Amiga adf", "game.adf", systems.CommodoreAmiga, true},
		{"Neo Geo Pocket", "game.ngp", systems.SNKNGP, true},
		{"case insensitive", "Game.NES", systems.NintendoNES, true},
		{"case insensitive upper", "Game.GBA", systems.NintendoGBA, true},
		{"ambiguous bin", "rom.bin", "", false},
		{"ambiguous zip", "rom.zip", "", false},
		{"ambiguous iso", "rom.iso", "", false},
		{"ambiguous chd", "rom.chd", "", false},
		{"ambiguous cue", "rom.cue", "", false},
		{"unknown extension", "readme.txt", "", false},
		{"no extension", "romfile", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSys, gotOK := DetectSystemByExtension(tt.filename)
			if gotOK != tt.wantOK {
				t.Errorf("DetectSystemByExtension(%q) ok = %v, want %v", tt.filename, gotOK, tt.wantOK)
			}
			if gotSys != tt.wantSys {
				t.Errorf("DetectSystemByExtension(%q) = %q, want %q", tt.filename, gotSys, tt.wantSys)
			}
		})
	}
}

func TestDetectMisplaced(t *testing.T) {
	tests := []struct {
		name       string
		files      []ScannedFile
		wantCount  int
		wantFirst  systems.SystemID // expected CorrectSystem for first misplaced
	}{
		{
			name: "NES file in SNES folder",
			files: []ScannedFile{
				{Path: "/roms/snes/Super Mario Bros.nes", System: systems.NintendoSNES},
			},
			wantCount: 1,
			wantFirst: systems.NintendoNES,
		},
		{
			name: "correct file not flagged",
			files: []ScannedFile{
				{Path: "/roms/nes/Super Mario Bros.nes", System: systems.NintendoNES},
			},
			wantCount: 0,
		},
		{
			name: "GBA file in GB folder",
			files: []ScannedFile{
				{Path: "/roms/gb/Pokemon.gba", System: systems.NintendoGB},
			},
			wantCount: 1,
			wantFirst: systems.NintendoGBA,
		},
		{
			name: "ambiguous extension not flagged",
			files: []ScannedFile{
				{Path: "/roms/genesis/game.bin", System: systems.SegaMD},
			},
			wantCount: 0,
		},
		{
			name: "multiple misplaced files",
			files: []ScannedFile{
				{Path: "/roms/snes/game.nes", System: systems.NintendoSNES},
				{Path: "/roms/nes/game.sfc", System: systems.NintendoNES},
				{Path: "/roms/nes/correct.nes", System: systems.NintendoNES},
			},
			wantCount: 2,
			wantFirst: systems.NintendoNES,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ScanResult{
				Files:    tt.files,
				BySystem: make(map[systems.SystemID][]ScannedFile),
			}
			for _, f := range tt.files {
				result.BySystem[f.System] = append(result.BySystem[f.System], f)
			}

			misplaced := DetectMisplaced(result)
			if len(misplaced) != tt.wantCount {
				t.Errorf("DetectMisplaced() found %d misplaced, want %d", len(misplaced), tt.wantCount)
			}
			if tt.wantCount > 0 && len(misplaced) > 0 {
				if misplaced[0].CorrectSystem != tt.wantFirst {
					t.Errorf("first misplaced CorrectSystem = %q, want %q", misplaced[0].CorrectSystem, tt.wantFirst)
				}
				if misplaced[0].Source != "extension" {
					t.Errorf("first misplaced Source = %q, want %q", misplaced[0].Source, "extension")
				}
			}
		})
	}
}

func TestRelocateMisplaced(t *testing.T) {
	result := &ScanResult{
		Files: []ScannedFile{
			{Path: "/roms/snes/game.nes", System: systems.NintendoSNES},
			{Path: "/roms/snes/correct.sfc", System: systems.NintendoSNES},
		},
		BySystem: map[systems.SystemID][]ScannedFile{
			systems.NintendoSNES: {
				{Path: "/roms/snes/game.nes", System: systems.NintendoSNES},
				{Path: "/roms/snes/correct.sfc", System: systems.NintendoSNES},
			},
		},
	}

	misplaced := []MisplacedFile{
		{Path: "/roms/snes/game.nes", CurrentSystem: systems.NintendoSNES, CorrectSystem: systems.NintendoNES},
	}

	RelocateMisplaced(result, misplaced)

	// Check the file was reassigned
	if result.Files[0].System != systems.NintendoNES {
		t.Errorf("Files[0].System = %q, want %q", result.Files[0].System, systems.NintendoNES)
	}
	// Check the correct file wasn't touched
	if result.Files[1].System != systems.NintendoSNES {
		t.Errorf("Files[1].System = %q, want %q", result.Files[1].System, systems.NintendoSNES)
	}
	// Check BySystem was rebuilt
	if len(result.BySystem[systems.NintendoNES]) != 1 {
		t.Errorf("BySystem[NES] has %d files, want 1", len(result.BySystem[systems.NintendoNES]))
	}
	if len(result.BySystem[systems.NintendoSNES]) != 1 {
		t.Errorf("BySystem[SNES] has %d files, want 1", len(result.BySystem[systems.NintendoSNES]))
	}
}

func TestResolveUnknown(t *testing.T) {
	result := &ScanResult{
		BySystem: make(map[systems.SystemID][]ScannedFile),
		Unresolved: []string{
			"/roms/unknown/game.gba",
			"/roms/unknown/game.nes",
			"/roms/unknown/game.bin",
			"/roms/unknown/readme.txt",
		},
	}

	ResolveUnknown(nil, result, nil)

	// .gba and .nes should be resolved, .bin and .txt should remain unresolved
	if len(result.Unresolved) != 2 {
		t.Errorf("Unresolved has %d files, want 2: %v", len(result.Unresolved), result.Unresolved)
	}
	if len(result.Files) != 2 {
		t.Errorf("Files has %d entries, want 2", len(result.Files))
	}
	if len(result.BySystem[systems.NintendoGBA]) != 1 {
		t.Errorf("BySystem[GBA] has %d files, want 1", len(result.BySystem[systems.NintendoGBA]))
	}
	if len(result.BySystem[systems.NintendoNES]) != 1 {
		t.Errorf("BySystem[NES] has %d files, want 1", len(result.BySystem[systems.NintendoNES]))
	}
}
