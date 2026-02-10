package organizer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ECM (Error Code Modeler) decompressor — pure Go, no external dependency.
// Decompresses .ecm files back to their original CD-ROM disc image format
// by reconstructing the EDC/ECC data that was stripped during compression.
//
// Based on the ECM format by Neill Corlett (public domain specification).

const ecmSectorSize = 2352

var (
	edcLUT  [256]uint32 // EDC (CRC-32 variant) lookup table
	eccFLUT [256]byte   // ECC forward: power → value (α^i in GF(2^8))
	eccBLUT [256]byte   // ECC backward: value → power (log_α in GF(2^8))
)

// CD-ROM sync pattern (12 bytes at start of every sector)
var ecmSync = [12]byte{
	0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00,
}

func init() {
	// Build EDC lookup table (CRC-32 with polynomial 0xD8018001)
	for i := 0; i < 256; i++ {
		edc := uint32(i)
		for j := 0; j < 8; j++ {
			if edc&1 != 0 {
				edc = (edc >> 1) ^ 0xD8018001
			} else {
				edc >>= 1
			}
		}
		edcLUT[i] = edc
	}

	// Build ECC GF(2^8) lookup tables
	// Primitive polynomial: x^8 + x^4 + x^3 + x^2 + 1 = 0x11D
	j := 1
	for i := 0; i < 255; i++ {
		eccFLUT[i] = byte(j)
		eccBLUT[j] = byte(i)
		j = (j << 1) ^ ((j >> 7) * 0x11D)
	}
}

// edcCompute calculates EDC over a byte slice.
func edcCompute(data []byte) uint32 {
	var edc uint32
	for _, b := range data {
		edc = (edc >> 8) ^ edcLUT[(edc^uint32(b))&0xFF]
	}
	return edc
}

// edcSet computes EDC over sector[start:start+length] and writes the
// 4-byte result at sector[start+length].
func edcSet(sector []byte, start, length int) {
	edc := edcCompute(sector[start : start+length])
	sector[start+length+0] = byte(edc)
	sector[start+length+1] = byte(edc >> 8)
	sector[start+length+2] = byte(edc >> 16)
	sector[start+length+3] = byte(edc >> 24)
}

// eccComputeBlock computes one block of Reed-Solomon ECC parity.
func eccComputeBlock(src []byte, majorCount, minorCount, majorMult, minorInc int, dest []byte) {
	size := majorCount * minorCount
	for major := 0; major < majorCount; major++ {
		idx := (major >> 1) * majorMult + (major & 1)
		var eccA, eccB byte
		for minor := 0; minor < minorCount; minor++ {
			temp := src[idx]
			idx += minorInc
			if idx >= size {
				idx -= size
			}
			eccA ^= temp
			eccB ^= temp
			eccA = eccFLUT[(int(eccBLUT[eccA])+1)%255]
		}
		eccA = eccFLUT[(255-int(eccBLUT[eccA]))%255]
		dest[major] = eccA
		dest[major+majorCount] = eccA ^ eccB
	}
}

// eccGenerate computes P and Q ECC parity for a CD-ROM sector.
// If zeroAddress is true, the address field (bytes 12-15) is temporarily
// zeroed for computation (required for Mode 2 sectors).
func eccGenerate(sector []byte, zeroAddress bool) {
	var addr [4]byte
	if zeroAddress {
		copy(addr[:], sector[12:16])
		sector[12], sector[13], sector[14], sector[15] = 0, 0, 0, 0
	}

	// P parity: 86 columns × 24 rows = 2064 bytes → 172 bytes parity
	eccComputeBlock(sector[12:], 86, 24, 2, 86, sector[2076:])
	// Q parity: 52 columns × 43 rows = 2236 bytes → 104 bytes parity
	eccComputeBlock(sector[12:], 52, 43, 86, 88, sector[2248:])

	if zeroAddress {
		copy(sector[12:16], addr[:])
	}
}

