package converter

import (
	"strings"
	"testing"
)

func TestParseProgress(t *testing.T) {
	input := "Compressing, 0.0% complete... \r" +
		"Compressing, 25.5% complete... \r" +
		"Compressing, 50.0% complete... \r" +
		"Compressing, 75.3% complete... \r" +
		"Compressing, 100.0% complete... \n"

	var values []float64
	parseProgress(strings.NewReader(input), func(pct float64) {
		values = append(values, pct)
	})

	expected := []float64{0.0, 25.5, 50.0, 75.3, 100.0}
	if len(values) != len(expected) {
		t.Fatalf("expected %d values, got %d: %v", len(expected), len(values), values)
	}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("values[%d] = %f, want %f", i, v, expected[i])
		}
	}
}

func TestParseProgress_NilCallback(t *testing.T) {
	input := "Compressing, 50.0% complete... \r"
	// Should not panic
	parseProgress(strings.NewReader(input), nil)
}

func TestDetectConvertType(t *testing.T) {
	tests := []struct {
		path     string
		expected ConvertType
	}{
		{"game.gdi", ConvertCD},
		{"game.GDI", ConvertCD},
		{"game.cue", ConvertCD},
		{"game.iso", ConvertDVD},
		{"game.bin", ConvertDVD},
	}
	for _, tt := range tests {
		result := DetectConvertType(tt.path)
		if result != tt.expected {
			t.Errorf("DetectConvertType(%s) = %d, want %d", tt.path, result, tt.expected)
		}
	}
}

func TestOutputPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"game.gdi", "game.chd"},
		{"game.cue", "game.chd"},
		{"game.iso", "game.chd"},
		{"/path/to/game.gdi", "/path/to/game.chd"},
	}
	for _, tt := range tests {
		result := OutputPath(tt.input)
		if result != tt.expected {
			t.Errorf("OutputPath(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestIsConvertible(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"game.gdi", true},
		{"game.cue", true},
		{"game.iso", true},
		{"game.chd", false},
		{"game.bin", false},
		{"game.zip", false},
	}
	for _, tt := range tests {
		result := IsConvertible(tt.path)
		if result != tt.expected {
			t.Errorf("IsConvertible(%s) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}
