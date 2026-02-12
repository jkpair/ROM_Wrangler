package transfer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestUSBBackend_Connect(t *testing.T) {
	dir := t.TempDir()
	backend := NewUSBBackend(dir)

	if err := backend.Connect(context.Background()); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
}

func TestUSBBackend_Connect_NotFound(t *testing.T) {
	backend := NewUSBBackend("/nonexistent/path")
	if err := backend.Connect(context.Background()); err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestUSBBackend_Upload(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(srcDir, "test.bin")
	content := []byte("test content for USB transfer")
	os.WriteFile(srcFile, content, 0644)

	backend := NewUSBBackend(dstDir)
	backend.Connect(context.Background())

	var lastWritten int64
	err := backend.Upload(context.Background(), srcFile, "/roms/test.bin", func(written int64) {
		lastWritten = written
	})
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if lastWritten != int64(len(content)) {
		t.Errorf("expected %d bytes written, got %d", len(content), lastWritten)
	}

	// Verify file was copied
	dstFile := filepath.Join(dstDir, "roms", "test.bin")
	data, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}
	if string(data) != string(content) {
		t.Error("content mismatch")
	}
}

func TestUSBBackend_FileExists(t *testing.T) {
	dir := t.TempDir()
	backend := NewUSBBackend(dir)

	// File doesn't exist yet
	exists, _ := backend.FileExists("/test.bin", 100)
	if exists {
		t.Error("expected file to not exist")
	}

	// Create file
	testFile := filepath.Join(dir, "test.bin")
	content := []byte("hello")
	os.WriteFile(testFile, content, 0644)

	// File exists with correct size
	exists, _ = backend.FileExists("/test.bin", int64(len(content)))
	if !exists {
		t.Error("expected file to exist with correct size")
	}

	// File exists with wrong size
	exists, _ = backend.FileExists("/test.bin", 999)
	if exists {
		t.Error("expected file to not match with wrong size")
	}
}

func TestBuildTransferPlan(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source files
	os.MkdirAll(filepath.Join(srcDir, "sega_dc"), 0755)
	os.WriteFile(filepath.Join(srcDir, "sega_dc", "game.chd"), []byte("CHD data"), 0644)
	os.WriteFile(filepath.Join(srcDir, "sega_dc", "game2.chd"), []byte("CHD data 2"), 0644)

	backend := NewUSBBackend(dstDir)
	backend.Connect(context.Background())

	plan, err := BuildTransferPlan(context.Background(), backend, srcDir, "/roms", false)
	if err != nil {
		t.Fatalf("BuildTransferPlan failed: %v", err)
	}

	if len(plan.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(plan.Items))
	}
}

func TestBuildTransferPlan_SyncMode(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source file
	os.MkdirAll(filepath.Join(srcDir, "nes"), 0755)
	content := []byte("NES ROM data")
	os.WriteFile(filepath.Join(srcDir, "nes", "game.nes"), content, 0644)

	// Create matching file on destination
	os.MkdirAll(filepath.Join(dstDir, "roms", "nes"), 0755)
	os.WriteFile(filepath.Join(dstDir, "roms", "nes", "game.nes"), content, 0644)

	backend := NewUSBBackend(dstDir)
	backend.Connect(context.Background())

	plan, err := BuildTransferPlan(context.Background(), backend, srcDir, "/roms", true)
	if err != nil {
		t.Fatal(err)
	}

	if plan.SkipCount != 1 {
		t.Errorf("expected 1 skip, got %d", plan.SkipCount)
	}
}

func TestUSBBackend_Upload_Cancelled(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcFile := filepath.Join(srcDir, "test.bin")
	os.WriteFile(srcFile, []byte("data"), 0644)

	backend := NewUSBBackend(dstDir)
	backend.Connect(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := backend.Upload(ctx, srcFile, "/roms/test.bin", nil)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
