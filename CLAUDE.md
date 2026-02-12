# Claude Code Skills: TUI Development, Go, and File Transfer Expertise

## Overview

This document establishes expertise in three interconnected domains:
1. Terminal User Interface (TUI) design and implementation using Bubble Tea
2. Go language best practices and idiomatic patterns
3. High-performance file transfer operations (USB and SFTP)

---

## Part 1: Bubble Tea TUI Development

### The Elm Architecture

Bubble Tea follows The Elm Architecture (TEA). Every program consists of three core concepts:

```go
// Model - your application state
type model struct {
    choices  []string
    cursor   int
    selected map[int]struct{}
}

// Init - returns initial state and optional command
func (m model) Init() tea.Cmd {
    return nil // or tea.Batch(...) for startup commands
}

// Update - handles messages, returns updated model and commands
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    }
    return m, nil
}

// View - renders the UI as a string
func (m model) View() string {
    return "Hello, TUI!"
}
```

### Project Structure for TUI Applications

```
myapp/
├── cmd/
│   └── myapp/
│       └── main.go           # Entry point, minimal logic
├── internal/
│   ├── tui/
│   │   ├── app.go            # Main model, Init, Update, View
│   │   ├── keys.go           # Key bindings using bubbles/key
│   │   ├── styles.go         # All Lip Gloss styles centralized
│   │   ├── commands.go       # Custom tea.Cmd functions
│   │   └── components/
│   │       ├── header.go     # Reusable header component
│   │       ├── statusbar.go  # Status bar component
│   │       ├── filelist.go   # File browser component
│   │       └── progress.go   # Progress indicator component
│   ├── transfer/
│   │   ├── sftp.go           # SFTP transfer logic
│   │   ├── usb.go            # USB/local transfer logic
│   │   └── queue.go          # Transfer queue management
│   └── config/
│       └── config.go         # Configuration handling
├── pkg/                      # Publicly importable packages (if any)
├── go.mod
├── go.sum
└── README.md
```

### Lip Gloss Styling Best Practices

Centralize all styles in a dedicated file:

```go
// internal/tui/styles.go
package tui

import "github.com/charmbracelet/lipgloss"

// Color palette - define once, use everywhere
var (
    ColorPrimary   = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
    ColorSecondary = lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}
    ColorSuccess   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
    ColorError     = lipgloss.AdaptiveColor{Light: "#FF5F87", Dark: "#FF6B6B"}
    ColorSubtle    = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}
)

// Base styles - compose these into more specific styles
var (
    BaseStyle = lipgloss.NewStyle().
            Padding(0, 1)

    BorderedStyle = BaseStyle.
            Border(lipgloss.RoundedBorder()).
            BorderForeground(ColorPrimary)

    TitleStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(ColorPrimary).
            MarginBottom(1)

    SelectedStyle = lipgloss.NewStyle().
            Foreground(ColorPrimary).
            Bold(true)

    DimStyle = lipgloss.NewStyle().
            Foreground(ColorSubtle)
)

// Component-specific styles
var (
    HeaderStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(ColorPrimary).
            Background(lipgloss.Color("#1a1a1a")).
            Padding(0, 2).
            Width(80)

    StatusBarStyle = lipgloss.NewStyle().
            Foreground(ColorSecondary).
            Background(lipgloss.Color("#2a2a2a")).
            Padding(0, 1)

    ProgressBarStyle = lipgloss.NewStyle().
            Foreground(ColorSuccess)

    ErrorStyle = lipgloss.NewStyle().
            Foreground(ColorError).
            Bold(true)
)

// Helper function for responsive widths
func (s Styles) WithWidth(width int) lipgloss.Style {
    return s.Container.Width(width)
}
```

### Key Bindings with Bubbles/Key

```go
// internal/tui/keys.go
package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
    Up       key.Binding
    Down     key.Binding
    Select   key.Binding
    Transfer key.Binding
    Quit     key.Binding
    Help     key.Binding
}

func newKeyMap() keyMap {
    return keyMap{
        Up: key.NewBinding(
            key.WithKeys("up", "k"),
            key.WithHelp("↑/k", "move up"),
        ),
        Down: key.NewBinding(
            key.WithKeys("down", "j"),
            key.WithHelp("↓/j", "move down"),
        ),
        Select: key.NewBinding(
            key.WithKeys("enter", " "),
            key.WithHelp("enter/space", "select"),
        ),
        Transfer: key.NewBinding(
            key.WithKeys("t"),
            key.WithHelp("t", "start transfer"),
        ),
        Quit: key.NewBinding(
            key.WithKeys("q", "ctrl+c"),
            key.WithHelp("q", "quit"),
        ),
        Help: key.NewBinding(
            key.WithKeys("?"),
            key.WithHelp("?", "toggle help"),
        ),
    }
}

// ShortHelp implements key.Map
func (k keyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Help, k.Quit}
}

// FullHelp implements key.Map
func (k keyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Up, k.Down, k.Select},
        {k.Transfer, k.Help, k.Quit},
    }
}
```

### Essential Bubbles Components

Always consider these official components before building custom ones:

| Component | Use Case |
|-----------|----------|
| `bubbles/list` | Filterable, paginated lists with delegation |
| `bubbles/table` | Tabular data display |
| `bubbles/textinput` | Single-line text input |
| `bubbles/textarea` | Multi-line text editing |
| `bubbles/viewport` | Scrollable content areas |
| `bubbles/progress` | Progress bars |
| `bubbles/spinner` | Loading indicators |
| `bubbles/filepicker` | File system navigation |
| `bubbles/help` | Dynamic help display |
| `bubbles/paginator` | Pagination controls |

### Component Pattern with Bubbles/List

```go
// internal/tui/components/filelist.go
package components

import (
    "github.com/charmbracelet/bubbles/list"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// Item implements list.Item
type FileItem struct {
    name     string
    path     string
    size     int64
    selected bool
}

func (i FileItem) Title() string       { return i.name }
func (i FileItem) Description() string { return formatSize(i.size) }
func (i FileItem) FilterValue() string { return i.name }

// FileList wraps bubbles/list with custom behavior
type FileList struct {
    list     list.Model
    selected map[string]struct{}
}

func NewFileList(items []FileItem, width, height int) FileList {
    delegate := list.NewDefaultDelegate()
    delegate.Styles.SelectedTitle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#7D56F4")).
        Bold(true)

    l := list.New(toListItems(items), delegate, width, height)
    l.Title = "Files"
    l.SetShowStatusBar(true)
    l.SetFilteringEnabled(true)

    return FileList{
        list:     l,
        selected: make(map[string]struct{}),
    }
}

func (f FileList) Update(msg tea.Msg) (FileList, tea.Cmd) {
    var cmd tea.Cmd
    f.list, cmd = f.list.Update(msg)
    return f, cmd
}

func (f FileList) View() string {
    return f.list.View()
}
```

