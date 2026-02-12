package systems

// ReplayOSFolders maps SystemID to the ReplayOS folder name under /roms/.
var ReplayOSFolders = map[SystemID]string{
	// Arcade
	ArcadeFBNeo:    "arcade_fbneo",
	ArcadeMAME:     "arcade_mame",
	ArcadeMAME2K3P: "arcade_mame_2k3p",
	ArcadeDC:       "arcade_dc",

	// Amstrad
	AmstradCPC: "amstrad_cpc",

	// Atari
	Atari2600:   "atari_2600",
	Atari5200:   "atari_5200",
	Atari7800:   "atari_7800",
	AtariJaguar: "atari_jaguar",
	AtariLynx:   "atari_lynx",

	// Commodore
	CommodoreC64:     "commodore_c64",
	CommodoreAmiga:   "commodore_amiga",
	CommodoreAmigaCD: "commodore_amigacd",

	// Microsoft
	MSX:  "microsoft_msx",
	MSX2: "microsoft_msx", // MSX2 shares MSX folder

	// NEC
	NECPCE:   "nec_pce",
	NECPCECD: "nec_pcecd",

	// Nintendo
	NintendoNES:  "nintendo_nes",
	NintendoFDS:  "nintendo_nes", // FDS shares NES folder
	NintendoSNES: "nintendo_snes",
	NintendoN64:  "nintendo_n64",
	NintendoGB:   "nintendo_gb",
	NintendoGBC:  "nintendo_gb", // GBC shares GB folder
	NintendoGBA:  "nintendo_gba",
	NintendoNDS:  "nintendo_ds",

	// Panasonic
	Panasonic3DO: "panasonic_3do",

	// Philips
	PhilipsCDi: "philips_cdi",

	// Sega
	SegaSG1000: "sega_sg",
	SegaMS:     "sega_sms",
	SegaMD:     "sega_smd",
	Sega32X:    "sega_32x",
	SegaCD:     "sega_cd",
	SegaSaturn: "sega_st",
	SegaDC:     "sega_dc",
	SegaGG:     "sega_gg",

	// Sharp
	SharpX68K: "sharp_x68k",

	// Sinclair
	SinclairZX: "sinclair_zx",

	// SNK
	SNKNeoGeo:   "snk_ng",
	SNKNeoGeoCD: "snk_ngcd",
	SNKNGP:      "snk_ngp",
	SNKNGPC:     "snk_ngp", // NGPC shares NGP folder

	// Sony
	SonyPSX: "sony_psx",

	// PC
	DOSBox:  "ibm_pc",
	ScummVM: "scummvm",

	// Media
	MediaPlayer: "media_player",
}

// FolderForSystem returns the ReplayOS folder name for a system.
func FolderForSystem(id SystemID) (string, bool) {
	folder, ok := ReplayOSFolders[id]
	return folder, ok
}
