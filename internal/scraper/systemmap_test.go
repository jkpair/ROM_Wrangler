package scraper

import (
	"testing"

	"github.com/kurlmarx/romwrangler/internal/systems"
)

func TestScreenScraperToSystemID(t *testing.T) {
	tests := []struct {
		name    string
		ssID    int
		wantSys systems.SystemID
		wantOK  bool
	}{
		{"Mega Drive", 1, systems.SegaMD, true},
		{"Master System", 2, systems.SegaMS, true},
		{"NES", 3, systems.NintendoNES, true},
		{"SNES", 4, systems.NintendoSNES, true},
		{"Game Boy", 9, systems.NintendoGB, true},
		{"Game Boy Color", 10, systems.NintendoGBC, true},
		{"GBA", 12, systems.NintendoGBA, true},
		{"N64", 14, systems.NintendoN64, true},
		{"NDS", 15, systems.NintendoNDS, true},
		{"Sega CD", 20, systems.SegaCD, true},
		{"Game Gear", 21, systems.SegaGG, true},
		{"Saturn", 22, systems.SegaSaturn, true},
		{"Dreamcast", 23, systems.SegaDC, true},
		{"Neo Geo Pocket", 25, systems.SNKNGP, true},
		{"Atari 2600", 26, systems.Atari2600, true},
		{"Atari Jaguar", 27, systems.AtariJaguar, true},
		{"Atari Lynx", 28, systems.AtariLynx, true},
		{"3DO", 29, systems.Panasonic3DO, true},
		{"PC Engine", 31, systems.NECPCE, true},
		{"PlayStation", 57, systems.SonyPSX, true},
		{"Amstrad CPC", 65, systems.AmstradCPC, true},
		{"FBNeo", 75, systems.ArcadeFBNeo, true},
		{"ZX Spectrum", 76, systems.SinclairZX, true},
		{"FDS", 106, systems.NintendoFDS, true},
		{"SG-1000", 109, systems.SegaSG1000, true},
		{"MSX", 113, systems.MSX, true},
		{"PC Engine CD", 114, systems.NECPCECD, true},
		{"Neo Geo", 142, systems.SNKNeoGeo, true},
		{"unknown ID 9999", 9999, "", false},
		{"unknown ID 0", 0, "", false},
		{"unknown ID -1", -1, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSys, gotOK := ScreenScraperToSystemID(tt.ssID)
			if gotOK != tt.wantOK {
				t.Errorf("ScreenScraperToSystemID(%d) ok = %v, want %v", tt.ssID, gotOK, tt.wantOK)
			}
			if gotSys != tt.wantSys {
				t.Errorf("ScreenScraperToSystemID(%d) = %q, want %q", tt.ssID, gotSys, tt.wantSys)
			}
		})
	}
}
