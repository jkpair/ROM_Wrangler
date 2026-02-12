package screens

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/multidisc"
	"github.com/kurlmarx/romwrangler/internal/systems"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type m3uPhase int

const (
	m3uPhaseScan m3uPhase = iota
	m3uPhasePreview
	m3uPhaseGenerating
	m3uPhaseResults
)

type m3uScanDoneMsg struct {
	sets []m3uSystemGroup
}

type m3uGenerateDoneMsg struct {
	created int
	skipped int
	errors  []error
}

// m3uSystemGroup holds the detected multi-disc sets for one system directory.
type m3uSystemGroup struct {
	systemID systems.SystemID
	dir      string
	sets     []multidisc.MultiDiscSet
}

type M3UScreen struct {
	cfg           *config.Config
	width, height int
	phase         m3uPhase

	// Scan results
	groups   []m3uSystemGroup
	totalSets int

	// Preview scroll
	scrollOffset int

	// Results
	created int
	skipped int
	errors  []error
}

func NewM3UScreen(cfg *config.Config, width, height int) *M3UScreen {
	return &M3UScreen{
		cfg:    cfg,
		width:  width,
		height: height,
	}
}

func (m *M3UScreen) Init() tea.Cmd {
	if len(m.cfg.SourceDirs) == 0 {
		return nil
	}
	m.phase = m3uPhaseScan
	dirs := m.cfg.ROMDirs()
	aliases := m.cfg.Aliases
	return func() tea.Msg {
		return m3uScanDoneMsg{sets: scanForMultiDisc(dirs, aliases)}
	}
}

