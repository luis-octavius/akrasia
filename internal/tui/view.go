package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// detailsMinWidth is the narrowest terminal that shows details beside the
// list; below it they go under the list.
const detailsMinWidth = 80

// View renders the whole screen in the alternate screen buffer.
func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m Model) render() string {
	header := m.renderTabs() + "\n"
	footer := m.renderFooter()

	bodyHeight := 0
	if m.height > 0 {
		bodyHeight = max(m.height-lipgloss.Height(header)-lipgloss.Height(footer)-1, 1)
	}

	var body string
	if m.mode == modeAddTodo {
		body = m.renderForm()
	} else {
		body = m.renderMain(bodyHeight)
	}

	// pad so the footer sticks to the bottom of the screen
	if gap := bodyHeight - lipgloss.Height(body); gap > 0 {
		body += strings.Repeat("\n", gap)
	}

	return header + "\n" + body + "\n" + footer
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
	line := " " + strings.Join(titles, "   ")

	var filters []string
	if m.search != "" {
		filters = append(filters, fmt.Sprintf("%s: %q", i18n.T("tui.status.filter"), m.search))
	}
	if m.priority != "" {
		filters = append(filters, i18n.T("tui.status.priority")+": "+m.priority)
	}
	if len(filters) > 0 {
		line += color.Muted().Render("   [" + strings.Join(filters, " · ") + "]")
	}
	return line
}

func (m Model) renderFooter() string {
	var lines []string

	switch m.mode {
	case modeNotes:
		lines = append(lines, m.renderPrompt(fmt.Sprintf(i18n.T("tui.prompt.notes"), m.target.Name)))
	case modeSearch:
		lines = append(lines, m.renderPrompt(i18n.T("tui.prompt.search")))
	case modeAddDump:
		lines = append(lines, m.renderPrompt(i18n.T("tui.prompt.dumpName")))
	case modeConfirmDelete:
		lines = append(lines, color.Warning().Render(" "+fmt.Sprintf(i18n.T("tui.prompt.delete"), m.target.Name)))
	}

	if m.status != "" {
		style := color.Success()
		if m.statusErr {
			style = color.Error()
		}
		lines = append(lines, style.Render(" "+m.status))
	}

	if m.mode == modeNormal {
		lines = append(lines, m.help.View(m.keys))
	} else {
		lines = append(lines, color.Muted().Render(m.promptHelp()))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderPrompt(label string) string {
	return color.Title().Render(" "+label) + " " + m.input.View()
}

// promptHelp lists the keys available while a prompt or the form is open.
func (m Model) promptHelp() string {
	switch m.mode {
	case modeConfirmDelete:
		return " y " + i18n.T("tui.help.confirm") + " • n/esc " + i18n.T("tui.help.cancel")
	case modeAddTodo:
		return " tab/↓ " + i18n.T("tui.help.nextField") + " • enter " + i18n.T("tui.help.confirm") + " • esc " + i18n.T("tui.help.cancel")
	default:
		return " enter " + i18n.T("tui.help.confirm") + " • esc " + i18n.T("tui.help.cancel")
	}
}

// renderMain is the list, with the details pane beside or under it.
func (m Model) renderMain(height int) string {
	if !m.details {
		return m.renderList(height)
	}

	r, ok := m.selected()
	if !ok {
		return m.renderList(height)
	}
	details := m.renderDetails(r)

	if m.width >= detailsMinWidth {
		listWidth := m.width * 55 / 100
		list := lipgloss.NewStyle().Width(listWidth).Render(m.renderList(height))
		pane := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			PaddingLeft(1).
			Render(details)
		return lipgloss.JoinHorizontal(lipgloss.Top, list, pane)
	}

	listHeight := max(height-lipgloss.Height(details)-1, 1)
	return m.renderList(listHeight) + "\n\n" + details
}

func (m Model) renderList(height int) string {
	st := m.tabs[m.active]
	rows := m.rows()

	switch {
	case st.err != nil:
		return color.Error().Render(" " + st.err.Error())
	case !st.loaded:
		return color.Muted().Render(" " + i18n.T("tui.loading"))
	case len(rows) == 0:
		return color.Muted().Render(" " + emptyMessage(m.active))
	}

	start, end := visibleRange(len(rows), st.cursor, height)

	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		lines = append(lines, renderRow(rows[i], i == st.cursor))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderForm() string {
	lines := []string{color.Title().Render(" " + i18n.T("tui.form.title")), ""}
	labelStyle := lipgloss.NewStyle().Width(labelWidth())
	for i := range m.form.fields {
		label := labelStyle.Render(fieldLabel(i))
		if i == m.form.focus {
			label = color.Title().Render(" › " + label)
		} else {
			label = color.Muted().Render("   " + label)
		}
		lines = append(lines, label+"  "+m.form.fields[i].View())
	}
	if m.form.err != nil {
		lines = append(lines, "", color.Error().Render(" "+m.form.err.Error()))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderDetails(r row) string {
	if r.dump != nil {
		lines := []string{color.Title().Render(r.dump.Name)}
		if !r.dump.CreatedAt.IsZero() {
			lines = append(lines, color.Muted().Render(r.dump.CreatedAt.Format(time.RFC822)))
		}
		return strings.Join(lines, "\n")
	}

	todo := r.todo
	field := func(label, value string) string {
		return color.Muted().Render(label+": ") + value
	}
	yesNo := func(b bool) string {
		if b {
			return i18n.T("tui.details.yes")
		}
		return i18n.T("tui.details.no")
	}

	status := i18n.T("tasks.tasks.notDone")
	if todo.Done() {
		status = i18n.T("tasks.tasks.done")
	}

	lines := []string{color.Title().Render(todo.Name), ""}
	if todo.Description.String != "" {
		lines = append(lines, field(i18n.T("tui.details.desc"), todo.Description.String))
	}
	if !todo.IsDaily {
		lines = append(lines, field(i18n.T("tui.details.expires"), todo.ExpiresAt.Format("Mon 02 Jan 2006 15:04")))
	}
	if todo.Priority != "" {
		lines = append(lines, field(i18n.T("tui.details.priority"), todo.Priority))
	}
	lines = append(lines,
		field(i18n.T("tui.details.daily"), yesNo(todo.IsDaily)),
		field(i18n.T("tui.details.status"), status),
	)
	if todo.IsDaily && m.streakKnown && m.streakName == todo.Name {
		lines = append(lines, field(i18n.T("tui.details.streak"), fmt.Sprintf(i18n.T("tui.details.days"), m.streakDays)))
	}
	return strings.Join(lines, "\n")
}

// visibleRange scrolls just enough to keep the cursor on screen; height 0
// means the size is unknown and everything is shown.
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
