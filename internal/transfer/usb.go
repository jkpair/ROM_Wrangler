package transfer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// USBBackend implements TransferBackend for local USB/SD card copy.
type USBBackend struct {
	MountPath string
}

func NewUSBBackend(mountPath string) *USBBackend {
	return &USBBackend{MountPath: mountPath}
}

func (u *USBBackend) Connect() error {
	info, err := os.Stat(u.MountPath)
	if err != nil {
		return fmt.Errorf("USB path not found: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("USB path is not a directory: %s", u.MountPath)
	}
	return nil
}

func (u *USBBackend) Close() error {
	return nil
}

func (u *USBBackend) MkdirAll(path string) error {
	fullPath := filepath.Join(u.MountPath, path)
	return os.MkdirAll(fullPath, 0755)
}

func (u *USBBackend) FileExists(path string, expectedSize int64) (bool, error) {
	fullPath := filepath.Join(u.MountPath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return false, nil
	}
	return info.Size() == expectedSize, nil
}

func (u *USBBackend) Upload(localPath, remotePath string, progressFn func(written int64)) error {
	destPath := filepath.Join(u.MountPath, remotePath)

	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	src, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer dst.Close()

	var writer io.Writer = dst
	var pw *ProgressWriter
	if progressFn != nil {
		pw = NewProgressWriter(dst, progressFn)
		writer = pw
	}

	if _, err := io.Copy(writer, src); err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	if pw != nil {
		pw.Flush()
	}

	return dst.Sync()
}
