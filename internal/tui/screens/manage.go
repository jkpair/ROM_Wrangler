package screens

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/converter"
	"github.com/kurlmarx/romwrangler/internal/devices"
	"github.com/kurlmarx/romwrangler/internal/organizer"
	"github.com/kurlmarx/romwrangler/internal/systems"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type managePhase int

const (
	managePhaseScan managePhase = iota
	managePhaseReview
	managePhaseExtractPreview
	managePhaseExtracting
	managePhaseReScan
	managePhaseConvertPreview
	managePhaseConverting
	managePhaseArchiving
	managePhasePlan
	managePhaseExecute
	managePhaseResults
)

type scanDoneMsg struct {
	result      *organizer.ScanResult
	chdmanPath  string
	extractable []organizer.ExtractableFile
}

type planProgressMsg struct {
	current  int
	total    int
	filename string
}

type planDoneMsg struct {
	result *organizer.PlanResult
}

type manageConvertProgressMsg struct {
	progress converter.BatchProgress
}

type manageConvertDoneMsg struct {
	results []converter.ConvertResult
}

type archiveDoneMsg struct {
	result *organizer.ArchiveResult
}

type extractProgressMsg struct {
	current  int
	total    int
	filename string
}

type extractDoneMsg struct {
	result    *organizer.ExtractResult
	processed []organizer.ExtractableFile
}

type reScanDoneMsg struct {
	result     *organizer.ScanResult
	chdmanPath string
}

type deleteArchiveDoneMsg struct {
	err error
}

type ManageScreen struct {
	cfg           *config.Config
	width, height int
	phase         managePhase

	// Scan results
	scanResult *organizer.ScanResult
	systemList []systems.SystemID

	// Extraction
	extractable        []organizer.ExtractableFile
	extractProcessed   []organizer.ExtractableFile // archives that were extracted (for deferred archiving)
	extractResult      *organizer.ExtractResult
	extractProgressCh  <-chan extractProgressMsg
	extractProgress    struct {
		current  int
		total    int
		filename string
	}

	// Archive deletion
	archiveDeleted   bool
	archiveDeleteErr error

	// Conversion
	chdmanPath        string
	convertResults    []converter.ConvertResult
	archiveResult     *organizer.ArchiveResult
	convertProgressCh <-chan converter.BatchProgress
	convertCancel     context.CancelFunc
	convertProgress   struct {
		currentFile string
		currentPct  float64
		filesDone   int
		totalFiles  int
	}

	// Plan
	sortPlan   *organizer.SortPlan
	planResult *organizer.PlanResult
	outputDir  string

	// UI state
	cursor     int
	progressCh <-chan planProgressMsg
	progress   struct {
		current  int
		total    int
		filename string
	}
}

func NewManageScreen(cfg *config.Config, width, height int) *ManageScreen {
	return &ManageScreen{
		cfg:    cfg,
		width:  width,
		height: height,
	}
}

func (m *ManageScreen) Init() tea.Cmd {
	if len(m.cfg.SourceDirs) == 0 {
		return nil
	}
	return m.startScan()
}

func (m *ManageScreen) startScan() tea.Cmd {
	m.phase = managePhaseScan
	dirs := m.cfg.SourceDirs
	aliases := m.cfg.Aliases
	chdmanCfg := m.cfg.ChdmanPath
	return func() tea.Msg {
		// Fix .cue FILE references (case mismatches) before scanning
		organizer.FixCueFileReferences(dirs)
		result := organizer.Scan(dirs, aliases)
		chdmanPath, _ := converter.FindChdman(chdmanCfg)
		extractable := organizer.FindExtractable(dirs, aliases)
		return scanDoneMsg{result: result, chdmanPath: chdmanPath, extractable: extractable}
	}
}

