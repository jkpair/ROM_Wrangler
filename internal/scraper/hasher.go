package scraper

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

// hashBufSize is the buffer size for hashing (256KB).
const hashBufSize = 256 * 1024

// HashFile computes CRC32, MD5, and SHA1 of a file in a single pass.
// It checks ctx for cancellation between read chunks.
func HashFile(ctx context.Context, path string) (FileHashes, error) {
	f, err := os.Open(path)
	if err != nil {
		return FileHashes{}, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return FileHashes{}, err
	}

	crc := crc32.NewIEEE()
	md := md5.New()
	sh := sha1.New()

	w := io.MultiWriter(crc, md, sh)
	buf := make([]byte, hashBufSize)

	for {
		select {
		case <-ctx.Done():
			return FileHashes{}, ctx.Err()
		default:
		}

		n, err := f.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return FileHashes{}, werr
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return FileHashes{}, err
		}
	}

	return FileHashes{
		CRC32: fmt.Sprintf("%08X", crc.Sum32()),
		MD5:   fmt.Sprintf("%x", md.Sum(nil)),
		SHA1:  fmt.Sprintf("%x", sh.Sum(nil)),
		Size:  info.Size(),
	}, nil
}
