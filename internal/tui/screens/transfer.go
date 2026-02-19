package screens

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/transfer"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type transferPhase int

const (
	transferPhaseMethod transferPhase = iota
	transferPhaseFolders
	transferPhaseConnect
	transferPhasePlan
	transferPhaseProgress
	transferPhaseResults
)

type transferConnectMsg struct {
	err error
}

type transferPlanMsg struct {
	plan *transfer.TransferPlan
	err  error
}

type transferProgressMsg struct {
	progress transfer.TransferProgress
}

type transferDoneMsg struct {
	err error
}

type transferFolder struct {
	label   string // "ROMs", "BIOS", "Saves", "Config"
	dirName string // "roms", "bios", "saves", "config"
	selected bool
}

type TransferScreen struct {
	cfg           *config.Config
	width, height int
	phase         transferPhase

	// Method selection
	methods []string
	cursor  int
	toolErr string // shown when a required tool (sshpass/rsync) is missing

	// Folder selection
	folderOptions []transferFolder
	folderCursor  int

	// Connection (USB path)
	backend    transfer.TransferBackend
	connectErr error

	// Plan (USB)
	plan             *transfer.TransferPlan
	planErr          error
	planFolderLabels []string

	// Bulk backend (rsync)
	bulkBackend transfer.BulkTransferBackend
	isBulk      bool

	// Progress
	progressCh      <-chan transfer.TransferProgress
	currentProgress transfer.TransferProgress

	// Cancellation
	cancel context.CancelFunc

	// Results
	itemsTransferred int // files (USB) or folders (rsync)
	totalErr         error
}

func NewTransferScreen(cfg *config.Config, width, height int) *TransferScreen {
	return &TransferScreen{
		cfg:    cfg,
		width:  width,
		height: height,
		methods: []string{
			"rsync - Fast incremental sync over SSH",
			"USB / SD Card",
			"Manual Instructions",
		},
	}
}

func (t *TransferScreen) Init() tea.Cmd { return nil }

func (t *TransferScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height

	case transferConnectMsg:
		if msg.err != nil {
			t.connectErr = msg.err
			t.phase = transferPhaseMethod
		} else {
			t.phase = transferPhasePlan
			return t, t.buildPlan()
		}

	case transferPlanMsg:
		if msg.err != nil {
			t.planErr = msg.err
			t.phase = transferPhaseMethod
		} else {
			t.plan = msg.plan
		}

	case transferProgressMsg:
		t.currentProgress = msg.progress
		if msg.progress.Done {
			t.itemsTransferred++
		}
		return t, listenTransferProgress(t.progressCh)

	case transferDoneMsg:
		t.totalErr = msg.err
		t.cancel = nil
		t.phase = transferPhaseResults

	case tea.KeyMsg:
		switch t.phase {
		case transferPhaseMethod:
			return t.updateMethod(msg)
		case transferPhaseFolders:
			return t.updateFolders(msg)
		case transferPhasePlan:
			return t.updatePlan(msg)
		case transferPhaseProgress:
			if key.Matches(msg, tui.Keys.Back) {
				if t.cancel != nil {
					t.cancel()
				}
			}
		case transferPhaseResults:
			if key.Matches(msg, tui.Keys.Back) || key.Matches(msg, tui.Keys.Enter) {
				t.closeBackends()
				return t, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		default:
			if key.Matches(msg, tui.Keys.Back) {
				t.closeBackends()
				return t, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	}
	return t, nil
}

func (t *TransferScreen) closeBackends() {
	if t.backend != nil {
		t.backend.Close()
	}
	if t.bulkBackend != nil {
		t.bulkBackend.Close()
	}
}

func (t *TransferScreen) updateMethod(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		return t, func() tea.Msg { return tui.NavigateBackMsg{} }
	case key.Matches(msg, tui.Keys.Up):
		if t.cursor > 0 {
			t.cursor--
		}
	case key.Matches(msg, tui.Keys.Down):
		if t.cursor < len(t.methods)-1 {
			t.cursor++
		}
	case key.Matches(msg, tui.Keys.Enter):
		t.toolErr = ""
		switch t.cursor {
		case 0: // rsync
			if _, err := transfer.FindTool("sshpass"); err != nil {
				t.toolErr = "Requires 'sshpass' \u2014 install: sudo pacman -S sshpass"
				return t, nil
			}
			if _, err := transfer.FindTool("rsync"); err != nil {
				t.toolErr = "Requires 'rsync' \u2014 install: sudo pacman -S rsync"
				return t, nil
			}
			backend := transfer.NewRsyncBackend(
				t.cfg.Device.Host,
				t.cfg.Device.Port,
				t.cfg.Device.User,
				t.cfg.Device.Password,
			)
			backend.Concurrency = t.cfg.Transfer.Concurrency
			t.bulkBackend = backend
			t.isBulk = true
			t.initFolderSelection()
			t.phase = transferPhaseFolders
		case 1: // USB
			t.backend = transfer.NewUSBBackend(t.cfg.Transfer.USBPath)
			t.isBulk = false
			t.initFolderSelection()
			t.phase = transferPhaseFolders
		case 2: // Manual
			// Instructions shown in viewMethod below the list
		}
	}
	return t, nil
}

func (t *TransferScreen) updatePlan(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		if t.backend != nil {
			t.backend.Close()
		}
		t.phase = transferPhaseMethod
	case key.Matches(msg, tui.Keys.Enter):
		if t.plan != nil {
			t.phase = transferPhaseProgress
			return t, t.startTransfer()
		}
	}
	return t, nil
}