// reconstructSector fills in the stripped fields (sync, mode, EDC, ECC)
// for a partially-read sector based on its type.
func reconstructSector(sector []byte, sectorType int) {
	// Set sync pattern
	copy(sector[0:12], ecmSync[:])

	switch sectorType {
	case 1: // Mode 1
		sector[0x0F] = 0x01
		// EDC over bytes [0x000..0x80F]
		edcSet(sector, 0, 0x810)
		// Zero reserved area [0x814..0x81B]
		for i := 0x814; i < 0x81C; i++ {
			sector[i] = 0
		}
		// ECC (address included in computation)
		eccGenerate(sector, false)

	case 2: // Mode 2 Form 1
		sector[0x0F] = 0x02
		// Duplicate subheader: [0x14..0x17] → [0x10..0x13]
		copy(sector[0x10:0x14], sector[0x14:0x18])
		// EDC over bytes [0x10..0x817]
		edcSet(sector, 0x10, 0x808)
		// ECC (address zeroed for computation)
		eccGenerate(sector, true)

	case 3: // Mode 2 Form 2
		sector[0x0F] = 0x02
		// Duplicate subheader: [0x14..0x17] → [0x10..0x13]
		copy(sector[0x10:0x14], sector[0x14:0x18])
		// EDC over bytes [0x10..0x92B]
		edcSet(sector, 0x10, 0x91C)
		// No ECC for Form 2
	}
}

// readEcmChunkHeader reads the variable-length encoded chunk header.
// Returns sector type (0-3), count, and whether the end marker was reached.
func readEcmChunkHeader(r io.Reader) (sectorType int, count int, eof bool, err error) {
	var buf [1]byte

	if _, err = io.ReadFull(r, buf[:]); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return 0, 0, true, nil
		}
		return 0, 0, false, err
	}

	c := buf[0]
	sectorType = int(c & 3)
	num := uint32(c>>2) & 0x1F // 5 bits from first byte

	if c&0x80 != 0 {
		if _, err = io.ReadFull(r, buf[:]); err != nil {
			return 0, 0, false, fmt.Errorf("ecm header byte 2: %w", err)
		}
		c = buf[0]
		num |= uint32(c&0x7F) << 5

		if c&0x80 != 0 {
			if _, err = io.ReadFull(r, buf[:]); err != nil {
				return 0, 0, false, fmt.Errorf("ecm header byte 3: %w", err)
			}
			c = buf[0]
			num |= uint32(c&0x7F) << 12

			if c&0x80 != 0 {
				if _, err = io.ReadFull(r, buf[:]); err != nil {
					return 0, 0, false, fmt.Errorf("ecm header byte 4: %w", err)
				}
				c = buf[0]
				num |= uint32(c&0x7F) << 19

				if c&0x80 != 0 {
					if _, err = io.ReadFull(r, buf[:]); err != nil {
						return 0, 0, false, fmt.Errorf("ecm header byte 5: %w", err)
					}
					c = buf[0]
					num |= uint32(c) << 26 // full byte (no mask)
				}
			}
		}
	}

	// End marker: num == 0xFFFFFFFF
	if num == 0xFFFFFFFF {
		return 0, 0, true, nil
	}

	return sectorType, int(num) + 1, false, nil
}

// decompressEcmStream reads an ECM stream (after the magic header) and
// writes the decompressed output.
func decompressEcmStream(in io.Reader, out io.Writer) error {
	sector := make([]byte, ecmSectorSize)

	for {
		sectorType, count, eof, err := readEcmChunkHeader(in)
		if err != nil {
			return err
		}
		if eof {
			return nil
		}

		switch sectorType {
		case 0:
			// Type 0: raw byte copy
			if _, err := io.CopyN(out, in, int64(count)); err != nil {
				return fmt.Errorf("type 0 copy (%d bytes): %w", count, err)
			}

		case 1:
			// Mode 1: 3 bytes address + 2048 bytes user data per sector
			for i := 0; i < count; i++ {
				clear(sector)
				if _, err := io.ReadFull(in, sector[0x0C:0x0C+3]); err != nil {
					return fmt.Errorf("mode 1 address: %w", err)
				}
				if _, err := io.ReadFull(in, sector[0x10:0x10+0x800]); err != nil {
					return fmt.Errorf("mode 1 data: %w", err)
				}
				reconstructSector(sector, 1)
				if _, err := out.Write(sector); err != nil {
					return fmt.Errorf("write sector: %w", err)
				}
			}

		case 2:
			// Mode 2 Form 1: 2052 bytes (subheader + user data) per sector
			for i := 0; i < count; i++ {
				clear(sector)
				if _, err := io.ReadFull(in, sector[0x14:0x14+0x804]); err != nil {
					return fmt.Errorf("mode 2/1 data: %w", err)
				}
				reconstructSector(sector, 2)
				if _, err := out.Write(sector); err != nil {
					return fmt.Errorf("write sector: %w", err)
				}
			}

		case 3:
			// Mode 2 Form 2: 2328 bytes (subheader + user data) per sector
			for i := 0; i < count; i++ {
				clear(sector)
				if _, err := io.ReadFull(in, sector[0x14:0x14+0x918]); err != nil {
					return fmt.Errorf("mode 2/2 data: %w", err)
				}
				reconstructSector(sector, 3)
				if _, err := out.Write(sector); err != nil {
					return fmt.Errorf("write sector: %w", err)
				}
			}
		}
	}
}

