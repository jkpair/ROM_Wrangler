package screens

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

// FileSelectedMsg is sent when a file/directory is selected.
type FileSelectedMsg struct {
	Path string
}

type fileEntry struct {
	name  string
	isDir bool
}

// FileBrowserScreen allows browsing and selecting files/directories.
type FileBrowserScreen struct {
	width, height int
	currentDir    string
	entries       []fileEntry
	cursor        int
	scroll        int
	dirOnly       bool
	extensions    []string // filter by extensions, empty = all
	err           error
}

func NewFileBrowserScreen(startDir string, dirOnly bool, extensions []string, width, height int) *FileBrowserScreen {
	fb := &FileBrowserScreen{
		width:      width,
		height:     height,
		currentDir: startDir,
		dirOnly:    dirOnly,
		extensions: extensions,
	}
	fb.loadDir()
	return fb
}

func (fb *FileBrowserScreen) Init() tea.Cmd { return nil }

func (fb *FileBrowserScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		fb.width = msg.Width
		fb.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Back):
			return fb, func() tea.Msg { return tui.NavigateBackMsg{} }
		case key.Matches(msg, tui.Keys.Up):
			if fb.cursor > 0 {
				fb.cursor--
				if fb.cursor < fb.scroll {
					fb.scroll = fb.cursor
				}
			}
		case key.Matches(msg, tui.Keys.Down):
			if fb.cursor < len(fb.entries)-1 {
				fb.cursor++
				maxVisible := fb.height - 8
				if fb.cursor >= fb.scroll+maxVisible {
					fb.scroll = fb.cursor - maxVisible + 1
				}
			}
		case key.Matches(msg, tui.Keys.Enter):
			if fb.cursor < len(fb.entries) {
				entry := fb.entries[fb.cursor]
				if entry.name == ".." {
					fb.currentDir = filepath.Dir(fb.currentDir)
					fb.loadDir()
					fb.cursor = 0
					fb.scroll = 0
				} else if entry.isDir {
					path := filepath.Join(fb.currentDir, entry.name)
					if fb.dirOnly {
						return fb, func() tea.Msg {
							return FileSelectedMsg{Path: path}
						}
					}
					fb.currentDir = path
					fb.loadDir()
					fb.cursor = 0
					fb.scroll = 0
				} else {
					path := filepath.Join(fb.currentDir, entry.name)
					return fb, func() tea.Msg {
						return FileSelectedMsg{Path: path}
					}
				}
			}
		case msg.String() == "s" && fb.dirOnly:
			// Select current directory
			return fb, func() tea.Msg {
				return FileSelectedMsg{Path: fb.currentDir}
			}
		}
	}
	return fb, nil
}

func (fb *FileBrowserScreen) loadDir() {
	fb.entries = nil
	fb.err = nil

	entries, err := os.ReadDir(fb.currentDir)
	if err != nil {
		fb.err = err
		return
	}

	// Always add parent directory
	if fb.currentDir != "/" {
		fb.entries = append(fb.entries, fileEntry{name: "..", isDir: true})
	}

	// Separate dirs and files, sort each
	var dirs, files []fileEntry
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue // skip hidden files
		}

		if e.IsDir() {
			dirs = append(dirs, fileEntry{name: e.Name(), isDir: true})
		} else if !fb.dirOnly {
			if len(fb.extensions) > 0 {
				ext := strings.ToLower(filepath.Ext(e.Name()))
				matched := false
				for _, filterExt := range fb.extensions {
					if ext == filterExt {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}
			files = append(files, fileEntry{name: e.Name(), isDir: false})
		}
	}

	sort.Slice(dirs, func(i, j int) bool { return dirs[i].name < dirs[j].name })
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })

	fb.entries = append(fb.entries, dirs...)
	fb.entries = append(fb.entries, files...)
}

func (fb *FileBrowserScreen) View() string {
	s := tui.StyleSubtitle.Render("Browse Files") + "\n"
	s += tui.StyleDim.Render(fb.currentDir) + "\n\n"

	if fb.err != nil {
		s += tui.StyleError.Render("Error: "+fb.err.Error()) + "\n"
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

	maxVisible := fb.height - 8
	if maxVisible < 5 {
		maxVisible = 5
	}

	end := fb.scroll + maxVisible
	if end > len(fb.entries) {
		end = len(fb.entries)
	}

	for i := fb.scroll; i < end; i++ {
		entry := fb.entries[i]
		cursor := "  "
		if i == fb.cursor {
			cursor = tui.StyleMenuCursor.String()
		}

		name := entry.name
		if entry.isDir {
			name = tui.StyleSubtitle.Render(name + "/")
		}

		s += cursor + name + "\n"
	}

	if len(fb.entries) > maxVisible {
		s += tui.StyleDim.Render(fmt.Sprintf("\n(%d items)", len(fb.entries)))
	}

	help := "enter: open/select  esc: back"
	if fb.dirOnly {
		help += "  s: select this dir"
	}
	s += "\n" + tui.StyleDim.Render(help)

	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (fb *FileBrowserScreen) ShortHelp() []key.Binding {
	return []key.Binding{tui.Keys.Up, tui.Keys.Down, tui.Keys.Enter, tui.Keys.Back}
}
