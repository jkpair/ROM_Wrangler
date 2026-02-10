package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

// ProgressItem represents a single item in the progress panel.
type ProgressItem struct {
	Name    string
	Percent float64
	Done    bool
	Error   error
}

// ProgressPanel displays multi-file progress.
type ProgressPanel struct {
	Title    string
	Items    []ProgressItem
	Width    int
	BarWidth int
}

func NewProgressPanel(title string, width int) *ProgressPanel {
	barWidth := 30
	if width > 0 && width-20 < barWidth {
		barWidth = width - 20
	}
	if barWidth < 10 {
		barWidth = 10
	}
	return &ProgressPanel{
		Title:    title,
		Width:    width,
		BarWidth: barWidth,
	}
}

func (p *ProgressPanel) SetItem(index int, item ProgressItem) {
	for len(p.Items) <= index {
		p.Items = append(p.Items, ProgressItem{})
	}
	p.Items[index] = item
}

func (p *ProgressPanel) View() string {
	s := tui.StyleSubtitle.Render(p.Title) + "\n\n"

	for _, item := range p.Items {
		var status string
		if item.Error != nil {
			status = tui.StyleError.Render("FAIL")
		} else if item.Done {
			status = tui.StyleSuccess.Render(" OK ")
		} else {
			status = renderBar(item.Percent, p.BarWidth)
		}

		name := item.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		s += fmt.Sprintf("%-30s %s\n", name, status)
	}

	return lipgloss.NewStyle().Width(p.Width).Render(s)
}

func renderBar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := tui.StyleProgressFilled.Render(strings.Repeat("█", filled))
	bar += tui.StyleProgressEmpty.Render(strings.Repeat("░", empty))
	return fmt.Sprintf("%s %5.1f%%", bar, pct)
}
