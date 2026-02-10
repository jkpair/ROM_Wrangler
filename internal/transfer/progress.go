package transfer

import "io"

// progressReportInterval controls how often progress callbacks fire.
// Reporting every 256KB avoids channel thrashing while still updating
// the UI smoothly.
const progressReportInterval = 256 * 1024

// ProgressWriter wraps an io.Writer and reports bytes written via a callback.
type ProgressWriter struct {
	Writer       io.Writer
	OnWrite      func(n int64)
	written      int64
	lastReported int64
}

func NewProgressWriter(w io.Writer, onWrite func(n int64)) *ProgressWriter {
	return &ProgressWriter{Writer: w, OnWrite: onWrite}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	pw.written += int64(n)
	if pw.OnWrite != nil && pw.written-pw.lastReported >= progressReportInterval {
		pw.OnWrite(pw.written)
		pw.lastReported = pw.written
	}
	return n, err
}

// Flush sends a final progress update with the current total.
func (pw *ProgressWriter) Flush() {
	if pw.OnWrite != nil && pw.written != pw.lastReported {
		pw.OnWrite(pw.written)
		pw.lastReported = pw.written
	}
}

// Written returns total bytes written so far.
func (pw *ProgressWriter) Written() int64 {
	return pw.written
}
