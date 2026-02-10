package transfer

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPBackend implements TransferBackend over SFTP.
type SFTPBackend struct {
	Host     string
	Port     int
	User     string
	Password string

	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func NewSFTPBackend(host string, port int, user, password string) *SFTPBackend {
	return &SFTPBackend{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	}
}

func (s *SFTPBackend) Connect() error {
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	s.sshClient = sshClient

	// Use larger max packet size (256KB) and more concurrent requests
	// to maximize throughput over the network.
	sftpClient, err := sftp.NewClient(sshClient,
		sftp.MaxPacketUnchecked(256*1024),
		sftp.MaxConcurrentRequestsPerFile(64),
	)
	if err != nil {
		sshClient.Close()
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

// uploadBufSize is the buffer size for file uploads (256KB).
// Larger buffers reduce syscall overhead and improve throughput.
const uploadBufSize = 256 * 1024

func (s *SFTPBackend) Upload(localPath, remotePath string, progressFn func(written int64)) error {
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

	// Use a large buffer for copying to reduce syscalls
	buf := make([]byte, uploadBufSize)
	if _, err := io.CopyBuffer(writer, localFile, buf); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	// Send final progress update
	if pw != nil {
		pw.Flush()
	}

	return nil
}
