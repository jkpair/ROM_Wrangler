package screens

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/converter"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type convertPhase int

const (
	convertPhaseSelect convertPhase = iota
	convertPhaseConfirm
	convertPhaseRunning
	convertPhaseResults
)

type convertInitDoneMsg struct {
	chdmanPath string
	files      []string
	err        error
}

type convertProgressMsg struct {
	progress converter.BatchProgress
}

type convertDoneMsg struct {
	results []converter.ConvertResult
}

type ConvertScreen struct {
	cfg           *config.Config
	width, height int

	phase      convertPhase
	chdmanPath string
	chdmanErr  error

	// File selection
	files    []string
	selected map[int]bool
	cursor   int

	// Progress
	progressCh <-chan converter.BatchProgress
	currentFile string
	currentPct  float64
	filesDone   int
	totalFiles  int

	// Results
	results []converter.ConvertResult

	cancel context.CancelFunc
}

func NewConvertScreen(cfg *config.Config, width, height int) *ConvertScreen {
	return &ConvertScreen{
		cfg:      cfg,
		width:    width,
		height:   height,
		selected: make(map[int]bool),
	}
}

func (c *ConvertScreen) Init() tea.Cmd {
	dirs := c.cfg.SourceDirs
	chdmanCfg := c.cfg.ChdmanPath
	return func() tea.Msg {
		path, err := converter.FindChdman(chdmanCfg)
		if err != nil {
			return convertInitDoneMsg{err: err}
		}

		var files []string
		for _, dir := range dirs {
			filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				if converter.IsConvertible(p) {
					files = append(files, p)
				}
				return nil
			})
		}
		return convertInitDoneMsg{chdmanPath: path, files: files}
	}
}

func (c *ConvertScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height

	case convertInitDoneMsg:
		if msg.err != nil {
			c.chdmanErr = msg.err
		} else {
			c.chdmanPath = msg.chdmanPath
			c.files = msg.files
		}

	case convertProgressMsg:
		p := msg.progress
		c.currentFile = p.Filename
		c.currentPct = p.Percent
		if p.Done {
			c.filesDone++
		}
		// Keep listening for more progress
		return c, listenConvertProgress(c.progressCh)

	case convertDoneMsg:
		c.results = msg.results
		c.phase = convertPhaseResults

	case tea.KeyMsg:
		switch c.phase {
		case convertPhaseSelect:
			return c.updateSelect(msg)
		case convertPhaseConfirm:
			return c.updateConfirm(msg)
		case convertPhaseRunning:
			if key.Matches(msg, tui.Keys.Back) {
				if c.cancel != nil {
					c.cancel()
				}
				c.phase = convertPhaseSelect
			}
		case convertPhaseResults:
			if key.Matches(msg, tui.Keys.Back) || key.Matches(msg, tui.Keys.Enter) {
				return c, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	}
	return c, nil
}

func (c *ConvertScreen) updateSelect(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		return c, func() tea.Msg { return tui.NavigateBackMsg{} }
	case key.Matches(msg, tui.Keys.Up):
		if c.cursor > 0 {
			c.cursor--
		}
	case key.Matches(msg, tui.Keys.Down):
		if c.cursor < len(c.files)-1 {
			c.cursor++
		}
	case key.Matches(msg, tui.Keys.Space):
		if c.selected[c.cursor] {
			delete(c.selected, c.cursor)
		} else {
			c.selected[c.cursor] = true
		}
	case key.Matches(msg, tui.Keys.Select):
		if len(c.selected) == len(c.files) {
			c.selected = make(map[int]bool)
		} else {
			for i := range c.files {
				c.selected[i] = true
			}
		}
	case key.Matches(msg, tui.Keys.Enter):
		if len(c.selected) > 0 {
			c.phase = convertPhaseConfirm
		}
	}
	return c, nil
}

func (c *ConvertScreen) updateConfirm(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		c.phase = convertPhaseSelect
	case key.Matches(msg, tui.Keys.Enter):
		c.phase = convertPhaseRunning
		return c, c.startConversion()
	}
	return c, nil
}

func (c *ConvertScreen) startConversion() tea.Cmd {
	var selected []string
	for i := range c.files {
		if c.selected[i] {
			selected = append(selected, c.files[i])
		}
	}
	c.totalFiles = len(selected)
	c.filesDone = 0

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	progressCh := make(chan converter.BatchProgress, 100)
	c.progressCh = progressCh

	chdmanPath := c.chdmanPath

	// Launch the conversion in a goroutine — it writes to progressCh and
	// closes it when done, then we pick up the results via convertDoneMsg.
	resultCh := make(chan []converter.ConvertResult, 1)
	go func() {
		results := converter.BatchConvert(ctx, chdmanPath, selected, 1, progressCh)
		resultCh <- results
	}()

	// Return two concurrent commands:
	// 1. Listen for the first progress message
	// 2. Wait for the final results
	return tea.Batch(
		listenConvertProgress(progressCh),
		waitConvertDone(resultCh),
	)
}

