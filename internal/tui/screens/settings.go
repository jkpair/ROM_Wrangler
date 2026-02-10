package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type settingsSection int

const (
	sectionMenu settingsSection = iota
	sectionGeneral
	sectionDevice
	sectionScraping
	sectionTransfer
)

type settingsField struct {
	label string
	value string
	input textinput.Model
}

type SettingsScreen struct {
	cfg           *config.Config
	width, height int

	section     settingsSection
	menuCursor  int
	fieldCursor int
	fields      []settingsField
	menuItems   []string
	saved       bool
	saveErr     error
}

func NewSettingsScreen(cfg *config.Config, width, height int) *SettingsScreen {
	s := &SettingsScreen{
		cfg:    cfg,
		width:  width,
		height: height,
		menuItems: []string{
			"General",
			"Device",
			"Scraping",
			"Transfer",
		},
	}
	return s
}

func (s *SettingsScreen) Init() tea.Cmd { return nil }

func (s *SettingsScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height

	case tea.KeyMsg:
		switch s.section {
		case sectionMenu:
			return s.updateMenu(msg)
		default:
			return s.updateFields(msg)
		}
	}

	// Update active text input
	if s.section != sectionMenu && s.fieldCursor < len(s.fields) {
		var cmd tea.Cmd
		s.fields[s.fieldCursor].input, cmd = s.fields[s.fieldCursor].input.Update(msg)
		return s, cmd
	}

	return s, nil
}

func (s *SettingsScreen) updateMenu(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		return s, func() tea.Msg { return tui.NavigateBackMsg{} }
	case key.Matches(msg, tui.Keys.Up):
		if s.menuCursor > 0 {
			s.menuCursor--
		}
	case key.Matches(msg, tui.Keys.Down):
		if s.menuCursor < len(s.menuItems)-1 {
			s.menuCursor++
		}
	case key.Matches(msg, tui.Keys.Enter):
		switch s.menuCursor {
		case 0:
			s.section = sectionGeneral
			s.buildGeneralFields()
		case 1:
			s.section = sectionDevice
			s.buildDeviceFields()
		case 2:
			s.section = sectionScraping
			s.buildScrapingFields()
		case 3:
			s.section = sectionTransfer
			s.buildTransferFields()
		}
		s.fieldCursor = 0
		if len(s.fields) > 0 {
			s.fields[0].input.Focus()
			return s, s.fields[0].input.Cursor.BlinkCmd()
		}
	}
	return s, nil
}

func (s *SettingsScreen) updateFields(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		s.section = sectionMenu
		s.fields = nil
		s.saved = false
	case msg.Type == tea.KeyTab || msg.Type == tea.KeyDown:
		if s.fieldCursor < len(s.fields)-1 {
			s.fields[s.fieldCursor].input.Blur()
			s.fieldCursor++
			s.fields[s.fieldCursor].input.Focus()
			return s, s.fields[s.fieldCursor].input.Cursor.BlinkCmd()
		}
	case msg.Type == tea.KeyShiftTab || msg.Type == tea.KeyUp:
		if s.fieldCursor > 0 {
			s.fields[s.fieldCursor].input.Blur()
			s.fieldCursor--
			s.fields[s.fieldCursor].input.Focus()
			return s, s.fields[s.fieldCursor].input.Cursor.BlinkCmd()
		}
	case msg.Type == tea.KeyCtrlS:
		s.applyFields()
		s.saveErr = config.Save(s.cfg, "")
		s.saved = true
	default:
		if s.fieldCursor < len(s.fields) {
			var cmd tea.Cmd
			s.fields[s.fieldCursor].input, cmd = s.fields[s.fieldCursor].input.Update(msg)
			return s, cmd
		}
	}
	return s, nil
}

func (s *SettingsScreen) buildGeneralFields() {
	s.fields = []settingsField{
		s.makeField("Source Directories", strings.Join(s.cfg.SourceDirs, ", ")),
		s.makeField("chdman Path", s.cfg.ChdmanPath),
		s.makeField("Auto-delete Archive", fmt.Sprintf("%v", s.cfg.DeleteArchive)),
	}
}

func (s *SettingsScreen) buildDeviceFields() {
	s.fields = []settingsField{
		s.makeField("Device Type", s.cfg.Device.Type),
		s.makeField("Host", s.cfg.Device.Host),
		s.makeField("Port", fmt.Sprintf("%d", s.cfg.Device.Port)),
		s.makeField("User", s.cfg.Device.User),
		s.makeField("Password", s.cfg.Device.Password),
		s.makeField("ROM Path", s.cfg.Device.ROMPath),
	}
	// Mask password
	s.fields[4].input.EchoMode = textinput.EchoPassword
}

