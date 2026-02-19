package transfer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// rsyncProgressRegex matches rsync --info=progress2 output like:
//
//	1,234,567  45%   10.00MB/s    0:01:23
var rsyncProgressRegex = regexp.MustCompile(`(\d+)%\s+([\d.]+\S+/s)`)

// RsyncBackend implements BulkTransferBackend using rsync over SSH.
// Requires sshpass for password-based authentication.
// Multiple folders are transferred in parallel — one rsync process per folder,
// limited by Concurrency (default: min(folders, NumCPU, 4)).
type RsyncBackend struct {
	Host        string
	User        string
	Password    string
	Port        int
	Concurrency int // 0 = auto
}

func NewRsyncBackend(host string, port int, user, password string) *RsyncBackend {
	return &RsyncBackend{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	}
}

func (r *RsyncBackend) Close() error { return nil }

// effectiveConcurrency returns the capped concurrency to use for n folders.
func (r *RsyncBackend) effectiveConcurrency(n int) int {
	c := r.Concurrency
	if c <= 0 {
		c = runtime.NumCPU()
		if c > 4 {
			c = 4
		}
	}
	if c > n {
		c = n
	}
	return c
}

func (r *RsyncBackend) TransferFolders(ctx context.Context, folders []FolderMapping, progressCh chan<- TransferProgress) error {
	defer func() {
		if progressCh != nil {
			close(progressCh)
		}
	}()

	if len(folders) == 0 {
		return nil
	}

	concurrency := r.effectiveConcurrency(len(folders))
	totalFolders := len(folders)

	var completed atomic.Int64
	var firstErr atomic.Value
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, folder := range folders {
		select {
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()
		default:
		}

		sem <- struct{}{} // acquire slot
		wg.Add(1)

		go func(f FolderMapping) {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			err := r.transferFolder(ctx, f, &completed, totalFolders, progressCh)
			if err != nil {
				firstErr.CompareAndSwap(nil, err)
			}
		}(folder)
	}

	wg.Wait()

	if v := firstErr.Load(); v != nil {
		return v.(error)
	}
	return nil
}

func (r *RsyncBackend) transferFolder(ctx context.Context, folder FolderMapping, completed *atomic.Int64, totalFolders int, progressCh chan<- TransferProgress) error {
	// Trailing slash on source makes rsync copy the contents, not the directory itself.
	localDir := strings.TrimRight(folder.LocalDir, "/") + "/"
	remoteDest := fmt.Sprintf("%s@%s:%s/", r.User, r.Host, folder.RemoteDir)

	// sshpass reads the password from the SSHPASS environment variable (-e flag).
	sshCmd := fmt.Sprintf("sshpass -e ssh -p %d -o StrictHostKeyChecking=no", r.Port)

	args := []string{
		"-a",
		"--info=progress2",
		"--no-perms",
		"--no-owner",
		"--no-group",
		"-e", sshCmd,
		localDir,
		remoteDest,
	}

	cmd := exec.CommandContext(ctx, "rsync", args...)
	cmd.Env = append(cmd.Environ(), "SSHPASS="+r.Password)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start rsync: %w", err)
	}

	// Parse progress from stdout; drain stderr to prevent blocking.
	var progressWg sync.WaitGroup
	progressWg.Add(1)
	go func() {
		defer progressWg.Done()
		r.parseProgress(stdout, completed, totalFolders, folder, progressCh)
	}()
	go io.Copy(io.Discard, stderr)

	cmdErr := cmd.Wait()
	progressWg.Wait() // ensure all progress is flushed before returning

	if cmdErr == nil {
		done := completed.Add(1)
		sendProgressNonBlocking(progressCh, TransferProgress{
			FileIndex:  int(done) - 1,
			TotalFiles: totalFolders,
			Filename:   filepath.Base(folder.LocalDir),
			BytesSent:  100,
			FileSize:   100,
			TotalSent:  done * 100,
			TotalSize:  int64(totalFolders) * 100,
			Done:       true,
		})
	}

	return cmdErr
}

func (r *RsyncBackend) parseProgress(rd io.Reader, completed *atomic.Int64, totalFolders int, folder FolderMapping, progressCh chan<- TransferProgress) {
	scanner := bufio.NewScanner(rd)
	scanner.Split(scanCRorLF)

	for scanner.Scan() {
		line := scanner.Text()
		matches := rsyncProgressRegex.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}

		pct, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}
		speed := matches[2]

		// Load completed count at each update so parallel progress is reflected.
		done := completed.Load()
		sendProgressNonBlocking(progressCh, TransferProgress{
			FileIndex:  int(done),
			TotalFiles: totalFolders,
			Filename:   fmt.Sprintf("%s (%s)", filepath.Base(folder.LocalDir), speed),
			BytesSent:  int64(pct),
			FileSize:   100,
			TotalSent:  done*100 + int64(pct),
			TotalSize:  int64(totalFolders) * 100,
		})
	}
}

// scanCRorLF splits on \r or \n to handle rsync's \r-based progress output.
func scanCRorLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i := 0; i < len(data); i++ {
		if data[i] == '\r' || data[i] == '\n' {
			return i + 1, data[:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
