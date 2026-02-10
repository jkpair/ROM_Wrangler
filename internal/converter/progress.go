package converter

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
)

// chdman outputs progress like: "Compressing, 45.2% complete... \r"
var progressRegex = regexp.MustCompile(`(\d+(?:\.\d+)?)%\s+complete`)

// parseProgress reads chdman stderr and calls progressFn with percentage values.
func parseProgress(r io.Reader, progressFn func(float64)) {
	if progressFn == nil {
		// Drain the reader even if no callback
		io.Copy(io.Discard, r)
		return
	}

	scanner := bufio.NewScanner(r)
	// chdman uses \r for progress, so split on any whitespace-delimited token
	scanner.Split(scanChdmanProgress)

	for scanner.Scan() {
		line := scanner.Text()
		matches := progressRegex.FindStringSubmatch(line)
		if len(matches) >= 2 {
			if pct, err := strconv.ParseFloat(matches[1], 64); err == nil {
				progressFn(pct)
			}
		}
	}
}

// scanChdmanProgress is a custom split function that handles \r-delimited output.
func scanChdmanProgress(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Look for \r or \n as delimiters
	for i := 0; i < len(data); i++ {
		if data[i] == '\r' || data[i] == '\n' {
			return i + 1, data[:i], nil
		}
	}

	if atEOF {
		return len(data), data, nil
	}
	// Request more data
	return 0, nil, nil
}
