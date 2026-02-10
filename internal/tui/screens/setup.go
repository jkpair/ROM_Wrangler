package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/organizer"
	"github.com/kurlmarx/romwrangler/internal/systems"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type setupPhase int

const (
	setupPhaseOverview setupPhase = iota
	setupPhaseConfirm
	setupPhaseDone
)

type setupDoneMsg struct {
	created int
	errs    []error
}

type SetupScreen struct {
	cfg           *config.Config
	width, height int
	phase         setupPhase

	baseDir  string
	statuses []organizer.FolderStatus

	// Results
	created int
	errs    []error

	// UI
	scroll int
}

func NewSetupScreen(cfg *config.Config, width, height int) *SetupScreen {
	s := &SetupScreen{
		cfg:    cfg,
		width:  width,
		height: height,
	}

	if len(cfg.SourceDirs) > 0 {
		s.baseDir = cfg.SourceDirs[0]
	}

	return s
}

func (s *SetupScreen) Init() tea.Cmd {
	if s.baseDir == "" {
		return nil
	}
	s.statuses = organizer.CheckFolders(s.baseDir)
	return nil
}

func (s *SetupScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height

	case setupDoneMsg:
		s.created = msg.created
		s.errs = msg.errs
		s.phase = setupPhaseDone
		// Refresh folder status
		s.statuses = organizer.CheckFolders(s.baseDir)

	case tea.KeyMsg:
		switch s.phase {
		case setupPhaseOverview:
			return s.updateOverview(msg)
		case setupPhaseConfirm:
			return s.updateConfirm(msg)
		case setupPhaseDone:
			if key.Matches(msg, tui.Keys.Back) || key.Matches(msg, tui.Keys.Enter) {
				return s, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	}
	return s, nil
}

func (s *SetupScreen) updateOverview(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		return s, func() tea.Msg { return tui.NavigateBackMsg{} }
	case key.Matches(msg, tui.Keys.Up):
		if s.scroll > 0 {
			s.scroll--
		}
	case key.Matches(msg, tui.Keys.Down):
		maxScroll := len(s.statuses) - (s.height - 12)
		if maxScroll < 0 {
			maxScroll = 0
		}
		if s.scroll < maxScroll {
			s.scroll++
		}
	case key.Matches(msg, tui.Keys.Enter):
		s.phase = setupPhaseConfirm
	}
	return s, nil
}

func (s *SetupScreen) updateConfirm(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		s.phase = setupPhaseOverview
	case key.Matches(msg, tui.Keys.Enter):
		baseDir := s.baseDir
		return s, func() tea.Msg {
			created, errs := organizer.GenerateAllFolders(baseDir)
			return setupDoneMsg{created: created, errs: errs}
		}
	}
	return s, nil
}

func (s *SetupScreen) View() string {
	if s.baseDir == "" {
		content := tui.StyleSubtitle.Render("Setup Folders") + "\n\n"
		content += tui.StyleWarning.Render("No source directory configured.") + "\n\n"
		content += tui.StyleDim.Render("Go to Settings and set a source directory first.")
		content += "\n\n" + tui.StyleDim.Render("esc: back")
		return lipgloss.NewStyle().Padding(1, 2).Render(content)
	}

	switch s.phase {
	case setupPhaseOverview:
		return s.viewOverview()
	case setupPhaseConfirm:
		return s.viewConfirm()
	case setupPhaseDone:
		return s.viewDone()
	}
	return ""
}

func (s *SetupScreen) viewOverview() string {
	content := tui.StyleSubtitle.Render("Setup Folders") + "\n\n"
	content += "Source directory: " + tui.StyleNormal.Render(s.baseDir) + "\n\n"

	var existing, missing int
	for _, st := range s.statuses {
		if st.Exists {
			existing++
		} else {
			missing++
		}
	}

	content += fmt.Sprintf("  %s %d folders exist\n",
		tui.StyleSuccess.Render(fmt.Sprintf("%d", existing)),
		existing)
	content += fmt.Sprintf("  %s %d folders missing\n\n",
		tui.StyleWarning.Render(fmt.Sprintf("%d", missing)),
		missing)

	// Show folder list with scroll
	maxVisible := s.height - 14
	if maxVisible < 5 {
		maxVisible = 5
	}

	end := s.scroll + maxVisible
	if end > len(s.statuses) {
		end = len(s.statuses)
	}

	for i := s.scroll; i < end; i++ {
		st := s.statuses[i]
		info, _ := systems.GetSystem(st.System)

		var icon string
		if st.Exists {
			if st.FileCount > 0 {
				icon = tui.StyleSuccess.Render("OK")
				content += fmt.Sprintf("  %s  %-25s %s (%d files)\n",
					icon, st.Folder,
					tui.StyleDim.Render(info.DisplayName),
					st.FileCount)
			} else {
				icon = tui.StyleDim.Render("--")
				content += fmt.Sprintf("  %s  %-25s %s (empty)\n",
					icon, st.Folder,
					tui.StyleDim.Render(info.DisplayName))
			}
		} else {
			icon = tui.StyleWarning.Render("++")
			content += fmt.Sprintf("  %s  %-25s %s\n",
				icon, st.Folder,
				tui.StyleDim.Render(info.DisplayName))
		}
	}

	if len(s.statuses) > maxVisible {
		content += tui.StyleDim.Render(fmt.Sprintf("\n  scroll: %d-%d of %d", s.scroll+1, end, len(s.statuses)))
	}

	if missing > 0 {
		content += "\n\n" + tui.StyleDim.Render("enter: create missing folders  esc: back")
	} else {
		content += "\n\n" + tui.StyleSuccess.Render("All folders exist!") + "  " + tui.StyleDim.Render("esc: back")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (s *SetupScreen) viewConfirm() string {
	var missing int
	for _, st := range s.statuses {
		if !st.Exists {
			missing++
		}
	}

	content := tui.StyleSubtitle.Render("Create Folders?") + "\n\n"
	content += fmt.Sprintf("This will create %d system folders in:\n", missing)
	content += tui.StyleNormal.Render(s.baseDir) + "\n\n"
	content += tui.StyleDim.Render("enter: yes, create them  esc: cancel")

	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (s *SetupScreen) viewDone() string {
	content := tui.StyleSubtitle.Render("Setup Complete") + "\n\n"

	content += fmt.Sprintf("%s %d folders created\n",
		tui.StyleSuccess.Render("OK"), s.created)

	if len(s.errs) > 0 {
		content += fmt.Sprintf("\n%s %d errors:\n",
			tui.StyleError.Render("!"), len(s.errs))
		for _, err := range s.errs {
			content += "  " + tui.StyleDim.Render(err.Error()) + "\n"
		}
	}

	content += "\n" + tui.StyleDim.Render("Your ROM folder structure is ready.") + "\n"
	content += tui.StyleDim.Render("Drop ROM files into the appropriate system folders,") + "\n"
	content += tui.StyleDim.Render("then use Manage ROMs to scan and organize them.") + "\n"

	content += "\n" + tui.StyleDim.Render("enter/esc: done")
	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (s *SetupScreen) ShortHelp() []key.Binding {
	return []key.Binding{tui.Keys.Up, tui.Keys.Down, tui.Keys.Enter, tui.Keys.Back}
}