func (m *ManageScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case scanDoneMsg:
		m.scanResult = msg.result
		m.chdmanPath = msg.chdmanPath
		m.extractable = msg.extractable
		m.phase = managePhaseReview
		m.buildSystemList()

	case extractProgressMsg:
		m.extractProgress.current = msg.current
		m.extractProgress.total = msg.total
		m.extractProgress.filename = msg.filename
		return m, listenExtractProgress(m.extractProgressCh)

	case extractDoneMsg:
		m.extractResult = msg.result
		m.extractProcessed = msg.processed
		// Go directly to re-scan — archiving is deferred until after conversion
		m.phase = managePhaseReScan
		return m, m.startReScan()

	case reScanDoneMsg:
		m.scanResult = msg.result
		m.chdmanPath = msg.chdmanPath
		m.extractable = nil // already extracted
		m.buildSystemList()
		// Skip review and go directly to convert, archive, or plan
		if m.hasConvertibleFiles() {
			m.phase = managePhaseConvertPreview
		} else if len(m.extractProcessed) > 0 {
			// No conversion needed but we have extracted archives to archive
			m.phase = managePhaseArchiving
			return m, m.startArchiveAll()
		} else {
			m.goToPlan()
		}

	case manageConvertProgressMsg:
		p := msg.progress
		m.convertProgress.currentFile = p.Filename
		m.convertProgress.currentPct = p.Percent
		if p.Done {
			m.convertProgress.filesDone++
		}
		return m, listenManageConvertProgress(m.convertProgressCh)

	case manageConvertDoneMsg:
		m.convertResults = msg.results
		// Update scan result BEFORE archiving, while GDI/CUE files still
		// exist on disk so CompanionFiles can read them to find tracks.
		m.scanResult.UpdateForConversions(m.convertResults)
		m.scanResult.RemoveFailedConversions(m.convertResults)
		m.phase = managePhaseArchiving
		return m, m.startArchiveAll()

	case archiveDoneMsg:
		m.archiveResult = msg.result
		m.goToPlan()

	case planProgressMsg:
		m.progress.current = msg.current
		m.progress.total = msg.total
		m.progress.filename = msg.filename
		return m, listenPlanProgress(m.progressCh)

	case planDoneMsg:
		m.planResult = msg.result
		m.phase = managePhaseResults
		// Auto-delete archive if configured
		if m.cfg.DeleteArchive && len(m.cfg.SourceDirs) > 0 {
			archiveDir := filepath.Join(m.cfg.SourceDirs[0], "_archive")
			return m, func() tea.Msg {
				err := organizer.DeleteArchiveDir(archiveDir)
				return deleteArchiveDoneMsg{err: err}
			}
		}

	case deleteArchiveDoneMsg:
		m.archiveDeleted = true
		m.archiveDeleteErr = msg.err

	case tea.KeyMsg:
		switch m.phase {
		case managePhaseReview:
			return m.updateReview(msg)
		case managePhaseExtractPreview:
			return m.updateExtractPreview(msg)
		case managePhaseConvertPreview:
			return m.updateConvertPreview(msg)
		case managePhaseConverting:
			if key.Matches(msg, tui.Keys.Back) {
				if m.convertCancel != nil {
					m.convertCancel()
				}
				m.goToPlan()
			}
		case managePhasePlan:
			return m.updatePlan(msg)
		case managePhaseResults:
			return m.updateResults(msg)
		default:
			if key.Matches(msg, tui.Keys.Back) {
				return m, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	}
	return m, nil
}

func (m *ManageScreen) buildSystemList() {
	m.systemList = nil
	for sysID := range m.scanResult.BySystem {
		m.systemList = append(m.systemList, sysID)
	}
	sort.Slice(m.systemList, func(i, j int) bool {
		return string(m.systemList[i]) < string(m.systemList[j])
	})
}

func (m *ManageScreen) hasConvertibleFiles() bool {
	return m.chdmanPath != "" && m.scanResult != nil && len(m.scanResult.Convertible) > 0
}

func (m *ManageScreen) hasExtractableFiles() bool {
	return len(m.extractable) > 0
}

func (m *ManageScreen) updateReview(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		return m, func() tea.Msg { return tui.NavigateBackMsg{} }
	case key.Matches(msg, tui.Keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, tui.Keys.Down):
		if m.cursor < len(m.systemList)-1 {
			m.cursor++
		}
	case key.Matches(msg, tui.Keys.Enter):
		if m.scanResult != nil && (len(m.scanResult.Files) > 0 || m.hasExtractableFiles()) {
			if m.hasExtractableFiles() {
				m.phase = managePhaseExtractPreview
			} else if m.hasConvertibleFiles() {
				m.phase = managePhaseConvertPreview
			} else {
				m.goToPlan()
			}
		}
	}
	return m, nil
}

