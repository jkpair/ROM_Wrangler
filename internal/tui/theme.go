package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Primary palette
	ColorViolet  = lipgloss.Color("#7C3AED")
	ColorCyan    = lipgloss.Color("#06B6D4")
	ColorGreen   = lipgloss.Color("#22C55E")
	ColorAmber   = lipgloss.Color("#F59E0B")
	ColorRed     = lipgloss.Color("#EF4444")
	ColorWhite   = lipgloss.Color("#F8FAFC")
	ColorGray    = lipgloss.Color("#94A3B8")
	ColorDimGray = lipgloss.Color("#475569")
	ColorBg      = lipgloss.Color("#0F172A")
	ColorSurface = lipgloss.Color("#1E293B")

	// Text styles
	StyleTitle = lipgloss.NewStyle().
			Foreground(ColorViolet).
			Bold(true)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(ColorCyan)

	StyleNormal = lipgloss.NewStyle().
			Foreground(ColorWhite)

	StyleDim = lipgloss.NewStyle().
			Foreground(ColorGray)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StyleWarning = lipgloss.NewStyle().
			Foreground(ColorAmber)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorRed)

	// UI element styles
	StyleSelected = lipgloss.NewStyle().
			Foreground(ColorViolet).
			Bold(true)

	StyleMenuCursor = lipgloss.NewStyle().
			Foreground(ColorCyan).
			SetString("▸ ")

	StyleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDimGray)

	StyleActiveBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorViolet)

	// Progress bar
	StyleProgressFilled = lipgloss.NewStyle().
				Foreground(ColorViolet)

	StyleProgressEmpty = lipgloss.NewStyle().
				Foreground(ColorDimGray)
)
