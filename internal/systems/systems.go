package systems

// SystemID is a unique identifier for a gaming system.
type SystemID string

const (
	// Arcade
	ArcadeFBNeo    SystemID = "arcade_fbneo"
	ArcadeMAME     SystemID = "arcade_mame"
	ArcadeMAME2K3P SystemID = "arcade_mame_2k3p"
	ArcadeDC       SystemID = "arcade_dc"

	// Amstrad
	AmstradCPC SystemID = "amstrad_cpc"

	// Atari
	Atari2600   SystemID = "atari_2600"
	Atari5200   SystemID = "atari_5200"
	Atari7800   SystemID = "atari_7800"
	AtariLynx   SystemID = "atari_lynx"
	AtariJaguar SystemID = "atari_jaguar"

	// Commodore
	CommodoreC64     SystemID = "commodore_c64"
	CommodoreAmiga   SystemID = "commodore_amiga"
	CommodoreAmigaCD SystemID = "commodore_amigacd"

	// Microsoft
	MSX  SystemID = "microsoft_msx"
	MSX2 SystemID = "microsoft_msx2"

	// NEC
	NECPCE   SystemID = "nec_pce"
	NECPCECD SystemID = "nec_pcecd"

	// Nintendo
	NintendoNES  SystemID = "nintendo_nes"
	NintendoFDS  SystemID = "nintendo_fds"
	NintendoSNES SystemID = "nintendo_snes"
	NintendoN64  SystemID = "nintendo_n64"
	NintendoGB   SystemID = "nintendo_gb"
	NintendoGBC  SystemID = "nintendo_gbc"
	NintendoGBA  SystemID = "nintendo_gba"
	NintendoNDS  SystemID = "nintendo_ds"

	// Panasonic
	Panasonic3DO SystemID = "panasonic_3do"

	// Philips
	PhilipsCDi SystemID = "philips_cdi"

	// Sega
	SegaSG1000 SystemID = "sega_sg"
	SegaMS     SystemID = "sega_sms"
	SegaMD     SystemID = "sega_smd"
	Sega32X    SystemID = "sega_32x"
	SegaCD     SystemID = "sega_cd"
	SegaSaturn SystemID = "sega_st"
	SegaDC     SystemID = "sega_dc"
	SegaGG     SystemID = "sega_gg"

	// Sharp
	SharpX68K SystemID = "sharp_x68k"

	// Sinclair
	SinclairZX SystemID = "sinclair_zx"

	// SNK
	SNKNeoGeo   SystemID = "snk_ng"
	SNKNeoGeoCD SystemID = "snk_ngcd"
	SNKNGP      SystemID = "snk_ngp"
	SNKNGPC     SystemID = "snk_ngpc"

	// Sony
	SonyPSX SystemID = "sony_psx"

	// PC
	DOSBox  SystemID = "ibm_pc"
	ScummVM SystemID = "scummvm"

	// Media
	MediaPlayer SystemID = "media_player"
)

// SystemInfo holds metadata about a gaming system.
type SystemInfo struct {
	ID          SystemID
	DisplayName string
	Company     string
	IsDiscBased bool
}

// AllSystems returns metadata for all known systems.
var AllSystems = map[SystemID]SystemInfo{
	// Arcade
	ArcadeFBNeo:    {ArcadeFBNeo, "Arcade (FBNeo)", "Various", false},
	ArcadeMAME:     {ArcadeMAME, "Arcade (MAME)", "Various", false},
	ArcadeMAME2K3P: {ArcadeMAME2K3P, "Arcade (MAME 2K3+)", "Various", false},
	ArcadeDC:       {ArcadeDC, "Arcade (Naomi/Atomiswave)", "Sega", false},

	// Amstrad
	AmstradCPC: {AmstradCPC, "Amstrad CPC", "Amstrad", false},

	// Atari
	Atari2600:   {Atari2600, "Atari 2600", "Atari", false},
	Atari5200:   {Atari5200, "Atari 5200", "Atari", false},
	Atari7800:   {Atari7800, "Atari 7800", "Atari", false},
	AtariLynx:   {AtariLynx, "Atari Lynx", "Atari", false},
	AtariJaguar: {AtariJaguar, "Atari Jaguar", "Atari", false},

	// Commodore
	CommodoreC64:     {CommodoreC64, "Commodore 64", "Commodore", false},
	CommodoreAmiga:   {CommodoreAmiga, "Amiga", "Commodore", false},
	CommodoreAmigaCD: {CommodoreAmigaCD, "Amiga CD32", "Commodore", true},

	// Microsoft
	MSX:  {MSX, "MSX", "Various", false},
	MSX2: {MSX2, "MSX2", "Various", false},

	// NEC
	NECPCE:   {NECPCE, "PC Engine / TurboGrafx-16", "NEC", false},
	NECPCECD: {NECPCECD, "PC Engine CD / TurboGrafx-CD", "NEC", true},

	// Nintendo
	NintendoNES:  {NintendoNES, "Nintendo Entertainment System", "Nintendo", false},
	NintendoFDS:  {NintendoFDS, "Famicom Disk System", "Nintendo", false},
	NintendoSNES: {NintendoSNES, "Super Nintendo", "Nintendo", false},
	NintendoN64:  {NintendoN64, "Nintendo 64", "Nintendo", false},
	NintendoGB:   {NintendoGB, "Game Boy", "Nintendo", false},
	NintendoGBC:  {NintendoGBC, "Game Boy Color", "Nintendo", false},
	NintendoGBA:  {NintendoGBA, "Game Boy Advance", "Nintendo", false},
	NintendoNDS:  {NintendoNDS, "Nintendo DS", "Nintendo", false},

	// Panasonic
	Panasonic3DO: {Panasonic3DO, "3DO", "Panasonic", true},

	// Philips
	PhilipsCDi: {PhilipsCDi, "CD-i", "Philips", true},

	// Sega
	SegaSG1000: {SegaSG1000, "SG-1000", "Sega", false},
	SegaMS:     {SegaMS, "Master System", "Sega", false},
	SegaMD:     {SegaMD, "Mega Drive / Genesis", "Sega", false},
	Sega32X:    {Sega32X, "32X", "Sega", false},
	SegaCD:     {SegaCD, "Sega CD / Mega CD", "Sega", true},
	SegaSaturn: {SegaSaturn, "Saturn", "Sega", true},
	SegaDC:     {SegaDC, "Dreamcast", "Sega", true},
	SegaGG:     {SegaGG, "Game Gear", "Sega", false},

	// Sharp
	SharpX68K: {SharpX68K, "Sharp X68000", "Sharp", false},

	// Sinclair
	SinclairZX: {SinclairZX, "ZX Spectrum", "Sinclair", false},

	// SNK
	SNKNeoGeo:   {SNKNeoGeo, "Neo Geo", "SNK", false},
	SNKNeoGeoCD: {SNKNeoGeoCD, "Neo Geo CD", "SNK", true},
	SNKNGP:      {SNKNGP, "Neo Geo Pocket", "SNK", false},
	SNKNGPC:     {SNKNGPC, "Neo Geo Pocket Color", "SNK", false},

	// Sony
	SonyPSX: {SonyPSX, "PlayStation", "Sony", true},

	// PC
	DOSBox:  {DOSBox, "DOS", "PC", false},
	ScummVM: {ScummVM, "ScummVM", "PC", false},

	// Media
	MediaPlayer: {MediaPlayer, "Alpha Player", "Media", false},
}

// GetSystem returns SystemInfo for a given ID, or false if not found.
func GetSystem(id SystemID) (SystemInfo, bool) {
	info, ok := AllSystems[id]
	return info, ok
}
