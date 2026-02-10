package components

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurlmarx/romwrangler/internal/tui"
)

// ConfirmResult is sent when the user responds to a confirmation dialog.
type ConfirmResult struct {
	ID        string
	Confirmed bool
}

// ConfirmModal is a confirmation dialog overlay.
type ConfirmModal struct {
	ID      string
	Title   string
	Message string
	cursor  int // 0=Yes, 1=No
	Active  bool
}

func NewConfirmModal(id, title, message string) ConfirmModal {
	return ConfirmModal{
		ID:      id,
		Title:   title,
		Message: message,
		cursor:  1, // default to No
		Active:  true,
	}
}

func (m ConfirmModal) Update(msg tea.Msg) (ConfirmModal, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, tui.Keys.Back):
			m.Active = false
			return m, func() tea.Msg {
				return ConfirmResult{ID: m.ID, Confirmed: false}
			}
		case msg.String() == "left" || msg.String() == "h":
			m.cursor = 0
		case msg.String() == "right" || msg.String() == "l":
			m.cursor = 1
		case key.Matches(msg, tui.Keys.Enter):
			m.Active = false
			return m, func() tea.Msg {
				return ConfirmResult{ID: m.ID, Confirmed: m.cursor == 0}
			}
		}
	}

	return m, nil
}

func (m ConfirmModal) View() string {
	if !m.Active {
		return ""
	}

	title := tui.StyleTitle.Render(m.Title)
	message := tui.StyleNormal.Render(m.Message)

	yesStyle := tui.StyleDim
	noStyle := tui.StyleDim
	if m.cursor == 0 {
		yesStyle = tui.StyleSelected
	} else {
		noStyle = tui.StyleSelected
	}

	buttons := yesStyle.Render("[ Yes ]") + "  " + noStyle.Render("[ No ]")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		message,
		"",
		buttons,
	)

	return tui.StyleActiveBorder.
		Padding(1, 3).
		Render(content)
}
