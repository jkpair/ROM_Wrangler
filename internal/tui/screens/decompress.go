package screens

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/organizer"
	"github.com/kurlmarx/romwrangler/internal/systems"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type decompressPhase int

const (
	decompressPhaseScan decompressPhase = iota
	decompressPhasePreview
	decompressPhaseExtracting
	decompressPhaseResults
)

type decompressScanDoneMsg struct {
	extractable []organizer.ExtractableFile
}

type decompressExtractDoneMsg struct {
	result    *organizer.ExtractResult
	processed []organizer.ExtractableFile
}

type DecompressScreen struct {
	cfg           *config.Config
	width, height int
	phase         decompressPhase

	// Scan
	extractable []organizer.ExtractableFile

	// Progress
	extractProgressCh <-chan extractProgressMsg
	extractProgress   struct {
		current  int
		total    int
		filename string
	}

	// Results
	extractResult    *organizer.ExtractResult
	extractProcessed []organizer.ExtractableFile
}

func NewDecompressScreen(cfg *config.Config, width, height int) *DecompressScreen {
	return &DecompressScreen{
		cfg:    cfg,
		width:  width,
		height: height,
	}
}

func (d *DecompressScreen) Init() tea.Cmd {
	if len(d.cfg.SourceDirs) == 0 {
		return nil
	}
	d.phase = decompressPhaseScan
	dirs := d.cfg.ROMDirs()
	aliases := d.cfg.Aliases
	return func() tea.Msg {
		extractable := organizer.FindExtractable(dirs, aliases)
		return decompressScanDoneMsg{extractable: extractable}
	}
}

func (d *DecompressScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height

	case decompressScanDoneMsg:
		d.extractable = msg.extractable
		d.phase = decompressPhasePreview

	case extractProgressMsg:
		d.extractProgress.current = msg.current
		d.extractProgress.total = msg.total
		d.extractProgress.filename = msg.filename
		return d, listenExtractProgress(d.extractProgressCh)

	case decompressExtractDoneMsg:
		d.extractResult = msg.result
		d.extractProcessed = msg.processed
		d.phase = decompressPhaseResults

	case tea.KeyMsg:
		switch d.phase {
		case decompressPhasePreview:
			return d.updatePreview(msg)
		case decompressPhaseResults:
			if key.Matches(msg, tui.Keys.Back) || key.Matches(msg, tui.Keys.Enter) {
				return d, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		default:
			if key.Matches(msg, tui.Keys.Back) {
				return d, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	}
	return d, nil
}

func (d *DecompressScreen) updatePreview(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		return d, func() tea.Msg { return tui.NavigateBackMsg{} }
	case key.Matches(msg, tui.Keys.Enter):
		if len(d.extractable) > 0 {
			d.phase = decompressPhaseExtracting
			return d, d.startExtraction()
		}
	}
	return d, nil
}

func (d *DecompressScreen) startExtraction() tea.Cmd {
	dirs := d.cfg.ROMDirs()
	aliases := d.cfg.Aliases

	progressCh := make(chan extractProgressMsg, 100)
	d.extractProgressCh = progressCh

	resultCh := make(chan decompressExtractDoneMsg, 1)
	go func() {
		result, processed := organizer.ExtractAll(dirs, aliases, func(current, total int, filename string) {
			progressCh <- extractProgressMsg{
				current:  current,
				total:    total,
				filename: filename,
			}
		})
		close(progressCh)
		resultCh <- decompressExtractDoneMsg{result: result, processed: processed}
	}()

	return tea.Batch(
		listenExtractProgress(progressCh),
		waitDecompressExtractDone(resultCh),
	)
}

func waitDecompressExtractDone(ch <-chan decompressExtractDoneMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func (d *DecompressScreen) View() string {
	switch d.phase {
	case decompressPhaseScan:
		return d.viewScan()
	case decompressPhasePreview:
		return d.viewPreview()
	case decompressPhaseExtracting:
		return d.viewExtracting()
	case decompressPhaseResults:
		return d.viewResults()
	}
	return ""
}

func (d *DecompressScreen) viewScan() string {
	s := tui.StyleSubtitle.Render("Decompress Files") + "\n\n"

	if len(d.cfg.SourceDirs) == 0 {
		s += tui.StyleWarning.Render("No root directory configured.") + "\n\n"
		s += tui.StyleDim.Render("Go to Settings to set a root directory.")
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	s += tui.StyleDim.Render("Scanning for compressed files...")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (d *DecompressScreen) viewPreview() string {
	s := tui.StyleSubtitle.Render("Decompress Files") + "\n\n"

	if len(d.extractable) == 0 {
		s += tui.StyleDim.Render("No compressed archives found.") + "\n"
		s += "\n" + tui.StyleDim.Render("esc: back")
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	// Group by system
	bySystem := make(map[systems.SystemID]int)
	for _, f := range d.extractable {
		bySystem[f.System]++
	}

	var sysIDs []systems.SystemID
	for id := range bySystem {
		sysIDs = append(sysIDs, id)
	}
	sort.Slice(sysIDs, func(i, j int) bool {
		return string(sysIDs[i]) < string(sysIDs[j])
	})

	s += fmt.Sprintf("%d archives to extract:\n\n", len(d.extractable))
	for _, sysID := range sysIDs {
		info, _ := systems.GetSystem(sysID)
		s += fmt.Sprintf("  %s: %d archives\n", info.DisplayName, bySystem[sysID])
	}

	s += "\nEach archive will be extracted into its own subfolder\n"
	s += "to prevent track file name collisions.\n"

	s += "\n" + tui.StyleDim.Render("enter: extract  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (d *DecompressScreen) viewExtracting() string {
	s := tui.StyleSubtitle.Render("Extracting Archives...") + "\n\n"

	p := d.extractProgress
	if p.total > 0 {
		pct := float64(p.current) / float64(p.total) * 100
		s += fmt.Sprintf("Progress: %d / %d\n", p.current, p.total)
		s += tui.StyleDim.Render(p.filename) + "\n"
		s += renderProgressBar(pct, 40) + "\n"
	} else {
		s += tui.StyleDim.Render("Preparing extraction...")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (d *DecompressScreen) viewResults() string {
	s := tui.StyleSubtitle.Render("Extraction Complete") + "\n\n"

	if d.extractResult != nil {
		s += fmt.Sprintf("%s %d archives extracted (%d files created)\n",
			tui.StyleSuccess.Render("OK"),
			d.extractResult.Extracted,
			d.extractResult.FilesCreated)
		if len(d.extractResult.Errors) > 0 {
			s += fmt.Sprintf("\n%s %d errors:\n",
				tui.StyleError.Render("!"),
				len(d.extractResult.Errors))
			for _, err := range d.extractResult.Errors {
				s += "  " + tui.StyleDim.Render(err.Error()) + "\n"
			}
		}
	}

	s += "\n" + tui.StyleDim.Render("enter/esc: done")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (d *DecompressScreen) ShortHelp() []key.Binding {
	return []key.Binding{tui.Keys.Enter, tui.Keys.Back}
}