// listenConvertProgress reads one progress message from the channel and
// returns it as a tea.Msg. The Update handler calls this again after each
// message to keep the loop going. When the channel closes, returns nil
// (no-op) — the convertDoneMsg from waitConvertDone takes over.
func listenConvertProgress(ch <-chan converter.BatchProgress) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return nil // channel closed, done msg will arrive separately
		}
		return convertProgressMsg{progress: p}
	}
}

func waitConvertDone(ch <-chan []converter.ConvertResult) tea.Cmd {
	return func() tea.Msg {
		results := <-ch
		return convertDoneMsg{results: results}
	}
}

func (c *ConvertScreen) View() string {
	if c.chdmanErr != nil {
		content := tui.StyleSubtitle.Render("Convert Files") + "\n\n"
		content += tui.StyleError.Render("chdman not found") + "\n\n"
		content += tui.StyleDim.Render(c.chdmanErr.Error()) + "\n\n"
		content += tui.StyleDim.Render("Press esc to go back")
		return lipgloss.NewStyle().Padding(1, 2).Render(content)
	}

	switch c.phase {
	case convertPhaseSelect:
		return c.viewSelect()
	case convertPhaseConfirm:
		return c.viewConfirm()
	case convertPhaseRunning:
		return c.viewRunning()
	case convertPhaseResults:
		return c.viewResults()
	}
	return ""
}

func (c *ConvertScreen) viewSelect() string {
	s := tui.StyleSubtitle.Render("Select files to convert to CHD") + "\n\n"

	if len(c.files) == 0 {
		s += tui.StyleDim.Render("No convertible files found (GDI/CUE/ISO).\n")
		s += tui.StyleDim.Render("Configure source directories in Settings first.")
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	for i, f := range c.files {
		cursor := "  "
		if i == c.cursor {
			cursor = tui.StyleMenuCursor.String()
		}

		check := "[ ] "
		if c.selected[i] {
			check = tui.StyleSuccess.Render("[x] ")
		}

		name := filepath.Base(f)
		dir := tui.StyleDim.Render(filepath.Dir(f))
		s += cursor + check + name + " " + dir + "\n"
	}

	s += "\n" + tui.StyleDim.Render(fmt.Sprintf("%d selected", len(c.selected)))
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (c *ConvertScreen) viewConfirm() string {
	s := tui.StyleSubtitle.Render("Confirm Conversion") + "\n\n"
	s += fmt.Sprintf("Convert %d file(s) to CHD format?\n\n", len(c.selected))

	for i := range c.files {
		if c.selected[i] {
			s += "  " + filepath.Base(c.files[i]) + "\n"
		}
	}

	s += "\n" + tui.StyleDim.Render("enter: start  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (c *ConvertScreen) viewRunning() string {
	s := tui.StyleSubtitle.Render("Converting...") + "\n\n"
	s += fmt.Sprintf("Progress: %d / %d files\n\n", c.filesDone, c.totalFiles)

	if c.currentFile != "" {
		s += tui.StyleDim.Render(filepath.Base(c.currentFile)) + "\n"
		s += renderProgressBar(c.currentPct, 40) + "\n"
	}

	s += "\n" + tui.StyleDim.Render("esc: cancel")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (c *ConvertScreen) viewResults() string {
	s := tui.StyleSubtitle.Render("Conversion Results") + "\n\n"

	var succeeded, failed int
	for _, r := range c.results {
		if r.Err != nil {
			failed++
			s += tui.StyleError.Render("FAIL") + " " + filepath.Base(r.InputPath) + "\n"
			s += "     " + tui.StyleDim.Render(r.Err.Error()) + "\n"
		} else {
			succeeded++
			s += tui.StyleSuccess.Render(" OK ") + " " + filepath.Base(r.InputPath) + " -> " + filepath.Base(r.OutputPath) + "\n"
		}
	}

	s += fmt.Sprintf("\n%s  %s",
		tui.StyleSuccess.Render(fmt.Sprintf("%d succeeded", succeeded)),
		tui.StyleError.Render(fmt.Sprintf("%d failed", failed)),
	)
	s += "\n\n" + tui.StyleDim.Render("enter/esc: done")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func renderProgressBar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := tui.StyleProgressFilled.Render(strings.Repeat("█", filled))
	bar += tui.StyleProgressEmpty.Render(strings.Repeat("░", empty))
	return fmt.Sprintf("%s %5.1f%%", bar, pct)
}

func (c *ConvertScreen) ShortHelp() []key.Binding {
	switch c.phase {
	case convertPhaseSelect:
		return []key.Binding{tui.Keys.Up, tui.Keys.Down, tui.Keys.Space, tui.Keys.Select, tui.Keys.Enter, tui.Keys.Back}
	default:
		return []key.Binding{tui.Keys.Back}
	}
}
