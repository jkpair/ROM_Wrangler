package organizer

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEdcCompute(t *testing.T) {
	// EDC of all zeros should be zero
	zeros := make([]byte, 2064)
	if edc := edcCompute(zeros); edc != 0 {
		t.Errorf("EDC of zeros = 0x%08X, want 0", edc)
	}

	// EDC of non-zero data should be non-zero
	data := []byte{0x01, 0x02, 0x03, 0x04}
	if edc := edcCompute(data); edc == 0 {
		t.Error("EDC of non-zero data should not be zero")
	}

	// EDC should be deterministic
	edc1 := edcCompute(data)
	edc2 := edcCompute(data)
	if edc1 != edc2 {
		t.Errorf("EDC not deterministic: 0x%08X != 0x%08X", edc1, edc2)
	}
}

func TestReadEcmChunkHeader_EndMarker(t *testing.T) {
	// End marker is 5 bytes of 0xFF → num = 0xFFFFFFFF
	data := bytes.NewReader([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	_, _, eof, err := readEcmChunkHeader(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eof {
		t.Error("expected EOF from end marker")
	}
}

func TestReadEcmChunkHeader_SingleByte(t *testing.T) {
	// Single byte, no continuation: 0x04 = type 0, num = 1, count = 2
	// bits: 0000_0100 → type = 0b00 = 0, num = 0b00001 = 1, no continuation
	data := bytes.NewReader([]byte{0x04})
	sectorType, count, eof, err := readEcmChunkHeader(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eof {
		t.Fatal("unexpected EOF")
	}
	if sectorType != 0 {
		t.Errorf("type = %d, want 0", sectorType)
	}
	if count != 2 { // num=1, count=num+1=2
		t.Errorf("count = %d, want 2", count)
	}
}

func TestReadEcmChunkHeader_Type1(t *testing.T) {
	// Type 1, count 1: type=1, num=0 → byte = (0 << 2) | 1 = 0x01
	data := bytes.NewReader([]byte{0x01})
	sectorType, count, eof, err := readEcmChunkHeader(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eof {
		t.Fatal("unexpected EOF")
	}
	if sectorType != 1 {
		t.Errorf("type = %d, want 1", sectorType)
	}
	if count != 1 { // num=0, count=0+1=1
		t.Errorf("count = %d, want 1", count)
	}
}

func TestReadEcmChunkHeader_EOF(t *testing.T) {
	data := bytes.NewReader([]byte{})
	_, _, eof, err := readEcmChunkHeader(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eof {
		t.Error("expected EOF from empty input")
	}
}

func TestReconstructSector_Mode1(t *testing.T) {
	sector := make([]byte, ecmSectorSize)

	// Set address bytes (MSF)
	sector[0x0C] = 0x00
	sector[0x0D] = 0x02
	sector[0x0E] = 0x00

	// Set some user data
	for i := 0x10; i < 0x10+0x800; i++ {
		sector[i] = byte(i & 0xFF)
	}

	reconstructSector(sector, 1)

	// Verify sync pattern
	if !bytes.Equal(sector[0:12], ecmSync[:]) {
		t.Error("sync pattern not set correctly")
	}

	// Verify mode byte
	if sector[0x0F] != 0x01 {
		t.Errorf("mode = 0x%02X, want 0x01", sector[0x0F])
	}

	// Verify EDC is non-zero (computed over non-zero data)
	edc := binary.LittleEndian.Uint32(sector[0x810:0x814])
	if edc == 0 {
		t.Error("EDC should be non-zero for non-zero data")
	}

	// Verify reserved area is zeroed
	for i := 0x814; i < 0x81C; i++ {
		if sector[i] != 0 {
			t.Errorf("reserved byte at 0x%X = 0x%02X, want 0x00", i, sector[i])
			break
		}
	}

	// Verify ECC P is non-zero
	allZeroP := true
	for i := 2076; i < 2076+172; i++ {
		if sector[i] != 0 {
			allZeroP = false
			break
		}
	}
	if allZeroP {
		t.Error("ECC P should be non-zero for non-zero data")
	}
}

func TestDecompressEcmStream_Type0(t *testing.T) {
	// Build a minimal ECM stream with a type 0 (raw copy) chunk
	var ecmData bytes.Buffer

	// Chunk header: type=0, count=5 (num=4)
	// byte = (4 << 2) | 0 = 0x10, no continuation (bit 7 clear)
	ecmData.WriteByte(0x10)
	// Raw data: 5 bytes
	ecmData.Write([]byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE})

	// End marker
	ecmData.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	var output bytes.Buffer
	if err := decompressEcmStream(&ecmData, &output); err != nil {
		t.Fatalf("decompressEcmStream: %v", err)
	}

	expected := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE}
	if !bytes.Equal(output.Bytes(), expected) {
		t.Errorf("output = %X, want %X", output.Bytes(), expected)
	}
}

func TestDecompressEcmStream_Mode1(t *testing.T) {
	// Build a minimal ECM stream with 1 Mode 1 sector
	var ecmData bytes.Buffer

	// Chunk header: type=1, count=1 (num=0)
	// byte = (0 << 2) | 1 = 0x01
	ecmData.WriteByte(0x01)

	// Mode 1 sector data: 3 bytes address + 2048 bytes user data
	ecmData.Write([]byte{0x00, 0x02, 0x00}) // MSF address
	userData := make([]byte, 2048)
	for i := range userData {
		userData[i] = byte(i & 0xFF)
	}
	ecmData.Write(userData)

	// End marker
	ecmData.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	var output bytes.Buffer
	if err := decompressEcmStream(&ecmData, &output); err != nil {
		t.Fatalf("decompressEcmStream: %v", err)
	}

	if output.Len() != ecmSectorSize {
		t.Fatalf("output size = %d, want %d", output.Len(), ecmSectorSize)
	}

	sector := output.Bytes()

	// Verify sync
	if !bytes.Equal(sector[0:12], ecmSync[:]) {
		t.Error("sync pattern incorrect")
	}
	// Verify mode
	if sector[0x0F] != 0x01 {
		t.Errorf("mode = 0x%02X, want 0x01", sector[0x0F])
	}
	// Verify address
	if sector[0x0C] != 0x00 || sector[0x0D] != 0x02 || sector[0x0E] != 0x00 {
		t.Error("address incorrect")
	}
	// Verify user data preserved
	if !bytes.Equal(sector[0x10:0x10+2048], userData) {
		t.Error("user data corrupted")
	}
}

func TestDecompressEcmNative_File(t *testing.T) {
	dir := t.TempDir()

	// Create a minimal ECM file
	ecmPath := filepath.Join(dir, "test.bin.ecm")
	var ecmFile bytes.Buffer

	// Magic header
	ecmFile.WriteString("ECM\x00")

	// Type 0 chunk: 10 bytes of raw data
	// num = 9 (count = 10), type = 0
	// byte = (9 << 2) | 0 = 0x24
	ecmFile.WriteByte(0x24)
	ecmFile.Write(bytes.Repeat([]byte{0x42}, 10))

	// End marker
	ecmFile.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	os.WriteFile(ecmPath, ecmFile.Bytes(), 0644)

	outputPath, err := decompressEcmNative(ecmPath)
	if err != nil {
		t.Fatalf("decompressEcmNative: %v", err)
	}

	expectedPath := filepath.Join(dir, "test.bin")
	if outputPath != expectedPath {
		t.Errorf("output path = %s, want %s", outputPath, expectedPath)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.Equal(data, bytes.Repeat([]byte{0x42}, 10)) {
		t.Errorf("output data = %X, want 10 bytes of 0x42", data)
	}
}

func TestDecompressEcmNative_BadMagic(t *testing.T) {
	dir := t.TempDir()
	ecmPath := filepath.Join(dir, "bad.bin.ecm")
	os.WriteFile(ecmPath, []byte("NOT_ECM_FILE"), 0644)

	_, err := decompressEcmNative(ecmPath)
	if err == nil {
		t.Error("expected error for bad magic")
	}
}

func TestDecompressEcm_Integration(t *testing.T) {
	// Test through the extract.go decompressEcm function
	dir := t.TempDir()
	ecmPath := filepath.Join(dir, "game.bin.ecm")

	var ecmFile bytes.Buffer
	ecmFile.WriteString("ECM\x00")
	ecmFile.WriteByte(0x24) // type 0, count 10
	ecmFile.Write(bytes.Repeat([]byte{0xAB}, 10))
	ecmFile.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	os.WriteFile(ecmPath, ecmFile.Bytes(), 0644)

	count, err := decompressEcm(ecmPath, "")
	if err != nil {
		t.Fatalf("decompressEcm: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	// Verify output file exists
	outputPath := filepath.Join(dir, "game.bin")
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("expected output file at %s", outputPath)
	}

	// Verify source .ecm was deleted (intermediate file)
	if _, err := os.Stat(ecmPath); !os.IsNotExist(err) {
		t.Error("expected .ecm file to be deleted after decompression")
	}
}

func TestDecompressEcm_FixesCueReferences(t *testing.T) {
	dir := t.TempDir()

	// Create a .cue that references .bin.ecm
	cuePath := filepath.Join(dir, "game.cue")
	os.WriteFile(cuePath, []byte("FILE \"game.bin.ecm\" BINARY\n  TRACK 01 MODE2/2352\n    INDEX 01 00:00:00\n"), 0644)

	// Create ECM file
	ecmPath := filepath.Join(dir, "game.bin.ecm")
	var ecmFile bytes.Buffer
	ecmFile.WriteString("ECM\x00")
	ecmFile.WriteByte(0x24) // type 0, count 10
	ecmFile.Write(bytes.Repeat([]byte{0xAB}, 10))
	ecmFile.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	os.WriteFile(ecmPath, ecmFile.Bytes(), 0644)

	_, err := decompressEcm(ecmPath, "")
	if err != nil {
		t.Fatalf("decompressEcm: %v", err)
	}

	// Verify .cue now references .bin instead of .bin.ecm
	data, err := os.ReadFile(cuePath)
	if err != nil {
		t.Fatalf("read cue: %v", err)
	}
	content := string(data)
	if strings.Contains(content, ".bin.ecm") {
		t.Error("cue still references .bin.ecm")
	}
	if !strings.Contains(content, "\"game.bin\"") {
		t.Errorf("cue should reference game.bin, got: %s", content)
	}
}

func TestFixSingleCueFile_CaseMismatch(t *testing.T) {
	dir := t.TempDir()

	// Create a .cue that references .BIN (uppercase)
	cuePath := filepath.Join(dir, "game.cue")
	os.WriteFile(cuePath, []byte(" FILE \"game.BIN\" BINARY\n  TRACK 01 MODE2/2352\n    INDEX 01 00:00:00\n"), 0644)

	// Create the actual file with lowercase extension
	os.WriteFile(filepath.Join(dir, "game.bin"), []byte("bindata"), 0644)

	modified := fixSingleCueFile(cuePath)
	if !modified {
		t.Error("expected cue to be modified")
	}

	data, _ := os.ReadFile(cuePath)
	content := string(data)
	if strings.Contains(content, "game.BIN") {
		t.Error("cue still has uppercase .BIN")
	}
	if !strings.Contains(content, "game.bin") {
		t.Errorf("cue should reference game.bin, got: %s", content)
	}
}

func TestFixSingleCueFile_NoChangeNeeded(t *testing.T) {
	dir := t.TempDir()

	// Create a .cue with correct case
	cuePath := filepath.Join(dir, "game.cue")
	original := "FILE \"game.bin\" BINARY\n  TRACK 01 MODE2/2352\n"
	os.WriteFile(cuePath, []byte(original), 0644)
	os.WriteFile(filepath.Join(dir, "game.bin"), []byte("bindata"), 0644)

	modified := fixSingleCueFile(cuePath)
	if modified {
		t.Error("cue should not be modified when case matches")
	}
}
