package screens

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type menuItem struct {
	title string
	desc  string
	screen tui.ScreenID
}

type HomeScreen struct {
	cfg           *config.Config
	items         []menuItem
	cursor        int
	width, height int
}

func NewHomeScreen(cfg *config.Config, width, height int) *HomeScreen {
	return &HomeScreen{
		cfg:    cfg,
		width:  width,
		height: height,
		items: []menuItem{
			{title: "Manage ROMs", desc: "Scan, identify, and organize ROM files", screen: tui.ScreenManage},
			{title: "Convert Files", desc: "Convert disc images to CHD format", screen: tui.ScreenConvert},
			{title: "Transfer", desc: "Send files to your gaming device", screen: tui.ScreenTransfer},
			{title: "Setup Folders", desc: "Generate system folder structure", screen: tui.ScreenSetup},
			{title: "Settings", desc: "Configure devices, paths, and options", screen: tui.ScreenSettings},
		},
	}
}

func (h *HomeScreen) Init() tea.Cmd { return nil }

func (h *HomeScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Up):
			if h.cursor > 0 {
				h.cursor--
			}
		case key.Matches(msg, tui.Keys.Down):
			if h.cursor < len(h.items)-1 {
				h.cursor++
			}
		case key.Matches(msg, tui.Keys.Enter):
			return h, func() tea.Msg {
				return tui.NavigateMsg{Screen: h.items[h.cursor].screen}
			}
		case key.Matches(msg, tui.Keys.Quit):
			return h, tea.Quit
		}
	}
	return h, nil
}

func (h *HomeScreen) View() string {
	s := "\n"
	s += tui.StyleSubtitle.Render("What would you like to do?") + "\n\n"

	for i, item := range h.items {
		cursor := "  "
		titleStyle := tui.StyleNormal
		if i == h.cursor {
			cursor = tui.StyleMenuCursor.String()
			titleStyle = tui.StyleSelected
		}

		title := titleStyle.Render(item.title)
		desc := tui.StyleDim.Render(item.desc)
		s += cursor + title + "\n"
		s += "  " + desc + "\n\n"
	}

	return lipgloss.NewStyle().Padding(0, 2).Render(s)
}

func (h *HomeScreen) ShortHelp() []key.Binding {
	return []key.Binding{tui.Keys.Up, tui.Keys.Down, tui.Keys.Enter, tui.Keys.Quit}
}
