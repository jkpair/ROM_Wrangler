package transfer

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// sftpBufSize is the buffer size for SFTP uploads (256KB).
const sftpBufSize = 256 * 1024

// SFTPBackend implements TransferBackend over SFTP.
type SFTPBackend struct {
	Host     string
	Port     int
	User     string
	Password string

	sshClient  *ssh.Client
	sftpClient *sftp.Client
	bufferPool sync.Pool
}

func NewSFTPBackend(host string, port int, user, password string) *SFTPBackend {
	return &SFTPBackend{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		bufferPool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, sftpBufSize)
				return &b
			},
		},
	}
}

func (s *SFTPBackend) Connect(ctx context.Context) error {
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	// Dial with context cancellation support
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	// Perform SSH handshake on the raw connection
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		return fmt.Errorf("SSH handshake failed: %w", err)
	}
	s.sshClient = ssh.NewClient(sshConn, chans, reqs)

	// Enable concurrent writes/reads for better throughput
	sftpClient, err := sftp.NewClient(s.sshClient,
		sftp.MaxPacketUnchecked(256*1024),
		sftp.MaxConcurrentRequestsPerFile(64),
		sftp.UseConcurrentWrites(true),
		sftp.UseConcurrentReads(true),
	)
	if err != nil {
		s.sshClient.Close()
		return fmt.Errorf("SFTP session failed: %w", err)
	}
	s.sftpClient = sftpClient

	return nil
}

func (s *SFTPBackend) Close() error {
	if s.sftpClient != nil {
		s.sftpClient.Close()
	}
	if s.sshClient != nil {
		s.sshClient.Close()
	}
	return nil
}

func (s *SFTPBackend) MkdirAll(remotePath string) error {
	// Split path and create each component
	parts := strings.Split(path.Clean(remotePath), "/")
	current := "/"
	for _, part := range parts {
		if part == "" {
			continue
		}
		current = path.Join(current, part)
		s.sftpClient.Mkdir(current) // ignore errors for existing dirs
	}
	return nil
}

func (s *SFTPBackend) FileExists(remotePath string, expectedSize int64) (bool, error) {
	info, err := s.sftpClient.Stat(remotePath)
	if err != nil {
		return false, nil // file doesn't exist
	}
	return info.Size() == expectedSize, nil
}

func (s *SFTPBackend) Upload(ctx context.Context, localPath, remotePath string, progressFn func(written int64)) error {
	// Check context before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open local file: %w", err)
	}
	defer localFile.Close()

	remoteFile, err := s.sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("create remote file: %w", err)
	}
	defer remoteFile.Close()

	var writer io.Writer = remoteFile
	var pw *ProgressWriter
	if progressFn != nil {
		pw = NewProgressWriter(remoteFile, progressFn)
		writer = pw
	}

	// Get buffer from pool
	bufp := s.bufferPool.Get().(*[]byte)
	defer s.bufferPool.Put(bufp)

	if _, err := io.CopyBuffer(writer, localFile, *bufp); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	// Send final progress update
	if pw != nil {
		pw.Flush()
	}

	return nil
}
