package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// notesChrome is the modal's border, padding, title and blank line around
// the scrolling area.
const notesChrome = 8

// startNotesView opens the scrollable notes window for the selected task
// and fetches all of its notes.
func (m Model) startNotesView() (tea.Model, tea.Cmd) {
	r, ok := m.selected()
	if !ok || r.todo == nil {
		m.setStatus(i18n.T("tui.prompt.noSelection"), nil)
		return m, nil
	}

	m.mode, m.target, m.status = modeNotesView, *r.todo, ""
	m.notesView = viewport.New(viewport.WithWidth(m.notesWidth()), viewport.WithHeight(1))
	m.notesView.SetContent(color.Muted().Render(i18n.T("tui.loading")))

	ctx, svc, name := m.ctx, m.svc, r.todo.Path()
	return m, func() tea.Msg {
		_, entries, err := svc.Notes(ctx, name, 0)
		return notesMsg{name: name, entries: entries, full: true, err: err}
	}
}

// handleNotesViewKey scrolls the window; esc, q or enter close it.
func (m Model) handleNotesViewKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "enter":
		m.mode = modeNormal
		return m, nil
	}

	var cmd tea.Cmd
	m.notesView, cmd = m.notesView.Update(msg)
	return m, cmd
}

// setNotesContent fills the window and sizes it to its content, up to what
// fits on screen.
func (m *Model) setNotesContent(entries []tasks.NoteEntry, err error) {
	var content string
	switch {
	case err != nil:
		content = color.Error().Render(err.Error())
	case len(entries) == 0:
		content = color.Muted().Render(i18n.T("tasks.tasks.noNotesYet"))
	default:
		content = strings.Join(noteLines(entries, m.notesWidth()), "\n")
	}

	m.notesView.SetWidth(m.notesWidth())
	m.notesView.SetContent(content)
	m.notesView.SetHeight(m.notesHeight(lipgloss.Height(content)))
	m.notesView.GotoTop()
}

func (m Model) notesWidth() int {
	return modalWidth(m.width) - 6
}

func (m Model) notesHeight(contentHeight int) int {
	if m.height <= 0 {
		return contentHeight
	}
	return max(min(contentHeight, m.height-notesChrome), 1)
}

// noteLines renders one line per day ("2026-10-19  what I learned"),
// wrapping long notes under their text column.
func noteLines(entries []tasks.NoteEntry, width int) []string {
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		date := color.Accent().UnsetBold().Render(e.Date) + "  "

		var text string
		switch {
		case e.Backfilled:
			text = color.Muted().Render(i18n.T("tasks.tasks.backfilled"))
		case e.Notes == "":
			text = color.Muted().Render(i18n.T("tasks.tasks.noNotes"))
		default:
			text = lipgloss.NewStyle().Width(max(width-lipgloss.Width(date), 10)).Render(e.Notes)
		}
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, date, text))
	}
	return lines
}

// renderNotesModal is the window opened with n.
func (m Model) renderNotesModal(box lipgloss.Style) string {
	title := fmt.Sprintf(i18n.T("tasks.tasks.notesTitle"), m.target.Name)
	return box.Render(color.Accent().Render(title) + "\n\n" + m.notesView.View())
}
