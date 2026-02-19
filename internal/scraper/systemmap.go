package scraper

import "github.com/kurlmarx/romwrangler/internal/systems"

// screenScraperSystems maps ScreenScraper API system IDs (integers)
// to internal SystemID constants. Only systems supported by ReplayOS
// are included.
//
// Reference: https://api.screenscraper.fr/api2/systemesListe.php
var screenScraperSystems = map[int]systems.SystemID{
	// Arcade
	75:  systems.ArcadeFBNeo,    // FBNeo
	137: systems.ArcadeMAME,     // MAME
	230: systems.ArcadeMAME2K3P, // MAME 2003+
	53:  systems.ArcadeDC,       // Naomi / Atomiswave

	// Amstrad
	65: systems.AmstradCPC,

	// Atari
	26: systems.Atari2600,
	40: systems.Atari5200,
	41: systems.Atari7800,
	28: systems.AtariLynx,
	27: systems.AtariJaguar,

	// Commodore
	66:  systems.CommodoreC64,
	64:  systems.CommodoreAmiga,
	130: systems.CommodoreAmigaCD,

	// Microsoft
	113: systems.MSX,
	116: systems.MSX2,

	// NEC
	31: systems.NECPCE,
	114: systems.NECPCECD,

	// Nintendo
	3:  systems.NintendoNES,
	106: systems.NintendoFDS,
	4:  systems.NintendoSNES,
	14: systems.NintendoN64,
	9:  systems.NintendoGB,
	10: systems.NintendoGBC,
	12: systems.NintendoGBA,
	15: systems.NintendoNDS,

	// Panasonic
	29: systems.Panasonic3DO,

	// Philips
	133: systems.PhilipsCDi,

	// Sega
	109: systems.SegaSG1000,
	2:   systems.SegaMS,
	1:   systems.SegaMD,
	19:  systems.Sega32X,
	20:  systems.SegaCD,
	22:  systems.SegaSaturn,
	23:  systems.SegaDC,
	21:  systems.SegaGG,

	// Sharp
	79: systems.SharpX68K,

	// Sinclair
	76: systems.SinclairZX,

	// SNK
	142: systems.SNKNeoGeo,
	70:  systems.SNKNeoGeoCD,
	25:  systems.SNKNGP,
	82:  systems.SNKNGPC,

	// Sony
	57: systems.SonyPSX,

	// PC
	135: systems.DOSBox,
	123: systems.ScummVM,
}

// ScreenScraperToSystemID converts a ScreenScraper API system ID (integer)
// to an internal SystemID. Returns false if the ID is unknown or unsupported.
func ScreenScraperToSystemID(ssID int) (systems.SystemID, bool) {
	sys, ok := screenScraperSystems[ssID]
	return sys, ok
}
