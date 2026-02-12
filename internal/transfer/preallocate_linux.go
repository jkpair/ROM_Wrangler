package transfer

import (
	"os"
	"syscall"
)

// preallocate uses fallocate on Linux to pre-allocate disk space,
// reducing fragmentation for large file copies.
func preallocate(f *os.File, size int64) error {
	return syscall.Fallocate(int(f.Fd()), 0, 0, size)
}
