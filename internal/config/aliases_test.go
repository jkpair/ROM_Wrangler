package config

import (
	"testing"

	"github.com/kurlmarx/romwrangler/internal/systems"
)

func TestResolveAlias_DefaultAliases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected systems.SystemID
	}{
		{"lowercase", "dreamcast", systems.SegaDC},
		{"uppercase", "DREAMCAST", systems.SegaDC},
		{"mixed case", "Dreamcast", systems.SegaDC},
		{"psx", "psx", systems.SonyPSX},
		{"ps1", "ps1", systems.SonyPSX},
		{"genesis", "genesis", systems.SegaMD},
		{"megadrive", "megadrive", systems.SegaMD},
		{"snes", "snes", systems.NintendoSNES},
		{"nes", "nes", systems.NintendoNES},
		{"gba", "gba", systems.NintendoGBA},
		{"mame", "mame", systems.ArcadeMAME},
		{"with spaces", "  dreamcast  ", systems.SegaDC},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := ResolveAlias(tt.input, nil)
			if !ok {
				t.Fatalf("expected alias %q to resolve, but it didn't", tt.input)
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestResolveAlias_ConfigOverride(t *testing.T) {
	overrides := map[string]string{
		"myroms": string(systems.SegaDC),
	}

	result, ok := ResolveAlias("myroms", overrides)
	if !ok {
		t.Fatal("expected config override to resolve")
	}
	if result != systems.SegaDC {
		t.Errorf("expected %s, got %s", systems.SegaDC, result)
	}
}

func TestResolveAlias_ConfigOverrideTakesPriority(t *testing.T) {
	overrides := map[string]string{
		"genesis": string(systems.SegaDC), // override genesis to point to DC
	}

	result, ok := ResolveAlias("genesis", overrides)
	if !ok {
		t.Fatal("expected to resolve")
	}
	if result != systems.SegaDC {
		t.Errorf("expected override to win: got %s", result)
	}
}

func TestResolveAlias_NotFound(t *testing.T) {
	_, ok := ResolveAlias("unknownsystem", nil)
	if ok {
		t.Error("expected unknown alias to not resolve")
	}
}
