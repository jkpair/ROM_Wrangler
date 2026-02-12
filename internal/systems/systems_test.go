package systems

import "testing"

func TestGetSystem(t *testing.T) {
	info, ok := GetSystem(SegaDC)
	if !ok {
		t.Fatal("expected SegaDC to exist")
	}
	if info.DisplayName != "Dreamcast" {
		t.Errorf("expected Dreamcast, got %s", info.DisplayName)
	}
	if info.Company != "Sega" {
		t.Errorf("expected Sega, got %s", info.Company)
	}
	if !info.IsDiscBased {
		t.Error("expected Dreamcast to be disc-based")
	}
}

func TestGetSystem_NotFound(t *testing.T) {
	_, ok := GetSystem("nonexistent")
	if ok {
		t.Error("expected nonexistent system to not be found")
	}
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		system SystemID
		ext    string
		valid  bool
	}{
		{SegaDC, ".chd", true},
		{SegaDC, ".gdi", true},
		{SegaDC, ".nes", false},
		{NintendoNES, ".nes", true},
		{NintendoNES, ".zip", false},
		{ArcadeFBNeo, ".zip", true},
		{SonyPSX, ".chd", true},
		{SonyPSX, ".gba", false},
	}

	for _, tt := range tests {
		result := IsValidFormat(tt.system, tt.ext)
		if result != tt.valid {
			t.Errorf("IsValidFormat(%s, %s) = %v, want %v", tt.system, tt.ext, result, tt.valid)
		}
	}
}

func TestFolderForSystem(t *testing.T) {
	folder, ok := FolderForSystem(SegaDC)
	if !ok {
		t.Fatal("expected folder for SegaDC")
	}
	if folder != "sega_dc" {
		t.Errorf("expected sega_dc, got %s", folder)
	}
}

func TestAllSystemsHaveFolders(t *testing.T) {
	for id := range AllSystems {
		_, ok := ReplayOSFolders[id]
		if !ok {
			t.Errorf("system %s has no ReplayOS folder mapping", id)
		}
	}
}

func TestAllSystemsHaveFormats(t *testing.T) {
	for id := range AllSystems {
		formats, ok := SupportedFormats[id]
		if !ok {
			t.Errorf("system %s has no supported formats", id)
		}
		if len(formats) == 0 {
			t.Errorf("system %s has empty format list", id)
		}
	}
}
