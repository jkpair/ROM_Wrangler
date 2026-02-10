package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/config"
)

// Screen is the interface all TUI screens implement.
type Screen interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string
	ShortHelp() []key.Binding
}

// ScreenFactory creates screens (set by main to break import cycle).
type ScreenFactory func(id ScreenID, cfg *config.Config, width, height int) Screen

// App is the root Bubbletea model.
type App struct {
	cfg           *config.Config
	screenStack   []Screen
	width, height int
	status        string
	factory       ScreenFactory
}

// NewApp creates a new root app model.
func NewApp(cfg *config.Config, factory ScreenFactory) *App {
	return &App{
		cfg:     cfg,
		factory: factory,
	}
}

func (a *App) Init() tea.Cmd {
	home := a.factory(ScreenHome, a.cfg, a.width, a.height)
	a.screenStack = []Screen{home}
	return home.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Pass to current screen
		if len(a.screenStack) > 0 {
			s, cmd := a.current().Update(msg)
			a.screenStack[len(a.screenStack)-1] = s
			return a, cmd
		}
		return a, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Quit) && len(a.screenStack) <= 1:
			return a, tea.Quit
		case key.Matches(msg, Keys.Back) && len(a.screenStack) > 1:
			a.screenStack = a.screenStack[:len(a.screenStack)-1]
			return a, nil
		}

	case NavigateMsg:
		screen := a.factory(msg.Screen, a.cfg, a.width, a.height)
		a.screenStack = append(a.screenStack, screen)
		return a, screen.Init()

	case NavigateBackMsg:
		if len(a.screenStack) > 1 {
			a.screenStack = a.screenStack[:len(a.screenStack)-1]
		}
		return a, nil

	case StatusMsg:
		a.status = msg.Text
		return a, nil

	case ErrorMsg:
		a.status = StyleError.Render("Error: " + msg.Error())
		return a, nil
	}

	if len(a.screenStack) > 0 {
		s, cmd := a.current().Update(msg)
		a.screenStack[len(a.screenStack)-1] = s
		return a, cmd
	}
	return a, nil
}

func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	var content string
	if len(a.screenStack) > 0 {
		content = a.current().View()
	}

	// Header
	header := StyleTitle.
		Width(a.width).
		Align(lipgloss.Center).
		Render("ROM Wrangler")

	// Status bar
	helpKeys := []string{}
	if len(a.screenStack) > 0 {
		for _, b := range a.current().ShortHelp() {
			helpKeys = append(helpKeys, StyleDim.Render(b.Help().Key)+" "+StyleGray(b.Help().Desc))
		}
	}
	statusLeft := lipgloss.JoinHorizontal(lipgloss.Top, joinWithSep(helpKeys, "  ")...)
	statusRight := a.status
	statusBar := lipgloss.NewStyle().
		Width(a.width).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				lipgloss.NewStyle().Width(a.width/2).Render(statusLeft),
				lipgloss.NewStyle().Width(a.width/2).Align(lipgloss.Right).Render(statusRight),
			),
		)

	// Layout
	contentHeight := a.height - lipgloss.Height(header) - lipgloss.Height(statusBar)
	body := lipgloss.NewStyle().
		Height(contentHeight).
		Width(a.width).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, header, body, statusBar)
}

func (a *App) current() Screen {
	return a.screenStack[len(a.screenStack)-1]
}

// StyleGray returns gray-styled text.
func StyleGray(s string) string {
	return StyleDim.Render(s)
}

func joinWithSep(items []string, sep string) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items)*2-1)
	for i, item := range items {
		if i > 0 {
			result = append(result, sep)
		}
		result = append(result, item)
	}
	return result
}