### Commands and Messages Pattern

```go
// internal/tui/commands.go
package tui

import (
    "time"
    tea "github.com/charmbracelet/bubbletea"
)

// Custom message types
type (
    TransferStartedMsg struct {
        TotalFiles int
        TotalBytes int64
    }

    TransferProgressMsg struct {
        CurrentFile string
        BytesDone   int64
        TotalBytes  int64
        FilesDone   int
        TotalFiles  int
    }

    TransferCompleteMsg struct {
        Duration time.Duration
        Errors   []error
    }

    ErrorMsg struct {
        Err error
    }
)

// Commands return tea.Cmd functions
func startTransfer(files []string, dest string) tea.Cmd {
    return func() tea.Msg {
        // This runs in a goroutine
        // Return messages to update the UI
        return TransferStartedMsg{
            TotalFiles: len(files),
        }
    }
}

// Tick command for progress updates
func tickEvery(d time.Duration) tea.Cmd {
    return tea.Every(d, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

type tickMsg time.Time
```

### Responsive Layout Pattern

```go
// Handle window resize in Update
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height

        // Update component dimensions
        headerHeight := 3
        footerHeight := 2
        contentHeight := m.height - headerHeight - footerHeight

        m.fileList.SetSize(m.width, contentHeight)
        m.viewport.Width = m.width
        m.viewport.Height = contentHeight

        return m, nil
    }
    // ...
}

// Compose layout in View
func (m model) View() string {
    header := m.renderHeader()
    content := m.fileList.View()
    footer := m.renderFooter()

    return lipgloss.JoinVertical(
        lipgloss.Left,
        header,
        content,
        footer,
    )
}
```

### Focus Management

```go
type focusState int

const (
    focusFileList focusState = iota
    focusDestination
    focusConfirm
)

type model struct {
    focus       focusState
    fileList    components.FileList
    destination textinput.Model
    // ...
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab":
            m.focus = (m.focus + 1) % 3
            return m, nil
        case "shift+tab":
            m.focus = (m.focus - 1 + 3) % 3
            return m, nil
        }
    }

    // Route updates to focused component
    var cmd tea.Cmd
    switch m.focus {
    case focusFileList:
        m.fileList, cmd = m.fileList.Update(msg)
    case focusDestination:
        m.destination, cmd = m.destination.Update(msg)
    }
    return m, cmd
}
```

---

## Part 2: Go Language Best Practices

### Error Handling

```go
// Always handle errors explicitly - never ignore them
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doing something: %w", err) // Wrap with context
}

// Use errors.Is and errors.As for error checking
if errors.Is(err, os.ErrNotExist) {
    // Handle missing file
}

var pathErr *os.PathError
if errors.As(err, &pathErr) {
    // Access pathErr.Path, pathErr.Op, etc.
}

// Define sentinel errors for expected conditions
var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
)

// Custom error types for rich error information
type TransferError struct {
    File   string
    Reason string
    Err    error
}

func (e *TransferError) Error() string {
    return fmt.Sprintf("transfer failed for %s: %s", e.File, e.Reason)
}

func (e *TransferError) Unwrap() error {
    return e.Err
}
```

### Naming Conventions

```go
// Package names: short, lowercase, no underscores
package transfer  // good
package fileTransfer  // bad

// Interfaces: verb-er pattern for single-method interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Transferable interface {
    Transfer(dest string) error
}

// Exported vs unexported
type Client struct {        // Exported - public API
    endpoint string         // Unexported - internal state
    Timeout  time.Duration  // Exported - configurable
}

// Getters don't use "Get" prefix
func (c *Client) Endpoint() string { return c.endpoint }  // good
func (c *Client) GetEndpoint() string { }                  // bad

// Acronyms: consistent casing
var httpClient *http.Client  // unexported
var HTTPClient *http.Client  // exported (all caps for acronym)
type URLParser struct{}      // URL is acronym, all caps
```

### Struct Design

```go
// Use functional options for configurable types
type TransferClient struct {
    host       string
    port       int
    timeout    time.Duration
    bufferSize int
    logger     *slog.Logger
}

type Option func(*TransferClient)

func WithTimeout(d time.Duration) Option {
    return func(c *TransferClient) {
        c.timeout = d
    }
}

func WithBufferSize(size int) Option {
    return func(c *TransferClient) {
        c.bufferSize = size
    }
}

func WithLogger(l *slog.Logger) Option {
    return func(c *TransferClient) {
        c.logger = l
    }
}

func NewTransferClient(host string, port int, opts ...Option) *TransferClient {
    c := &TransferClient{
        host:       host,
        port:       port,
        timeout:    30 * time.Second,  // sensible defaults
        bufferSize: 32 * 1024,
        logger:     slog.Default(),
    }
    for _, opt := range opts {
        opt(c)
    }
    return c
}

// Usage
client := NewTransferClient("example.com", 22,
    WithTimeout(60*time.Second),
    WithBufferSize(64*1024),
)
```

### Concurrency Patterns

```go
// Use context for cancellation and timeouts
func (c *Client) Transfer(ctx context.Context, files []string) error {
    for _, file := range files {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := c.transferFile(ctx, file); err != nil {
                return err
            }
        }
    }
    return nil
}

// Worker pool pattern for parallel processing
func TransferFiles(ctx context.Context, files []string, workers int) error {
    g, ctx := errgroup.WithContext(ctx)
    fileChan := make(chan string)

    // Spawn workers
    for i := 0; i < workers; i++ {
        g.Go(func() error {
            for file := range fileChan {
                if err := transferFile(ctx, file); err != nil {
                    return err
                }
            }
            return nil
        })
    }

    // Send files to workers
    g.Go(func() error {
        defer close(fileChan)
        for _, file := range files {
            select {
            case fileChan <- file:
            case <-ctx.Done():
                return ctx.Err()
            }
        }
        return nil
    })

    return g.Wait()
}

// Protect shared state with sync.Mutex
type TransferStats struct {
    mu          sync.Mutex
    bytesTotal  int64
    filesDone   int
    errors      []error
}

func (s *TransferStats) RecordFile(bytes int64) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.bytesTotal += bytes
    s.filesDone++
}
```

