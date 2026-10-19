package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// chromeHeight is the number of lines used by the tab bar and help footer.
const chromeHeight = 3

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")
	body := m.renderBody()
	b.WriteString(body)
	b.WriteString("\n")

	// pad so the help footer sticks to the bottom of the screen
	if h := m.bodyHeight(); h > 0 {
		if gap := h - lipgloss.Height(body); gap > 0 {
			b.WriteString(strings.Repeat("\n", gap))
		}
	}
	b.WriteString(m.help.View(m.keys))

	return b.String()
}

func (m Model) renderTabs() string {
	titles := make([]string, tabCount)
	for t := tab(0); t < tabCount; t++ {
		if t == m.active {
			titles[t] = color.Title().Underline(true).Render(t.title())
		} else {
			titles[t] = color.Muted().Render(t.title())
		}
	}
	return " " + strings.Join(titles, "   ")
}

func (m Model) renderBody() string {
	st := m.tabs[m.active]

	switch {
	case st.err != nil:
		return color.Error().Render(" " + st.err.Error())
	case !st.loaded:
		return color.Muted().Render(" " + i18n.T("tui.loading"))
	case len(st.rows) == 0:
		return color.Muted().Render(" " + emptyMessage(m.active))
	}

	start, end := visibleRange(len(st.rows), st.cursor, m.bodyHeight())

	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		lines = append(lines, renderRow(st.rows[i], i == st.cursor))
	}
	return strings.Join(lines, "\n")
}

// bodyHeight is how many rows fit between the tab bar and the footer; 0
// means the size is unknown yet and everything is shown.
func (m Model) bodyHeight() int {
	if m.height == 0 {
		return 0
	}
	return max(m.height-chromeHeight, 1)
}

// visibleRange scrolls just enough to keep the cursor on screen.
func visibleRange(total, cursor, height int) (int, int) {
	if height <= 0 || total <= height {
		return 0, total
	}
	start := 0
	if cursor >= height {
		start = cursor - height + 1
	}
	return start, start + height
}

func emptyMessage(t tab) string {
	switch t {
	case tabToday:
		return i18n.T("tasks.info.noPendingTasks")
	case tabDump:
		return i18n.T("tasks.info.noDump")
	default:
		return i18n.T("tui.empty")
	}
}

func renderRow(r row, selected bool) string {
	switch {
	case r.header != "":
		return color.Title().Render(" " + r.header)
	case r.todo != nil:
		return cursorPrefix(selected) + renderTodoLine(*r.todo, selected)
	default:
		return cursorPrefix(selected) + renderDumpLine(*r.dump, selected)
	}
}

func cursorPrefix(selected bool) string {
	if selected {
		return color.Title().Render(" › ")
	}
	return "   "
}

// renderTodoLine is the compact one-line form of a task.
func renderTodoLine(todo tasks.TodoView, selected bool) string {
	icon, state := "○", color.Muted()
	switch {
	case todo.Done():
		icon, state = "✔", color.Success()
	case tasks.IsExpiringSoon(todo.ExpiresAt) && !todo.IsDaily:
		state = color.Warning()
	}

	name := lipgloss.NewStyle().Bold(selected).Render(todo.Name)

	meta := make([]string, 0, 2)
	if !todo.IsDaily {
		meta = append(meta, todo.ExpiresAt.Format("02 Jan 15:04"))
	}
	if todo.Priority != "" {
		meta = append(meta, todo.Priority)
	}

	line := state.Render(icon) + " " + name
	if len(meta) > 0 {
		line += color.Muted().Render("  · " + strings.Join(meta, " · "))
	}
	return line
}

func renderDumpLine(d tasks.DumpEntry, selected bool) string {
	line := lipgloss.NewStyle().Bold(selected).Render(d.Name)
	if !d.CreatedAt.IsZero() {
		line += color.Muted().Render("  · " + d.CreatedAt.Format("02 Jan 15:04"))
	}
	return line
}
