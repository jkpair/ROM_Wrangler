package systems

// ReplayOSFolders maps SystemID to the ReplayOS folder name under /roms/.
var ReplayOSFolders = map[SystemID]string{
	// Atari
	Atari2600:   "atari_2600",
	Atari5200:   "atari_5200",
	Atari7800:   "atari_7800",
	AtariLynx:   "atari_lynx",
	AtariJaguar: "atari_jaguar",
	AtariST:     "atari_st",

	// Bandai
	BandaiWS:  "bandai_ws",
	BandaiWSC: "bandai_wsc",

	// Coleco
	ColecoVision: "coleco_vision",

	// Commodore
	CommodoreC64:   "commodore_c64",
	CommodoreAmiga: "commodore_amiga",

	// GCE
	GCEVectrex: "gce_vectrex",

	// Magnavox / Philips
	MagnavoxOdyssey2: "magnavox_odyssey2",
	PhilipsCDi:       "philips_cdi",

	// Mattel
	MattelIntv: "mattel_intv",

	// Microsoft
	MicrosoftXbox: "microsoft_xbox",

	// NEC
	NECPCE:   "nec_pce",
	NECPCECD: "nec_pce_cd",
	NECSGRFX: "nec_sgrfx",
	NECPCFX:  "nec_pcfx",

	// Nintendo
	NintendoNES:      "nintendo_nes",
	NintendoFDS:      "nintendo_fds",
	NintendoSNES:     "nintendo_snes",
	NintendoN64:      "nintendo_n64",
	NintendoGC:       "nintendo_gc",
	NintendoWii:      "nintendo_wii",
	NintendoGB:       "nintendo_gb",
	NintendoGBC:      "nintendo_gbc",
	NintendoGBA:      "nintendo_gba",
	NintendoNDS:      "nintendo_nds",
	NintendoVB:       "nintendo_vb",
	NintendoPokeMini: "nintendo_pokemini",

	// Panasonic
	Panasonic3DO: "panasonic_3do",

	// Sega
	SegaSG1000: "sega_sg1000",
	SegaMS:     "sega_ms",
	SegaMD:     "sega_md",
	Sega32X:    "sega_32x",
	SegaCD:     "sega_cd",
	SegaSaturn: "sega_saturn",
	SegaDC:     "sega_dc",
	SegaGG:     "sega_gg",

	// SNK
	SNKNeoGeo:   "snk_neogeo",
	SNKNeoGeoCD: "snk_neogeo_cd",
	SNKNGP:      "snk_ngp",
	SNKNGPC:     "snk_ngpc",

	// Sony
	SonyPSX: "sony_psx",
	SonyPS2: "sony_ps2",
	SonyPSP: "sony_psp",

	// Misc
	DOSBox:  "pc_dos",
	ScummVM: "pc_scummvm",
	MSX:     "msx",
	MSX2:    "msx2",

	// Arcade
	Arcade: "arcade",
}

// FolderForSystem returns the ReplayOS folder name for a system.
func FolderForSystem(id SystemID) (string, bool) {
	folder, ok := ReplayOSFolders[id]
	return folder, ok
}