func (m *ManageScreen) updateExtractPreview(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		m.phase = managePhaseReview
	case key.Matches(msg, tui.Keys.Enter):
		m.phase = managePhaseExtracting
		return m, m.startExtraction()
	case msg.String() == "s":
		// Skip extraction, proceed to convert or plan
		m.extractable = nil
		if m.hasConvertibleFiles() {
			m.phase = managePhaseConvertPreview
		} else {
			m.goToPlan()
		}
	}
	return m, nil
}

func (m *ManageScreen) updateConvertPreview(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		m.phase = managePhaseReview
	case key.Matches(msg, tui.Keys.Enter):
		m.phase = managePhaseConverting
		return m, m.startConversion()
	case msg.String() == "s":
		m.goToPlan()
	}
	return m, nil
}

func (m *ManageScreen) updateResults(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back), key.Matches(msg, tui.Keys.Enter):
		return m, func() tea.Msg { return tui.NavigateBackMsg{} }
	case msg.String() == "d":
		if !m.archiveDeleted && len(m.cfg.SourceDirs) > 0 {
			archiveDir := filepath.Join(m.cfg.SourceDirs[0], "_archive")
			return m, func() tea.Msg {
				err := organizer.DeleteArchiveDir(archiveDir)
				return deleteArchiveDoneMsg{err: err}
			}
		}
	}
	return m, nil
}

func (m *ManageScreen) goToPlan() {
	dev := devices.NewReplayOS()
	if len(m.cfg.SourceDirs) > 0 {
		m.outputDir = m.cfg.SourceDirs[0]
	} else {
		m.outputDir = "."
	}
	m.sortPlan = organizer.BuildSortPlan(m.scanResult, dev, m.outputDir, true)
	m.phase = managePhasePlan
}

func (m *ManageScreen) startExtraction() tea.Cmd {
	dirs := m.cfg.SourceDirs
	aliases := m.cfg.Aliases

	progressCh := make(chan extractProgressMsg, 100)
	m.extractProgressCh = progressCh
	m.extractProgress.current = 0
	m.extractProgress.total = 0
	m.extractProgress.filename = ""

	resultCh := make(chan extractDoneMsg, 1)
	go func() {
		result, processed := organizer.ExtractAll(dirs, aliases, func(current, total int, filename string) {
			progressCh <- extractProgressMsg{
				current:  current,
				total:    total,
				filename: filename,
			}
		})
		close(progressCh)
		resultCh <- extractDoneMsg{result: result, processed: processed}
	}()

	return tea.Batch(
		listenExtractProgress(progressCh),
		waitExtractDone(resultCh),
	)
}

func (m *ManageScreen) startReScan() tea.Cmd {
	dirs := m.cfg.SourceDirs
	aliases := m.cfg.Aliases
	chdmanCfg := m.cfg.ChdmanPath
	return func() tea.Msg {
		// Fix .cue FILE references (case mismatches, .ecm references)
		// before scanning so chdman can find the referenced files.
		organizer.FixCueFileReferences(dirs)
		result := organizer.Scan(dirs, aliases)
		chdmanPath, _ := converter.FindChdman(chdmanCfg)
		return reScanDoneMsg{result: result, chdmanPath: chdmanPath}
	}
}

func (m *ManageScreen) startConversion() tea.Cmd {
	var inputs []string
	for _, f := range m.scanResult.Convertible {
		inputs = append(inputs, f.Path)
	}
	m.convertProgress.totalFiles = len(inputs)
	m.convertProgress.filesDone = 0
	m.convertProgress.currentFile = ""
	m.convertProgress.currentPct = 0

	ctx, cancel := context.WithCancel(context.Background())
	m.convertCancel = cancel

	progressCh := make(chan converter.BatchProgress, 100)
	m.convertProgressCh = progressCh

	chdmanPath := m.chdmanPath

	resultCh := make(chan []converter.ConvertResult, 1)
	go func() {
		results := converter.BatchConvert(ctx, chdmanPath, inputs, 1, progressCh)
		resultCh <- results
	}()

	return tea.Batch(
		listenManageConvertProgress(progressCh),
		waitManageConvertDone(resultCh),
	)
}

