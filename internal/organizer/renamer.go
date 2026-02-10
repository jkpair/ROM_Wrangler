package organizer

import (
	"regexp"
	"strings"
)

// Tag patterns to strip from filenames
var (
	// Bad dump, verified, alternate, hack indicators
	stripPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\[!\]`),                         // verified dump
		regexp.MustCompile(`\[b\d*\]`),                      // bad dump
		regexp.MustCompile(`\[a\d*\]`),                      // alternate
		regexp.MustCompile(`\[h\d*[^\]]*\]`),                // hack
		regexp.MustCompile(`\[o\d*\]`),                      // overdump
		regexp.MustCompile(`\[t\d*\]`),                      // trained
		regexp.MustCompile(`\[f\d*\]`),                      // fixed
		regexp.MustCompile(`\[p\d*\]`),                      // pirate
		regexp.MustCompile(`\[T[+-][^\]]*\]`),               // translation
		regexp.MustCompile(`\[SLUS-\d+\]`),                  // Sony serial
		regexp.MustCompile(`\[SLES-\d+\]`),                  // Sony serial EU
		regexp.MustCompile(`\[SLPM-\d+\]`),                  // Sony serial JP
		regexp.MustCompile(`\[SCUS-\d+\]`),                  // Sony serial
		regexp.MustCompile(`\[SCES-\d+\]`),                  // Sony serial
		regexp.MustCompile(`\[SCPS-\d+\]`),                  // Sony serial
		regexp.MustCompile(`\[GDI-\d+\]`),                   // GDI serial
		regexp.MustCompile(`\[\d+M\]`),                      // size markers
	}

	// Preserve these patterns (region, disc info)
	// (USA), (Europe), (Japan), (Disc 1), etc. are kept

	// Clean up whitespace
	multiSpace = regexp.MustCompile(`\s{2,}`)
)

// CleanFilename strips dump tags and standardizes spacing while preserving
// region and disc information.
func CleanFilename(name string) string {
	result := name

	for _, p := range stripPatterns {
		result = p.ReplaceAllString(result, "")
	}

	// Clean up whitespace
	result = multiSpace.ReplaceAllString(result, " ")
	result = strings.TrimSpace(result)

	return result
}
