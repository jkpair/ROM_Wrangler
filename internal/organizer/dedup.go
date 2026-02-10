package organizer

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/kurlmarx/romwrangler/internal/multidisc"
	"github.com/kurlmarx/romwrangler/internal/systems"
)

// variantPatterns match parenthesized tags that distinguish game variants.
// These are stripped to produce a "base game name" for grouping.
var variantPatterns = []*regexp.Regexp{
	// Regions
	regexp.MustCompile(`(?i)\((USA|US|America)\)`),
	regexp.MustCompile(`(?i)\((Japan|JP|JPN)\)`),
	regexp.MustCompile(`(?i)\((Europe|EU|EUR)\)`),
	regexp.MustCompile(`(?i)\((World)\)`),
	regexp.MustCompile(`(?i)\((Korea|KR)\)`),
	regexp.MustCompile(`(?i)\((Asia)\)`),
	regexp.MustCompile(`(?i)\((Australia)\)`),
	regexp.MustCompile(`(?i)\((Brazil)\)`),
	regexp.MustCompile(`(?i)\((Canada)\)`),
	regexp.MustCompile(`(?i)\((China)\)`),
	regexp.MustCompile(`(?i)\((France|Fr)\)`),
	regexp.MustCompile(`(?i)\((Germany|De)\)`),
	regexp.MustCompile(`(?i)\((Italy|It)\)`),
	regexp.MustCompile(`(?i)\((Spain|Es)\)`),
	regexp.MustCompile(`(?i)\((Sweden|Sv)\)`),
	regexp.MustCompile(`(?i)\((Netherlands|Nl)\)`),
	regexp.MustCompile(`(?i)\((Russia|Ru)\)`),

	// Multi-region combos
	regexp.MustCompile(`(?i)\((USA,\s*Europe)\)`),

	// Video standards
	regexp.MustCompile(`(?i)\((NTSC|PAL|SECAM|NTSC-J|NTSC-U|PAL-E)\)`),

	// Language codes (single and comma-separated)
	regexp.MustCompile(`(?i)\((En|Fr|De|Es|It|Nl|Sv|No|Da|Fi|Pt|Ru|Ja|Zh|Ko)(,\s*(En|Fr|De|Es|It|Nl|Sv|No|Da|Fi|Pt|Ru|Ja|Zh|Ko))*\)`),

	// Revisions and versions
	regexp.MustCompile(`(?i)\(Rev\s*[A-Z0-9.]+\)`),
	regexp.MustCompile(`(?i)\(v\d[\d.]*[a-z]?\)`),

	// Pre-release / special variants
	regexp.MustCompile(`(?i)\((Beta|Proto|Sample|Demo|Promo|Preview|Kiosk|Debug|Unl)\)`),
	regexp.MustCompile(`(?i)\(Beta\s*\d+\)`),

	// Year tags
	regexp.MustCompile(`\(\d{4}\)`),

	// Virtual Console / rerelease markers
	regexp.MustCompile(`(?i)\((Virtual Console|VC|Switch Online|Classic Mini)\)`),
}

// regionPriority assigns preference scores to files based on region tags.
// Lower score = higher preference. Used to pre-select the best variant.
var regionPriority = []struct {
	pattern *regexp.Regexp
	score   int
}{
	{regexp.MustCompile(`(?i)\(USA\)`), 0},
	{regexp.MustCompile(`(?i)\(World\)`), 1},
	{regexp.MustCompile(`(?i)\(USA,\s*Europe\)`), 2},
	{regexp.MustCompile(`(?i)\(Europe\)`), 3},
	{regexp.MustCompile(`(?i)\(Japan\)`), 4},
}

// VariantGroup represents a set of files that are different versions
// of the same game (different regions, revisions, etc.).
type VariantGroup struct {
	BaseName string
	System   systems.SystemID
	Files    []ScannedFile
}

// BaseGameName strips region, version, revision, and other variant tags
// from a filename (without extension) to produce a grouping key.
func BaseGameName(filename string) string {
	result := CleanFilename(filename)
	result = multidisc.StripDiscPattern(result)
	for _, p := range variantPatterns {
		result = p.ReplaceAllString(result, "")
	}
	result = multiSpace.ReplaceAllString(result, " ")
	result = strings.TrimSpace(result)
	return result
}

// DetectVariants groups scanned files by base game name + system,
// returning only groups with 2 or more variants (actual duplicates).
// Each group's Files slice is sorted by region preference (USA first).
func DetectVariants(scanResult *ScanResult) []VariantGroup {
	type groupKey struct {
		baseName string
		system   systems.SystemID
	}

	groups := make(map[groupKey][]ScannedFile)

	for _, f := range scanResult.Files {
		name := filepath.Base(f.Path)
		nameNoExt := strings.TrimSuffix(name, filepath.Ext(name))
		base := BaseGameName(nameNoExt)

		key := groupKey{baseName: base, system: f.System}
		groups[key] = append(groups[key], f)
	}

	var result []VariantGroup
	for key, files := range groups {
		if len(files) < 2 {
			continue
		}
		sort.SliceStable(files, func(i, j int) bool {
			return regionScore(files[i].Path) < regionScore(files[j].Path)
		})
		result = append(result, VariantGroup{
			BaseName: key.baseName,
			System:   key.system,
			Files:    files,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].System != result[j].System {
			return string(result[i].System) < string(result[j].System)
		}
		return result[i].BaseName < result[j].BaseName
	})

	return result
}

// regionScore returns a numeric priority for a file path based on region tags.
// Lower = preferred. Files with no recognized region get a high default score.
func regionScore(path string) int {
	name := filepath.Base(path)
	for _, rp := range regionPriority {
		if rp.pattern.MatchString(name) {
			return rp.score
		}
	}
	return 100
}

// RemoveFiles removes the specified file paths from the scan result,
// rebuilding Files, BySystem, and Convertible.
func (sr *ScanResult) RemoveFiles(paths []string) {
	removeSet := make(map[string]bool, len(paths))
	for _, p := range paths {
		removeSet[p] = true
	}

	var newFiles []ScannedFile
	for _, f := range sr.Files {
		if !removeSet[f.Path] {
			newFiles = append(newFiles, f)
		}
	}
	sr.Files = newFiles

	sr.BySystem = make(map[systems.SystemID][]ScannedFile)
	for _, f := range sr.Files {
		sr.BySystem[f.System] = append(sr.BySystem[f.System], f)
	}

	var newConvertible []ScannedFile
	for _, f := range sr.Convertible {
		if !removeSet[f.Path] {
			newConvertible = append(newConvertible, f)
		}
	}
	sr.Convertible = newConvertible
}
