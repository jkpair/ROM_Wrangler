package screens

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/organizer"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type biosPhase int

const (
	biosPhaseScan biosPhase = iota
	biosPhaseOverview
	biosPhaseConfirm
	biosPhaseDone
)

type biosScanDoneMsg struct {
	statuses []organizer.BIOSFolderStatus
	matches  []organizer.BIOSFileMatch
}

type biosSetupDoneMsg struct {
	foldersCreated int
	filesMoved     int
	errs           []error
}

type BIOSSetupScreen struct {
	cfg           *config.Config
	width, height int
	phase         biosPhase

	biosDir  string
	statuses []organizer.BIOSFolderStatus
	matches  []organizer.BIOSFileMatch

	// Results
	foldersCreated int
	filesMoved     int
	errs           []error

	// UI
	scroll int
}

func NewBIOSSetupScreen(cfg *config.Config, width, height int) *BIOSSetupScreen {
	s := &BIOSSetupScreen{
		cfg:    cfg,
		width:  width,
		height: height,
	}

	if len(cfg.SourceDirs) > 0 {
		s.biosDir = filepath.Join(cfg.SourceDirs[0], "bios")
	}

	return s
}

func (b *BIOSSetupScreen) Init() tea.Cmd {
	if b.biosDir == "" {
		return nil
	}
	b.phase = biosPhaseScan
	biosDir := b.biosDir
	romDirs := b.cfg.ROMDirs()
	return func() tea.Msg {
		statuses := organizer.CheckBIOSFolders(biosDir)
		var matches []organizer.BIOSFileMatch
		for _, dir := range romDirs {
			matches = append(matches, organizer.ScanBIOSFiles(dir)...)
		}
		return biosScanDoneMsg{statuses: statuses, matches: matches}
	}
}

func (b *BIOSSetupScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height

	case biosScanDoneMsg:
		b.statuses = msg.statuses
		b.matches = msg.matches
		b.phase = biosPhaseOverview

	case biosSetupDoneMsg:
		b.foldersCreated = msg.foldersCreated
		b.filesMoved = msg.filesMoved
		b.errs = msg.errs
		b.phase = biosPhaseDone
		b.statuses = organizer.CheckBIOSFolders(b.biosDir)

	case tea.KeyMsg:
		switch b.phase {
		case biosPhaseOverview:
			return b.updateOverview(msg)
		case biosPhaseConfirm:
			return b.updateConfirm(msg)
		case biosPhaseDone:
			if key.Matches(msg, tui.Keys.Back) || key.Matches(msg, tui.Keys.Enter) {
				return b, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		default:
			if key.Matches(msg, tui.Keys.Back) {
				return b, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	}
	return b, nil
}

func (b *BIOSSetupScreen) updateOverview(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		return b, func() tea.Msg { return tui.NavigateBackMsg{} }
	case key.Matches(msg, tui.Keys.Up):
		if b.scroll > 0 {
			b.scroll--
		}
	case key.Matches(msg, tui.Keys.Down):
		maxScroll := len(b.statuses) - (b.height - 16)
		if maxScroll < 0 {
			maxScroll = 0
		}
		if b.scroll < maxScroll {
			b.scroll++
		}
	case key.Matches(msg, tui.Keys.Enter):
		b.phase = biosPhaseConfirm
	}
	return b, nil
}

func (b *BIOSSetupScreen) updateConfirm(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		b.phase = biosPhaseOverview
	case key.Matches(msg, tui.Keys.Enter):
		biosDir := b.biosDir
		matches := b.matches
		return b, func() tea.Msg {
			foldersCreated, errs := organizer.GenerateBIOSFolders(biosDir)
			filesMoved, moveErrs := organizer.OrganizeBIOSFiles(matches, biosDir)
			errs = append(errs, moveErrs...)
			return biosSetupDoneMsg{
				foldersCreated: foldersCreated,
				filesMoved:     filesMoved,
				errs:           errs,
			}
		}
	}
	return b, nil
}

func (b *BIOSSetupScreen) View() string {
	if b.biosDir == "" {
		content := tui.StyleSubtitle.Render("Setup BIOS Folders") + "\n\n"
		content += tui.StyleWarning.Render("No root directory configured.") + "\n\n"
		content += tui.StyleDim.Render("Go to Settings and set a root directory first.")
		content += "\n\n" + tui.StyleDim.Render("esc: back")
		return lipgloss.NewStyle().Padding(1, 2).Render(content)
	}

	switch b.phase {
	case biosPhaseScan:
		return b.viewScan()
	case biosPhaseOverview:
		return b.viewOverview()
	case biosPhaseConfirm:
		return b.viewConfirm()
	case biosPhaseDone:
		return b.viewDone()
	}
	return ""
}

