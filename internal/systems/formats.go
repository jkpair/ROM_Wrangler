package systems

// SupportedFormats maps each system to its accepted file extensions.
// These match the ReplayOS documentation exactly.
var SupportedFormats = map[SystemID][]string{
	// Arcade — zip is the native ROM format (no extraction)
	ArcadeFBNeo:    {".zip"},
	ArcadeMAME:     {".zip"},
	ArcadeMAME2K3P: {".zip"},
	ArcadeDC:       {".zip"},

	// Amstrad
	AmstradCPC: {".dsk", ".sna", ".tap", ".cdt", ".voc", ".cpr", ".m3u"},

	// Atari
	Atari2600:   {".a26", ".bin"},
	Atari5200:   {".a52", ".bin"},
	Atari7800:   {".a78", ".bin", ".cdf"},
	AtariJaguar: {".j64", ".jag"},
	AtariLynx:   {".lnx"},

	// Commodore
	CommodoreC64:     {".d64", ".d71", ".d80", ".d81", ".d82", ".g64", ".g41", ".x64", ".t64", ".tap", ".prg", ".p00", ".crt", ".bin", ".gz", ".m3u"},
	CommodoreAmiga:   {".adf", ".adz", ".dms", ".fdi", ".raw", ".hdf", ".hdz", ".lha", ".slave", ".info", ".uae", ".m3u"},
	CommodoreAmigaCD: {".cue", ".ccd", ".nrg", ".mds", ".iso", ".chd", ".m3u"},

	// Microsoft
	MSX:  {".rom", ".ri", ".mx1", ".mx2", ".dsk", ".col", ".sg", ".sc", ".sf", ".cas", ".m3u"},
	MSX2: {".rom", ".ri", ".mx1", ".mx2", ".dsk", ".col", ".sg", ".sc", ".sf", ".cas", ".m3u"},

	// NEC
	NECPCE:   {".pce", ".sgx", ".toc"},
	NECPCECD: {".cue", ".ccd", ".chd", ".m3u"},

	// Nintendo
	NintendoNES:  {".fds", ".nes", ".unf", ".unif"},
	NintendoFDS:  {".fds", ".nes", ".unf", ".unif"},
	NintendoSNES: {".smc", ".sfc", ".swc", ".fig", ".bs", ".st"},
	NintendoN64:  {".n64", ".v64", ".z64", ".bin", ".u1"},
	NintendoGB:   {".gb", ".sgb"},
	NintendoGBC:  {".gbc", ".sgbc"},
	NintendoGBA:  {".gba"},
	NintendoNDS:  {".nds"},

	// Panasonic
	Panasonic3DO: {".iso", ".chd", ".cue"},

	// Philips
	PhilipsCDi: {".iso", ".chd", ".cue"},

	// Sega
	SegaSG1000: {".sg"},
	SegaGG:     {".gg"},
	SegaMS:     {".sms"},
	SegaMD:     {".md", ".smd", ".gen", ".bin"},
	SegaCD:     {".m3u", ".cue", ".iso", ".chd"},
	Sega32X:    {".32x"},
	SegaSaturn: {".cue", ".ccd", ".chd", ".toc", ".m3u"},
	SegaDC:     {".chd", ".cdi", ".elf", ".cue", ".gdi", ".lst", ".dat", ".m3u"},

	// Sharp
	SharpX68K: {".dim", ".img", ".d88", ".88d", ".hdm", ".dup", ".2hd", ".xdf", ".hdf", ".cmd", ".m3u"},

	// Sinclair
	SinclairZX: {".tzx", ".tap", ".z80", ".rzx", ".scl", ".trd", ".dsk", ".dck", ".sna", ".szx"},

	// SNK — zip is the native ROM format for Neo Geo
	SNKNeoGeo:   {".zip"},
	SNKNeoGeoCD: {".cue", ".chd"},
	SNKNGP:      {".ngp", ".ngc", ".ngpc", ".npc"},
	SNKNGPC:     {".ngp", ".ngc", ".ngpc", ".npc"},

	// Sony
	SonyPSX: {".exe", ".psexe", ".cue", ".img", ".iso", ".chd", ".pbp", ".mds", ".psf", ".m3u"},

	// PC — zip is a native format for DOS
	DOSBox:  {".zip", ".dosz", ".exe", ".com", ".bat", ".iso", ".cue", ".img", ".m3u", ".m3u8"},
	ScummVM: {".scummvm", ".svm"},

	// Media
	MediaPlayer: {".mkv", ".avi", ".f4v", ".mp4", ".mp3", ".flac", ".ogg", ".wav"},
}

// IsValidFormat checks if a file extension is supported for a system.
func IsValidFormat(system SystemID, ext string) bool {
	formats, ok := SupportedFormats[system]
	if !ok {
		return false
	}
	for _, f := range formats {
		if f == ext {
			return true
		}
	}
	return false
}
