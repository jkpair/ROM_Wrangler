package screens

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/converter"
	"github.com/kurlmarx/romwrangler/internal/devices"
	"github.com/kurlmarx/romwrangler/internal/organizer"
	"github.com/kurlmarx/romwrangler/internal/romdb"
	"github.com/kurlmarx/romwrangler/internal/scraper"
	"github.com/kurlmarx/romwrangler/internal/systems"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type managePhase int

const (
	managePhaseScan managePhase = iota
	managePhaseResolve // resolve misplaced/unresolved ROMs
	managePhaseAssign  // manual system assignment for remaining unresolved
	managePhaseReview
	managePhaseExtracting
	managePhaseReScan
	managePhaseDedupFilter
	managePhaseConverting
	managePhaseSorting   // build sort plan + execute automatically
	managePhaseArchiving
	managePhaseTransfer // transfer option before results
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

type resolveProgressMsg struct {
	current  int
	total    int
	filename string
}

type resolveDoneMsg struct {
	misplaced   []organizer.MisplacedFile
	resolvedN   int // number of unresolved files resolved by extension
	ssResolvedN int // number of unresolved files resolved by ScreenScraper
}

type ManageScreen struct {
	cfg           *config.Config
	width, height int
	phase         managePhase

	// Scan results
	scanResult *organizer.ScanResult
	systemList []systems.SystemID

	// Resolve misplaced/unknown ROMs
	misplaced         []organizer.MisplacedFile
	resolvedN         int // number of unresolved files resolved by extension
	ssResolvedN       int // number resolved by ScreenScraper
	resolveDone       bool
	resolveProgressCh <-chan resolveProgressMsg
	resolveProgress   struct {
		current  int
		total    int
		filename string
	}

	// Manual system assignment for remaining unresolved files
	assignFiles     []string                       // unresolved file paths
	assignCursor    int                            // cursor in file list
	assignSysCursor int                            // cursor in system picker
	assignPicking   bool                           // true = system picker focused
	assignSystems   []systems.SystemID             // filtered systems for current file
	assignChoices   map[string]systems.SystemID    // user assignments: path -> system

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

	// Duplicate/variant filter
	variantGroups []organizer.VariantGroup
	dedupSelected map[string]bool // key = file path, true = keep
	dedupCursor   int
	dedupFiltered []string // paths of files removed by dedup filter (for archiving)

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

	// Transfer option
	transferCursor         int  // 0=rsync, 1=USB, 2=Manual, 3=Skip
	showManualInstructions bool

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
	dirs := m.cfg.ROMDirs()
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
		m.phase = managePhaseResolve
		return m, m.startResolve()

	case resolveProgressMsg:
		m.resolveProgress.current = msg.current
		m.resolveProgress.total = msg.total
		m.resolveProgress.filename = msg.filename
		return m, listenResolveProgress(m.resolveProgressCh)

	case resolveDoneMsg:
		m.resolveDone = true
		m.misplaced = msg.misplaced
		m.resolvedN = msg.resolvedN
		m.ssResolvedN = msg.ssResolvedN
		hasChanges := len(m.misplaced) > 0 || m.resolvedN > 0 || m.ssResolvedN > 0
		if hasChanges {
			m.buildSystemList()
			// Stay on resolve phase so user can review
		} else if len(m.scanResult.Unresolved) > 0 {
			// Nothing auto-resolved but still unresolved — go to manual assign
			m.buildSystemList()
			m.initAssign()
			m.phase = managePhaseAssign
		} else {
			// Nothing to resolve, skip to review
			m.phase = managePhaseReview
			m.buildSystemList()
		}

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
		// Check for variant groups before proceeding
		m.variantGroups = organizer.DetectVariants(m.scanResult)
		if len(m.variantGroups) > 0 {
			m.initDedupFilter()
			m.phase = managePhaseDedupFilter
		} else {
			return m.advanceFromDedup()
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
		return m.advanceToSorting()

	case archiveDoneMsg:
		m.archiveResult = msg.result
		m.transferCursor = 0
		m.showManualInstructions = false
		m.phase = managePhaseTransfer

	case planProgressMsg:
		m.progress.current = msg.current
		m.progress.total = msg.total
		m.progress.filename = msg.filename
		return m, listenPlanProgress(m.progressCh)

	case planDoneMsg:
		m.planResult = msg.result
		m.phase = managePhaseArchiving
		return m, m.startArchiveAll()

	case deleteArchiveDoneMsg:
		m.archiveDeleted = true
		m.archiveDeleteErr = msg.err

	case tea.KeyMsg:
		switch m.phase {
		case managePhaseResolve:
			return m.updateResolve(msg)
		case managePhaseAssign:
			return m.updateAssign(msg)
		case managePhaseReview:
			return m.updateReview(msg)
		case managePhaseDedupFilter:
			return m.updateDedupFilter(msg)
		case managePhaseConverting:
			if key.Matches(msg, tui.Keys.Back) {
				if m.convertCancel != nil {
					m.convertCancel()
				}
				return m.advanceToSorting()
			}
		case managePhaseTransfer:
			return m.updateTransfer(msg)
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
				m.phase = managePhaseExtracting
				return m, m.startExtraction()
			}
			return m.advanceFromReview()
		}
	}
	return m, nil
}