### Interface Design

```go
// Accept interfaces, return concrete types
type FileSource interface {
    List() ([]FileInfo, error)
    Read(path string) (io.ReadCloser, error)
}

type FileDestination interface {
    Write(path string, r io.Reader, size int64) error
    Mkdir(path string) error
}

// Concrete implementation
type SFTPDestination struct {
    client *sftp.Client
}

func NewSFTPDestination(client *sftp.Client) *SFTPDestination {
    return &SFTPDestination{client: client}
}

func (d *SFTPDestination) Write(path string, r io.Reader, size int64) error {
    f, err := d.client.Create(path)
    if err != nil {
        return fmt.Errorf("creating remote file: %w", err)
    }
    defer f.Close()

    _, err = io.Copy(f, r)
    return err
}

// Transferer works with any source/destination
type Transferer struct {
    src  FileSource
    dest FileDestination
}

func (t *Transferer) TransferAll(ctx context.Context) error {
    files, err := t.src.List()
    if err != nil {
        return err
    }
    // ... transfer logic using interfaces
}
```

### Testing

```go
// Table-driven tests
func TestFormatSize(t *testing.T) {
    tests := []struct {
        name     string
        bytes    int64
        expected string
    }{
        {"zero", 0, "0 B"},
        {"bytes", 500, "500 B"},
        {"kilobytes", 1024, "1.0 KB"},
        {"megabytes", 1048576, "1.0 MB"},
        {"gigabytes", 1073741824, "1.0 GB"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := FormatSize(tt.bytes)
            if result != tt.expected {
                t.Errorf("FormatSize(%d) = %s, want %s", tt.bytes, result, tt.expected)
            }
        })
    }
}

// Use interfaces for testability
type mockDestination struct {
    files   map[string][]byte
    mkdirs  []string
    writeErr error
}

func (m *mockDestination) Write(path string, r io.Reader, size int64) error {
    if m.writeErr != nil {
        return m.writeErr
    }
    data, _ := io.ReadAll(r)
    m.files[path] = data
    return nil
}

func TestTransfer(t *testing.T) {
    dest := &mockDestination{files: make(map[string][]byte)}
    // Test with mock
}
```

---

## Part 3: High-Performance File Transfer

### SFTP Transfer with SSH

```go
// internal/transfer/sftp.go
package transfer

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sync"

    "github.com/pkg/sftp"
    "golang.org/x/crypto/ssh"
)

type SFTPConfig struct {
    Host           string
    Port           int
    User           string
    PrivateKeyPath string
    Password       string  // fallback if no key
    BufferSize     int     // default 32KB
    MaxConcurrent  int     // parallel transfers
}

type SFTPClient struct {
    config     SFTPConfig
    sshClient  *ssh.Client
    sftpClient *sftp.Client
    bufferPool sync.Pool
}

func NewSFTPClient(cfg SFTPConfig) (*SFTPClient, error) {
    if cfg.BufferSize == 0 {
        cfg.BufferSize = 32 * 1024
    }
    if cfg.MaxConcurrent == 0 {
        cfg.MaxConcurrent = 4
    }

    // Build auth methods
    var authMethods []ssh.AuthMethod
    if cfg.PrivateKeyPath != "" {
        key, err := os.ReadFile(cfg.PrivateKeyPath)
        if err != nil {
            return nil, fmt.Errorf("reading private key: %w", err)
        }
        signer, err := ssh.ParsePrivateKey(key)
        if err != nil {
            return nil, fmt.Errorf("parsing private key: %w", err)
        }
        authMethods = append(authMethods, ssh.PublicKeys(signer))
    }
    if cfg.Password != "" {
        authMethods = append(authMethods, ssh.Password(cfg.Password))
    }

    sshConfig := &ssh.ClientConfig{
        User:            cfg.User,
        Auth:            authMethods,
        HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: proper host key verification
    }

    addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
    sshClient, err := ssh.Dial("tcp", addr, sshConfig)
    if err != nil {
        return nil, fmt.Errorf("SSH dial: %w", err)
    }

    sftpClient, err := sftp.NewClient(sshClient,
        sftp.UseConcurrentWrites(true),
        sftp.UseConcurrentReads(true),
        sftp.MaxConcurrentRequestsPerFile(64),
    )
    if err != nil {
        sshClient.Close()
        return nil, fmt.Errorf("SFTP client: %w", err)
    }

    return &SFTPClient{
        config:     cfg,
        sshClient:  sshClient,
        sftpClient: sftpClient,
        bufferPool: sync.Pool{
            New: func() interface{} {
                return make([]byte, cfg.BufferSize)
            },
        },
    }, nil
}

func (c *SFTPClient) Close() error {
    c.sftpClient.Close()
    return c.sshClient.Close()
}

// TransferFiles transfers multiple files with progress reporting
func (c *SFTPClient) TransferFiles(ctx context.Context, files []FileTransfer, progress chan<- Progress) error {
    g, ctx := errgroup.WithContext(ctx)
    sem := make(chan struct{}, c.config.MaxConcurrent)
    var stats TransferStats

    for _, file := range files {
        file := file // capture for goroutine
        g.Go(func() error {
            sem <- struct{}{}        // acquire
            defer func() { <-sem }() // release

            err := c.transferFile(ctx, file, &stats, progress)
            if err != nil {
                return fmt.Errorf("transferring %s: %w", file.Source, err)
            }
            return nil
        })
    }

    return g.Wait()
}

func (c *SFTPClient) transferFile(ctx context.Context, file FileTransfer, stats *TransferStats, progress chan<- Progress) error {
    // Check context before starting
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    // Open source file
    src, err := os.Open(file.Source)
    if err != nil {
        return fmt.Errorf("opening source: %w", err)
    }
    defer src.Close()

    srcInfo, err := src.Stat()
    if err != nil {
        return err
    }

    // Create destination directory
    destDir := filepath.Dir(file.Destination)
    if err := c.sftpClient.MkdirAll(destDir); err != nil {
        return fmt.Errorf("creating directory %s: %w", destDir, err)
    }

    // Create destination file
    dest, err := c.sftpClient.Create(file.Destination)
    if err != nil {
        return fmt.Errorf("creating destination: %w", err)
    }
    defer dest.Close()

    // Get buffer from pool
    buf := c.bufferPool.Get().([]byte)
    defer c.bufferPool.Put(buf)

    // Copy with progress
    var written int64
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        n, err := src.Read(buf)
        if n > 0 {
            nw, werr := dest.Write(buf[:n])
            if werr != nil {
                return werr
            }
            written += int64(nw)

            // Report progress
            if progress != nil {
                select {
                case progress <- Progress{
                    File:       file.Source,
                    BytesDone:  written,
                    TotalBytes: srcInfo.Size(),
                }:
                default: // non-blocking
                }
            }
        }
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
    }

    // Preserve permissions
    if err := dest.Chmod(srcInfo.Mode()); err != nil {
        // Log but don't fail
    }

    stats.RecordFile(written)
    return nil
}

type FileTransfer struct {
    Source      string
    Destination string
}

type Progress struct {
    File       string
    BytesDone  int64
    TotalBytes int64
}

type TransferStats struct {
    mu         sync.Mutex
    BytesTotal int64
    FilesDone  int
    Errors     []error
}

func (s *TransferStats) RecordFile(bytes int64) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.BytesTotal += bytes
    s.FilesDone++
}
```