func (b *BIOSSetupScreen) viewScan() string {
	s := tui.StyleSubtitle.Render("Setup BIOS Folders") + "\n\n"
	s += tui.StyleDim.Render("Scanning for BIOS folders and files...")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (b *BIOSSetupScreen) viewOverview() string {
	s := tui.StyleSubtitle.Render("Setup BIOS Folders") + "\n\n"
	s += "BIOS directory: " + tui.StyleNormal.Render(b.biosDir) + "\n\n"

	var existing, missing int
	for _, st := range b.statuses {
		if st.Exists {
			existing++
		} else {
			missing++
		}
	}

	s += fmt.Sprintf("  %s %d folders exist\n",
		tui.StyleSuccess.Render(fmt.Sprintf("%d", existing)),
		existing)
	s += fmt.Sprintf("  %s %d folders missing\n",
		tui.StyleWarning.Render(fmt.Sprintf("%d", missing)),
		missing)

	if len(b.matches) > 0 {
		s += fmt.Sprintf("  %s %d BIOS files detected in source dirs\n",
			tui.StyleAccent.Render(fmt.Sprintf("%d", len(b.matches))),
			len(b.matches))
	}
	s += "\n"

	// Show folder list with scroll
	maxVisible := b.height - 18
	if maxVisible < 5 {
		maxVisible = 5
	}

	end := b.scroll + maxVisible
	if end > len(b.statuses) {
		end = len(b.statuses)
	}

	for i := b.scroll; i < end; i++ {
		st := b.statuses[i]
		var icon string
		if st.Exists {
			if st.FileCount > 0 {
				icon = tui.StyleSuccess.Render("OK")
				s += fmt.Sprintf("  %s  %-40s (%d files)\n",
					icon, st.Folder, st.FileCount)
			} else {
				icon = tui.StyleDim.Render("--")
				s += fmt.Sprintf("  %s  %-40s (empty)\n",
					icon, st.Folder)
			}
		} else {
			icon = tui.StyleWarning.Render("++")
			s += fmt.Sprintf("  %s  %s\n", icon, st.Folder)
		}
	}

	if len(b.statuses) > maxVisible {
		s += tui.StyleDim.Render(fmt.Sprintf("\n  scroll: %d-%d of %d",
			b.scroll+1, end, len(b.statuses)))
	}

	if missing > 0 || len(b.matches) > 0 {
		s += "\n\n" + tui.StyleDim.Render("enter: create folders & organize files  esc: back")
	} else {
		s += "\n\n" + tui.StyleSuccess.Render("All BIOS folders exist!") + "  " + tui.StyleDim.Render("esc: back")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (b *BIOSSetupScreen) viewConfirm() string {
	var missing int
	for _, st := range b.statuses {
		if !st.Exists {
			missing++
		}
	}

	s := tui.StyleSubtitle.Render("Setup BIOS Folders?") + "\n\n"

	if missing > 0 {
		s += fmt.Sprintf("This will create %d BIOS folders in:\n", missing)
		s += tui.StyleNormal.Render(b.biosDir) + "\n\n"
	}

	if len(b.matches) > 0 {
		s += fmt.Sprintf("This will organize %d detected BIOS files.\n\n", len(b.matches))
	}

	s += tui.StyleDim.Render("enter: yes, proceed  esc: cancel")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (b *BIOSSetupScreen) viewDone() string {
	s := tui.StyleSubtitle.Render("BIOS Setup Complete") + "\n\n"

	s += fmt.Sprintf("%s %d folders created\n",
		tui.StyleSuccess.Render("OK"), b.foldersCreated)

	if b.filesMoved > 0 {
		s += fmt.Sprintf("%s %d BIOS files organized\n",
			tui.StyleSuccess.Render("OK"), b.filesMoved)
	}

	if len(b.errs) > 0 {
		s += fmt.Sprintf("\n%s %d errors:\n",
			tui.StyleError.Render("!"), len(b.errs))
		for _, err := range b.errs {
			s += "  " + tui.StyleDim.Render(err.Error()) + "\n"
		}
	}

	s += "\n" + tui.StyleDim.Render("Your BIOS folder structure is ready.") + "\n"
	s += tui.StyleDim.Render("Copy BIOS files into the appropriate folders,") + "\n"
	s += tui.StyleDim.Render("then transfer them to your device.") + "\n"

	s += "\n" + tui.StyleDim.Render("enter/esc: done")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (b *BIOSSetupScreen) ShortHelp() []key.Binding {
	return []key.Binding{tui.Keys.Up, tui.Keys.Down, tui.Keys.Enter, tui.Keys.Back}
}
