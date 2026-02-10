package multidisc

import (
	"path/filepath"
	"sort"
	"strings"
)

// MultiDiscSet represents a group of disc files that belong to the same game.
type MultiDiscSet struct {
	BaseName string   // Common game name without disc indicator
	Files    []DiscFile
}

// DiscFile is a single disc in a multi-disc set.
type DiscFile struct {
	Path    string
	DiscNum int
}

// DetectSets groups files into multi-disc sets based on filename patterns.
// Returns sets for multi-disc games and a list of standalone files.
func DetectSets(files []string) (sets []MultiDiscSet, standalone []string) {
	groups := make(map[string][]DiscFile)

	for _, f := range files {
		name := filepath.Base(f)
		nameNoExt := strings.TrimSuffix(name, filepath.Ext(name))

		if !HasDiscPattern(nameNoExt) {
			standalone = append(standalone, f)
			continue
		}

		baseName := strings.TrimSpace(StripDiscPattern(nameNoExt))
		discNum := ExtractDiscNumber(nameNoExt)

		groups[baseName] = append(groups[baseName], DiscFile{
			Path:    f,
			DiscNum: discNum,
		})
	}

	for baseName, discs := range groups {
		if len(discs) < 2 {
			// Single disc with disc pattern — still standalone
			standalone = append(standalone, discs[0].Path)
			continue
		}

		// Sort by disc number
		sort.Slice(discs, func(i, j int) bool {
			return discs[i].DiscNum < discs[j].DiscNum
		})

		sets = append(sets, MultiDiscSet{
			BaseName: baseName,
			Files:    discs,
		})
	}

	// Sort sets by name for deterministic output
	sort.Slice(sets, func(i, j int) bool {
		return sets[i].BaseName < sets[j].BaseName
	})

	return sets, standalone
}
