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
	// Atari
	"2600":           systems.Atari2600,
	"atari2600":      systems.Atari2600,
	"atari 2600":     systems.Atari2600,
	"vcs":            systems.Atari2600,
	"5200":           systems.Atari5200,
	"atari5200":      systems.Atari5200,
	"atari 5200":     systems.Atari5200,
	"7800":           systems.Atari7800,
	"atari7800":      systems.Atari7800,
	"atari 7800":     systems.Atari7800,
	"lynx":           systems.AtariLynx,
	"atarilynx":      systems.AtariLynx,
	"atari lynx":     systems.AtariLynx,
	"jaguar":         systems.AtariJaguar,
	"atarijaguar":    systems.AtariJaguar,
	"atari jaguar":   systems.AtariJaguar,
	"atarist":        systems.AtariST,
	"atari st":       systems.AtariST,

	// Bandai
	"wonderswan":       systems.BandaiWS,
	"ws":               systems.BandaiWS,
	"wonderswancolor":  systems.BandaiWSC,
	"wonderswan color": systems.BandaiWSC,
	"wsc":              systems.BandaiWSC,

	// Coleco
	"colecovision": systems.ColecoVision,
	"coleco":       systems.ColecoVision,

	// Commodore
	"c64":            systems.CommodoreC64,
	"commodore64":    systems.CommodoreC64,
	"commodore 64":   systems.CommodoreC64,
	"amiga":          systems.CommodoreAmiga,

	// GCE
	"vectrex": systems.GCEVectrex,

	// Magnavox
	"odyssey2":  systems.MagnavoxOdyssey2,
	"odyssey 2": systems.MagnavoxOdyssey2,
	"cdi":       systems.PhilipsCDi,
	"cd-i":      systems.PhilipsCDi,

	// Mattel
	"intellivision": systems.MattelIntv,
	"intv":          systems.MattelIntv,

	// Microsoft
	"xbox": systems.MicrosoftXbox,

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
	"supergrafx":    systems.NECSGRFX,
	"sgrfx":         systems.NECSGRFX,
	"pcfx":          systems.NECPCFX,
	"pc-fx":         systems.NECPCFX,

	// Nintendo
	"nes":         systems.NintendoNES,
	"famicom":     systems.NintendoNES,
	"fc":          systems.NintendoNES,
	"nintendo":    systems.NintendoNES,
	"fds":         systems.NintendoFDS,
	"snes":        systems.NintendoSNES,
	"supernes":    systems.NintendoSNES,
	"superfamicom": systems.NintendoSNES,
	"sfc":         systems.NintendoSNES,
	"n64":         systems.NintendoN64,
	"nintendo64":  systems.NintendoN64,
	"nintendo 64": systems.NintendoN64,
	"gamecube":    systems.NintendoGC,
	"gc":          systems.NintendoGC,
	"ngc":         systems.NintendoGC,
	"wii":         systems.NintendoWii,
	"gb":          systems.NintendoGB,
	"gameboy":     systems.NintendoGB,
	"game boy":    systems.NintendoGB,
	"gbc":         systems.NintendoGBC,
	"gameboycolor": systems.NintendoGBC,
	"game boy color": systems.NintendoGBC,
	"gba":         systems.NintendoGBA,
	"gameboyadvance": systems.NintendoGBA,
	"game boy advance": systems.NintendoGBA,
	"nds":         systems.NintendoNDS,
	"ds":          systems.NintendoNDS,
	"nintendods":  systems.NintendoNDS,
	"nintendo ds": systems.NintendoNDS,
	"virtualboy":  systems.NintendoVB,
	"virtual boy": systems.NintendoVB,
	"vb":          systems.NintendoVB,
	"pokemini":    systems.NintendoPokeMini,
	"pokemon mini": systems.NintendoPokeMini,

	// Panasonic
	"3do": systems.Panasonic3DO,

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

	// SNK
	"neogeo":      systems.SNKNeoGeo,
	"neo geo":     systems.SNKNeoGeo,
	"neo-geo":     systems.SNKNeoGeo,
	"neogeocd":    systems.SNKNeoGeoCD,
	"neo geo cd":  systems.SNKNeoGeoCD,
	"ngp":         systems.SNKNGP,
	"neopocket":   systems.SNKNGP,
	"ngpc":        systems.SNKNGPC,
	"neopocketcolor": systems.SNKNGPC,

	// Sony
	"psx":          systems.SonyPSX,
	"ps1":          systems.SonyPSX,
	"playstation":  systems.SonyPSX,
	"playstation1": systems.SonyPSX,
	"playstation 1": systems.SonyPSX,
	"ps2":          systems.SonyPS2,
	"playstation2": systems.SonyPS2,
	"playstation 2": systems.SonyPS2,
	"psp":          systems.SonyPSP,

	// Misc
	"dos":     systems.DOSBox,
	"dosbox":  systems.DOSBox,
	"scummvm": systems.ScummVM,
	"msx":     systems.MSX,
	"msx2":    systems.MSX2,

	// Arcade
	"arcade":  systems.Arcade,
	"mame":    systems.Arcade,
	"fbneo":   systems.Arcade,
	"fba":     systems.Arcade,
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
