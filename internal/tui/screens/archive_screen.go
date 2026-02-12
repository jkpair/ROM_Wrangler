package screens

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/organizer"
	"github.com/kurlmarx/romwrangler/internal/systems"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type archiveScreenPhase int

const (
	archiveScreenPhaseScan archiveScreenPhase = iota
	archiveScreenPhaseVariants
	archiveScreenPhasePreview
	archiveScreenPhaseArchiving
	archiveScreenPhaseResults
)

type archiveScreenScanDoneMsg struct {
	scanResult  *organizer.ScanResult
	superseded  []string
	extracted   []string
	variants    []organizer.VariantGroup
}

type archiveScreenArchiveDoneMsg struct {
	result *organizer.ArchiveResult
}

type ArchiveScreen struct {
	cfg           *config.Config
	width, height int
	phase         archiveScreenPhase

	// Scan results
	scanResult      *organizer.ScanResult
	supersededFiles []string // disc images where CHD exists
	extractedFiles  []string // archives where extracted folder exists

	// Dedup/variant filter
	variantGroups []organizer.VariantGroup
	dedupSelected map[string]bool
	dedupCursor   int
	dedupFiltered []string

	// Results
	archiveResult *organizer.ArchiveResult
}

func NewArchiveScreen(cfg *config.Config, width, height int) *ArchiveScreen {
	return &ArchiveScreen{
		cfg:    cfg,
		width:  width,
		height: height,
	}
}

func (a *ArchiveScreen) Init() tea.Cmd {
	if len(a.cfg.SourceDirs) == 0 {
		return nil
	}
	a.phase = archiveScreenPhaseScan
	dirs := a.cfg.ROMDirs()
	aliases := a.cfg.Aliases
	return func() tea.Msg {
		// Fix cue references before scanning
		organizer.FixCueFileReferences(dirs)
		scanResult := organizer.Scan(dirs, aliases)
		superseded := organizer.FindSupersededDiscImages(dirs, aliases)
		extracted := organizer.FindExtractedArchives(dirs, aliases)
		variants := organizer.DetectVariants(scanResult)
		return archiveScreenScanDoneMsg{
			scanResult: scanResult,
			superseded: superseded,
			extracted:  extracted,
			variants:   variants,
		}
	}
}

