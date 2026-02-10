package systems

// SupportedFormats maps each system to its accepted file extensions.
var SupportedFormats = map[SystemID][]string{
	// Atari
	Atari2600:   {".a26", ".bin", ".rom", ".zip", ".7z"},
	Atari5200:   {".a52", ".bin", ".rom", ".zip", ".7z"},
	Atari7800:   {".a78", ".bin", ".rom", ".zip", ".7z"},
	AtariLynx:   {".lnx", ".o", ".zip", ".7z"},
	AtariJaguar: {".j64", ".jag", ".rom", ".zip", ".7z"},
	AtariST:     {".st", ".stx", ".msa", ".dim", ".zip", ".7z"},

	// Bandai
	BandaiWS:  {".ws", ".zip", ".7z"},
	BandaiWSC: {".wsc", ".ws", ".zip", ".7z"},

	// Coleco
	ColecoVision: {".col", ".rom", ".bin", ".zip", ".7z"},

	// Commodore
	CommodoreC64:   {".d64", ".t64", ".tap", ".prg", ".crt", ".zip", ".7z"},
	CommodoreAmiga: {".adf", ".adz", ".dms", ".hdf", ".hdz", ".lha", ".zip", ".7z"},

	// GCE
	GCEVectrex: {".vec", ".gam", ".bin", ".zip", ".7z"},

	// Magnavox / Philips
	MagnavoxOdyssey2: {".o2", ".bin", ".zip", ".7z"},
	PhilipsCDi:       {".chd", ".cue", ".bin", ".iso"},

	// Mattel
	MattelIntv: {".int", ".bin", ".rom", ".zip", ".7z"},

	// Microsoft
	MicrosoftXbox: {".iso", ".chd"},

	// NEC
	NECPCE:   {".pce", ".sgx", ".zip", ".7z"},
	NECPCECD: {".chd", ".cue", ".bin", ".iso"},
	NECSGRFX: {".pce", ".sgx", ".zip", ".7z"},
	NECPCFX:  {".chd", ".cue", ".bin", ".iso"},

	// Nintendo
	NintendoNES:      {".nes", ".unf", ".unif", ".fds", ".zip", ".7z"},
	NintendoFDS:      {".fds", ".zip", ".7z"},
	NintendoSNES:     {".sfc", ".smc", ".zip", ".7z"},
	NintendoN64:      {".n64", ".z64", ".v64", ".zip", ".7z"},
	NintendoGC:       {".iso", ".gcm", ".chd", ".rvz", ".ciso"},
	NintendoWii:      {".iso", ".wbfs", ".chd", ".rvz", ".ciso"},
	NintendoGB:       {".gb", ".zip", ".7z"},
	NintendoGBC:      {".gbc", ".gb", ".zip", ".7z"},
	NintendoGBA:      {".gba", ".zip", ".7z"},
	NintendoNDS:      {".nds", ".zip", ".7z"},
	NintendoVB:       {".vb", ".vboy", ".zip", ".7z"},
	NintendoPokeMini: {".min", ".zip", ".7z"},

	// Panasonic
	Panasonic3DO: {".chd", ".cue", ".bin", ".iso"},

	// Sega
	SegaSG1000: {".sg", ".bin", ".zip", ".7z"},
	SegaMS:     {".sms", ".bin", ".zip", ".7z"},
	SegaMD:     {".md", ".bin", ".gen", ".smd", ".zip", ".7z"},
	Sega32X:    {".32x", ".bin", ".zip", ".7z"},
	SegaCD:     {".chd", ".cue", ".bin", ".iso"},
	SegaSaturn: {".chd", ".cue", ".bin", ".iso"},
	SegaDC:     {".chd", ".gdi", ".cue", ".bin", ".cdi"},
	SegaGG:     {".gg", ".bin", ".zip", ".7z"},

	// SNK
	SNKNeoGeo:   {".zip", ".7z"},
	SNKNeoGeoCD: {".chd", ".cue", ".bin", ".iso"},
	SNKNGP:      {".ngp", ".zip", ".7z"},
	SNKNGPC:     {".ngc", ".ngpc", ".zip", ".7z"},

	// Sony
	SonyPSX: {".chd", ".cue", ".bin", ".iso", ".pbp", ".img", ".mdf", ".ecm"},
	SonyPS2: {".chd", ".iso", ".bin", ".cue", ".gz"},
	SonyPSP: {".iso", ".cso", ".pbp"},

	// Misc
	DOSBox:  {".zip", ".7z", ".exe", ".com", ".bat", ".conf"},
	ScummVM: {".zip", ".7z"},
	MSX:     {".rom", ".mx1", ".mx2", ".dsk", ".cas", ".zip", ".7z"},
	MSX2:    {".rom", ".mx1", ".mx2", ".dsk", ".cas", ".zip", ".7z"},

	// Arcade
	Arcade: {".zip", ".7z"},
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
