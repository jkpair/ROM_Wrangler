package transfer

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// usbBufSize is the buffer size for USB uploads (1MB).
// USB benefits from larger blocks than network transfers.
const usbBufSize = 1024 * 1024

// USBBackend implements TransferBackend for local USB/SD card copy.
type USBBackend struct {
	MountPath  string
	bufferPool sync.Pool
}

func NewUSBBackend(mountPath string) *USBBackend {
	return &USBBackend{
		MountPath: mountPath,
		bufferPool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, usbBufSize)
				return &b
			},
		},
	}
}

func (u *USBBackend) Connect(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

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

func (u *USBBackend) Upload(ctx context.Context, localPath, remotePath string, progressFn func(written int64)) error {
	// Check context before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

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

	srcInfo, err := src.Stat()
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer dst.Close()

	// Pre-allocate space on destination for large files
	if srcInfo.Size() > 0 {
		preallocate(dst, srcInfo.Size()) // best-effort, ignore error
	}

	var writer io.Writer = dst
	var pw *ProgressWriter
	if progressFn != nil {
		pw = NewProgressWriter(dst, progressFn)
		writer = pw
	}

	// Get buffer from pool (1MB for USB throughput)
	bufp := u.bufferPool.Get().(*[]byte)
	defer u.bufferPool.Put(bufp)

	if _, err := io.CopyBuffer(writer, src, *bufp); err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	if pw != nil {
		pw.Flush()
	}

	return dst.Sync()
}
