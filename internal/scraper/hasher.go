package scraper

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

// HashFile computes CRC32, MD5, and SHA1 of a file in a single pass.
func HashFile(path string) (FileHashes, error) {
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
	if _, err := io.Copy(w, f); err != nil {
		return FileHashes{}, err
	}

	return FileHashes{
		CRC32: fmt.Sprintf("%08X", crc.Sum32()),
		MD5:   fmt.Sprintf("%x", md.Sum(nil)),
		SHA1:  fmt.Sprintf("%x", sh.Sum(nil)),
		Size:  info.Size(),
	}, nil
}