func (m *ManageScreen) advanceFromReview() (tui.Screen, tea.Cmd) {
	m.variantGroups = organizer.DetectVariants(m.scanResult)
	if len(m.variantGroups) > 0 {
		m.initDedupFilter()
		m.phase = managePhaseDedupFilter
		return m, nil
	}
	return m.advanceFromDedup()
}

func (m *ManageScreen) advanceToSorting() (tui.Screen, tea.Cmd) {
	dev := devices.NewReplayOS()
	if len(m.cfg.SourceDirs) > 0 {
		m.outputDir = m.cfg.ROMDirs()[0]
	} else {
		m.outputDir = "."
	}
	m.sortPlan = organizer.BuildSortPlan(m.scanResult, dev, m.outputDir, true)
	m.phase = managePhaseSorting
	return m, m.executePlan()
}

func (m *ManageScreen) updateTransfer(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Up):
		if m.transferCursor > 0 {
			m.transferCursor--
		}
	case key.Matches(msg, tui.Keys.Down):
		if m.transferCursor < 3 {
			m.transferCursor++
		}
	case key.Matches(msg, tui.Keys.Enter):
		switch m.transferCursor {
		case 0, 1: // rsync, USB → navigate to Transfer screen
			return m, func() tea.Msg { return tui.NavigateMsg{Screen: tui.ScreenTransfer} }
		case 2: // Manual instructions
			m.showManualInstructions = true
			m.phase = managePhaseResults
		case 3: // Skip
			m.phase = managePhaseResults
		}
	case key.Matches(msg, tui.Keys.Back):
		m.phase = managePhaseResults
	}
	return m, nil
}

func (m *ManageScreen) updateResults(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back), key.Matches(msg, tui.Keys.Enter):
		return m, func() tea.Msg { return tui.NavigateBackMsg{} }
	case msg.String() == "d":
		if !m.archiveDeleted && len(m.cfg.SourceDirs) > 0 {
			archiveDir := filepath.Join(m.cfg.ROMDirs()[0], "_archive")
			return m, func() tea.Msg {
				err := organizer.DeleteArchiveDir(archiveDir)
				return deleteArchiveDoneMsg{err: err}
			}
		}
	}
	return m, nil
}