// startArchiveAll archives both extracted archives (.rar/.zip/.7z/.ecm) and
// conversion originals (.cue/.gdi + track files) in a single step.
func (m *ManageScreen) startArchiveAll() tea.Cmd {
	convertResults := m.convertResults
	extractProcessed := m.extractProcessed
	sourceRoots := m.cfg.SourceDirs
	archiveDir := filepath.Join(sourceRoots[0], "_archive")

	return func() tea.Msg {
		combined := &organizer.ArchiveResult{}

		// Archive conversion originals (cue/gdi + companion tracks)
		if len(convertResults) > 0 {
			plan := organizer.BuildArchivePlan(sourceRoots, convertResults, archiveDir)
			result := organizer.ExecuteArchive(plan, nil)
			combined.FilesMoved += result.FilesMoved
			combined.Errors = append(combined.Errors, result.Errors...)
		}

		// Archive extracted archives (rar/zip/7z/ecm files)
		if len(extractProcessed) > 0 {
			result := organizer.ArchiveExtractedZips(extractProcessed, sourceRoots, archiveDir)
			combined.FilesMoved += result.FilesMoved
			combined.Errors = append(combined.Errors, result.Errors...)
		}

		return archiveDoneMsg{result: combined}
	}
}

func listenManageConvertProgress(ch <-chan converter.BatchProgress) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return nil
		}
		return manageConvertProgressMsg{progress: p}
	}
}

func waitManageConvertDone(ch <-chan []converter.ConvertResult) tea.Cmd {
	return func() tea.Msg {
		results := <-ch
		return manageConvertDoneMsg{results: results}
	}
}

func listenExtractProgress(ch <-chan extractProgressMsg) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return nil
		}
		return p
	}
}

func waitExtractDone(ch <-chan extractDoneMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func (m *ManageScreen) updatePlan(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		m.phase = managePhaseReview
	case key.Matches(msg, tui.Keys.Enter):
		m.phase = managePhaseExecute
		return m, m.executePlan()
	}
	return m, nil
}

func (m *ManageScreen) executePlan() tea.Cmd {
	plan := m.sortPlan

	ch := make(chan planProgressMsg, 100)
	m.progressCh = ch

	resultCh := make(chan *organizer.PlanResult, 1)
	go func() {
		result, _ := organizer.ExecutePlan(plan, true, func(current, total int, filename string) {
			ch <- planProgressMsg{
				current:  current,
				total:    total,
				filename: filename,
			}
		})
		close(ch)
		resultCh <- result
	}()

	return tea.Batch(
		listenPlanProgress(ch),
		waitPlanDone(resultCh),
	)
}

func listenPlanProgress(ch <-chan planProgressMsg) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return nil
		}
		return p
	}
}

func waitPlanDone(ch <-chan *organizer.PlanResult) tea.Cmd {
	return func() tea.Msg {
		result := <-ch
		return planDoneMsg{result: result}
	}
}

func (m *ManageScreen) View() string {
	switch m.phase {
	case managePhaseScan:
		return m.viewScan()
	case managePhaseReview:
		return m.viewReview()
	case managePhaseExtractPreview:
		return m.viewExtractPreview()
	case managePhaseExtracting:
		return m.viewExtracting()
	case managePhaseReScan:
		return m.viewReScan()
	case managePhaseConvertPreview:
		return m.viewConvertPreview()
	case managePhaseConverting:
		return m.viewConverting()
	case managePhaseArchiving:
		return m.viewArchiving()
	case managePhasePlan:
		return m.viewPlan()
	case managePhaseExecute:
		return m.viewExecute()
	case managePhaseResults:
		return m.viewResults()
	}
	return ""
}

