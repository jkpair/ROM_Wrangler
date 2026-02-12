package config

import (
	"strings"

	"github.com/kurlmarx/romwrangler/internal/systems"
)

func init() {
	// Automatically register every ReplayOS folder name as an alias.
	// This means folders like "nintendo_snes", "sega_dc", "sony_psx"
	// are recognized in addition to short names like "snes", "dc", "psx".
	for id, folder := range systems.ReplayOSFolders {
		if _, exists := DefaultAliases[folder]; !exists {
			DefaultAliases[folder] = id
		}
	}
}

// DefaultAliases maps common folder/directory names to SystemIDs.
var DefaultAliases = map[string]systems.SystemID{
	// Arcade — 4 separate systems
	"arcade":      systems.ArcadeFBNeo, // default arcade → fbneo
	"fbneo":       systems.ArcadeFBNeo,
	"fba":         systems.ArcadeFBNeo,
	"mame":        systems.ArcadeMAME,
	"mame2003":    systems.ArcadeMAME2K3P,
	"mame2003plus": systems.ArcadeMAME2K3P,
	"mame 2003":   systems.ArcadeMAME2K3P,
	"mame2k3p":    systems.ArcadeMAME2K3P,
	"naomi":       systems.ArcadeDC,
	"atomiswave":  systems.ArcadeDC,

	// Amstrad
	"amstrad": systems.AmstradCPC,
	"cpc":     systems.AmstradCPC,

	// Atari
	"2600":         systems.Atari2600,
	"atari2600":    systems.Atari2600,
	"atari 2600":   systems.Atari2600,
	"vcs":          systems.Atari2600,
	"5200":         systems.Atari5200,
	"atari5200":    systems.Atari5200,
	"atari 5200":   systems.Atari5200,
	"7800":         systems.Atari7800,
	"atari7800":    systems.Atari7800,
	"atari 7800":   systems.Atari7800,
	"lynx":         systems.AtariLynx,
	"atarilynx":    systems.AtariLynx,
	"atari lynx":   systems.AtariLynx,
	"jaguar":       systems.AtariJaguar,
	"atarijaguar":  systems.AtariJaguar,
	"atari jaguar": systems.AtariJaguar,

	// Commodore
	"c64":          systems.CommodoreC64,
	"commodore64":  systems.CommodoreC64,
	"commodore 64": systems.CommodoreC64,
	"amiga":        systems.CommodoreAmiga,
	"amigacd":      systems.CommodoreAmigaCD,
	"amigacd32":    systems.CommodoreAmigaCD,
	"amiga cd32":   systems.CommodoreAmigaCD,
	"cd32":         systems.CommodoreAmigaCD,

	// Microsoft
	"msx":  systems.MSX,
	"msx2": systems.MSX2,

	// NEC
	"pce":           systems.NECPCE,
	"pcengine":      systems.NECPCE,
	"pc engine":     systems.NECPCE,
	"turbografx":    systems.NECPCE,
	"turbografx16":  systems.NECPCE,
	"turbografx-16": systems.NECPCE,
	"tg16":          systems.NECPCE,
	"pcecd":         systems.NECPCECD,
	"pcenginecd":    systems.NECPCECD,
	"turbografxcd":  systems.NECPCECD,
	"tgcd":          systems.NECPCECD,

	// Nintendo
	"nes":          systems.NintendoNES,
	"famicom":      systems.NintendoNES,
	"fc":           systems.NintendoNES,
	"nintendo":     systems.NintendoNES,
	"fds":          systems.NintendoFDS,
	"snes":         systems.NintendoSNES,
	"supernes":     systems.NintendoSNES,
	"superfamicom": systems.NintendoSNES,
	"sfc":          systems.NintendoSNES,
	"n64":          systems.NintendoN64,
	"nintendo64":   systems.NintendoN64,
	"nintendo 64":  systems.NintendoN64,
	"gb":           systems.NintendoGB,
	"gameboy":      systems.NintendoGB,
	"game boy":     systems.NintendoGB,
	"gbc":          systems.NintendoGBC,
	"gameboycolor":     systems.NintendoGBC,
	"game boy color":   systems.NintendoGBC,
	"gba":              systems.NintendoGBA,
	"gameboyadvance":   systems.NintendoGBA,
	"game boy advance": systems.NintendoGBA,
	"nds":          systems.NintendoNDS,
	"ds":           systems.NintendoNDS,
	"nintendods":    systems.NintendoNDS,
	"nintendo ds":  systems.NintendoNDS,

	// Panasonic
	"3do": systems.Panasonic3DO,

	// Philips
	"cdi":  systems.PhilipsCDi,
	"cd-i": systems.PhilipsCDi,

	// Sega
	"sg1000":        systems.SegaSG1000,
	"sg-1000":       systems.SegaSG1000,
	"mastersystem":  systems.SegaMS,
	"master system": systems.SegaMS,
	"sms":           systems.SegaMS,
	"megadrive":     systems.SegaMD,
	"mega drive":    systems.SegaMD,
	"genesis":       systems.SegaMD,
	"md":            systems.SegaMD,
	"gen":           systems.SegaMD,
	"smd":           systems.SegaMD,
	"32x":           systems.Sega32X,
	"sega32x":       systems.Sega32X,
	"segacd":        systems.SegaCD,
	"sega cd":       systems.SegaCD,
	"megacd":        systems.SegaCD,
	"mega cd":       systems.SegaCD,
	"saturn":        systems.SegaSaturn,
	"segasaturn":    systems.SegaSaturn,
	"sega saturn":   systems.SegaSaturn,
	"dreamcast":     systems.SegaDC,
	"dc":            systems.SegaDC,
	"gamegear":      systems.SegaGG,
	"game gear":     systems.SegaGG,
	"gg":            systems.SegaGG,

	// Sharp
	"x68000": systems.SharpX68K,
	"x68k":   systems.SharpX68K,

	// Sinclair
	"zxspectrum":  systems.SinclairZX,
	"zx spectrum": systems.SinclairZX,
	"spectrum":    systems.SinclairZX,

	// SNK
	"neogeo":         systems.SNKNeoGeo,
	"neo geo":        systems.SNKNeoGeo,
	"neo-geo":        systems.SNKNeoGeo,
	"neogeocd":       systems.SNKNeoGeoCD,
	"neo geo cd":     systems.SNKNeoGeoCD,
	"ngp":            systems.SNKNGP,
	"neopocket":      systems.SNKNGP,
	"ngpc":           systems.SNKNGPC,
	"neopocketcolor": systems.SNKNGPC,

	// Sony
	"psx":            systems.SonyPSX,
	"ps1":            systems.SonyPSX,
	"playstation":    systems.SonyPSX,
	"playstation1":   systems.SonyPSX,
	"playstation 1":  systems.SonyPSX,

	// PC
	"dos":    systems.DOSBox,
	"dosbox": systems.DOSBox,

	// Media
	"mediaplayer": systems.MediaPlayer,
	"media":       systems.MediaPlayer,
}

// ResolveAlias resolves a folder name to a SystemID using config overrides first,
// then default aliases. Matching is case-insensitive.
func ResolveAlias(name string, configAliases map[string]string) (systems.SystemID, bool) {
	normalized := strings.ToLower(strings.TrimSpace(name))

	// Check config overrides first
	if configAliases != nil {
		for alias, systemStr := range configAliases {
			if strings.ToLower(alias) == normalized {
				sid := systems.SystemID(systemStr)
				if _, ok := systems.AllSystems[sid]; ok {
					return sid, true
				}
			}
		}
	}

	// Fall back to defaults
	if sid, ok := DefaultAliases[normalized]; ok {
		return sid, true
	}
	return "", false
}