func (m *ManageScreen) startExtraction() tea.Cmd {
	dirs := m.cfg.ROMDirs()
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
	dirs := m.cfg.ROMDirs()
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

// startArchiveAll archives extracted archives (.rar/.zip/.7z/.ecm),
// conversion originals (.cue/.gdi + track files), and dedup-filtered
// variant files in a single step.
func (m *ManageScreen) startArchiveAll() tea.Cmd {
	convertResults := m.convertResults
	extractProcessed := m.extractProcessed
	dedupFiltered := m.dedupFiltered
	sourceRoots := m.cfg.ROMDirs()
	archiveDir := filepath.Join(sourceRoots[0], "_archive")

	// Capture unsupported + unresolved before goroutine
	var unsupported, unresolved []string
	if m.scanResult != nil {
		unsupported = m.scanResult.Unsupported
		unresolved = m.scanResult.Unresolved
	}

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

		// Archive dedup-filtered files (unselected variants)
		if len(dedupFiltered) > 0 {
			result := organizer.ArchiveFilteredFiles(dedupFiltered, sourceRoots, archiveDir)
			combined.FilesMoved += result.FilesMoved
			combined.Errors = append(combined.Errors, result.Errors...)
		}

		// Archive unsupported + unresolved files
		if len(unsupported) > 0 || len(unresolved) > 0 {
			result := organizer.ArchiveUnsupported(unresolved, unsupported, sourceRoots, archiveDir)
			combined.FilesMoved += result.FilesMoved
			combined.Errors = append(combined.Errors, result.Errors...)
		}

		return archiveDoneMsg{result: combined}
	}
}

func listenResolveProgress(ch <-chan resolveProgressMsg) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return nil
		}
		return p
	}
}

func waitResolveDone(ch <-chan resolveDoneMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
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
	case managePhaseResolve:
		return m.viewResolve()
	case managePhaseAssign:
		return m.viewAssign()
	case managePhaseReview:
		return m.viewReview()
	case managePhaseExtracting:
		return m.viewExtracting()
	case managePhaseReScan:
		return m.viewReScan()
	case managePhaseDedupFilter:
		return m.viewDedupFilter()
	case managePhaseConverting:
		return m.viewConverting()
	case managePhaseSorting:
		return m.viewSorting()
	case managePhaseArchiving:
		return m.viewArchiving()
	case managePhaseTransfer:
		return m.viewTransfer()
	case managePhaseResults:
		return m.viewResults()
	}
	return ""
}

func (m *ManageScreen) startResolve() tea.Cmd {
	scanResult := m.scanResult
	cfg := m.cfg
	hasSS := cfg.Scraping.ScreenScraperUser != ""

	if !hasSS {
		// Simple path: extension-only resolve (instant, no progress needed)
		return func() tea.Msg {
			misplaced := organizer.DetectMisplaced(scanResult)
			unresolvedBefore := len(scanResult.Unresolved)
			organizer.ResolveUnknown(context.Background(), scanResult, nil)
			resolvedN := unresolvedBefore - len(scanResult.Unresolved)
			return resolveDoneMsg{misplaced: misplaced, resolvedN: resolvedN}
		}
	}

	// SS path: async with progress reporting
	progressCh := make(chan resolveProgressMsg, 100)
	m.resolveProgressCh = progressCh
	m.resolveProgress.current = 0
	m.resolveProgress.total = 0
	m.resolveProgress.filename = ""

	resultCh := make(chan resolveDoneMsg, 1)
	go func() {
		// Extension-based first (instant)
		misplaced := organizer.DetectMisplaced(scanResult)
		unresolvedBefore := len(scanResult.Unresolved)
		organizer.ResolveUnknown(context.Background(), scanResult, nil)
		resolvedN := unresolvedBefore - len(scanResult.Unresolved)

		// SS resolve for remaining unresolved files
		ssResolvedN := 0
		if len(scanResult.Unresolved) > 0 {
			ssClient := scraper.NewScreenScraperClient(
				cfg.Scraping.ScreenScraperUser,
				cfg.Scraping.ScreenScraperPass,
			)
			var cache scraper.Cache
			if db, err := romdb.Open(""); err == nil {
				cache = db
			}
			identifier := scraper.NewIdentifier(nil, ssClient, cache)

			total := len(scanResult.Unresolved)
			var stillUnresolved []string
			for i, path := range scanResult.Unresolved {
				progressCh <- resolveProgressMsg{
					current:  i + 1,
					total:    total,
					filename: filepath.Base(path),
				}
				match, err := identifier.Identify(context.Background(), path, "")
				if err == nil && match.Matched && match.Game != nil && match.Game.System != "" {
					sysID := match.Game.System
					sf := organizer.ScannedFile{Path: path, System: sysID, Resolved: true}
					scanResult.Files = append(scanResult.Files, sf)
					scanResult.BySystem[sysID] = append(scanResult.BySystem[sysID], sf)
					ssResolvedN++
					continue
				}
				stillUnresolved = append(stillUnresolved, path)
			}
			scanResult.Unresolved = stillUnresolved
		}

		close(progressCh)
		resultCh <- resolveDoneMsg{
			misplaced:   misplaced,
			resolvedN:   resolvedN,
			ssResolvedN: ssResolvedN,
		}
	}()

	return tea.Batch(
		listenResolveProgress(progressCh),
		waitResolveDone(resultCh),
	)
}

