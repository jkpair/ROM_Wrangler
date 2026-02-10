package scraper

import "github.com/kurlmarx/romwrangler/internal/systems"

// FileHashes holds computed checksums for a file.
type FileHashes struct {
	CRC32  string
	MD5    string
	SHA1   string
	Size   int64
}

// GameInfo holds identified game metadata.
type GameInfo struct {
	Name        string
	System      systems.SystemID
	Region      string
	Serial      string
	Description string
	Publisher   string
	Year        string
	Source      string // "dat", "screenscraper", "manual"
}

// ROMMatch pairs a file with its identified game info.
type ROMMatch struct {
	FilePath string
	Hashes   FileHashes
	Game     *GameInfo
	Matched  bool
}