### USB/Local Transfer with High Performance

```go
// internal/transfer/usb.go
package transfer

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "runtime"
    "syscall"

    "golang.org/x/sync/errgroup"
)

type USBConfig struct {
    BufferSize    int  // Default 1MB for USB
    MaxConcurrent int  // Parallel copy operations
    UseDirectIO   bool // Bypass OS cache for large files
    SyncAfterCopy bool // fsync each file after copy
}

type USBTransfer struct {
    config     USBConfig
    bufferPool sync.Pool
}

func NewUSBTransfer(cfg USBConfig) *USBTransfer {
    if cfg.BufferSize == 0 {
        cfg.BufferSize = 1024 * 1024 // 1MB default for USB
    }
    if cfg.MaxConcurrent == 0 {
        cfg.MaxConcurrent = runtime.NumCPU()
    }

    return &USBTransfer{
        config: cfg,
        bufferPool: sync.Pool{
            New: func() interface{} {
                return make([]byte, cfg.BufferSize)
            },
        },
    }
}

// TransferToUSB copies files to USB with maximum throughput
func (u *USBTransfer) TransferToUSB(ctx context.Context, files []FileTransfer, progress chan<- Progress) error {
    g, ctx := errgroup.WithContext(ctx)
    sem := make(chan struct{}, u.config.MaxConcurrent)

    for _, file := range files {
        file := file
        g.Go(func() error {
            sem <- struct{}{}
            defer func() { <-sem }()
            return u.copyFile(ctx, file, progress)
        })
    }

    return g.Wait()
}

func (u *USBTransfer) copyFile(ctx context.Context, file FileTransfer, progress chan<- Progress) error {
    src, err := os.Open(file.Source)
    if err != nil {
        return fmt.Errorf("opening source: %w", err)
    }
    defer src.Close()

    srcInfo, err := src.Stat()
    if err != nil {
        return err
    }

    // Create destination directory
    if err := os.MkdirAll(filepath.Dir(file.Destination), 0755); err != nil {
        return fmt.Errorf("creating directory: %w", err)
    }

    // Create destination with same permissions
    dest, err := os.OpenFile(file.Destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
    if err != nil {
        return fmt.Errorf("creating destination: %w", err)
    }
    defer dest.Close()

    // Pre-allocate space on destination (huge performance boost for large files)
    if srcInfo.Size() > 0 {
        if err := preallocate(dest, srcInfo.Size()); err != nil {
            // Log but continue - preallocation is optimization only
        }
    }

    // Get buffer from pool
    buf := u.bufferPool.Get().([]byte)
    defer u.bufferPool.Put(buf)

    var written int64
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        n, err := src.Read(buf)
        if n > 0 {
            nw, werr := dest.Write(buf[:n])
            if werr != nil {
                return werr
            }
            written += int64(nw)

            if progress != nil {
                select {
                case progress <- Progress{
                    File:       file.Source,
                    BytesDone:  written,
                    TotalBytes: srcInfo.Size(),
                }:
                default:
                }
            }
        }
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
    }

    // Sync to ensure data is on disk (important for USB)
    if u.config.SyncAfterCopy {
        if err := dest.Sync(); err != nil {
            return fmt.Errorf("syncing file: %w", err)
        }
    }

    return nil
}

// preallocate uses fallocate on Linux to pre-allocate space
func preallocate(f *os.File, size int64) error {
    if runtime.GOOS == "linux" {
        return syscall.Fallocate(int(f.Fd()), 0, 0, size)
    }
    // On other systems, seek and write a byte (less efficient)
    if _, err := f.Seek(size-1, 0); err != nil {
        return err
    }
    if _, err := f.Write([]byte{0}); err != nil {
        return err
    }
    _, err := f.Seek(0, 0)
    return err
}
```

### Transfer Queue Management

```go
// internal/transfer/queue.go
package transfer

import (
    "context"
    "sync"
    "time"
)

type TransferStatus int

const (
    StatusPending TransferStatus = iota
    StatusInProgress
    StatusComplete
    StatusFailed
)

type TransferJob struct {
    ID          string
    Source      string
    Destination string
    Size        int64
    Status      TransferStatus
    Progress    float64
    Error       error
    StartTime   time.Time
    EndTime     time.Time
}

type TransferQueue struct {
    mu       sync.RWMutex
    jobs     map[string]*TransferJob
    pending  chan *TransferJob
    workers  int
    running  bool
    progress chan Progress
}

func NewTransferQueue(workers int) *TransferQueue {
    return &TransferQueue{
        jobs:     make(map[string]*TransferJob),
        pending:  make(chan *TransferJob, 1000),
        workers:  workers,
        progress: make(chan Progress, 100),
    }
}

func (q *TransferQueue) Add(job *TransferJob) {
    q.mu.Lock()
    q.jobs[job.ID] = job
    q.mu.Unlock()
    q.pending <- job
}

func (q *TransferQueue) GetProgress() <-chan Progress {
    return q.progress
}

func (q *TransferQueue) Start(ctx context.Context, transferFn func(context.Context, *TransferJob, chan<- Progress) error) {
    q.running = true
    var wg sync.WaitGroup

    for i := 0; i < q.workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                case job, ok := <-q.pending:
                    if !ok {
                        return
                    }
                    q.processJob(ctx, job, transferFn)
                }
            }
        }()
    }

    wg.Wait()
}

func (q *TransferQueue) processJob(ctx context.Context, job *TransferJob, transferFn func(context.Context, *TransferJob, chan<- Progress) error) {
    q.mu.Lock()
    job.Status = StatusInProgress
    job.StartTime = time.Now()
    q.mu.Unlock()

    err := transferFn(ctx, job, q.progress)

    q.mu.Lock()
    job.EndTime = time.Now()
    if err != nil {
        job.Status = StatusFailed
        job.Error = err
    } else {
        job.Status = StatusComplete
        job.Progress = 100.0
    }
    q.mu.Unlock()
}

func (q *TransferQueue) Stats() (pending, inProgress, complete, failed int) {
    q.mu.RLock()
    defer q.mu.RUnlock()
    for _, job := range q.jobs {
        switch job.Status {
        case StatusPending:
            pending++
        case StatusInProgress:
            inProgress++
        case StatusComplete:
            complete++
        case StatusFailed:
            failed++
        }
    }
    return
}
```