func (m *ManageScreen) viewScan() string {
	s := tui.StyleSubtitle.Render("Manage ROMs") + "\n\n"

	if len(m.cfg.SourceDirs) == 0 {
		s += tui.StyleWarning.Render("No source directories configured.") + "\n\n"
		s += tui.StyleDim.Render("Go to Settings to add source directories.")
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	s += tui.StyleDim.Render("Scanning source directories...")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) viewReview() string {
	s := tui.StyleSubtitle.Render("Scan Results") + "\n\n"

	if m.scanResult == nil || (len(m.scanResult.Files) == 0 && !m.hasExtractableFiles()) {
		s += tui.StyleDim.Render("No ROM files found in source directories.")
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	s += fmt.Sprintf("Found %d files across %d systems\n\n",
		len(m.scanResult.Files), len(m.systemList))

	for i, sysID := range m.systemList {
		cursor := "  "
		if i == m.cursor {
			cursor = tui.StyleMenuCursor.String()
		}

		info, _ := systems.GetSystem(sysID)
		count := len(m.scanResult.BySystem[sysID])
		s += fmt.Sprintf("%s%s (%d files)\n", cursor,
			tui.StyleNormal.Render(info.DisplayName),
			count)
	}

	if m.hasExtractableFiles() {
		s += fmt.Sprintf("\n%s %d disc image archives to extract\n",
			tui.StyleSuccess.Render("+"),
			len(m.extractable))
	}

	if len(m.scanResult.Convertible) > 0 && m.chdmanPath != "" {
		s += fmt.Sprintf("\n%s %d files can be converted to CHD\n",
			tui.StyleSuccess.Render("+"),
			len(m.scanResult.Convertible))
	}

	if len(m.scanResult.Unresolved) > 0 {
		s += fmt.Sprintf("\n%s %d unresolved files\n",
			tui.StyleWarning.Render("!"),
			len(m.scanResult.Unresolved))
	}

	if len(m.scanResult.Errors) > 0 {
		s += fmt.Sprintf("\n%s %d errors\n",
			tui.StyleError.Render("!"),
			len(m.scanResult.Errors))
	}

	s += "\n" + tui.StyleDim.Render("enter: organize  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) viewExtractPreview() string {
	s := tui.StyleSubtitle.Render("Extract Disc Archives") + "\n\n"

	// Group by system
	bySystem := make(map[systems.SystemID]int)
	for _, f := range m.extractable {
		bySystem[f.System]++
	}

	var sysIDs []systems.SystemID
	for id := range bySystem {
		sysIDs = append(sysIDs, id)
	}
	sort.Slice(sysIDs, func(i, j int) bool {
		return string(sysIDs[i]) < string(sysIDs[j])
	})

	s += fmt.Sprintf("%d archives to extract:\n\n", len(m.extractable))
	for _, sysID := range sysIDs {
		info, _ := systems.GetSystem(sysID)
		s += fmt.Sprintf("  %s: %d archives\n", info.DisplayName, bySystem[sysID])
	}

	s += "\nEach archive will be extracted into its own subfolder\n"
	s += "to prevent track file name collisions.\n"

	s += "\n" + tui.StyleDim.Render("enter: extract  s: skip  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) viewExtracting() string {
	s := tui.StyleSubtitle.Render("Extracting Archives...") + "\n\n"

	p := m.extractProgress
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

func (m *ManageScreen) viewReScan() string {
	s := tui.StyleSubtitle.Render("Re-scanning...") + "\n\n"

	if m.extractResult != nil {
		s += fmt.Sprintf("%s %d archives extracted (%d files created)\n",
			tui.StyleSuccess.Render("OK"),
			m.extractResult.Extracted,
			m.extractResult.FilesCreated)
		if len(m.extractResult.Errors) > 0 {
			s += fmt.Sprintf("%s %d extraction errors\n",
				tui.StyleError.Render("!"),
				len(m.extractResult.Errors))
		}
		s += "\n"
	}

	s += tui.StyleDim.Render("Scanning for extracted files...")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) viewConvertPreview() string {
	s := tui.StyleSubtitle.Render("Convert to CHD") + "\n\n"

	// Show extract summary if we just extracted
	if m.extractResult != nil {
		s += fmt.Sprintf("%s %d archives extracted (%d files created)\n",
			tui.StyleSuccess.Render("OK"),
			m.extractResult.Extracted,
			m.extractResult.FilesCreated)
		s += "\n"
	}

	// Group convertible files by system
	bySystem := make(map[systems.SystemID]int)
	for _, f := range m.scanResult.Convertible {
		bySystem[f.System]++
	}

	var sysIDs []systems.SystemID
	for id := range bySystem {
		sysIDs = append(sysIDs, id)
	}
	sort.Slice(sysIDs, func(i, j int) bool {
		return string(sysIDs[i]) < string(sysIDs[j])
	})

	s += fmt.Sprintf("%d files to convert:\n\n", len(m.scanResult.Convertible))
	for _, sysID := range sysIDs {
		info, _ := systems.GetSystem(sysID)
		s += fmt.Sprintf("  %s: %d files\n", info.DisplayName, bySystem[sysID])
	}

	archiveDir := filepath.Join(m.cfg.SourceDirs[0], "_archive")
	s += fmt.Sprintf("\nOriginals will be archived to:\n  %s\n", tui.StyleDim.Render(archiveDir))

	s += "\n" + tui.StyleDim.Render("enter: convert  s: skip  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) viewConverting() string {
	s := tui.StyleSubtitle.Render("Converting to CHD...") + "\n\n"
	s += fmt.Sprintf("Progress: %d / %d files\n\n",
		m.convertProgress.filesDone, m.convertProgress.totalFiles)

	if m.convertProgress.currentFile != "" {
		s += tui.StyleDim.Render(filepath.Base(m.convertProgress.currentFile)) + "\n"
		s += renderProgressBar(m.convertProgress.currentPct, 40) + "\n"
	}

	s += "\n" + tui.StyleDim.Render("esc: cancel")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) viewArchiving() string {
	s := tui.StyleSubtitle.Render("Archiving Originals...") + "\n\n"
	s += tui.StyleDim.Render("Moving original disc images to archive...")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) viewPlan() string {
	s := tui.StyleSubtitle.Render("Sort Plan") + "\n\n"

	if m.sortPlan == nil {
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	// Show extraction summary if we extracted
	if m.extractResult != nil {
		s += fmt.Sprintf("%s %d archives extracted (%d files created)\n",
			tui.StyleSuccess.Render("OK"),
			m.extractResult.Extracted,
			m.extractResult.FilesCreated)
	}

	// Show conversion/archive summary if we did conversions
	if m.convertResults != nil {
		var succeeded, failed int
		for _, r := range m.convertResults {
			if r.Err != nil {
				failed++
			} else {
				succeeded++
			}
		}
		if succeeded > 0 {
			s += fmt.Sprintf("%s %d files converted to CHD\n",
				tui.StyleSuccess.Render("OK"), succeeded)
		}
		if failed > 0 {
			s += fmt.Sprintf("%s %d conversions failed:\n",
				tui.StyleError.Render("!"), failed)
			for _, r := range m.convertResults {
				if r.Err != nil {
					s += "  " + tui.StyleDim.Render(filepath.Base(r.InputPath)+": "+r.Err.Error()) + "\n"
				}
			}
		}
		if m.archiveResult != nil && m.archiveResult.FilesMoved > 0 {
			s += fmt.Sprintf("%s %d originals archived\n",
				tui.StyleSuccess.Render("OK"), m.archiveResult.FilesMoved)
		}
	}

	if m.extractResult != nil || m.convertResults != nil {
		s += "\n"
	}

	s += fmt.Sprintf("Files to move: %d\n", len(m.sortPlan.Files))
	s += fmt.Sprintf("M3U playlists: %d\n", len(m.sortPlan.M3Us))
	s += fmt.Sprintf("Directories:   %d\n", len(m.sortPlan.DirsToCreate))
	s += fmt.Sprintf("Output:        %s\n", m.outputDir)

	s += "\n" + tui.StyleDim.Render("enter: execute  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) viewExecute() string {
	s := tui.StyleSubtitle.Render("Organizing...") + "\n\n"
	if m.progress.total > 0 {
		pct := float64(m.progress.current) / float64(m.progress.total) * 100
		s += fmt.Sprintf("Progress: %d / %d\n", m.progress.current, m.progress.total)
		s += tui.StyleDim.Render(m.progress.filename) + "\n"
		s += renderProgressBar(pct, 40) + "\n"
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) viewResults() string {
	s := tui.StyleSubtitle.Render("Organization Complete") + "\n\n"

	// Extraction results
	if m.extractResult != nil {
		s += fmt.Sprintf("%s %d archives extracted (%d files created)\n",
			tui.StyleSuccess.Render("OK"),
			m.extractResult.Extracted,
			m.extractResult.FilesCreated)
		if len(m.extractResult.Errors) > 0 {
			s += fmt.Sprintf("%s %d extraction errors:\n",
				tui.StyleError.Render("!"),
				len(m.extractResult.Errors))
			for _, err := range m.extractResult.Errors {
				s += "  " + tui.StyleDim.Render(err.Error()) + "\n"
			}
		}
		s += "\n"
	}

	// Conversion results
	if m.convertResults != nil {
		var succeeded, failed int
		for _, r := range m.convertResults {
			if r.Err != nil {
				failed++
			} else {
				succeeded++
			}
		}
		if succeeded > 0 {
			s += fmt.Sprintf("%s %d files converted to CHD\n",
				tui.StyleSuccess.Render("OK"), succeeded)
		}
		if failed > 0 {
			s += fmt.Sprintf("%s %d conversions failed:\n",
				tui.StyleError.Render("!"), failed)
			for _, r := range m.convertResults {
				if r.Err != nil {
					s += "  " + tui.StyleDim.Render(filepath.Base(r.InputPath)+": "+r.Err.Error()) + "\n"
				}
			}
		}
		if m.archiveResult != nil && m.archiveResult.FilesMoved > 0 {
			s += fmt.Sprintf("%s %d originals archived\n",
				tui.StyleSuccess.Render("OK"), m.archiveResult.FilesMoved)
		}
		s += "\n"
	}

	if m.planResult != nil {
		if m.planResult.FilesMoved > 0 {
			s += fmt.Sprintf("%s %d files moved\n",
				tui.StyleSuccess.Render("OK"), m.planResult.FilesMoved)
		}
		if m.planResult.FilesCopied > 0 {
			s += fmt.Sprintf("%s %d files copied\n",
				tui.StyleSuccess.Render("OK"), m.planResult.FilesCopied)
		}
		s += fmt.Sprintf("%s %d M3U playlists created\n",
			tui.StyleSuccess.Render("OK"), m.planResult.M3UsWritten)
		s += fmt.Sprintf("%s %d directories created\n",
			tui.StyleSuccess.Render("OK"), m.planResult.DirsCreated)

		if len(m.planResult.Errors) > 0 {
			s += fmt.Sprintf("\n%s %d errors:\n",
				tui.StyleError.Render("!"), len(m.planResult.Errors))
			for _, err := range m.planResult.Errors {
				s += "  " + tui.StyleDim.Render(err.Error()) + "\n"
			}
		}
	}

	// Archive deletion status
	if m.archiveDeleted {
		if m.archiveDeleteErr != nil {
			s += fmt.Sprintf("\n%s Archive deletion failed: %s\n",
				tui.StyleError.Render("!"), m.archiveDeleteErr.Error())
		} else {
			s += fmt.Sprintf("\n%s Archive deleted\n",
				tui.StyleSuccess.Render("OK"))
		}
	}

	helpText := "enter/esc: done"
	if !m.archiveDeleted {
		helpText = "d: delete archive  enter/esc: done"
	}
	s += "\n" + tui.StyleDim.Render(helpText)
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) ShortHelp() []key.Binding {
	return []key.Binding{tui.Keys.Up, tui.Keys.Down, tui.Keys.Enter, tui.Keys.Back}
}
