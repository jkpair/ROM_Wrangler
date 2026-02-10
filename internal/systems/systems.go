package systems

// SystemID is a unique identifier for a gaming system.
type SystemID string

const (
	// Atari
	Atari2600   SystemID = "atari_2600"
	Atari5200   SystemID = "atari_5200"
	Atari7800   SystemID = "atari_7800"
	AtariLynx   SystemID = "atari_lynx"
	AtariJaguar SystemID = "atari_jaguar"
	AtariST     SystemID = "atari_st"

	// Bandai
	BandaiWS  SystemID = "bandai_ws"
	BandaiWSC SystemID = "bandai_wsc"

	// Coleco
	ColecoVision SystemID = "coleco_vision"

	// Commodore
	CommodoreC64   SystemID = "commodore_c64"
	CommodoreAmiga SystemID = "commodore_amiga"

	// GCE
	GCEVectrex SystemID = "gce_vectrex"

	// Magnavox / Philips
	MagnavoxOdyssey2 SystemID = "magnavox_odyssey2"
	PhilipsCDi       SystemID = "philips_cdi"

	// Mattel
	MattelIntv SystemID = "mattel_intv"

	// Microsoft
	MicrosoftXbox SystemID = "microsoft_xbox"

	// NEC
	NECPCE    SystemID = "nec_pce"
	NECPCECD  SystemID = "nec_pce_cd"
	NECSGRFX  SystemID = "nec_sgrfx"
	NECPCFX   SystemID = "nec_pcfx"

	// Nintendo
	NintendoNES    SystemID = "nintendo_nes"
	NintendoFDS    SystemID = "nintendo_fds"
	NintendoSNES   SystemID = "nintendo_snes"
	NintendoN64    SystemID = "nintendo_n64"
	NintendoGC     SystemID = "nintendo_gc"
	NintendoWii    SystemID = "nintendo_wii"
	NintendoGB     SystemID = "nintendo_gb"
	NintendoGBC    SystemID = "nintendo_gbc"
	NintendoGBA    SystemID = "nintendo_gba"
	NintendoNDS    SystemID = "nintendo_nds"
	NintendoVB     SystemID = "nintendo_vb"
	NintendoPokeMini SystemID = "nintendo_pokemini"

	// Panasonic
	Panasonic3DO SystemID = "panasonic_3do"

	// Sega
	SegaSG1000 SystemID = "sega_sg1000"
	SegaMS     SystemID = "sega_ms"
	SegaMD     SystemID = "sega_md"
	Sega32X    SystemID = "sega_32x"
	SegaCD     SystemID = "sega_cd"
	SegaSaturn SystemID = "sega_saturn"
	SegaDC     SystemID = "sega_dc"
	SegaGG     SystemID = "sega_gg"

	// SNK
	SNKNeoGeo   SystemID = "snk_neogeo"
	SNKNeoGeoCD SystemID = "snk_neogeo_cd"
	SNKNGP      SystemID = "snk_ngp"
	SNKNGPC     SystemID = "snk_ngpc"

	// Sony
	SonyPSX SystemID = "sony_psx"
	SonyPS2 SystemID = "sony_ps2"
	SonyPSP SystemID = "sony_psp"

	// Misc/Computer
	DOSBox  SystemID = "pc_dos"
	ScummVM SystemID = "pc_scummvm"
	MSX     SystemID = "msx"
	MSX2    SystemID = "msx2"

	// Arcade
	Arcade SystemID = "arcade"
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
	// Atari
	Atari2600:   {Atari2600, "Atari 2600", "Atari", false},
	Atari5200:   {Atari5200, "Atari 5200", "Atari", false},
	Atari7800:   {Atari7800, "Atari 7800", "Atari", false},
	AtariLynx:   {AtariLynx, "Atari Lynx", "Atari", false},
	AtariJaguar: {AtariJaguar, "Atari Jaguar", "Atari", false},
	AtariST:     {AtariST, "Atari ST", "Atari", false},

	// Bandai
	BandaiWS:  {BandaiWS, "WonderSwan", "Bandai", false},
	BandaiWSC: {BandaiWSC, "WonderSwan Color", "Bandai", false},

	// Coleco
	ColecoVision: {ColecoVision, "ColecoVision", "Coleco", false},

	// Commodore
	CommodoreC64:   {CommodoreC64, "Commodore 64", "Commodore", false},
	CommodoreAmiga: {CommodoreAmiga, "Amiga", "Commodore", false},

	// GCE
	GCEVectrex: {GCEVectrex, "Vectrex", "GCE", false},

	// Magnavox / Philips
	MagnavoxOdyssey2: {MagnavoxOdyssey2, "Odyssey 2", "Magnavox", false},
	PhilipsCDi:       {PhilipsCDi, "CD-i", "Philips", true},

	// Mattel
	MattelIntv: {MattelIntv, "Intellivision", "Mattel", false},

	// Microsoft
	MicrosoftXbox: {MicrosoftXbox, "Xbox", "Microsoft", true},

	// NEC
	NECPCE:   {NECPCE, "PC Engine / TurboGrafx-16", "NEC", false},
	NECPCECD: {NECPCECD, "PC Engine CD / TurboGrafx-CD", "NEC", true},
	NECSGRFX: {NECSGRFX, "SuperGrafx", "NEC", false},
	NECPCFX:  {NECPCFX, "PC-FX", "NEC", true},

	// Nintendo
	NintendoNES:      {NintendoNES, "Nintendo Entertainment System", "Nintendo", false},
	NintendoFDS:      {NintendoFDS, "Famicom Disk System", "Nintendo", false},
	NintendoSNES:     {NintendoSNES, "Super Nintendo", "Nintendo", false},
	NintendoN64:      {NintendoN64, "Nintendo 64", "Nintendo", false},
	NintendoGC:       {NintendoGC, "GameCube", "Nintendo", true},
	NintendoWii:      {NintendoWii, "Wii", "Nintendo", true},
	NintendoGB:       {NintendoGB, "Game Boy", "Nintendo", false},
	NintendoGBC:      {NintendoGBC, "Game Boy Color", "Nintendo", false},
	NintendoGBA:      {NintendoGBA, "Game Boy Advance", "Nintendo", false},
	NintendoNDS:      {NintendoNDS, "Nintendo DS", "Nintendo", false},
	NintendoVB:       {NintendoVB, "Virtual Boy", "Nintendo", false},
	NintendoPokeMini: {NintendoPokeMini, "Pokemon Mini", "Nintendo", false},

	// Panasonic
	Panasonic3DO: {Panasonic3DO, "3DO", "Panasonic", true},

	// Sega
	SegaSG1000: {SegaSG1000, "SG-1000", "Sega", false},
	SegaMS:     {SegaMS, "Master System", "Sega", false},
	SegaMD:     {SegaMD, "Mega Drive / Genesis", "Sega", false},
	Sega32X:    {Sega32X, "32X", "Sega", false},
	SegaCD:     {SegaCD, "Sega CD / Mega CD", "Sega", true},
	SegaSaturn: {SegaSaturn, "Saturn", "Sega", true},
	SegaDC:     {SegaDC, "Dreamcast", "Sega", true},
	SegaGG:     {SegaGG, "Game Gear", "Sega", false},

	// SNK
	SNKNeoGeo:   {SNKNeoGeo, "Neo Geo", "SNK", false},
	SNKNeoGeoCD: {SNKNeoGeoCD, "Neo Geo CD", "SNK", true},
	SNKNGP:      {SNKNGP, "Neo Geo Pocket", "SNK", false},
	SNKNGPC:     {SNKNGPC, "Neo Geo Pocket Color", "SNK", false},

	// Sony
	SonyPSX: {SonyPSX, "PlayStation", "Sony", true},
	SonyPS2: {SonyPS2, "PlayStation 2", "Sony", true},
	SonyPSP: {SonyPSP, "PlayStation Portable", "Sony", false},

	// Misc
	DOSBox:  {DOSBox, "DOS", "PC", false},
	ScummVM: {ScummVM, "ScummVM", "PC", false},
	MSX:     {MSX, "MSX", "Various", false},
	MSX2:    {MSX2, "MSX2", "Various", false},

	// Arcade
	Arcade: {Arcade, "Arcade", "Various", false},
}

// GetSystem returns SystemInfo for a given ID, or false if not found.
func GetSystem(id SystemID) (SystemInfo, bool) {
	info, ok := AllSystems[id]
	return info, ok
}