// fixCueEcmReferences finds .cue files in the same directory as the
// decompressed file and replaces references to the .ecm filename with the
// decompressed filename. Some distributions have .cue files that reference
// .bin.ecm instead of .bin.
func fixCueEcmReferences(ecmPath, outputPath string) {
	dir := filepath.Dir(ecmPath)
	ecmBase := filepath.Base(ecmPath)
	outputBase := filepath.Base(outputPath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".cue" {
			continue
		}
		cuePath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(cuePath)
		if err != nil {
			continue
		}
		content := string(data)
		if strings.Contains(content, ecmBase) {
			newContent := strings.ReplaceAll(content, ecmBase, outputBase)
			os.WriteFile(cuePath, []byte(newContent), 0644)
		}
	}
}

// FixCueFileReferences fixes FILE references in .cue files to match actual
// files on disk. Handles two common issues from Windows-era distributions:
// 1. Case mismatches (e.g. .BIN vs .bin on case-sensitive Linux filesystems)
// 2. References to .ecm files that have been decompressed
func FixCueFileReferences(dirs []string) int {
	fixed := 0
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			subPath := filepath.Join(dir, entry.Name())
			fixCueFilesInDir(subPath, &fixed)
		}
	}
	return fixed
}

// fixCueFilesInDir recursively fixes .cue FILE references in a directory tree.
func fixCueFilesInDir(dir string, fixed *int) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) == ".cue" {
			if fixSingleCueFile(path) {
				*fixed++
			}
		}
		return nil
	})
}

// fixSingleCueFile reads a .cue file and fixes FILE references that don't
// match actual files on disk (case mismatches). Returns true if the file
// was modified.
func fixSingleCueFile(cuePath string) bool {
	data, err := os.ReadFile(cuePath)
	if err != nil {
		return false
	}

	dir := filepath.Dir(cuePath)
	content := string(data)
	newContent := content
	modified := false

	// Build case-insensitive map of files in the directory
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	lowerToActual := make(map[string]string)
	for _, e := range dirEntries {
		if !e.IsDir() {
			lowerToActual[strings.ToLower(e.Name())] = e.Name()
		}
	}

	// Find FILE "filename" references and fix case mismatches
	lines := strings.Split(newContent, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToUpper(trimmed), "FILE ") {
			continue
		}

		// Extract quoted filename
		start := strings.Index(trimmed, "\"")
		if start < 0 {
			continue
		}
		end := strings.Index(trimmed[start+1:], "\"")
		if end < 0 {
			continue
		}
		refName := trimmed[start+1 : start+1+end]

		// Check if the referenced file exists
		refPath := filepath.Join(dir, refName)
		if _, err := os.Stat(refPath); err == nil {
			continue // file exists, no fix needed
		}

		// Try case-insensitive match
		actual, ok := lowerToActual[strings.ToLower(refName)]
		if !ok {
			continue // no match found
		}

		// Replace the reference with the actual filename
		lines[i] = strings.Replace(line, refName, actual, 1)
		modified = true
	}

	if modified {
		newContent = strings.Join(lines, "\n")
		os.WriteFile(cuePath, []byte(newContent), 0644)
	}

	return modified
}

// decompressEcmNative decompresses an ECM file using pure Go.
// Returns the output file path.
func decompressEcmNative(ecmPath string) (string, error) {
	in, err := os.Open(ecmPath)
	if err != nil {
		return "", fmt.Errorf("open ecm: %w", err)
	}
	defer in.Close()

	// Verify magic header "ECM\x00"
	magic := make([]byte, 4)
	if _, err := io.ReadFull(in, magic); err != nil {
		return "", fmt.Errorf("read ecm header: %w", err)
	}
	if string(magic) != "ECM\x00" {
		return "", fmt.Errorf("not a valid ECM file (bad magic)")
	}

	// Output path: strip .ecm extension (e.g. game.bin.ecm → game.bin)
	outputPath := strings.TrimSuffix(ecmPath, filepath.Ext(ecmPath))

	out, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("create output: %w", err)
	}

	if err := decompressEcmStream(in, out); err != nil {
		out.Close()
		os.Remove(outputPath) // clean up partial output
		return "", err
	}

	if err := out.Close(); err != nil {
		os.Remove(outputPath)
		return "", fmt.Errorf("close output: %w", err)
	}

	return outputPath, nil
}