func (m *ManageScreen) updateResolve(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	if !m.resolveDone {
		// Still resolving, only allow back
		if key.Matches(msg, tui.Keys.Back) {
			return m, func() tea.Msg { return tui.NavigateBackMsg{} }
		}
		return m, nil
	}
	switch {
	case key.Matches(msg, tui.Keys.Back):
		return m, func() tea.Msg { return tui.NavigateBackMsg{} }
	case key.Matches(msg, tui.Keys.Enter):
		// Accept relocations
		if len(m.misplaced) > 0 {
			organizer.RelocateMisplaced(m.scanResult, m.misplaced)
			m.buildSystemList()
		}
		if len(m.scanResult.Unresolved) > 0 {
			m.initAssign()
			m.phase = managePhaseAssign
		} else {
			m.phase = managePhaseReview
		}
	case msg.String() == "s":
		// Skip relocations
		if len(m.scanResult.Unresolved) > 0 {
			m.initAssign()
			m.phase = managePhaseAssign
		} else {
			m.phase = managePhaseReview
		}
	}
	return m, nil
}

func (m *ManageScreen) viewResolve() string {
	s := tui.StyleSubtitle.Render("Detect Misplaced ROMs") + "\n\n"

	if !m.resolveDone {
		if m.resolveProgress.total > 0 {
			// ScreenScraper lookup in progress
			s += fmt.Sprintf("Resolving via ScreenScraper... (%d / %d)\n",
				m.resolveProgress.current, m.resolveProgress.total)
			s += tui.StyleDim.Render("Hashing: "+m.resolveProgress.filename) + "\n"
		} else {
			s += tui.StyleDim.Render("Analyzing file placements...")
		}
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	hasChanges := len(m.misplaced) > 0 || m.resolvedN > 0 || m.ssResolvedN > 0

	if !hasChanges {
		s += tui.StyleDim.Render("All files are in the correct folders.")
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	if m.resolvedN > 0 {
		s += fmt.Sprintf("%s Resolved %d unknown files by extension\n\n",
			tui.StyleSuccess.Render("+"), m.resolvedN)
	}

	if m.ssResolvedN > 0 {
		s += fmt.Sprintf("%s Resolved %d files via ScreenScraper\n\n",
			tui.StyleSuccess.Render("+"), m.ssResolvedN)
	}

	if len(m.misplaced) > 0 {
		s += fmt.Sprintf("Detected %d misplaced files:\n\n", len(m.misplaced))

		limit := len(m.misplaced)
		if limit > 20 {
			limit = 20
		}
		for _, mf := range m.misplaced[:limit] {
			name := filepath.Base(mf.Path)
			fromInfo, _ := systems.GetSystem(mf.CurrentSystem)
			toInfo, _ := systems.GetSystem(mf.CorrectSystem)
			s += fmt.Sprintf("  %s\n", tui.StyleNormal.Render(name))
			s += fmt.Sprintf("    %s %s %s (%s)\n",
				tui.StyleDim.Render(fromInfo.DisplayName),
				tui.StyleWarning.Render("->"),
				tui.StyleSuccess.Render(toInfo.DisplayName),
				mf.Source)
		}
		if len(m.misplaced) > 20 {
			s += fmt.Sprintf("\n  ... and %d more\n", len(m.misplaced)-20)
		}
	}

	if len(m.scanResult.Unresolved) > 0 {
		s += fmt.Sprintf("\n%s %d files still unresolved (manual assignment next)\n",
			tui.StyleWarning.Render("!"), len(m.scanResult.Unresolved))
	}

	s += "\n" + tui.StyleDim.Render("enter: accept  s: skip  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) viewScan() string {
	s := tui.StyleSubtitle.Render("Manage ROMs") + "\n\n"

	if len(m.cfg.SourceDirs) == 0 {
		s += tui.StyleWarning.Render("No root directory configured.") + "\n\n"
		s += tui.StyleDim.Render("Go to Settings to set a root directory.")
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

func (m *ManageScreen) viewSorting() string {
	s := tui.StyleSubtitle.Render("Organizing...") + "\n\n"
	if m.progress.total > 0 {
		pct := float64(m.progress.current) / float64(m.progress.total) * 100
		s += fmt.Sprintf("Progress: %d / %d\n", m.progress.current, m.progress.total)
		s += tui.StyleDim.Render(m.progress.filename) + "\n"
		s += renderProgressBar(pct, 40) + "\n"
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) viewTransfer() string {
	s := tui.StyleSubtitle.Render("Transfer to Device") + "\n\n"
	s += "How would you like to transfer your organized ROMs?\n\n"
	options := []string{
		"Transfer via rsync",
		"Transfer via USB",
		"Show manual instructions",
		"Skip",
	}
	for i, opt := range options {
		cursor := "  "
		if i == m.transferCursor {
			cursor = tui.StyleMenuCursor.String()
		}
		s += cursor + tui.StyleNormal.Render(opt) + "\n"
	}
	s += "\n" + tui.StyleDim.Render("enter: select  esc: skip to summary")
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

	// Archive results
	if m.archiveResult != nil && m.archiveResult.FilesMoved > 0 {
		s += fmt.Sprintf("\n%s %d files archived\n",
			tui.StyleSuccess.Render("OK"), m.archiveResult.FilesMoved)
		if len(m.archiveResult.Errors) > 0 {
			s += fmt.Sprintf("%s %d archive errors:\n",
				tui.StyleError.Render("!"), len(m.archiveResult.Errors))
			for _, err := range m.archiveResult.Errors {
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

	// Manual transfer instructions
	if m.showManualInstructions {
		s += "\n" + tui.StyleSubtitle.Render("Manual Transfer Instructions") + "\n\n"
		s += "1. Copy the roms/ folder to a USB drive (exFAT format)\n"
		s += "2. Insert into your ReplayOS device\n"
		s += "3. Or use SFTP: sftp root@<device-ip>\n"
		s += "   Default credentials: root:replayos (port 22)\n"
	}

	helpText := "enter/esc: done"
	if !m.archiveDeleted {
		helpText = "d: delete archive  enter/esc: done"
	}
	s += "\n" + tui.StyleDim.Render(helpText)
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

// --- Manual system assignment methods ---

func (m *ManageScreen) initAssign() {
	m.assignFiles = make([]string, len(m.scanResult.Unresolved))
	copy(m.assignFiles, m.scanResult.Unresolved)
	m.assignCursor = 0
	m.assignSysCursor = 0
	m.assignPicking = false
	m.assignChoices = make(map[string]systems.SystemID)
	m.updateAssignSystems()
}

func (m *ManageScreen) updateAssignSystems() {
	if m.assignCursor >= 0 && m.assignCursor < len(m.assignFiles) {
		path := m.assignFiles[m.assignCursor]
		ext := strings.ToLower(filepath.Ext(path))
		m.assignSystems = systemsForExtension(ext)
		if m.assignSysCursor >= len(m.assignSystems) {
			m.assignSysCursor = 0
		}
	}
}

func systemsForExtension(ext string) []systems.SystemID {
	var result []systems.SystemID
	for sysID, formats := range systems.SupportedFormats {
		for _, f := range formats {
			if f == ext {
				result = append(result, sysID)
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		a, _ := systems.GetSystem(result[i])
		b, _ := systems.GetSystem(result[j])
		return a.DisplayName < b.DisplayName
	})
	return result
}

func (m *ManageScreen) updateAssign(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		return m, func() tea.Msg { return tui.NavigateBackMsg{} }
	case key.Matches(msg, tui.Keys.Tab):
		m.assignPicking = !m.assignPicking
		if m.assignPicking {
			m.updateAssignSystems()
		}
	case key.Matches(msg, tui.Keys.Up):
		if m.assignPicking {
			if m.assignSysCursor > 0 {
				m.assignSysCursor--
			}
		} else {
			if m.assignCursor > 0 {
				m.assignCursor--
				m.updateAssignSystems()
				m.assignSysCursor = 0
			}
		}
	case key.Matches(msg, tui.Keys.Down):
		if m.assignPicking {
			if m.assignSysCursor < len(m.assignSystems)-1 {
				m.assignSysCursor++
			}
		} else {
			if m.assignCursor < len(m.assignFiles)-1 {
				m.assignCursor++
				m.updateAssignSystems()
				m.assignSysCursor = 0
			}
		}
	case key.Matches(msg, tui.Keys.Enter):
		if m.assignPicking && len(m.assignSystems) > 0 {
			// Assign system to current file
			path := m.assignFiles[m.assignCursor]
			m.assignChoices[path] = m.assignSystems[m.assignSysCursor]
			m.assignPicking = false
			// Auto-advance to next unassigned file
			for i := m.assignCursor + 1; i < len(m.assignFiles); i++ {
				if _, ok := m.assignChoices[m.assignFiles[i]]; !ok {
					m.assignCursor = i
					m.updateAssignSystems()
					m.assignSysCursor = 0
					break
				}
			}
		} else if !m.assignPicking {
			// Switch to system picker
			m.assignPicking = true
			m.updateAssignSystems()
		}
	case msg.String() == "a":
		// Apply all assignments and proceed to review
		m.applyAssignments()
		m.buildSystemList()
		m.phase = managePhaseReview
	case msg.String() == "s":
		// Skip, proceed to review without assigning
		m.buildSystemList()
		m.phase = managePhaseReview
	}
	return m, nil
}

func (m *ManageScreen) applyAssignments() {
	if len(m.assignChoices) == 0 {
		return
	}
	assigned := make(map[string]bool)
	for path, sysID := range m.assignChoices {
		sf := organizer.ScannedFile{Path: path, System: sysID, Resolved: true}
		m.scanResult.Files = append(m.scanResult.Files, sf)
		m.scanResult.BySystem[sysID] = append(m.scanResult.BySystem[sysID], sf)
		assigned[path] = true
	}
	var remaining []string
	for _, path := range m.scanResult.Unresolved {
		if !assigned[path] {
			remaining = append(remaining, path)
		}
	}
	m.scanResult.Unresolved = remaining
}

func (m *ManageScreen) viewAssign() string {
	s := tui.StyleSubtitle.Render("Assign System to Unresolved Files") + "\n\n"

	assignedCount := len(m.assignChoices)
	s += fmt.Sprintf("%d files, %d assigned\n\n", len(m.assignFiles), assignedCount)

	// File list
	viewportHeight := m.height - 14
	if viewportHeight < 8 {
		viewportHeight = 8
	}
	fileHeight := viewportHeight * 2 / 3
	if fileHeight < 3 {
		fileHeight = 3
	}

	startIdx := 0
	if m.assignCursor > fileHeight-2 {
		startIdx = m.assignCursor - fileHeight + 2
	}

	for i := startIdx; i < len(m.assignFiles) && i < startIdx+fileHeight; i++ {
		path := m.assignFiles[i]
		name := filepath.Base(path)

		cursor := "  "
		if i == m.assignCursor {
			if !m.assignPicking {
				cursor = tui.StyleMenuCursor.String()
			} else {
				cursor = tui.StyleDim.Render("> ")
			}
		}

		assignment := tui.StyleDim.Render("(unassigned)")
		if sysID, ok := m.assignChoices[path]; ok {
			info, _ := systems.GetSystem(sysID)
			assignment = tui.StyleSuccess.Render(info.DisplayName)
		}

		s += fmt.Sprintf("%s%s  %s %s\n", cursor,
			tui.StyleNormal.Render(name),
			tui.StyleDim.Render("->"),
			assignment)
	}

	// System picker
	s += "\n"
	if len(m.assignSystems) > 0 {
		currentPath := ""
		if m.assignCursor < len(m.assignFiles) {
			currentPath = m.assignFiles[m.assignCursor]
		}
		ext := strings.ToLower(filepath.Ext(currentPath))
		s += tui.StyleDim.Render(fmt.Sprintf("Systems for %s:", ext)) + "\n"

		sysHeight := viewportHeight - fileHeight - 1
		if sysHeight < 3 {
			sysHeight = 3
		}
		sysStart := 0
		if m.assignSysCursor > sysHeight-2 {
			sysStart = m.assignSysCursor - sysHeight + 2
		}

		for i := sysStart; i < len(m.assignSystems) && i < sysStart+sysHeight; i++ {
			sysID := m.assignSystems[i]
			info, _ := systems.GetSystem(sysID)

			cursor := "    "
			if i == m.assignSysCursor && m.assignPicking {
				cursor = "  " + tui.StyleMenuCursor.String()
			}

			s += cursor + info.DisplayName + "\n"
		}
	} else {
		s += tui.StyleDim.Render("  No compatible systems found") + "\n"
	}

	s += "\n" + tui.StyleDim.Render("tab: switch focus  enter: assign  a: apply  s: skip  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

// --- Dedup filter methods ---

// dedupFlatItem represents one row in the dedup filter UI.
type dedupFlatItem struct {
	isHeader  bool
	groupIdx  int
	path      string
	groupName string
}

func (m *ManageScreen) buildDedupFlatList() []dedupFlatItem {
	var items []dedupFlatItem
	for gi, g := range m.variantGroups {
		info, _ := systems.GetSystem(g.System)
		items = append(items, dedupFlatItem{
			isHeader:  true,
			groupIdx:  gi,
			groupName: g.BaseName + " (" + info.DisplayName + ")",
		})
		for _, f := range g.Files {
			items = append(items, dedupFlatItem{
				isHeader: false,
				groupIdx: gi,
				path:     f.Path,
			})
		}
	}
	return items
}

func (m *ManageScreen) initDedupFilter() {
	m.dedupSelected = make(map[string]bool)
	m.dedupCursor = 0
	m.dedupFiltered = nil

	for _, g := range m.variantGroups {
		// Pre-select the first file (highest region priority, e.g. USA)
		for i, f := range g.Files {
			if i == 0 {
				m.dedupSelected[f.Path] = true
			}
		}
	}
}

func (m *ManageScreen) updateDedupFilter(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	items := m.buildDedupFlatList()
	maxIdx := len(items) - 1

	switch {
	case key.Matches(msg, tui.Keys.Back):
		m.phase = managePhaseReview
	case key.Matches(msg, tui.Keys.Up):
		if m.dedupCursor > 0 {
			m.dedupCursor--
		}
	case key.Matches(msg, tui.Keys.Down):
		if m.dedupCursor < maxIdx {
			m.dedupCursor++
		}
	case key.Matches(msg, tui.Keys.Space):
		if m.dedupCursor >= 0 && m.dedupCursor <= maxIdx {
			item := items[m.dedupCursor]
			if !item.isHeader {
				if m.dedupSelected[item.path] {
					delete(m.dedupSelected, item.path)
				} else {
					m.dedupSelected[item.path] = true
				}
			}
		}
	case key.Matches(msg, tui.Keys.Select): // 'a' key
		if m.dedupCursor >= 0 && m.dedupCursor <= maxIdx {
			item := items[m.dedupCursor]
			gi := item.groupIdx
			g := m.variantGroups[gi]
			allSelected := true
			for _, f := range g.Files {
				if !m.dedupSelected[f.Path] {
					allSelected = false
					break
				}
			}
			for _, f := range g.Files {
				if allSelected {
					delete(m.dedupSelected, f.Path)
				} else {
					m.dedupSelected[f.Path] = true
				}
			}
		}
	case msg.String() == "s":
		// Skip filtering entirely (keep all variants)
		m.dedupFiltered = nil
		return m.advanceFromDedup()
	case key.Matches(msg, tui.Keys.Enter):
		m.applyDedupFilter()
		return m.advanceFromDedup()
	}
	return m, nil
}

func (m *ManageScreen) applyDedupFilter() {
	var toRemove []string
	for _, g := range m.variantGroups {
		for _, f := range g.Files {
			if !m.dedupSelected[f.Path] {
				toRemove = append(toRemove, f.Path)
			}
		}
	}
	if len(toRemove) > 0 {
		m.dedupFiltered = toRemove
		m.scanResult.RemoveFiles(toRemove)
		m.buildSystemList()
	}
}

func (m *ManageScreen) advanceFromDedup() (tui.Screen, tea.Cmd) {
	if m.hasConvertibleFiles() {
		m.phase = managePhaseConverting
		return m, m.startConversion()
	}
	return m.advanceToSorting()
}

func (m *ManageScreen) viewDedupFilter() string {
	s := tui.StyleSubtitle.Render("Filter Duplicate Versions") + "\n\n"

	items := m.buildDedupFlatList()

	// Count stats
	totalFiles := 0
	selectedCount := 0
	for _, g := range m.variantGroups {
		totalFiles += len(g.Files)
		for _, f := range g.Files {
			if m.dedupSelected[f.Path] {
				selectedCount++
			}
		}
	}

	s += fmt.Sprintf("Found %d games with %d total variants\n",
		len(m.variantGroups), totalFiles)
	s += fmt.Sprintf("Keeping %d, archiving %d\n\n",
		selectedCount, totalFiles-selectedCount)

	// Render list with scrolling
	viewportHeight := m.height - 12
	if viewportHeight < 5 {
		viewportHeight = 5
	}
	startIdx := 0
	if m.dedupCursor > viewportHeight-3 {
		startIdx = m.dedupCursor - viewportHeight + 3
	}

	for i := startIdx; i < len(items) && i < startIdx+viewportHeight; i++ {
		item := items[i]
		if item.isHeader {
			cursor := "  "
			if i == m.dedupCursor {
				cursor = tui.StyleMenuCursor.String()
			}
			s += cursor + tui.StyleSelected.Render(item.groupName) + "\n"
		} else {
			cursor := "    "
			if i == m.dedupCursor {
				cursor = "  " + tui.StyleMenuCursor.String()
			}

			check := "[ ] "
			if m.dedupSelected[item.path] {
				check = tui.StyleSuccess.Render("[x] ")
			}

			name := filepath.Base(item.path)
			s += cursor + check + name + "\n"
		}
	}

	s += "\n" + tui.StyleDim.Render("space: toggle  a: toggle group  enter: confirm  s: skip  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (m *ManageScreen) ShortHelp() []key.Binding {
	return []key.Binding{tui.Keys.Up, tui.Keys.Down, tui.Keys.Enter, tui.Keys.Back}
}
