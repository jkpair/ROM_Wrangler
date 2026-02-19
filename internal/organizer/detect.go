package organizer

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/kurlmarx/romwrangler/internal/scraper"
	"github.com/kurlmarx/romwrangler/internal/systems"
)

// uniqueExtensions maps file extensions that belong to exactly one system.
// Only extensions that unambiguously identify a system are included here;
// shared extensions like .bin, .zip, .iso, .chd, .cue, .m3u are excluded.
var uniqueExtensions = map[string]systems.SystemID{
	// Nintendo
	".nes":  systems.NintendoNES,
	".unf":  systems.NintendoNES,
	".unif": systems.NintendoNES,
	".fds":  systems.NintendoFDS,
	".sfc":  systems.NintendoSNES,
	".smc":  systems.NintendoSNES,
	".swc":  systems.NintendoSNES,
	".fig":  systems.NintendoSNES,
	".n64":  systems.NintendoN64,
	".z64":  systems.NintendoN64,
	".v64":  systems.NintendoN64,
	".gb":   systems.NintendoGB,
	".sgb":  systems.NintendoGB,
	".gbc":  systems.NintendoGBC,
	".sgbc": systems.NintendoGBC,
	".gba":  systems.NintendoGBA,
	".nds":  systems.NintendoNDS,

	// Atari
	".a26": systems.Atari2600,
	".a52": systems.Atari5200,
	".a78": systems.Atari7800,
	".lnx": systems.AtariLynx,
	".j64": systems.AtariJaguar,
	".jag": systems.AtariJaguar,

	// NEC
	".pce": systems.NECPCE,
	".sgx": systems.NECPCE,

	// Sega
	".32x": systems.Sega32X,
	".gg":  systems.SegaGG,
	".sms": systems.SegaMS,
	".md":  systems.SegaMD,
	".smd": systems.SegaMD,
	".gen": systems.SegaMD,

	// Dreamcast-specific (gdi/cdi are unique to DC among supported systems)
	".gdi": systems.SegaDC,
	".cdi": systems.SegaDC,

	// SNK
	".ngp":  systems.SNKNGP,
	".ngc":  systems.SNKNGP,
	".ngpc": systems.SNKNGPC,
	".npc":  systems.SNKNGPC,

	// PC
	".scummvm": systems.ScummVM,
	".svm":     systems.ScummVM,
	".dosz":    systems.DOSBox,

	// Commodore
	".d64": systems.CommodoreC64,
	".t64": systems.CommodoreC64,
	".crt": systems.CommodoreC64,
	".adf": systems.CommodoreAmiga,
	".adz": systems.CommodoreAmiga,
}

// DetectSystemByExtension returns the system for a file based solely on
// its extension. Only works for extensions unique to a single system.
func DetectSystemByExtension(filename string) (systems.SystemID, bool) {
	ext := strings.ToLower(filepath.Ext(filename))
	sys, ok := uniqueExtensions[ext]
	return sys, ok
}

// MisplacedFile describes a ROM file that is in the wrong system folder.
type MisplacedFile struct {
	Path          string
	CurrentSystem systems.SystemID
	CorrectSystem systems.SystemID
	Source        string // "extension" or "screenscraper"
}

// DetectMisplaced finds files in the scan result that are in the wrong
// system folder, based on extension detection.
func DetectMisplaced(result *ScanResult) []MisplacedFile {
	var misplaced []MisplacedFile
	for _, f := range result.Files {
		detected, ok := DetectSystemByExtension(filepath.Base(f.Path))
		if !ok {
			continue
		}
		if detected != f.System {
			misplaced = append(misplaced, MisplacedFile{
				Path:          f.Path,
				CurrentSystem: f.System,
				CorrectSystem: detected,
				Source:        "extension",
			})
		}
	}
	return misplaced
}

// RelocateMisplaced moves misplaced files in the scan result to the correct
// system. It removes them from their current system bucket and adds them
// to the correct one.
func RelocateMisplaced(result *ScanResult, misplaced []MisplacedFile) {
	if len(misplaced) == 0 {
		return
	}

	relocateMap := make(map[string]systems.SystemID, len(misplaced))
	for _, m := range misplaced {
		relocateMap[m.Path] = m.CorrectSystem
	}

	// Update system assignment in Files slice
	for i, f := range result.Files {
		if newSys, ok := relocateMap[f.Path]; ok {
			result.Files[i].System = newSys
		}
	}

	// Rebuild BySystem
	result.BySystem = make(map[systems.SystemID][]ScannedFile)
	for _, f := range result.Files {
		result.BySystem[f.System] = append(result.BySystem[f.System], f)
	}
}

// ResolveUnknown attempts to assign a system to unresolved files (files in
// unrecognized directories or at the source root). It tries extension-based
// detection first, then optionally uses the ScreenScraper API via the
// identifier. Successfully resolved files are added to the scan result.
func ResolveUnknown(ctx context.Context, result *ScanResult, identifier *scraper.Identifier) {
	var stillUnresolved []string

	for _, path := range result.Unresolved {
		// Try extension-based detection first (instant, no API)
		if sysID, ok := DetectSystemByExtension(filepath.Base(path)); ok {
			ext := strings.ToLower(filepath.Ext(path))
			if systems.IsValidFormat(sysID, ext) {
				sf := ScannedFile{Path: path, System: sysID, Resolved: true}
				result.Files = append(result.Files, sf)
				result.BySystem[sysID] = append(result.BySystem[sysID], sf)
				continue
			}
		}

		// Try ScreenScraper if identifier is available
		if identifier != nil {
			match, err := identifier.Identify(ctx, path, "")
			if err == nil && match.Matched && match.Game != nil && match.Game.System != "" {
				sysID := match.Game.System
				sf := ScannedFile{Path: path, System: sysID, Resolved: true}
				result.Files = append(result.Files, sf)
				result.BySystem[sysID] = append(result.BySystem[sysID], sf)
				continue
			}
		}

		stillUnresolved = append(stillUnresolved, path)
	}
	result.Unresolved = stillUnresolved
}
