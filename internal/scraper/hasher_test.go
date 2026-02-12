package scraper

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHashFile(t *testing.T) {
	// Create a temp file with known content
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")
	content := []byte("Hello, ROM Wrangler!")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	hashes, err := HashFile(context.Background(), path)
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}

	if hashes.Size != int64(len(content)) {
		t.Errorf("size = %d, want %d", hashes.Size, len(content))
	}

	// All hash fields should be non-empty
	if hashes.CRC32 == "" {
		t.Error("CRC32 is empty")
	}
	if hashes.MD5 == "" {
		t.Error("MD5 is empty")
	}
	if hashes.SHA1 == "" {
		t.Error("SHA1 is empty")
	}

	// CRC32 should be uppercase hex
	for _, c := range hashes.CRC32 {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
			t.Errorf("CRC32 contains non-uppercase-hex char: %c", c)
		}
	}
}

func TestHashFile_NotFound(t *testing.T) {
	_, err := HashFile(context.Background(), "/nonexistent/file.bin")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
