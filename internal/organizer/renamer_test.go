package organizer

import "testing"

func TestCleanFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Super Mario Bros. (World) [!]", "Super Mario Bros. (World)"},
		{"Game [b1]", "Game"},
		{"Game [a2]", "Game"},
		{"Game [h1C]", "Game"},
		{"Game [SLUS-12345]", "Game"},
		{"Game [SLES-12345]", "Game"},
		{"Game [t1]", "Game"},
		{"Game [f1]", "Game"},
		{"Game (USA) (Disc 1)", "Game (USA) (Disc 1)"},          // preserves region+disc
		{"Game (Europe)", "Game (Europe)"},                        // preserves region
		{"Game  [!]  [b1]  extra", "Game extra"},                  // cleans multiple tags
		{"Game [T+Eng]", "Game"},                                  // translation
		{"Game [64M]", "Game"},                                    // size marker
	}
	for _, tt := range tests {
		result := CleanFilename(tt.input)
		if result != tt.expected {
			t.Errorf("CleanFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