func (t *TransferScreen) initFolderSelection() {
	t.folderOptions = []transferFolder{
		{label: "ROMs", dirName: "roms", selected: true},
		{label: "BIOS", dirName: "bios", selected: false},
		{label: "Saves", dirName: "saves", selected: false},
		{label: "Config", dirName: "config", selected: false},
	}
	t.folderCursor = 0
}

func (t *TransferScreen) updateFolders(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		if t.backend != nil {
			t.backend.Close()
			t.backend = nil
		}
		if t.bulkBackend != nil {
			t.bulkBackend.Close()
			t.bulkBackend = nil
		}
		t.isBulk = false
		t.phase = transferPhaseMethod
	case key.Matches(msg, tui.Keys.Up):
		if t.folderCursor > 0 {
			t.folderCursor--
		}
	case key.Matches(msg, tui.Keys.Down):
		if t.folderCursor < len(t.folderOptions)-1 {
			t.folderCursor++
		}
	case msg.Type == tea.KeySpace:
		t.folderOptions[t.folderCursor].selected = !t.folderOptions[t.folderCursor].selected
	case key.Matches(msg, tui.Keys.Enter):
		if len(t.selectedFolders()) > 0 {
			if t.isBulk {
				t.phase = transferPhaseProgress
				return t, t.startBulkTransfer()
			}
			t.phase = transferPhaseConnect
			return t, t.connect()
		}
	}
	return t, nil
}

func (t *TransferScreen) selectedFolders() []string {
	var folders []string
	for _, f := range t.folderOptions {
		if f.selected {
			folders = append(folders, f.dirName)
		}
	}
	return folders
}

func (t *TransferScreen) selectedLabels() []string {
	var labels []string
	for _, f := range t.folderOptions {
		if f.selected {
			labels = append(labels, f.label)
		}
	}
	return labels
}

func (t *TransferScreen) connect() tea.Cmd {
	backend := t.backend
	return func() tea.Msg {
		err := backend.Connect(context.Background())
		return transferConnectMsg{err: err}
	}
}

func (t *TransferScreen) buildPlan() tea.Cmd {
	backend := t.backend
	cfg := t.cfg
	selected := t.selectedFolders()
	t.planFolderLabels = t.selectedLabels()

	return func() tea.Msg {
		rootDir := ""
		if len(cfg.SourceDirs) > 0 {
			rootDir = cfg.SourceDirs[0]
		}

		var plans []*transfer.TransferPlan
		for _, folder := range selected {
			localDir := filepath.Join(rootDir, folder)
			remoteBase := folder // USB: relative path
			plan, err := transfer.BuildTransferPlan(context.Background(), backend, localDir, remoteBase, cfg.Transfer.SyncMode)
			if err != nil {
				continue
			}
			plans = append(plans, plan)
		}
		return transferPlanMsg{plan: transfer.MergeTransferPlans(plans...)}
	}
}

