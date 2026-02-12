package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type StatusBar struct {
	Width int
	Help  string
	Info  string
}

func NewStatusBar(width int) StatusBar {
	return StatusBar{Width: width}
}

func (s StatusBar) View() string {
	left := tui.StyleDim.Render(s.Help)
	right := tui.StyleDim.Render(s.Info)

	leftW := s.Width / 2
	rightW := s.Width - leftW

	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(leftW).Render(left),
		lipgloss.NewStyle().Width(rightW).Align(lipgloss.Right).Render(right),
	)
	return lipgloss.NewStyle().
		Width(s.Width).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(tui.ColorBorderDim).
		Render(row)
}
