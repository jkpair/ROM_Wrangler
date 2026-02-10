package components

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

// FileItem represents a file in the list.
type FileItem struct {
	Path     string
	Size     int64
	System   string
	Selected bool
}

// FileList displays a styled file list with selection support.
type FileList struct {
	Items    []FileItem
	Cursor   int
	Width    int
	Height   int
	Scroll   int
}

func NewFileList(width, height int) *FileList {
	return &FileList{Width: width, Height: height}
}

func (fl *FileList) View() string {
	if len(fl.Items) == 0 {
		return tui.StyleDim.Render("  No files")
	}

	maxVisible := fl.Height
	if maxVisible < 1 {
		maxVisible = 10
	}

	end := fl.Scroll + maxVisible
	if end > len(fl.Items) {
		end = len(fl.Items)
	}

	var s string
	for i := fl.Scroll; i < end; i++ {
		item := fl.Items[i]

		cursor := "  "
		if i == fl.Cursor {
			cursor = tui.StyleMenuCursor.String()
		}

		check := "[ ] "
		if item.Selected {
			check = tui.StyleSuccess.Render("[x] ")
		}

		name := filepath.Base(item.Path)
		size := formatSize(item.Size)
		sys := ""
		if item.System != "" {
			sys = tui.StyleDim.Render(" [" + item.System + "]")
		}

		line := fmt.Sprintf("%s%s%s %s%s", cursor, check, name,
			tui.StyleDim.Render(size), sys)
		s += line + "\n"
	}

	if len(fl.Items) > maxVisible {
		s += tui.StyleDim.Render(fmt.Sprintf("\n  %d/%d items", fl.Cursor+1, len(fl.Items)))
	}

	return lipgloss.NewStyle().Width(fl.Width).Render(s)
}

func (fl *FileList) CursorUp() {
	if fl.Cursor > 0 {
		fl.Cursor--
		if fl.Cursor < fl.Scroll {
			fl.Scroll = fl.Cursor
		}
	}
}

func (fl *FileList) CursorDown() {
	if fl.Cursor < len(fl.Items)-1 {
		fl.Cursor++
		maxVisible := fl.Height
		if maxVisible < 1 {
			maxVisible = 10
		}
		if fl.Cursor >= fl.Scroll+maxVisible {
			fl.Scroll = fl.Cursor - maxVisible + 1
		}
	}
}

func (fl *FileList) ToggleCurrent() {
	if fl.Cursor < len(fl.Items) {
		fl.Items[fl.Cursor].Selected = !fl.Items[fl.Cursor].Selected
	}
}

func (fl *FileList) SelectAll() {
	allSelected := true
	for _, item := range fl.Items {
		if !item.Selected {
			allSelected = false
			break
		}
	}
	for i := range fl.Items {
		fl.Items[i].Selected = !allSelected
	}
}

func (fl *FileList) SelectedCount() int {
	count := 0
	for _, item := range fl.Items {
		if item.Selected {
			count++
		}
	}
	return count
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
