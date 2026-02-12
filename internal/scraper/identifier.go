package scraper

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/kurlmarx/romwrangler/internal/systems"
)

// Identifier orchestrates ROM identification via multiple sources.
type Identifier struct {
	datIndices    []*DATIndex
	ssClient      *ScreenScraperClient
	cache         Cache
}

// Cache is an interface for storing/retrieving identification results.
type Cache interface {
	GetByHash(sha1 string) (*GameInfo, bool)
	Put(sha1 string, info *GameInfo) error
}

// NewIdentifier creates a new Identifier.
func NewIdentifier(datIndices []*DATIndex, ssClient *ScreenScraperClient, cache Cache) *Identifier {
	return &Identifier{
		datIndices: datIndices,
		ssClient:   ssClient,
		cache:      cache,
	}
}

// Identify attempts to identify a ROM file. Order:
// 1. Check cache
// 2. Hash the file
// 3. Try DAT files
// 4. Try ScreenScraper API
// 5. Cache and return result
func (id *Identifier) Identify(ctx context.Context, filePath string, systemID systems.SystemID) (*ROMMatch, error) {
	match := &ROMMatch{FilePath: filePath}

	// Hash the file
	hashes, err := HashFile(ctx, filePath)
	if err != nil {
		return nil, err
	}
	match.Hashes = hashes

	// Check cache
	if id.cache != nil {
		if info, ok := id.cache.GetByHash(hashes.SHA1); ok {
			match.Game = info
			match.Matched = true
			return match, nil
		}
	}

	// Try DAT files
	for _, idx := range id.datIndices {
		if entry, ok := idx.Lookup(hashes); ok {
			info := &GameInfo{
				Name:   cleanGameName(entry.GameName),
				System: systemID,
				Source: "dat",
			}
			match.Game = info
			match.Matched = true
			if id.cache != nil {
				id.cache.Put(hashes.SHA1, info)
			}
			return match, nil
		}
	}

	// Try ScreenScraper
	if id.ssClient != nil {
		info, err := id.ssClient.Identify(ctx, hashes, systemID)
		if err != nil {
			// Non-fatal: just means we couldn't identify
			return match, nil
		}
		if info != nil {
			match.Game = info
			match.Matched = true
			if id.cache != nil {
				id.cache.Put(hashes.SHA1, info)
			}
			return match, nil
		}
	}

	return match, nil
}

// cleanGameName normalizes a game name from DAT entry.
func cleanGameName(name string) string {
	name = strings.TrimSpace(name)
	// Remove file extension if present
	ext := filepath.Ext(name)
	if ext != "" {
		name = strings.TrimSuffix(name, ext)
	}
	return name
}
