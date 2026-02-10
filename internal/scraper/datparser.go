package scraper

import (
	"encoding/xml"
	"io"
	"os"
	"strings"
)

// Logiqx XML DAT format structures
type datFile struct {
	XMLName xml.Name  `xml:"datafile"`
	Header  datHeader `xml:"header"`
	Games   []datGame `xml:"game"`
}

type datHeader struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
}

type datGame struct {
	Name        string   `xml:"name,attr"`
	Description string   `xml:"description"`
	ROMs        []datROM `xml:"rom"`
}

type datROM struct {
	Name   string `xml:"name,attr"`
	Size   string `xml:"size,attr"`
	CRC    string `xml:"crc,attr"`
	MD5    string `xml:"md5,attr"`
	SHA1   string `xml:"sha1,attr"`
	Status string `xml:"status,attr"`
}

// DATIndex provides hash-based lookup into a parsed DAT file.
type DATIndex struct {
	Name   string
	ByCRC  map[string]*DATEntry
	ByMD5  map[string]*DATEntry
	BySHA1 map[string]*DATEntry
}

// DATEntry is a single ROM entry from a DAT file.
type DATEntry struct {
	GameName    string
	ROMName     string
	CRC         string
	MD5         string
	SHA1        string
}

// ParseDAT parses a Logiqx XML DAT file and builds an index.
func ParseDAT(path string) (*DATIndex, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseDATReader(f)
}

// ParseDATReader parses a DAT file from a reader.
func ParseDATReader(r io.Reader) (*DATIndex, error) {
	var dat datFile
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&dat); err != nil {
		return nil, err
	}

	idx := &DATIndex{
		Name:   dat.Header.Name,
		ByCRC:  make(map[string]*DATEntry),
		ByMD5:  make(map[string]*DATEntry),
		BySHA1: make(map[string]*DATEntry),
	}

	for _, game := range dat.Games {
		for _, rom := range game.ROMs {
			entry := &DATEntry{
				GameName: game.Name,
				ROMName:  rom.Name,
				CRC:      strings.ToUpper(rom.CRC),
				MD5:      strings.ToLower(rom.MD5),
				SHA1:     strings.ToLower(rom.SHA1),
			}

			if entry.CRC != "" {
				idx.ByCRC[entry.CRC] = entry
			}
			if entry.MD5 != "" {
				idx.ByMD5[entry.MD5] = entry
			}
			if entry.SHA1 != "" {
				idx.BySHA1[entry.SHA1] = entry
			}
		}
	}

	return idx, nil
}

// Lookup tries to find a match by SHA1, then MD5, then CRC32.
func (idx *DATIndex) Lookup(hashes FileHashes) (*DATEntry, bool) {
	if hashes.SHA1 != "" {
		if entry, ok := idx.BySHA1[strings.ToLower(hashes.SHA1)]; ok {
			return entry, true
		}
	}
	if hashes.MD5 != "" {
		if entry, ok := idx.ByMD5[strings.ToLower(hashes.MD5)]; ok {
			return entry, true
		}
	}
	if hashes.CRC32 != "" {
		if entry, ok := idx.ByCRC[strings.ToUpper(hashes.CRC32)]; ok {
			return entry, true
		}
	}
	return nil, false
}