### Batch File Discovery

```go
// internal/transfer/discover.go
package transfer

import (
    "io/fs"
    "os"
    "path/filepath"
    "strings"
)

type DiscoverOptions struct {
    Recursive   bool
    FollowLinks bool
    Patterns    []string // glob patterns to include
    Exclude     []string // glob patterns to exclude
    MinSize     int64
    MaxSize     int64
}

// DiscoverFiles finds all files matching criteria
func DiscoverFiles(root string, opts DiscoverOptions) ([]FileTransfer, error) {
    var files []FileTransfer

    walkFn := func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        // Skip directories if not recursive
        if d.IsDir() {
            if path != root && !opts.Recursive {
                return fs.SkipDir
            }
            return nil
        }

        // Check exclusions
        for _, pattern := range opts.Exclude {
            if matched, _ := filepath.Match(pattern, d.Name()); matched {
                return nil
            }
        }

        // Check inclusions
        if len(opts.Patterns) > 0 {
            matched := false
            for _, pattern := range opts.Patterns {
                if m, _ := filepath.Match(pattern, d.Name()); m {
                    matched = true
                    break
                }
            }
            if !matched {
                return nil
            }
        }

        // Check size constraints
        if opts.MinSize > 0 || opts.MaxSize > 0 {
            info, err := d.Info()
            if err != nil {
                return nil
            }
            if opts.MinSize > 0 && info.Size() < opts.MinSize {
                return nil
            }
            if opts.MaxSize > 0 && info.Size() > opts.MaxSize {
                return nil
            }
        }

        files = append(files, FileTransfer{
            Source: path,
        })
        return nil
    }

    if err := filepath.WalkDir(root, walkFn); err != nil {
        return nil, err
    }

    return files, nil
}

// SetDestinations updates all files with destination paths
func SetDestinations(files []FileTransfer, srcRoot, destRoot string) {
    for i := range files {
        rel, _ := filepath.Rel(srcRoot, files[i].Source)
        files[i].Destination = filepath.Join(destRoot, rel)
    }
}
```

### Performance Tips Summary

| Technique | When to Use | Benefit |
|-----------|-------------|---------|
| Buffer pooling | Always | Reduces GC pressure |
| Pre-allocation | Large files | Reduces fragmentation |
| Parallel workers | Multiple files | Saturates bandwidth |
| Context cancellation | Always | Clean shutdown |
| Direct I/O | Very large files | Bypasses OS cache |
| fsync | USB/removable | Ensures data persists |
| Concurrent SFTP | Always | Multiplexes SSH connection |

### USB Detection (Linux)

```go
// internal/transfer/usb_linux.go
package transfer

import (
    "bufio"
    "os"
    "path/filepath"
    "strings"
)

type USBDevice struct {
    Path       string
    MountPoint string
    Label      string
    Size       int64
    Removable  bool
}

func DetectUSBDevices() ([]USBDevice, error) {
    var devices []USBDevice

    // Read /proc/mounts
    file, err := os.Open("/proc/mounts")
    if err != nil {
        return nil, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        fields := strings.Fields(scanner.Text())
        if len(fields) < 2 {
            continue
        }

        device, mountPoint := fields[0], fields[1]

        // Check if removable
        if strings.HasPrefix(device, "/dev/sd") {
            // Extract device letter (e.g., "sda" from "/dev/sda1")
            base := filepath.Base(device)
            if len(base) >= 3 {
                letter := base[:3]
                removablePath := filepath.Join("/sys/block", letter, "removable")
                if data, err := os.ReadFile(removablePath); err == nil {
                    if strings.TrimSpace(string(data)) == "1" {
                        devices = append(devices, USBDevice{
                            Path:       device,
                            MountPoint: mountPoint,
                            Removable:  true,
                        })
                    }
                }
            }
        }
    }

    return devices, scanner.Err()
}
```

---

## Quick Reference Commands

### Building and Running

```bash
# Build with optimizations
go build -ldflags="-s -w" -o myapp ./cmd/myapp

# Run with race detector during development
go run -race ./cmd/myapp

# Test with coverage
go test -cover -race ./...

# Generate test coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Benchmark
go test -bench=. -benchmem ./...
```

### Useful Tools

```bash
# Format code
gofmt -w .
goimports -w .

# Lint
golangci-lint run

# Find inefficient allocations
go build -gcflags='-m' ./...

# Profile CPU
go test -cpuprofile=cpu.out -bench=.
go tool pprof cpu.out
```

---

## Checklist for TUI File Transfer App

- [ ] Use Bubble Tea with proper Model/Update/View separation
- [ ] Centralize styles in `styles.go`
- [ ] Define key bindings with `bubbles/key`
- [ ] Use `bubbles/list` for file selection with filtering
- [ ] Use `bubbles/progress` for transfer progress
- [ ] Handle `tea.WindowSizeMsg` for responsive layout
- [ ] Implement proper context cancellation
- [ ] Use buffer pools to reduce allocations
- [ ] Parallelize transfers with worker pools
- [ ] Pre-allocate destination files for large transfers
- [ ] Report progress via channels (non-blocking sends)
- [ ] Sync files after copy for USB safety
- [ ] Wrap errors with context using `%w`
- [ ] Write table-driven tests with mocks

---

## Part 4: ReplayOS ROM Handling

### Overview

ReplayOS is a lightweight, optimized Linux distribution for Raspberry Pi hardware (Zero 2, 3A/3B/3B+, 4B, 5B, Pi 500, CM5) built around libretro/RetroArch cores. It emphasizes low-latency emulation with support for both LCD and CRT displays.

