package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

type Header struct {
	Width      int
	Breadcrumb []string
}

func NewHeader(width int, breadcrumb ...string) Header {
	return Header{Width: width, Breadcrumb: breadcrumb}
}

func (h Header) View() string {
	title := tui.StyleTitle.Render("ROM Wrangler")

	var crumb string
	if len(h.Breadcrumb) > 0 {
		parts := make([]string, len(h.Breadcrumb))
		for i, p := range h.Breadcrumb {
			if i == len(h.Breadcrumb)-1 {
				parts[i] = tui.StyleSubtitle.Render(p)
			} else {
				parts[i] = tui.StyleDim.Render(p)
			}
		}
		crumb = " " + tui.StyleDim.Render("›") + " " + strings.Join(parts, " "+tui.StyleDim.Render("›")+" ")
	}

	return lipgloss.NewStyle().
		Width(h.Width).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(tui.ColorDimGray).
		Render(title + crumb)
}