func (s *SettingsScreen) buildScrapingFields() {
	s.fields = []settingsField{
		s.makeField("ScreenScraper User", s.cfg.Scraping.ScreenScraperUser),
		s.makeField("ScreenScraper Pass", s.cfg.Scraping.ScreenScraperPass),
		s.makeField("DAT Directories", strings.Join(s.cfg.Scraping.DATDirs, ", ")),
	}
	s.fields[1].input.EchoMode = textinput.EchoPassword
}

func (s *SettingsScreen) buildTransferFields() {
	s.fields = []settingsField{
		s.makeField("Method", s.cfg.Transfer.Method),
		s.makeField("Sync Mode", fmt.Sprintf("%v", s.cfg.Transfer.SyncMode)),
		s.makeField("USB Path", s.cfg.Transfer.USBPath),
		s.makeField("Concurrency", fmt.Sprintf("%d", s.cfg.Transfer.Concurrency)),
	}
}

func (s *SettingsScreen) makeField(label, value string) settingsField {
	ti := textinput.New()
	ti.SetValue(value)
	ti.Prompt = ""
	ti.CharLimit = 256
	ti.Width = 40
	return settingsField{label: label, value: value, input: ti}
}

func (s *SettingsScreen) applyFields() {
	switch s.section {
	case sectionGeneral:
		dirs := strings.Split(s.fields[0].input.Value(), ",")
		s.cfg.SourceDirs = nil
		for _, d := range dirs {
			d = strings.TrimSpace(d)
			if d != "" {
				s.cfg.SourceDirs = append(s.cfg.SourceDirs, d)
			}
		}
		s.cfg.ChdmanPath = s.fields[1].input.Value()
		s.cfg.DeleteArchive = s.fields[2].input.Value() == "true"

	case sectionDevice:
		s.cfg.Device.Type = s.fields[0].input.Value()
		s.cfg.Device.Host = s.fields[1].input.Value()
		fmt.Sscanf(s.fields[2].input.Value(), "%d", &s.cfg.Device.Port)
		s.cfg.Device.User = s.fields[3].input.Value()
		s.cfg.Device.Password = s.fields[4].input.Value()
		s.cfg.Device.ROMPath = s.fields[5].input.Value()

	case sectionScraping:
		s.cfg.Scraping.ScreenScraperUser = s.fields[0].input.Value()
		s.cfg.Scraping.ScreenScraperPass = s.fields[1].input.Value()
		dirs := strings.Split(s.fields[2].input.Value(), ",")
		s.cfg.Scraping.DATDirs = nil
		for _, d := range dirs {
			d = strings.TrimSpace(d)
			if d != "" {
				s.cfg.Scraping.DATDirs = append(s.cfg.Scraping.DATDirs, d)
			}
		}

	case sectionTransfer:
		s.cfg.Transfer.Method = s.fields[0].input.Value()
		s.cfg.Transfer.SyncMode = s.fields[1].input.Value() == "true"
		s.cfg.Transfer.USBPath = s.fields[2].input.Value()
		fmt.Sscanf(s.fields[3].input.Value(), "%d", &s.cfg.Transfer.Concurrency)
	}
}

func (s *SettingsScreen) View() string {
	switch s.section {
	case sectionMenu:
		return s.viewMenu()
	default:
		return s.viewFields()
	}
}

func (s *SettingsScreen) viewMenu() string {
	content := tui.StyleSubtitle.Render("Settings") + "\n\n"

	for i, item := range s.menuItems {
		cursor := "  "
		style := tui.StyleNormal
		if i == s.menuCursor {
			cursor = tui.StyleMenuCursor.String()
			style = tui.StyleSelected
		}
		content += cursor + style.Render(item) + "\n"
	}

	content += "\n" + tui.StyleDim.Render("Config: "+config.DefaultPath())
	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (s *SettingsScreen) viewFields() string {
	title := s.menuItems[s.menuCursor]
	content := tui.StyleSubtitle.Render(title+" Settings") + "\n\n"

	for i, field := range s.fields {
		labelStyle := tui.StyleDim
		if i == s.fieldCursor {
			labelStyle = tui.StyleSubtitle
		}
		content += labelStyle.Render(field.label+":") + "\n"
		content += "  " + field.input.View() + "\n\n"
	}

	if s.saved {
		if s.saveErr != nil {
			content += tui.StyleError.Render("Save failed: "+s.saveErr.Error()) + "\n"
		} else {
			content += tui.StyleSuccess.Render("Settings saved!") + "\n"
		}
	}

	content += "\n" + tui.StyleDim.Render("tab/↓: next field  ctrl+s: save  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (s *SettingsScreen) ShortHelp() []key.Binding {
	if s.section == sectionMenu {
		return []key.Binding{tui.Keys.Up, tui.Keys.Down, tui.Keys.Enter, tui.Keys.Back}
	}
	return []key.Binding{tui.Keys.Tab, tui.Keys.Back}
}
