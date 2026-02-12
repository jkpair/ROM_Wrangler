package screens

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

const logo = "KurlMarx Utilities\n" + "" +
	"▄▄▄▄▄▄▄\n" +
	"███▀▀███▄\n" +
	"███▄▄███▀ ▄███▄ ███▄███▄\n" +
	"███▀▀██▄  ██ ██ ██ ██ ██\n" +
	"███  ▀███ ▀███▀ ██ ██ ██\n" +
	"\n" +
	"▄▄▄▄  ▄▄▄  ▄▄▄▄                        ▄▄\n" +
	"▀███  ███  ███▀                        ██\n" +
	" ███  ███  ███ ████▄  ▀▀█▄ ████▄ ▄████ ██ ▄█▀█▄ ████▄\n" +
	" ███▄▄███▄▄███ ██ ▀▀ ▄█▀██ ██ ██ ██ ██ ██ ██▄█▀ ██ ▀▀\n" +
	"  ▀████▀████▀  ██    ▀█▄██ ██ ██ ▀████ ██ ▀█▄▄▄ ██\n" +
	"                                    ██\n" +
	"                                  ▀▀▀"

const tagline = "                              -FOR USE WITH REPLAYOS-"

type menuItem struct {
	title  string
	desc   string
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
			{title: "Manage ROMs", desc: "Full pipeline: scan, organize, convert, and transfer", screen: tui.ScreenManage},
			{title: "Decompress Files", desc: "Extract .zip, .7z, .rar, and .ecm archives", screen: tui.ScreenDecompress},
			{title: "Archive Redundant Files", desc: "Clean up duplicates, superseded disc images, and spent archives", screen: tui.ScreenArchive},
			{title: "Convert Files", desc: "Convert disc images to CHD format", screen: tui.ScreenConvert},
			{title: "Generate M3U Files", desc: "Generate M3U playlists for multi-disc games", screen: tui.ScreenM3U},
			{title: "Transfer", desc: "Send files to your gaming device", screen: tui.ScreenTransfer},
			{title: "Settings", desc: "Configure devices, paths, and options", screen: tui.ScreenSettings},
			{title: "About ReplayOS", desc: "Learn more about ReplayOS and support the project", screen: tui.ScreenReplayOS},
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
	logoStyle := lipgloss.NewStyle().Foreground(tui.ColorGray)
	s := logoStyle.Render(logo) + "\n"
	s += tui.StyleDim.Render(tagline) + "\n\n"

	for i, item := range h.items {
		cursor := "  "
		titleStyle := tui.StyleNormal
		if i == h.cursor {
			cursor = tui.StyleMenuCursor.String()
			titleStyle = tui.StyleSelected
		}
		s += cursor + titleStyle.Render(item.title) + "\n"
	}

	// Show description of selected item
	s += "\n" + tui.StyleDim.Render(h.items[h.cursor].desc)

	return lipgloss.NewStyle().Padding(0, 2).Render(s)
}

func (h *HomeScreen) ShortHelp() []key.Binding {
	return []key.Binding{tui.Keys.Up, tui.Keys.Down, tui.Keys.Enter, tui.Keys.Quit}
}