func (t *TransferScreen) startTransfer() tea.Cmd {
	backend := t.backend
	plan := t.plan
	concurrency := t.cfg.Transfer.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	progressCh := make(chan transfer.TransferProgress, 100)
	t.progressCh = progressCh

	errCh := make(chan error, 1)
	go func() {
		err := transfer.Execute(ctx, backend, plan, concurrency, progressCh)
		errCh <- err
	}()

	return tea.Batch(
		listenTransferProgress(progressCh),
		waitTransferDone(errCh),
	)
}

func (t *TransferScreen) startBulkTransfer() tea.Cmd {
	folders := t.buildFolderMappings()
	backend := t.bulkBackend

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	progressCh := make(chan transfer.TransferProgress, 100)
	t.progressCh = progressCh

	errCh := make(chan error, 1)
	go func() {
		err := backend.TransferFolders(ctx, folders, progressCh)
		errCh <- err
	}()

	return tea.Batch(
		listenTransferProgress(progressCh),
		waitTransferDone(errCh),
	)
}

func (t *TransferScreen) buildFolderMappings() []transfer.FolderMapping {
	rootDir := ""
	if len(t.cfg.SourceDirs) > 0 {
		rootDir = t.cfg.SourceDirs[0]
	}

	var mappings []transfer.FolderMapping
	for _, folder := range t.selectedFolders() {
		mappings = append(mappings, transfer.FolderMapping{
			LocalDir:  filepath.Join(rootDir, folder),
			RemoteDir: path.Join(t.cfg.Device.RootPath, folder),
		})
	}
	return mappings
}

func listenTransferProgress(ch <-chan transfer.TransferProgress) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return nil
		}
		return transferProgressMsg{progress: p}
	}
}

func waitTransferDone(ch <-chan error) tea.Cmd {
	return func() tea.Msg {
		return transferDoneMsg{err: <-ch}
	}
}

func (t *TransferScreen) View() string {
	switch t.phase {
	case transferPhaseMethod:
		return t.viewMethod()
	case transferPhaseFolders:
		return t.viewFolders()
	case transferPhaseConnect:
		return t.viewConnect()
	case transferPhasePlan:
		return t.viewPlan()
	case transferPhaseProgress:
		return t.viewProgress()
	case transferPhaseResults:
		return t.viewResults()
	}
	return ""
}