func (m *M3UScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case m3uScanDoneMsg:
		m.groups = msg.sets
		m.totalSets = 0
		for _, g := range m.groups {
			m.totalSets += len(g.sets)
		}
		m.phase = m3uPhasePreview

	case m3uGenerateDoneMsg:
		m.created = msg.created
		m.skipped = msg.skipped
		m.errors = msg.errors
		m.phase = m3uPhaseResults

	case tea.KeyMsg:
		switch m.phase {
		case m3uPhasePreview:
			return m.updatePreview(msg)
		case m3uPhaseResults:
			if key.Matches(msg, tui.Keys.Back) || key.Matches(msg, tui.Keys.Enter) {
				return m, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		default:
			if key.Matches(msg, tui.Keys.Back) {
				return m, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	}
	return m, nil
}

func (m *M3UScreen) updatePreview(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		return m, func() tea.Msg { return tui.NavigateBackMsg{} }
	case key.Matches(msg, tui.Keys.Up):
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
	case key.Matches(msg, tui.Keys.Down):
		m.scrollOffset++
	case key.Matches(msg, tui.Keys.Enter):
		if m.totalSets > 0 {
			m.phase = m3uPhaseGenerating
			return m, m.startGeneration()
		}
	}
	return m, nil
}

func (m *M3UScreen) startGeneration() tea.Cmd {
	groups := m.groups
	return func() tea.Msg {
		var created, skipped int
		var errors []error

		for _, g := range groups {
			for _, set := range g.sets {
				m3uPath := filepath.Join(g.dir, set.BaseName+".m3u")

				// Skip if M3U already exists
				if _, err := os.Stat(m3uPath); err == nil {
					skipped++
					continue
				}

				if _, err := multidisc.WriteM3U(g.dir, set, "", false); err != nil {
					errors = append(errors, fmt.Errorf("%s: %w", set.BaseName, err))
				} else {
					created++
				}
			}
		}

		return m3uGenerateDoneMsg{created: created, skipped: skipped, errors: errors}
	}
}

func (m *M3UScreen) View() string {
	switch m.phase {
	case m3uPhaseScan:
		return m.viewScan()
	case m3uPhasePreview:
		return m.viewPreview()
	case m3uPhaseGenerating:
		return m.viewGenerating()
	case m3uPhaseResults:
		return m.viewResults()
	}
	return ""
}

func (m *M3UScreen) viewScan() string {
	s := tui.StyleSubtitle.Render("Generate M3U Files") + "\n\n"

	if len(m.cfg.SourceDirs) == 0 {
		s += tui.StyleWarning.Render("No root directory configured.") + "\n\n"
		s += tui.StyleDim.Render("Go to Settings to set a root directory.")
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	s += tui.StyleDim.Render("Scanning for multi-disc games...")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *M3UScreen) viewPreview() string {
	s := tui.StyleSubtitle.Render("Generate M3U Files") + "\n\n"

	if m.totalSets == 0 {
		s += tui.StyleDim.Render("No multi-disc games found.") + "\n"
		s += "\n" + tui.StyleDim.Render("esc: back")
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	s += fmt.Sprintf("Found %d multi-disc games:\n\n", m.totalSets)

	// Build display lines
	var lines []string
	for _, g := range m.groups {
		info, ok := systems.GetSystem(g.systemID)
		displayName := string(g.systemID)
		if ok {
			displayName = info.DisplayName
		}
		lines = append(lines, tui.StyleSubtitle.Render(displayName))
		for _, set := range g.sets {
			lines = append(lines, fmt.Sprintf("  %s (%d discs)", set.BaseName, len(set.Files)))
		}
		lines = append(lines, "")
	}

	// Scrollable area — reserve space for header + footer
	maxVisible := m.height - 12
	if maxVisible < 5 {
		maxVisible = 5
	}
	if m.scrollOffset > len(lines)-maxVisible {
		m.scrollOffset = len(lines) - maxVisible
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	end := m.scrollOffset + maxVisible
	if end > len(lines) {
		end = len(lines)
	}
	for _, line := range lines[m.scrollOffset:end] {
		s += line + "\n"
	}

	if len(lines) > maxVisible {
		s += tui.StyleDim.Render(fmt.Sprintf("(%d more, use arrows to scroll)", len(lines)-maxVisible)) + "\n"
	}

	s += "\n" + tui.StyleDim.Render("enter: generate  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *M3UScreen) viewGenerating() string {
	s := tui.StyleSubtitle.Render("Generating M3U Files...") + "\n\n"
	s += tui.StyleDim.Render("Writing playlists...")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *M3UScreen) viewResults() string {
	s := tui.StyleSubtitle.Render("M3U Generation Complete") + "\n\n"

	if m.created > 0 {
		s += fmt.Sprintf("%s %d M3U files created\n", tui.StyleSuccess.Render("OK"), m.created)
	}
	if m.skipped > 0 {
		s += fmt.Sprintf("%s %d already exist (skipped)\n", tui.StyleDim.Render("--"), m.skipped)
	}
	if m.created == 0 && m.skipped > 0 {
		s += "\n" + tui.StyleDim.Render("All M3U files already exist. Delete existing files to regenerate.") + "\n"
	}
	if len(m.errors) > 0 {
		s += fmt.Sprintf("\n%s %d errors:\n", tui.StyleError.Render("!"), len(m.errors))
		for _, err := range m.errors {
			s += "  " + tui.StyleDim.Render(err.Error()) + "\n"
		}
	}

	s += "\n" + tui.StyleDim.Render("enter/esc: done")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *M3UScreen) ShortHelp() []key.Binding {
	return []key.Binding{tui.Keys.Enter, tui.Keys.Back}
}

// scanForMultiDisc walks ROM directories and finds multi-disc sets per system.
func scanForMultiDisc(romDirs []string, aliases map[string]string) []m3uSystemGroup {
	var groups []m3uSystemGroup

	for _, dir := range romDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == "_archive" {
				continue
			}

			systemID, ok := config.ResolveAlias(entry.Name(), aliases)
			if !ok {
				continue
			}

			info, ok := systems.GetSystem(systemID)
			if !ok || !info.IsDiscBased {
				continue
			}

			sysDir := filepath.Join(dir, entry.Name())
			var discFiles []string

			filepath.Walk(sysDir, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return nil
				}
				ext := strings.ToLower(filepath.Ext(path))
				// Only consider disc image formats, not .m3u itself
				if ext == ".m3u" {
					return nil
				}
				if systems.IsValidFormat(systemID, ext) && multidisc.HasDiscPattern(fi.Name()) {
					discFiles = append(discFiles, path)
				}
				return nil
			})

			if len(discFiles) == 0 {
				continue
			}

			sets, _ := multidisc.DetectSets(discFiles)
			if len(sets) == 0 {
				continue
			}

			groups = append(groups, m3uSystemGroup{
				systemID: systemID,
				dir:      sysDir,
				sets:     sets,
			})
		}
	}

	// Sort by system name
	sort.Slice(groups, func(i, j int) bool {
		return string(groups[i].systemID) < string(groups[j].systemID)
	})

	return groups
}