func (a *ArchiveScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case archiveScreenScanDoneMsg:
		a.scanResult = msg.scanResult
		a.supersededFiles = msg.superseded
		a.extractedFiles = msg.extracted
		a.variantGroups = msg.variants
		if len(a.variantGroups) > 0 {
			a.initDedupFilter()
			a.phase = archiveScreenPhaseVariants
		} else {
			a.phase = archiveScreenPhasePreview
		}

	case archiveScreenArchiveDoneMsg:
		a.archiveResult = msg.result
		a.phase = archiveScreenPhaseResults

	case tea.KeyMsg:
		switch a.phase {
		case archiveScreenPhaseVariants:
			return a.updateVariants(msg)
		case archiveScreenPhasePreview:
			return a.updatePreview(msg)
		case archiveScreenPhaseResults:
			if key.Matches(msg, tui.Keys.Back) || key.Matches(msg, tui.Keys.Enter) {
				return a, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		default:
			if key.Matches(msg, tui.Keys.Back) {
				return a, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	}
	return a, nil
}

// --- Dedup/variant filter ---

func (a *ArchiveScreen) initDedupFilter() {
	a.dedupSelected = make(map[string]bool)
	a.dedupCursor = 0
	a.dedupFiltered = nil

	for _, g := range a.variantGroups {
		// Pre-select the first file (highest region priority)
		for i, f := range g.Files {
			if i == 0 {
				a.dedupSelected[f.Path] = true
			}
		}
	}
}

type archiveDedupItem struct {
	isHeader  bool
	groupIdx  int
	path      string
	groupName string
}

func (a *ArchiveScreen) buildDedupFlatList() []archiveDedupItem {
	var items []archiveDedupItem
	for gi, g := range a.variantGroups {
		info, _ := systems.GetSystem(g.System)
		items = append(items, archiveDedupItem{
			isHeader:  true,
			groupIdx:  gi,
			groupName: g.BaseName + " (" + info.DisplayName + ")",
		})
		for _, f := range g.Files {
			items = append(items, archiveDedupItem{
				isHeader: false,
				groupIdx: gi,
				path:     f.Path,
			})
		}
	}
	return items
}

func (a *ArchiveScreen) updateVariants(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	items := a.buildDedupFlatList()
	maxIdx := len(items) - 1

	switch {
	case key.Matches(msg, tui.Keys.Back):
		return a, func() tea.Msg { return tui.NavigateBackMsg{} }
	case key.Matches(msg, tui.Keys.Up):
		if a.dedupCursor > 0 {
			a.dedupCursor--
		}
	case key.Matches(msg, tui.Keys.Down):
		if a.dedupCursor < maxIdx {
			a.dedupCursor++
		}
	case key.Matches(msg, tui.Keys.Space):
		if a.dedupCursor >= 0 && a.dedupCursor <= maxIdx {
			item := items[a.dedupCursor]
			if !item.isHeader {
				if a.dedupSelected[item.path] {
					delete(a.dedupSelected, item.path)
				} else {
					a.dedupSelected[item.path] = true
				}
			}
		}
	case key.Matches(msg, tui.Keys.Select): // 'a' key
		if a.dedupCursor >= 0 && a.dedupCursor <= maxIdx {
			item := items[a.dedupCursor]
			gi := item.groupIdx
			g := a.variantGroups[gi]
			allSelected := true
			for _, f := range g.Files {
				if !a.dedupSelected[f.Path] {
					allSelected = false
					break
				}
			}
			for _, f := range g.Files {
				if allSelected {
					delete(a.dedupSelected, f.Path)
				} else {
					a.dedupSelected[f.Path] = true
				}
			}
		}
	case msg.String() == "s":
		// Skip variant filtering (keep all variants)
		a.dedupFiltered = nil
		a.phase = archiveScreenPhasePreview
	case key.Matches(msg, tui.Keys.Enter):
		a.applyDedupFilter()
		a.phase = archiveScreenPhasePreview
	}
	return a, nil
}

func (a *ArchiveScreen) applyDedupFilter() {
	var toRemove []string
	for _, g := range a.variantGroups {
		for _, f := range g.Files {
			if !a.dedupSelected[f.Path] {
				toRemove = append(toRemove, f.Path)
			}
		}
	}
	if len(toRemove) > 0 {
		a.dedupFiltered = toRemove
	}
}

// --- Preview ---

func (a *ArchiveScreen) updatePreview(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		return a, func() tea.Msg { return tui.NavigateBackMsg{} }
	case key.Matches(msg, tui.Keys.Enter):
		totalFiles := len(a.supersededFiles) + len(a.extractedFiles) + len(a.dedupFiltered)
		if totalFiles > 0 {
			a.phase = archiveScreenPhaseArchiving
			return a, a.startArchive()
		}
	}
	return a, nil
}

func (a *ArchiveScreen) startArchive() tea.Cmd {
	superseded := a.supersededFiles
	extracted := a.extractedFiles
	dedupFiltered := a.dedupFiltered
	sourceRoots := a.cfg.ROMDirs()
	archiveDir := filepath.Join(sourceRoots[0], "_archive")

	return func() tea.Msg {
		combined := &organizer.ArchiveResult{}

		// Combine all redundant file paths
		var allPaths []string
		allPaths = append(allPaths, superseded...)
		allPaths = append(allPaths, extracted...)
		allPaths = append(allPaths, dedupFiltered...)

		if len(allPaths) > 0 {
			result := organizer.ArchiveFilteredFiles(allPaths, sourceRoots, archiveDir)
			combined.FilesMoved += result.FilesMoved
			combined.Errors = append(combined.Errors, result.Errors...)
		}

		return archiveScreenArchiveDoneMsg{result: combined}
	}
}

// --- Views ---

func (a *ArchiveScreen) View() string {
	switch a.phase {
	case archiveScreenPhaseScan:
		return a.viewScan()
	case archiveScreenPhaseVariants:
		return a.viewVariants()
	case archiveScreenPhasePreview:
		return a.viewPreview()
	case archiveScreenPhaseArchiving:
		return a.viewArchiving()
	case archiveScreenPhaseResults:
		return a.viewResults()
	}
	return ""
}

func (a *ArchiveScreen) viewScan() string {
	s := tui.StyleSubtitle.Render("Archive Redundant Files") + "\n\n"

	if len(a.cfg.SourceDirs) == 0 {
		s += tui.StyleWarning.Render("No root directory configured.") + "\n\n"
		s += tui.StyleDim.Render("Go to Settings to set a root directory.")
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	s += tui.StyleDim.Render("Scanning for redundant files...")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (a *ArchiveScreen) viewVariants() string {
	s := tui.StyleSubtitle.Render("Select Versions to Keep") + "\n\n"

	items := a.buildDedupFlatList()

	// Count stats
	totalFiles := 0
	selectedCount := 0
	for _, g := range a.variantGroups {
		totalFiles += len(g.Files)
		for _, f := range g.Files {
			if a.dedupSelected[f.Path] {
				selectedCount++
			}
		}
	}

	s += fmt.Sprintf("Found %d games with %d total variants\n",
		len(a.variantGroups), totalFiles)
	s += fmt.Sprintf("Keeping %d, archiving %d\n\n",
		selectedCount, totalFiles-selectedCount)

	// Render list with scrolling
	viewportHeight := a.height - 12
	if viewportHeight < 5 {
		viewportHeight = 5
	}
	startIdx := 0
	if a.dedupCursor > viewportHeight-3 {
		startIdx = a.dedupCursor - viewportHeight + 3
	}

	for i := startIdx; i < len(items) && i < startIdx+viewportHeight; i++ {
		item := items[i]
		if item.isHeader {
			cursor := "  "
			if i == a.dedupCursor {
				cursor = tui.StyleMenuCursor.String()
			}
			s += cursor + tui.StyleSelected.Render(item.groupName) + "\n"
		} else {
			cursor := "    "
			if i == a.dedupCursor {
				cursor = "  " + tui.StyleMenuCursor.String()
			}

			check := "[ ] "
			if a.dedupSelected[item.path] {
				check = tui.StyleSuccess.Render("[x] ")
			}

			name := filepath.Base(item.path)
			s += cursor + check + name + "\n"
		}
	}

	s += "\n" + tui.StyleDim.Render("space: toggle  a: toggle group  enter: confirm  s: skip  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (a *ArchiveScreen) viewPreview() string {
	s := tui.StyleSubtitle.Render("Archive Redundant Files") + "\n\n"

	totalFiles := len(a.supersededFiles) + len(a.extractedFiles) + len(a.dedupFiltered)

	if totalFiles == 0 {
		s += tui.StyleDim.Render("No redundant files found.") + "\n"
		s += "\n" + tui.StyleDim.Render("esc: back")
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	if len(a.supersededFiles) > 0 {
		s += fmt.Sprintf("%s %d superseded disc images (CHD exists)\n",
			tui.StyleSuccess.Render("+"), len(a.supersededFiles))
	}
	if len(a.extractedFiles) > 0 {
		s += fmt.Sprintf("%s %d extracted archives\n",
			tui.StyleSuccess.Render("+"), len(a.extractedFiles))
	}
	if len(a.dedupFiltered) > 0 {
		s += fmt.Sprintf("%s %d duplicate versions\n",
			tui.StyleSuccess.Render("+"), len(a.dedupFiltered))
	}

	s += fmt.Sprintf("\nTotal: %d files to archive\n", totalFiles)

	archiveDir := filepath.Join(a.cfg.ROMDirs()[0], "_archive")
	s += fmt.Sprintf("\nDestination:\n  %s\n", tui.StyleDim.Render(archiveDir))

	s += "\n" + tui.StyleDim.Render("enter: archive  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (a *ArchiveScreen) viewArchiving() string {
	s := tui.StyleSubtitle.Render("Archiving...") + "\n\n"
	s += tui.StyleDim.Render("Moving redundant files to archive...")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (a *ArchiveScreen) viewResults() string {
	s := tui.StyleSubtitle.Render("Archive Complete") + "\n\n"

	if a.archiveResult != nil {
		s += fmt.Sprintf("%s %d files archived\n",
			tui.StyleSuccess.Render("OK"), a.archiveResult.FilesMoved)
		if len(a.archiveResult.Errors) > 0 {
			s += fmt.Sprintf("\n%s %d errors:\n",
				tui.StyleError.Render("!"),
				len(a.archiveResult.Errors))
			for _, err := range a.archiveResult.Errors {
				s += "  " + tui.StyleDim.Render(err.Error()) + "\n"
			}
		}
	}

	s += "\n" + tui.StyleDim.Render("enter/esc: done")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (a *ArchiveScreen) ShortHelp() []key.Binding {
	if a.phase == archiveScreenPhaseVariants {
		return []key.Binding{tui.Keys.Up, tui.Keys.Down, tui.Keys.Space, tui.Keys.Select, tui.Keys.Enter, tui.Keys.Back}
	}
	return []key.Binding{tui.Keys.Enter, tui.Keys.Back}
}