func (t *TransferScreen) viewMethod() string {
	s := tui.StyleSubtitle.Render("Transfer Method") + "\n\n"

	if t.connectErr != nil {
		s += tui.StyleError.Render("Connection failed: "+t.connectErr.Error()) + "\n\n"
	}
	if t.planErr != nil {
		s += tui.StyleError.Render("Plan failed: "+t.planErr.Error()) + "\n\n"
	}

	for i, method := range t.methods {
		cursor := "  "
		style := tui.StyleNormal
		if i == t.cursor {
			cursor = tui.StyleMenuCursor.String()
			style = tui.StyleSelected
		}
		s += cursor + style.Render(method) + "\n"
	}

	if t.toolErr != "" {
		s += "\n" + tui.StyleError.Render(t.toolErr) + "\n"
	}

	if t.cursor == 2 { // Manual Instructions
		s += "\n" + tui.StyleDim.Render("Manual transfer instructions:") + "\n"
		s += tui.StyleDim.Render("1. Copy the desired folders (roms/, bios/, saves/, config/) to a USB drive") + "\n"
		s += tui.StyleDim.Render("2. Insert into your ReplayOS device") + "\n"
		s += tui.StyleDim.Render("3. Merge folders at the device root") + "\n"
	}

	s += "\n" + tui.StyleDim.Render("enter: select  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (t *TransferScreen) viewFolders() string {
	s := tui.StyleSubtitle.Render("Select Folders") + "\n\n"

	for i, f := range t.folderOptions {
		check := "[ ]"
		if f.selected {
			check = "[x]"
		}
		cursor := "  "
		style := tui.StyleNormal
		if i == t.folderCursor {
			cursor = tui.StyleMenuCursor.String()
			style = tui.StyleSelected
		}
		s += cursor + style.Render(fmt.Sprintf("%s %-10s %s/", check, f.label, f.dirName)) + "\n"
	}

	s += "\n" + tui.StyleDim.Render("space: toggle  enter: confirm  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (t *TransferScreen) viewConnect() string {
	s := tui.StyleSubtitle.Render("Connecting...") + "\n\n"
	s += tui.StyleDim.Render(fmt.Sprintf("Mounting: %s", t.cfg.Transfer.USBPath))
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (t *TransferScreen) viewPlan() string {
	s := tui.StyleSubtitle.Render("Transfer Plan") + "\n\n"

	if t.plan == nil {
		s += tui.StyleDim.Render("Building plan...")
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	if len(t.planFolderLabels) > 0 {
		s += fmt.Sprintf("Folders:           %s\n", strings.Join(t.planFolderLabels, ", "))
	}
	total := len(t.plan.Items)
	s += fmt.Sprintf("Files to transfer: %d\n", total-t.plan.SkipCount)
	if t.plan.SkipCount > 0 {
		s += fmt.Sprintf("Files to skip:     %d (already exist)\n", t.plan.SkipCount)
	}
	s += fmt.Sprintf("Total size:        %s\n", formatBytes(t.plan.TotalSize))

	s += "\n" + tui.StyleDim.Render("enter: start transfer  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (t *TransferScreen) viewProgress() string {
	s := tui.StyleSubtitle.Render("Transferring...") + "\n\n"

	p := t.currentProgress
	if p.TotalFiles > 0 {
		if t.isBulk {
			// rsync: folder-level progress (BytesSent=pct 0-100, FileSize=100)
			s += fmt.Sprintf("Folder %d / %d: %s\n", p.FileIndex+1, p.TotalFiles, p.Filename)
			if p.FileSize > 0 {
				pct := float64(p.BytesSent) / float64(p.FileSize) * 100
				s += renderProgressBar(pct, 40) + "\n"
			}
			s += "\n"
			if p.TotalSize > 0 {
				totalPct := float64(p.TotalSent) / float64(p.TotalSize) * 100
				completedFolders := p.TotalSent / 100
				s += fmt.Sprintf("Overall: %d / %d folders\n", completedFolders, p.TotalFiles)
				s += renderProgressBar(totalPct, 40) + "\n"
			}
		} else {
			// USB: file-level progress with byte counts
			s += fmt.Sprintf("File %d / %d: %s\n", p.FileIndex+1, p.TotalFiles, p.Filename)
			if p.FileSize > 0 {
				pct := float64(p.BytesSent) / float64(p.FileSize) * 100
				s += renderProgressBar(pct, 40) + "\n"
			}
			s += "\n"
			if p.TotalSize > 0 {
				totalPct := float64(p.TotalSent) / float64(p.TotalSize) * 100
				s += fmt.Sprintf("Overall: %s / %s\n", formatBytes(p.TotalSent), formatBytes(p.TotalSize))
				s += renderProgressBar(totalPct, 40) + "\n"
			}
		}
	}

	s += "\n" + tui.StyleDim.Render("esc: cancel transfer")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (t *TransferScreen) viewResults() string {
	s := tui.StyleSubtitle.Render("Transfer Complete") + "\n\n"

	if t.isBulk {
		s += fmt.Sprintf("%s %d folders transferred\n",
			tui.StyleSuccess.Render("OK"), t.itemsTransferred)
	} else {
		s += fmt.Sprintf("%s %d files transferred\n",
			tui.StyleSuccess.Render("OK"), t.itemsTransferred)
	}

	if t.totalErr != nil {
		s += tui.StyleError.Render("Error: "+t.totalErr.Error()) + "\n"
	}

	s += "\n" + tui.StyleDim.Render("enter/esc: done")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func (t *TransferScreen) ShortHelp() []key.Binding {
	return []key.Binding{tui.Keys.Up, tui.Keys.Down, tui.Keys.Enter, tui.Keys.Back}
}
