package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Everforest Dark Hard palette - core colors
	ColorFg           = lipgloss.Color("#D3C6AA") // Default foreground text
	ColorBg           = lipgloss.Color("#1E2326") // Hard background (deepest)
ColorBgSurface    = lipgloss.Color("#272E33") // Slightly lighter surface / panels
ColorBgDim        = lipgloss.Color("#1E2326") // Often same as bg for hard variant
ColorGray         = lipgloss.Color("#859289") // Dim / comment gray
ColorGreen        = lipgloss.Color("#A7C080") // Main success / green accent
ColorAqua         = lipgloss.Color("#83C092") // Cyan-ish green for secondary accents
ColorYellow       = lipgloss.Color("#DBBC7F") // Warm yellow / warning
ColorOrange       = lipgloss.Color("#E69875") // Orange for operators / attention
ColorRed          = lipgloss.Color("#E67E80") // Error / delete
ColorPurple       = lipgloss.Color("#D699B6") // Violet / special (less used)

// Additional useful shades
ColorStatusline   = lipgloss.Color("#A7C080") // Active status / green tint
ColorBorderDim    = lipgloss.Color("#475258") // Subtle borders
ColorBorderActive = lipgloss.Color("#A7C080") // Active border accent

// Text styles
StyleTitle = lipgloss.NewStyle().
Foreground(ColorGreen).
Bold(true)

StyleSubtitle = lipgloss.NewStyle().
Foreground(ColorAqua)

StyleNormal = lipgloss.NewStyle().
Foreground(ColorFg)

StyleDim = lipgloss.NewStyle().
Foreground(ColorGray)

StyleSuccess = lipgloss.NewStyle().
Foreground(ColorGreen)

StyleWarning = lipgloss.NewStyle().
Foreground(ColorYellow)

StyleError = lipgloss.NewStyle().
Foreground(ColorRed)

StyleAccent = lipgloss.NewStyle(). // Replaces old StyleCyan
Foreground(ColorAqua)

// UI element styles
StyleSelected = lipgloss.NewStyle().
Foreground(ColorGreen).
Bold(true)

StyleMenuCursor = lipgloss.NewStyle().
Foreground(ColorAqua).
SetString("▸ ")

StyleBorder = lipgloss.NewStyle().
Border(lipgloss.RoundedBorder()).
BorderForeground(ColorBorderDim)

StyleActiveBorder = lipgloss.NewStyle().
Border(lipgloss.RoundedBorder()).
BorderForeground(ColorBorderActive)

// Progress bar
StyleProgressFilled = lipgloss.NewStyle().
Foreground(ColorGreen)

StyleProgressEmpty = lipgloss.NewStyle().
Foreground(ColorBorderDim)
)
