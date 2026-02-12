//go:build !linux

package transfer

import "os"

// preallocate is a no-op on non-Linux platforms.
func preallocate(_ *os.File, _ int64) error {
	return nil
}
