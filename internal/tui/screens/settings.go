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
	sectionMenu     settingsSection = iota
	sectionGeneral                  // fields: source dirs, chdman, delete archive
	sectionTransfer                 // sub-menu: SFTP, USB, Concurrency
	sectionNetwork                  // fields: host, port, user, password, ROM path
	sectionUSB                      // fields: USB path
	sectionConcurrency              // fields: concurrency
)

type settingsField struct {
	label string
	value string
	input textinput.Model
}

type SettingsScreen struct {
	cfg           *config.Config
	width, height int

	section       settingsSection
	parentSection settingsSection
	sectionTitle  string
	menuCursor    int
	fieldCursor   int
	fields        []settingsField
	menuItems     []string
	saved         bool
	saveErr       error
}

var mainMenuItems = []string{"General", "Transfer", "Setup ROM Folders", "Setup BIOS Folders"}
var transferMenuItems = []string{"Network Settings", "USB Settings", "Concurrency"}

func NewSettingsScreen(cfg *config.Config, width, height int) *SettingsScreen {
	s := &SettingsScreen{
		cfg:       cfg,
		width:     width,
		height:    height,
		menuItems: mainMenuItems,
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
		case sectionTransfer:
			return s.updateTransferMenu(msg)
		default:
			return s.updateFields(msg)
		}
	}

	// Update active text input
	if s.section != sectionMenu && s.section != sectionTransfer && s.fieldCursor < len(s.fields) {
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
			s.parentSection = sectionMenu
			s.sectionTitle = "General"
			s.buildGeneralFields()
			s.fieldCursor = 0
			if len(s.fields) > 0 {
				s.fields[0].input.Focus()
				return s, s.fields[0].input.Cursor.BlinkCmd()
			}
		case 1:
			s.section = sectionTransfer
			s.menuItems = transferMenuItems
			s.menuCursor = 0
		case 2:
			return s, func() tea.Msg {
				return tui.NavigateMsg{Screen: tui.ScreenSetup}
			}
		case 3:
			return s, func() tea.Msg {
				return tui.NavigateMsg{Screen: tui.ScreenBIOS}
			}
		}
	}
	return s, nil
}

func (s *SettingsScreen) updateTransferMenu(msg tea.KeyMsg) (tui.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.Keys.Back):
		s.section = sectionMenu
		s.menuItems = mainMenuItems
		s.menuCursor = 1 // return to Transfer highlighted
	case key.Matches(msg, tui.Keys.Up):
		if s.menuCursor > 0 {
			s.menuCursor--
		}
	case key.Matches(msg, tui.Keys.Down):
		if s.menuCursor < len(s.menuItems)-1 {
			s.menuCursor++
		}
	case key.Matches(msg, tui.Keys.Enter):
		s.parentSection = sectionTransfer
		s.fieldCursor = 0
		switch s.menuCursor {
		case 0:
			s.section = sectionNetwork
			s.sectionTitle = "Network"
			s.buildNetworkFields()
		case 1:
			s.section = sectionUSB
			s.sectionTitle = "USB"
			s.buildUSBFields()
		case 2:
			s.section = sectionConcurrency
			s.sectionTitle = "Concurrency"
			s.buildConcurrencyFields()
		}
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
		s.fields = nil
		s.saved = false
		if s.parentSection == sectionTransfer {
			s.section = sectionTransfer
			s.menuItems = transferMenuItems
		} else {
			s.section = sectionMenu
			s.menuItems = mainMenuItems
		}
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
		s.makeField("Root Directory", strings.Join(s.cfg.SourceDirs, ", ")),
		s.makeField("chdman Path", s.cfg.ChdmanPath),
		s.makeField("Auto-delete Archive", fmt.Sprintf("%v", s.cfg.DeleteArchive)),
	}
}

func (s *SettingsScreen) buildNetworkFields() {
	s.fields = []settingsField{
		s.makeField("Host", s.cfg.Device.Host),
		s.makeField("Port", fmt.Sprintf("%d", s.cfg.Device.Port)),
		s.makeField("User", s.cfg.Device.User),
		s.makeField("Password", s.cfg.Device.Password),
		s.makeField("Root Path", s.cfg.Device.RootPath),
	}
	s.fields[3].input.EchoMode = textinput.EchoPassword
}

func (s *SettingsScreen) buildUSBFields() {
	s.fields = []settingsField{
		s.makeField("USB Path", s.cfg.Transfer.USBPath),
	}
}

func (s *SettingsScreen) buildConcurrencyFields() {
	s.fields = []settingsField{
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

	case sectionNetwork:
		s.cfg.Device.Host = s.fields[0].input.Value()
		fmt.Sscanf(s.fields[1].input.Value(), "%d", &s.cfg.Device.Port)
		s.cfg.Device.User = s.fields[2].input.Value()
		s.cfg.Device.Password = s.fields[3].input.Value()
		s.cfg.Device.RootPath = s.fields[4].input.Value()

	case sectionUSB:
		s.cfg.Transfer.USBPath = s.fields[0].input.Value()

	case sectionConcurrency:
		fmt.Sscanf(s.fields[0].input.Value(), "%d", &s.cfg.Transfer.Concurrency)
	}
}

func (s *SettingsScreen) View() string {
	switch s.section {
	case sectionMenu:
		return s.viewMenu("Settings", mainMenuItems)
	case sectionTransfer:
		return s.viewMenu("Transfer Settings", transferMenuItems)
	default:
		return s.viewFields()
	}
}

func (s *SettingsScreen) viewMenu(title string, items []string) string {
	content := tui.StyleSubtitle.Render(title) + "\n\n"

	for i, item := range items {
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
	content := tui.StyleSubtitle.Render(s.sectionTitle+" Settings") + "\n\n"

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

	content += "\n" + tui.StyleDim.Render("tab/\u2193: next field  ctrl+s: save  esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (s *SettingsScreen) ShortHelp() []key.Binding {
	if s.section == sectionMenu || s.section == sectionTransfer {
		return []key.Binding{tui.Keys.Up, tui.Keys.Down, tui.Keys.Enter, tui.Keys.Back}
	}
	return []key.Binding{tui.Keys.Tab, tui.Keys.Back}
}
