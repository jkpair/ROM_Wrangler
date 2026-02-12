package screens

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type ReplayOSScreen struct {
	width, height int
}

func NewReplayOSScreen(width, height int) *ReplayOSScreen {
	return &ReplayOSScreen{width: width, height: height}
}

func (r *ReplayOSScreen) Init() tea.Cmd { return nil }

func (r *ReplayOSScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height
	case tea.KeyMsg:
		if key.Matches(msg, tui.Keys.Back) || key.Matches(msg, tui.Keys.Enter) {
			return r, func() tea.Msg { return tui.NavigateBackMsg{} }
		}
	}
	return r, nil
}

func (r *ReplayOSScreen) View() string {
	s := tui.StyleSubtitle.Render("About ReplayOS") + "\n\n"

	s += tui.StyleNormal.Render("ReplayOS") +
		tui.StyleDim.Render(" is a retro gaming operating system for the Raspberry Pi.") + "\n"
	s += tui.StyleDim.Render("Rom Wrangler is an independently maintained companion tool.") + "\n\n"

	s += tui.StyleSelected.Render("Website") + "\n"
	s += "  " + tui.StyleAccent.Render("https://www.replayos.com") + "\n"
	s += "  " + tui.StyleDim.Render("Visit for how-to guides, documentation, and community support.") + "\n\n"

	s += tui.StyleSelected.Render("Patreon") + "\n"
	s += "  " + tui.StyleAccent.Render("https://www.patreon.com/c/RePlayOS/home") + "\n"
	s += "  " + tui.StyleDim.Render("Download the latest version of ReplayOS and support development.") + "\n\n"

	s += tui.StyleDim.Render("esc: back")
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (r *ReplayOSScreen) ShortHelp() []key.Binding {
	return []key.Binding{tui.Keys.Back}
}