**Key Constraints:**
- Raspberry Pi hardware only (no PC/x86, no handhelds)
- Fixed set of libretro cores (no custom core support)
- BIOS files required for many systems
- Specific ROM formats per system

### Folder Structure
```
├── bios
├── captures
├── config
│   ├── input
│   │   ├── game
│   │   │   ├── crt
│   │   │   └── lcd
│   │   └── system
│   │       ├── crt
│   │       └── lcd
│   └── settings
│       ├── game
│       │   ├── crt
│       │   └── lcd
│       └── system
│           ├── crt
│           └── lcd
├── roms
│   ├── amstrad_cpc
│   ├── arcade_dc
│   ├── arcade_fbneo
│   ├── arcade_mame
│   ├── arcade_mame_2k3p
│   ├── atari_2600
│   ├── atari_5200
│   ├── atari_7800
│   ├── atari_jaguar
│   ├── atari_lynx
│   ├── _autostart
│   ├── commodore_amiga
│   ├── commodore_amigacd
│   ├── commodore_c64
│   ├── _extra
│   ├── _favorites
│   ├── ibm_pc
│   ├── media_player
│   ├── microsoft_msx
│   ├── nec_pce
│   ├── nec_pcecd
│   ├── nintendo_ds
│   ├── nintendo_gb
│   ├── nintendo_gba
│   ├── nintendo_n64
│   ├── nintendo_nes
│   ├── nintendo_snes
│   ├── panasonic_3do
│   ├── philips_cdi
│   ├── _recent
│   ├── scummvm
│   ├── sega_32x
│   ├── sega_cd
│   ├── sega_dc
│   ├── sega_gg
│   ├── sega_sg
│   ├── sega_smd
│   ├── sega_sms
│   ├── sega_st
│   ├── sharp_x68k
│   ├── sinclair_zx
│   ├── snk_ng
│   ├── snk_ngcd
│   ├── snk_ngp
│   └── sony_psx
└── saves
```

### Supported Systems and File Formats

