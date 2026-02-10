package screens

import (
	"fmt"

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

type TransferScreen struct {
	cfg           *config.Config
	width, height int
	phase         transferPhase

	// Method selection
	methods []string
	cursor  int

	// Connection
	backend    transfer.TransferBackend
	connectErr error

	// Plan
	plan    *transfer.TransferPlan
	planErr error

	// Progress
	progressCh      <-chan transfer.TransferProgress
	currentProgress transfer.TransferProgress

	// Results
	filesTransferred int
	totalErr         error
}

func NewTransferScreen(cfg *config.Config, width, height int) *TransferScreen {
	return &TransferScreen{
		cfg:    cfg,
		width:  width,
		height: height,
		methods: []string{"SFTP (Network)", "USB / SD Card", "Manual Instructions"},
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
			t.filesTransferred++
		}
		// Keep listening for more progress
		return t, listenTransferProgress(t.progressCh)

	case transferDoneMsg:
		t.totalErr = msg.err
		t.phase = transferPhaseResults

	case tea.KeyMsg:
		switch t.phase {
		case transferPhaseMethod:
			return t.updateMethod(msg)
		case transferPhasePlan:
			return t.updatePlan(msg)
		case transferPhaseResults:
			if key.Matches(msg, tui.Keys.Back) || key.Matches(msg, tui.Keys.Enter) {
				if t.backend != nil {
					t.backend.Close()
				}
				return t, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		default:
			if key.Matches(msg, tui.Keys.Back) {
				if t.backend != nil {
					t.backend.Close()
				}
				return t, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	}
	return t, nil
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
		switch t.cursor {
		case 0: // SFTP
			t.backend = transfer.NewSFTPBackend(
				t.cfg.Device.Host,
				t.cfg.Device.Port,
				t.cfg.Device.User,
				t.cfg.Device.Password,
			)
			t.phase = transferPhaseConnect
			return t, t.connect()
		case 1: // USB
			t.backend = transfer.NewUSBBackend(t.cfg.Transfer.USBPath)
			t.phase = transferPhaseConnect
			return t, t.connect()
		case 2: // Manual
			// Stay on this screen showing instructions
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

func (t *TransferScreen) connect() tea.Cmd {
	backend := t.backend
	return func() tea.Msg {
		err := backend.Connect()
		return transferConnectMsg{err: err}
	}
}

func (t *TransferScreen) buildPlan() tea.Cmd {
	backend := t.backend
	cfg := t.cfg
	return func() tea.Msg {
		var sourceDir string
		if len(cfg.SourceDirs) > 0 {
			sourceDir = cfg.SourceDirs[0]
		}
		plan, err := transfer.BuildTransferPlan(backend, sourceDir, cfg.Device.ROMPath, cfg.Transfer.SyncMode)
		return transferPlanMsg{plan: plan, err: err}
	}
}

func (t *TransferScreen) startTransfer() tea.Cmd {
	backend := t.backend
	plan := t.plan

	progressCh := make(chan transfer.TransferProgress, 100)
	t.progressCh = progressCh

	errCh := make(chan error, 1)
	go func() {
		err := transfer.Execute(backend, plan, progressCh)
		errCh <- err
	}()

	return tea.Batch(
		listenTransferProgress(progressCh),
		waitTransferDone(errCh),
	)
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
		err := <-ch
		return transferDoneMsg{err: err}
	}
}

func (t *TransferScreen) View() string {
	switch t.phase {
	case transferPhaseMethod:
		return t.viewMethod()
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

	if t.cursor == 2 {
		s += "\n" + tui.StyleDim.Render("Manual transfer instructions:") + "\n"
		s += tui.StyleDim.Render("1. Copy the 'organized' folder to a USB drive or SD card") + "\n"
		s += tui.StyleDim.Render("2. Insert into your ReplayOS device") + "\n"
		s += tui.StyleDim.Render("3. Copy folders to /roms/ on the device") + "\n"
	}

	s += "\n" + tui.StyleDim.Render("enter: select  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (t *TransferScreen) viewConnect() string {
	s := tui.StyleSubtitle.Render("Connecting...") + "\n\n"
	s += tui.StyleDim.Render(fmt.Sprintf("Host: %s:%d", t.cfg.Device.Host, t.cfg.Device.Port))
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (t *TransferScreen) viewPlan() string {
	s := tui.StyleSubtitle.Render("Transfer Plan") + "\n\n"

	if t.plan == nil {
		s += tui.StyleDim.Render("Building plan...")
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
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

	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (t *TransferScreen) viewResults() string {
	s := tui.StyleSubtitle.Render("Transfer Complete") + "\n\n"

	s += fmt.Sprintf("%s %d files transferred\n",
		tui.StyleSuccess.Render("OK"), t.filesTransferred)

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
