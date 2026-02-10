package multidisc

import "regexp"

// discPatterns are regexes that match disc number indicators in filenames.
// Each has a named capture group "num" for the disc number.
var discPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\(Disc\s*(?P<num>\d+)\s+of\s+\d+\)`), // (Disc 1 of 2)
	regexp.MustCompile(`(?i)\(Disc\s*(?P<num>\d+)\)`),             // (Disc 1), (Disc 2)
	regexp.MustCompile(`(?i)\(CD\s*(?P<num>\d+)\)`),               // (CD1), (CD 2)
	regexp.MustCompile(`(?i)\(Disk\s*(?P<num>\d+)\)`),             // (Disk 1)
	regexp.MustCompile(`(?i)[\s._-]d(?P<num>\d+)(?:[\s._]|$)`),   // _d1, .d2, -d1
	regexp.MustCompile(`(?i)[\s._-]cd(?P<num>\d+)(?:[\s._]|$)`),  // _cd1, .cd2
	regexp.MustCompile(`(?i)[\s._-]disc(?P<num>\d+)(?:[\s._]|$)`), // _disc1
}

// HasDiscPattern returns true if the filename contains a disc number pattern.
func HasDiscPattern(filename string) bool {
	for _, p := range discPatterns {
		if p.MatchString(filename) {
			return true
		}
	}
	return false
}

// StripDiscPattern removes the disc number pattern from a filename to get the base game name.
func StripDiscPattern(filename string) string {
	result := filename
	for _, p := range discPatterns {
		result = p.ReplaceAllString(result, "")
	}
	return result
}

// ExtractDiscNumber returns the disc number from a filename, or 0 if not found.
func ExtractDiscNumber(filename string) int {
	for _, p := range discPatterns {
		match := p.FindStringSubmatch(filename)
		if match == nil {
			continue
		}
		for i, name := range p.SubexpNames() {
			if name == "num" && i < len(match) {
				var num int
				for _, c := range match[i] {
					num = num*10 + int(c-'0')
				}
				return num
			}
		}
	}
	return 0
}