| System | Folder | Formats | Notes |
|--------|--------|---------|-------|
| Arcade (FBNeo) | arcade_fbneo | zip | Must match FBNeo romset version |
| Arcade (MAME) | arcade_mame | zip | Must match MAME romset version |
| Arcade (MAME 2K3+) | arcade_mame_2k3p | zip | MAME 0.78 romset |
| Arcade (Naomi/Atomiswave) | arcade_dc | zip | Requires dc/ BIOS files |
| Atari 2600 | atari_2600 | a26, bin | No BIOS required |
| Atari 5200 | atari_5200 | a52, bin | Requires 5200.rom |
| Atari 7800 | atari_7800 | a78, bin, cdf | Requires 7800 BIOS (U).rom |
| Atari Jaguar | atari_jaguar | j64, jag | No BIOS required |
| Atari Lynx | atari_lynx | lnx | Requires lynxboot.img |
| NEC PC Engine | nec_pce | pce, sgx, toc | No BIOS required |
| NEC PC Engine CD | nec_pcecd | cue, ccd, chd, m3u | Requires syscard1-3.pce |
| Nintendo NES | nintendo_nes | fds, nes, unf, unif | disksys.rom for FDS |
| Nintendo SNES | nintendo_snes | smc, sfc, swc, fig, bs, st | No BIOS required |
| Nintendo 64 | nintendo_n64 | n64, v64, z64, bin, u1 | No BIOS required |
| Nintendo GB | nintendo_gb | gb, sgb | Optional: gb_bios.bin |
| Nintendo GBC | nintendo_gb | gbc, sgbc | Optional: gbc_bios.bin |
| Nintendo GBA | nintendo_gba | gba | Requires gba_bios.bin |
| Nintendo DS | nintendo_ds | nds | Requires melonDS DS/ folder |
| SEGA SG-1000 | sega_sg | sg | No BIOS required |
| SEGA Game Gear | sega_gg | gg | No BIOS required |
| SEGA Master System | sega_sms | sms | No BIOS required |
| SEGA Genesis | sega_smd | md, smd, gen, bin | No BIOS required |
| SEGA CD | sega_cd | m3u, cue, iso, chd | Requires bios_CD_{E,J,U}.bin |
| SEGA 32X | sega_32x | 32x | No BIOS required |
| SEGA Saturn | sega_st | cue, ccd, chd, toc, m3u | Requires sega_101.bin, mpr-17933.bin |
| SEGA Dreamcast | sega_dc | chd, cdi, elf, cue, gdi, m3u | Requires dc/dc_boot.bin |
| SNK Neo Geo | snk_ng | zip | Requires fbneo/neogeo.zip |
| SNK Neo Geo CD | snk_ngcd | cue, chd | Requires neocd/neocd_z.rom |
| SNK Neo Geo Pocket | snk_ngp | ngp, ngc, ngpc, npc | No BIOS required |
| Sony PlayStation | sony_psx | cue, iso, chd, pbp, m3u | Requires scph5500-5502.bin |
| Panasonic 3DO | panasonic_3do | iso, chd, cue | Requires panafz10.bin |
| Philips CD-i | philips_cdi | iso, chd, cue | Requires same_cdi/bios/*.zip |
| Amstrad CPC | amstrad_cpc | dsk, sna, tap, cdt, m3u | No BIOS required |
| Commodore 64 | commodore_c64 | d64, t64, tap, prg, crt, m3u | No BIOS required |
| Commodore Amiga | commodore_amiga | adf, adz, hdf, m3u | Requires kick*.A500/A600/A1200 |
| Commodore Amiga CD32 | commodore_amigacd | cue, iso, chd, m3u | Requires kick40060.CD32* |
| Sharp X68000 | sharp_x68k | dim, img, d88, m3u | Requires keropi/*.dat |
| Microsoft MSX | microsoft_msx | rom, dsk, cas, m3u | Requires Machines/Shared Roms/*.ROM |
| Sinclair ZX Spectrum | sinclair_zx | tzx, tap, z80, sna, szx | No BIOS required |
| IBM PC (DOS) | ibm_pc | zip, dosz, exe, iso, m3u | Optional: MT32/SC-55 soundfonts |
| ScummVM | scummvm | scummvm, svm | Optional: MT32/SC-55 soundfonts |

### BIOS File Requirements

ReplayOS validates BIOS files before launching games. Missing BIOS = blocked launch.

**Critical BIOS Files by System:**

```
bios/
├── 5200.rom                          # Atari 5200
├── 7800 BIOS (U).rom                 # Atari 7800
├── lynxboot.img                      # Atari Lynx
├── gba_bios.bin                      # Game Boy Advance
├── gb_bios.bin                       # Game Boy (optional)
├── gbc_bios.bin                      # Game Boy Color (optional)
├── disksys.rom                       # Famicom Disk System
├── syscard1.pce                      # PC Engine CD
├── syscard2.pce                      # PC Engine CD
├── syscard3.pce                      # PC Engine CD (preferred)
├── gexpress.pce                      # PC Engine CD (Game Express)
├── bios_CD_E.bin                     # SEGA CD (Europe)
├── bios_CD_J.bin                     # SEGA CD (Japan)
├── bios_CD_U.bin                     # SEGA CD (USA)
├── sega_101.bin                      # SEGA Saturn
├── mpr-17933.bin                     # SEGA Saturn
├── scph5500.bin                      # PlayStation (Japan)
├── scph5501.bin                      # PlayStation (USA)
├── scph5502.bin                      # PlayStation (Europe)
├── panafz10.bin                      # 3DO
├── dc/
│   ├── dc_boot.bin                   # Dreamcast
│   ├── naomi.zip                     # Naomi arcade
│   ├── naomi2.zip                    # Naomi 2 arcade
│   ├── awbios.zip                    # Atomiswave
│   └── ...
├── fbneo/
│   ├── neogeo.zip                    # Neo Geo (REQUIRED for all NG games)
│   └── samples/                      # Arcade sound samples
├── neocd/
│   └── neocd_z.rom                   # Neo Geo CD
├── melonDS DS/
│   ├── bios7.bin                     # Nintendo DS
│   ├── bios9.bin                     # Nintendo DS
│   ├── firmware.bin                  # Nintendo DS
│   ├── dsi_bios7.bin                 # DSi (optional)
│   ├── dsi_bios9.bin                 # DSi (optional)
│   ├── dsi_firmware.bin              # DSi (optional)
│   └── dsi_nand.bin                  # DSi (optional)
├── same_cdi/bios/
│   ├── cdibios.zip                   # CD-i
│   ├── cdimono1.zip                  # CD-i
│   └── cdimono2.zip                  # CD-i
├── keropi/
│   ├── cgrom.dat                     # Sharp X68000
│   ├── iplrom.dat                    # Sharp X68000
│   ├── iplrom30.dat                  # Sharp X68000
│   ├── iplromco.dat                  # Sharp X68000
│   └── iplromxv.dat                  # Sharp X68000
├── Machines/Shared Roms/
│   ├── MSX.ROM                       # MSX
│   ├── MSX2.ROM                      # MSX2
│   ├── MSX2EXT.ROM                   # MSX2
│   ├── MSX2P.ROM                     # MSX2+
│   ├── MSX2PEXT.ROM                  # MSX2+
│   ├── FMPAC.ROM                     # MSX FM-PAC
│   └── KANJI.ROM                     # MSX Kanji
├── kick33180.A500                    # Amiga 500
├── kick34005.A500                    # Amiga 500
├── kick37175.A500                    # Amiga 500
├── kick37350.A600                    # Amiga 600
├── kick39106.A1200                   # Amiga 1200
├── kick39106.A4000                   # Amiga 4000
├── kick40060.CD32                    # Amiga CD32
├── kick40060.CD32.ext                # Amiga CD32 Extended
└── scummvm/extra/
    ├── Roland_SC-55.sf2              # Roland SC-55 Soundfont
    ├── MT32_CONTROL.ROM              # MT-32 Control ROM
    ├── MT32_PCM.ROM                  # MT-32 PCM ROM
    ├── CM32L_CONTROL.ROM             # CM-32L Control ROM
    └── CM32L_PCM.ROM                 # CM-32L PCM ROM
```

### Arcade ROM Handling

Arcade emulation requires **exact romset version matching**.

#### ROMset Types

| Type | Description | Recommended |
|------|-------------|-------------|
| Non-merged | Each ZIP is standalone, includes parent files | ✅ Yes |
| Full non-merged | Like non-merged, BIOS included in each ZIP | ✅ Yes |
| Split | Clones reference parent ZIPs separately | ⚠️ Works |
| Merged | Multiple games per ZIP | ❌ No |

**Always use non-merged or full non-merged sets for ReplayOS.**

#### ROMset Versions

| ReplayOS Core | Required ROMset |
|---------------|-----------------|
| arcade_fbneo | Latest FBNeo (sync with current libretro) |
| arcade_mame | Latest MAME (check ReplayOS version) |
| arcade_mame_2k3p | MAME 0.78 |

#### Parent/Clone Relationships

Arcade ROMs use short codenames, not full titles:

```
sf2.zip        → Street Fighter II (parent)
sf2ce.zip      → Street Fighter II Champion Edition (clone)
sf2hf.zip      → Street Fighter II Hyper Fighting (clone)
```

**Best practice:** Use parent ROMs unless you specifically need a regional variant.

#### Neo Geo Special Handling

Neo Geo games **always** require `neogeo.zip` BIOS in `bios/fbneo/`:

```
bios/fbneo/neogeo.zip     # REQUIRED - contains system BIOS
roms/snk_ng/mslug.zip     # Metal Slug
roms/snk_ng/kof98.zip     # King of Fighters '98
```

### CHD Conversion

CHD (Compressed Hunks of Data) is the preferred format for CD-based games.

#### Converting to CHD

```bash
# Install chdman (comes with MAME)
# On Arch/CachyOS:
sudo pacman -S mame-tools

# Convert BIN/CUE to CHD
chdman createcd -i game.cue -o game.chd

# Convert GDI to CHD (Dreamcast)
chdman createcd -i game.gdi -o game.chd

# Convert ISO to CHD
chdman createcd -i game.iso -o game.chd

# Batch convert all CUE files in directory
for f in *.cue; do
    chdman createcd -i "$f" -o "${f%.cue}.chd"
done
```

#### Converting from CHD (if needed)

```bash
# Extract CHD back to BIN/CUE
chdman extractcd -i game.chd -o game.cue -ob game.bin
```

#### CHD Benefits
- Single file per disc (no BIN/CUE pairs)
- Lossless compression (typically 40-60% size reduction)
- Widely supported by libretro cores
- Works with M3U playlists

### M3U Multi-Disc Playlists

M3U files organize multi-disc games and hide individual disc files from the frontend.

#### Creating M3U Files

```bash
# Example: Final Fantasy VII (3 discs)
# File: roms/sony_psx/Final Fantasy VII.m3u

Final Fantasy VII (Disc 1).chd
Final Fantasy VII (Disc 2).chd
Final Fantasy VII (Disc 3).chd
```

**Critical M3U Rules:**
1. **Plain text, UTF-8 or ASCII encoding**
2. **Unix line endings (LF, not CRLF)** - use `dos2unix` if needed
3. **One disc per line, in order**
4. **Relative paths from M3U location**
5. **No extra blank lines or spaces**
6. **Filename must NOT include disc number**

#### M3U with Subdirectory Organization

```bash
# Hide disc files in subfolder
roms/sony_psx/
├── Final Fantasy VII.m3u
└── .hidden/
    ├── Final Fantasy VII (Disc 1).chd
    ├── Final Fantasy VII (Disc 2).chd
    └── Final Fantasy VII (Disc 3).chd

# M3U contents:
.hidden/Final Fantasy VII (Disc 1).chd
.hidden/Final Fantasy VII (Disc 2).chd
.hidden/Final Fantasy VII (Disc 3).chd
```

#### Batch M3U Generation Script

```bash
#!/bin/bash
# generate_m3u.sh - Create M3U files for multi-disc games

cd "$1" || exit 1

# Find all disc 1 files and create M3U
find . -maxdepth 1 \( -name "*Disc 1*.chd" -o -name "*Disc 1*.cue" -o -name "*CD1*.chd" \) | while read -r disc1; do
    # Extract base name (remove disc identifier)
    base=$(echo "$disc1" | sed -E 's/[[:space:]]?\(?Disc [0-9]+\)?//g; s/[[:space:]]?\(?CD[0-9]+\)?//g; s/\.(chd|cue)$//')
    base=$(basename "$base")
    
    m3u_file="${base}.m3u"
    
    # Skip if M3U exists
    [[ -f "$m3u_file" ]] && continue
    
    # Find all discs for this game
    ext="${disc1##*.}"
    find . -maxdepth 1 -name "${base}*Disc*.${ext}" -o -name "${base}*CD*.${ext}" | sort > "$m3u_file"
    
    # Convert to relative paths
    sed -i 's|^\./||' "$m3u_file"
    
    echo "Created: $m3u_file"
done
```

### ROM Validation with DAT Files

Use ClrMamePro or other ROM managers to validate your sets.

#### Getting DAT Files

```bash
# FBNeo DAT files (official)
# https://github.com/libretro/FBNeo/tree/master/dats

# MAME DAT files
# Included with MAME releases, or from Pleasuredome

# No-Intro DAT files (console ROMs)
# https://datomatic.no-intro.org/
```

#### Basic ClrMamePro Workflow

1. Load the appropriate DAT file
2. Point to your ROM folder as "ROM path"
3. Point to source folders as "Rebuild" paths
4. Run Scanner to identify missing/wrong files
5. Run Rebuilder to fix issues

### File Transfer to ReplayOS

#### Recommended Methods

| Method | Speed | Setup | Best For |
|--------|-------|-------|----------|
| USB Drive | Fast | Easy | Large collections |
| SFTP over WiFi | Slow | Medium | Small updates |
| NFS Share | Fast | Complex | Permanent setup |
| Direct SD Access | Fastest | Easy | Initial setup |

#### USB Drive Transfer

```bash
# Format USB as exFAT (best compatibility)
sudo mkfs.exfat -n REPLAY /dev/sdX1

# Mount and copy
sudo mount /dev/sdX1 /mnt/usb
cp -rv ~/roms/* /mnt/usb/roms/
cp -rv ~/bios/* /mnt/usb/bios/
sync  # CRITICAL: ensure all writes complete
sudo umount /mnt/usb
```

#### SFTP Transfer

```bash
# Connect to ReplayOS (default credentials vary by version)
sftp pi@replayos.local

# Or with explicit IP
sftp pi@192.168.1.100

# Batch upload
sftp> put -r /local/roms/* /media/sd/roms/
sftp> put -r /local/bios/* /media/sd/bios/
```

#### High-Performance Batch Transfer

Use the techniques from Part 3 (File Transfer) for large ROM collections:

```go
// Example: Transfer entire ROM collection to ReplayOS via SFTP
func TransferROMCollection(sftpClient *SFTPClient, localRoot, remoteRoot string) error {
    files, err := DiscoverFiles(localRoot, DiscoverOptions{
        Recursive: true,
        Patterns:  []string{"*.zip", "*.chd", "*.bin", "*.cue", "*.m3u"},
    })
    if err != nil {
        return err
    }
    
    SetDestinations(files, localRoot, remoteRoot)
    
    progress := make(chan Progress, 100)
    go func() {
        for p := range progress {
            fmt.Printf("\r%s: %.1f%%", p.File, float64(p.BytesDone)/float64(p.TotalBytes)*100)
        }
    }()
    
    return sftpClient.TransferFiles(context.Background(), files, progress)
}
```

### ReplayOS Configuration

Configuration file: `/media/sd/config/replay.cfg`

#### Key Settings for ROM Management

```ini
# Storage location (sd, usb, nfs)
system_storage = "usb"

# NFS configuration
nfs_server = "192.168.1.50"
nfs_share = "/export/roms"
nfs_version = "4"

# Boot directly to specific system
system_boot_to_system = "arcade_fbneo"

# Regenerate folder listings on boot
system_folder_regen = "true"
```

### Troubleshooting ROM Issues

| Symptom | Likely Cause | Solution |
|---------|--------------|----------|
| Game won't launch, returns to menu | Missing BIOS | Check BIOS requirements for system |
| Arcade game "ROM not found" | Wrong romset version | Rebuild with correct DAT |
| CHD won't load | Corrupt or wrong version | Re-convert from source |
| M3U shows all discs | Wrong M3U format | Check encoding (UTF-8), line endings (LF) |
| Save not persisting | Wrong folder | Check saves/ directory exists |
| Slow performance | Wrong Pi model | Some systems need Pi 4/5 |

### Quick Reference: ROM Preparation Workflow

1. **Verify source files** - Check against No-Intro or Redump DATs
2. **Convert CD games to CHD** - `chdman createcd`
3. **Create M3U for multi-disc** - Plain text, one file per line
4. **Verify arcade romsets** - Match to FBNeo/MAME version
5. **Collect BIOS files** - Check system requirements
6. **Organize into correct folders** - Match ReplayOS structure
7. **Transfer to device** - USB recommended for large sets
8. **Verify on device** - Test representative games per system
