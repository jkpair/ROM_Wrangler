package transfer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// TransferBackend is the interface for transfer methods.
type TransferBackend interface {
	Connect(ctx context.Context) error
	Close() error
	MkdirAll(path string) error
	FileExists(path string, expectedSize int64) (bool, error)
	Upload(ctx context.Context, localPath, remotePath string, progressFn func(written int64)) error
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
func BuildTransferPlan(ctx context.Context, backend TransferBackend, localDir, remoteBase string, syncMode bool) (*TransferPlan, error) {
	plan := &TransferPlan{}

	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
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

// Execute runs the transfer plan with the given concurrency level.
// Use concurrency=1 for sequential execution.
func Execute(ctx context.Context, backend TransferBackend, plan *TransferPlan, concurrency int, progressCh chan<- TransferProgress) error {
	if concurrency < 1 {
		concurrency = 1
	}

	// Collect non-skipped items
	var items []TransferItem
	for _, item := range plan.Items {
		if !item.Skip {
			items = append(items, item)
		}
	}
	totalFiles := len(items)
	if totalFiles == 0 {
		if progressCh != nil {
			close(progressCh)
		}
		return nil
	}

	var totalSent atomic.Int64
	var firstErr atomic.Value
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for idx, item := range items {
		select {
		case <-ctx.Done():
			// Wait for in-flight uploads before closing channel
			wg.Wait()
			if progressCh != nil {
				close(progressCh)
			}
			return ctx.Err()
		default:
		}

		sem <- struct{}{} // acquire
		wg.Add(1)

		go func(fileIdx int, item TransferItem) {
			defer wg.Done()
			defer func() { <-sem }() // release

			// Ensure remote directory exists
			remoteDir := filepath.ToSlash(filepath.Dir(item.RemotePath))
			if err := backend.MkdirAll(remoteDir); err != nil {
				sendProgressNonBlocking(progressCh, TransferProgress{
					FileIndex: fileIdx, TotalFiles: totalFiles,
					Filename: filepath.Base(item.LocalPath), Err: err,
				})
				return
			}

			sendProgressNonBlocking(progressCh, TransferProgress{
				FileIndex:  fileIdx,
				TotalFiles: totalFiles,
				Filename:   filepath.Base(item.LocalPath),
				FileSize:   item.Size,
				TotalSent:  totalSent.Load(),
				TotalSize:  plan.TotalSize,
			})

			err := backend.Upload(ctx, item.LocalPath, item.RemotePath, func(written int64) {
				sendProgressNonBlocking(progressCh, TransferProgress{
					FileIndex:  fileIdx,
					TotalFiles: totalFiles,
					Filename:   filepath.Base(item.LocalPath),
					BytesSent:  written,
					FileSize:   item.Size,
					TotalSent:  totalSent.Load() + written,
					TotalSize:  plan.TotalSize,
				})
			})

			if err != nil {
				firstErr.CompareAndSwap(nil, err)
				sendProgressNonBlocking(progressCh, TransferProgress{
					FileIndex: fileIdx, TotalFiles: totalFiles,
					Filename: filepath.Base(item.LocalPath), Err: err,
				})
			} else {
				totalSent.Add(item.Size)
				sendProgressNonBlocking(progressCh, TransferProgress{
					FileIndex:  fileIdx,
					TotalFiles: totalFiles,
					Filename:   filepath.Base(item.LocalPath),
					BytesSent:  item.Size,
					FileSize:   item.Size,
					TotalSent:  totalSent.Load(),
					TotalSize:  plan.TotalSize,
					Done:       true,
				})
			}
		}(idx, item)
	}

	wg.Wait()
	if progressCh != nil {
		close(progressCh)
	}

	if v := firstErr.Load(); v != nil {
		return v.(error)
	}
	return nil
}

// MergeTransferPlans combines multiple plans into one.
func MergeTransferPlans(plans ...*TransferPlan) *TransferPlan {
	merged := &TransferPlan{}
	for _, p := range plans {
		if p == nil {
			continue
		}
		merged.Items = append(merged.Items, p.Items...)
		merged.TotalSize += p.TotalSize
		merged.SkipCount += p.SkipCount
	}
	return merged
}

// sendProgressNonBlocking sends progress without blocking if the consumer is slow.
func sendProgressNonBlocking(ch chan<- TransferProgress, p TransferProgress) {
	if ch == nil {
		return
	}
	select {
	case ch <- p:
	default:
	}
}
