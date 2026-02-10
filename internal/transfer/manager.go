package transfer

import (
	"fmt"
	"os"
	"path/filepath"
)

// TransferBackend is the interface for transfer methods.
type TransferBackend interface {
	Connect() error
	Close() error
	MkdirAll(path string) error
	FileExists(path string, expectedSize int64) (bool, error)
	Upload(localPath, remotePath string, progressFn func(written int64)) error
}

// TransferItem describes a single file to transfer.
type TransferItem struct {
	LocalPath  string
	RemotePath string
	Size       int64
	Skip       bool // set by sync mode
}

// TransferPlan is a list of files to transfer.
type TransferPlan struct {
	Items     []TransferItem
	TotalSize int64
	SkipCount int
}

// TransferProgress reports progress of a transfer operation.
type TransferProgress struct {
	FileIndex   int
	TotalFiles  int
	Filename    string
	BytesSent   int64
	FileSize    int64
	TotalSent   int64
	TotalSize   int64
	Done        bool
	Err         error
}

// BuildTransferPlan builds a list of files to transfer.
// In sync mode, it checks if files already exist on the destination.
func BuildTransferPlan(backend TransferBackend, localDir, remoteBase string, syncMode bool) (*TransferPlan, error) {
	plan := &TransferPlan{}

	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == "_archive" {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(localDir, path)
		remotePath := filepath.ToSlash(filepath.Join(remoteBase, relPath))

		item := TransferItem{
			LocalPath:  path,
			RemotePath: remotePath,
			Size:       info.Size(),
		}

		if syncMode && backend != nil {
			exists, err := backend.FileExists(remotePath, info.Size())
			if err == nil && exists {
				item.Skip = true
				plan.SkipCount++
			}
		}

		if !item.Skip {
			plan.TotalSize += info.Size()
		}
		plan.Items = append(plan.Items, item)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to build transfer plan: %w", err)
	}

	return plan, nil
}

// Execute runs the transfer plan.
func Execute(backend TransferBackend, plan *TransferPlan, progressCh chan<- TransferProgress) error {
	var totalSent int64
	transferIdx := 0

	for i, item := range plan.Items {
		if item.Skip {
			continue
		}

		// Ensure remote directory exists
		remoteDir := filepath.ToSlash(filepath.Dir(item.RemotePath))
		if err := backend.MkdirAll(remoteDir); err != nil {
			sendProgress(progressCh, TransferProgress{
				FileIndex: transferIdx, TotalFiles: len(plan.Items) - plan.SkipCount,
				Filename: filepath.Base(item.LocalPath), Err: err,
			})
			continue
		}

		if progressCh != nil {
			progressCh <- TransferProgress{
				FileIndex:  transferIdx,
				TotalFiles: len(plan.Items) - plan.SkipCount,
				Filename:   filepath.Base(item.LocalPath),
				FileSize:   item.Size,
				TotalSent:  totalSent,
				TotalSize:  plan.TotalSize,
			}
		}

		err := backend.Upload(item.LocalPath, item.RemotePath, func(written int64) {
			sendProgress(progressCh, TransferProgress{
				FileIndex:  transferIdx,
				TotalFiles: len(plan.Items) - plan.SkipCount,
				Filename:   filepath.Base(plan.Items[i].LocalPath),
				BytesSent:  written,
				FileSize:   item.Size,
				TotalSent:  totalSent + written,
				TotalSize:  plan.TotalSize,
			})
		})

		if err != nil {
			sendProgress(progressCh, TransferProgress{
				FileIndex: transferIdx, TotalFiles: len(plan.Items) - plan.SkipCount,
				Filename: filepath.Base(item.LocalPath), Err: err,
			})
		} else {
			totalSent += item.Size
			sendProgress(progressCh, TransferProgress{
				FileIndex:  transferIdx,
				TotalFiles: len(plan.Items) - plan.SkipCount,
				Filename:   filepath.Base(item.LocalPath),
				BytesSent:  item.Size,
				FileSize:   item.Size,
				TotalSent:  totalSent,
				TotalSize:  plan.TotalSize,
				Done:       true,
			})
		}
		transferIdx++
	}

	if progressCh != nil {
		close(progressCh)
	}
	return nil
}

func sendProgress(ch chan<- TransferProgress, p TransferProgress) {
	if ch != nil {
		ch <- p
	}
}
