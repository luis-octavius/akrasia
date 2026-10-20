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

const (
	// detailsMinWidth is the narrowest terminal that shows details beside
	// the list; below it they go under the list.
	detailsMinWidth = 80
	// defaultWidth is used before the first WindowSizeMsg (and in tests).
	defaultWidth = 80
	// maxModalWidth caps prompts and the form on wide terminals.
	maxModalWidth = 64
	// detailsTextWidth wraps notes in the details pane; the pane itself
	// clips anything wider.
	detailsTextWidth = 40
	// historyLimit is how many streak periods the details pane lists.
	historyLimit = 5
	// progressMaxWidth and progressMinWidth bound the progress bar; below
	// the minimum only the "Today 3/5" label is shown.
	progressMaxWidth = 40
	progressMinWidth = 12
)

// View renders the whole screen in the alternate screen buffer.
func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m Model) render() string {
	width := m.width
	if width <= 0 {
		width = defaultWidth
	}

	header := m.renderHeader(width) + "\n" + m.renderProgress(width)
	footer := m.renderFooter(width)

	bodyHeight := 0
	if m.height > 0 {
		bodyHeight = max(m.height-lipgloss.Height(header)-lipgloss.Height(footer), 3)
	}
	body := m.renderBody(width, bodyHeight)

	// nothing may spill past the terminal edge, or JoinVertical would pad
	// every other line to match it
	clip := lipgloss.NewStyle().MaxWidth(width)
	screen := lipgloss.JoinVertical(lipgloss.Left, clip.Render(header), clip.Render(body), clip.Render(footer))

	modal := m.renderModal(width)
	if modal == "" {
		return screen
	}

	// center the modal over the screen
	height := max(m.height, lipgloss.Height(screen))
	x := max((width-lipgloss.Width(modal))/2, 0)
	y := max((height-lipgloss.Height(modal))/2, 0)
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(screen),
		lipgloss.NewLayer(modal).X(x).Y(y).Z(1),
	).Render()
}

func modalWidth(width int) int {
	if width <= 0 {
		width = defaultWidth
	}
	return max(min(maxModalWidth, width-4), 20)
}

// --- header ------------------------------------------------------------

func (m Model) renderHeader(width int) string {
	tabs := make([]string, 0, tabCount+2)
	for t := tab(0); t < tabCount; t++ {
		tabs = append(tabs, tabStyle(t == m.active).Render(m.tabLabel(t)))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)

	// the badge goes first when there is room for it
	badge := badgeStyle().Render("akrasia")
	if lipgloss.Width(badge)+1+lipgloss.Width(row) <= width {
		row = lipgloss.JoinHorizontal(lipgloss.Center, badge, " ", row)
	}

	if filters := m.filterSummary(); filters != "" {
		label := color.Muted().Render("  " + filters)
		if lipgloss.Width(row)+lipgloss.Width(label) <= width {
			row = lipgloss.JoinHorizontal(lipgloss.Center, row, label)
		}
	}
	return row
}

// renderProgress is "Today 3/5  ▌▌▌▌▌░░░  60%" below the tabs; it stays
// blank until the first count arrives so the layout doesn't jump.
func (m Model) renderProgress(width int) string {
	if !m.progressKnown {
		return ""
	}

	label := fmt.Sprintf(i18n.T("tasks.tasks.progress"), m.progress.Done, m.progress.Total)
	if m.progress.Total == 0 {
		label = i18n.T("tasks.tasks.progressClear")
	}
	label = " " + color.Title().Render(label) + "  "

	barWidth := min(progressMaxWidth, width-lipgloss.Width(label)-1)
	if barWidth < progressMinWidth {
		return label
	}
	return label + color.ProgressBar(m.progress.Ratio(), barWidth)
}

func (m Model) tabLabel(t tab) string {
	st := m.tabs[t]
	if !st.loaded {
		return t.title()
	}
	count := 0
	for _, r := range st.rows {
		if r.selectable() {
			count++
		}
	}
	return fmt.Sprintf("%s %d", t.title(), count)
}

func (m Model) filterSummary() string {
	var parts []string
	if m.search != "" {
		parts = append(parts, fmt.Sprintf("%s: %q", i18n.T("tui.status.filter"), m.search))
	}
	if m.priority != "" {
		parts = append(parts, i18n.T("tui.status.priority")+": "+m.priority)
	}
	return strings.Join(parts, " · ")
}

// --- body --------------------------------------------------------------

// renderBody is the list panel, with the details panel beside or under it.
func (m Model) renderBody(width, height int) string {
	r, ok := m.selected()
	if !m.details || !ok {
		return m.renderListPanel(width, height)
	}

	if width >= detailsMinWidth {
		listWidth := width * 58 / 100
		return lipgloss.JoinHorizontal(lipgloss.Top,
			m.renderListPanel(listWidth, height),
			m.renderDetailsPanel(r, width-listWidth, height),
		)
	}

	details := m.renderDetailsPanel(r, width, 0)
	listHeight := 0
	if height > 0 {
		listHeight = max(height-lipgloss.Height(details), 3)
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.renderListPanel(width, listHeight), details)
}

// panel draws content in a rounded box of exactly width x height (height
// 0 means "fit the content").
func panel(content string, focused bool, width, height int) string {
	st := panelStyle(focused).Width(width)
	if height > 0 {
		// Height only pads; content taller than the box would push the
		// rest of the screen down, so cut it to the inner height
		st = st.Height(height)
		content = clipLines(content, height-2)
	}
	return st.Render(content)
}

// clipLines keeps the first n lines of s, ending with "…" when it cut.
func clipLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if n <= 0 || len(lines) <= n {
		return s
	}
	lines = append(lines[:n-1], color.Muted().Render("…"))
	return strings.Join(lines, "\n")
}

func (m Model) renderListPanel(width, height int) string {
	innerWidth := max(width-4, 10)
	innerHeight := 0
	if height > 0 {
		innerHeight = max(height-2, 1)
	}
	return panel(m.renderList(innerWidth, innerHeight), true, width, height)
}

func (m Model) renderList(width, height int) string {
	st := m.tabs[m.active]
	rows := m.rows()

	var message string
	switch {
	case st.err != nil:
		message = color.Error().Render(st.err.Error())
	case !st.loaded:
		message = color.Muted().Render(i18n.T("tui.loading"))
	case len(rows) == 0:
		message = color.Muted().Render(emptyMessage(m.active))
	}
	if message != "" {
		if height <= 0 {
			return message
		}
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, message)
	}

	start, end := visibleRange(len(rows), st.cursor, height)

	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		lines = append(lines, renderRow(rows, i, i == st.cursor, width))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderDetailsPanel(r row, width, height int) string {
	innerWidth := max(width-4, 10)
	content := lipgloss.NewStyle().Width(innerWidth).Render(m.renderDetails(r))
	return panel(content, false, width, height)
}

func (m Model) renderDetails(r row) string {
	field := func(label, value string) string {
		return color.Muted().Render(label+"  ") + value
	}

	if r.dump != nil {
		lines := []string{color.Accent().Render(r.dump.Name)}
		if !r.dump.CreatedAt.IsZero() {
			lines = append(lines, "", color.Muted().Render(r.dump.CreatedAt.Format(time.RFC822)))
		}
		return strings.Join(lines, "\n")
	}

	todo := r.todo
	yesNo := func(b bool) string {
		if b {
			return i18n.T("tui.details.yes")
		}
		return i18n.T("tui.details.no")
	}

	status := color.Warning().Render(i18n.T("tasks.tasks.notDone"))
	if todo.Done() {
		status = color.Success().Render(i18n.T("tasks.tasks.done"))
	}

	lines := []string{color.Accent().Render(todo.Name), ""}
	if todo.Description.String != "" {
		lines = append(lines, todo.Description.String, "")
	}
	if !todo.IsDaily {
		lines = append(lines, field(i18n.T("tui.details.expires"), todo.ExpiresAt.Format("Mon 02 Jan 2006 15:04")))
	}
	if todo.Priority != "" {
		lines = append(lines, field(i18n.T("tui.details.priority"), priorityBadge(todo.Priority)))
	}
	lines = append(lines,
		field(i18n.T("tui.details.daily"), yesNo(todo.IsDaily)),
		field(i18n.T("tui.details.status"), status),
	)

	if todo.IsDaily && m.streakKnown && m.detailsName == todo.Name {
		days := color.Accent().Render(fmt.Sprintf(i18n.T("tui.details.days"), m.streakDays))
		lines = append(lines, field(i18n.T("tui.details.streak"), days))

		if len(m.streakHistory) > 0 {
			lines = append(lines, "", color.Title().Render(i18n.T("tui.details.history")))
			for _, p := range m.streakHistory[:min(len(m.streakHistory), historyLimit)] {
				lines = append(lines, color.Muted().Render("• ")+fmt.Sprintf(i18n.T("tui.details.period"), p.StartDate, p.EndDate, p.Days))
			}
		}
	}
	if m.notesKnown && m.detailsName == todo.Name && len(m.detailNotes) > 0 {
		lines = append(lines, "", color.Title().Render(i18n.T("tui.details.notes")))
		lines = append(lines, noteLines(m.detailNotes, detailsTextWidth)...)
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

// --- rows --------------------------------------------------------------

func renderRow(rows []row, i int, selected bool, width int) string {
	r := rows[i]
	switch {
	case r.header != "":
		return renderSectionHeader(r.header, sectionCount(rows, i), width)
	case r.todo != nil:
		return renderTodoLine(*r.todo, selected, width)
	default:
		return renderDumpLine(*r.dump, selected, width)
	}
}

// sectionCount counts the items under the header at index i.
func sectionCount(rows []row, i int) int {
	n := 0
	for _, r := range rows[i+1:] {
		if !r.selectable() {
			break
		}
		n++
	}
	return n
}

// renderSectionHeader draws "TITLE 3 ────" across the panel.
func renderSectionHeader(title string, count, width int) string {
	label := color.Accent().Render(title) + " " + color.Muted().Render(fmt.Sprint(count)) + " "
	rule := max(width-lipgloss.Width(label), 0)
	return label + lipgloss.NewStyle().Foreground(color.BorderMutedColor()).Render(strings.Repeat("─", rule))
}

// line joins left and right parts across width, painting the gap and every
// part with base so a selected row is highlighted end to end.
func line(base lipgloss.Style, left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		right, gap = "", max(width-lipgloss.Width(left), 0)
	}
	out := left + base.Render(strings.Repeat(" ", gap)) + right
	return lipgloss.NewStyle().MaxWidth(width).Render(out)
}

func rowBase(selected bool) lipgloss.Style {
	if selected {
		return color.Selection()
	}
	return lipgloss.NewStyle()
}

func cursorMark(base lipgloss.Style, selected bool) string {
	if selected {
		return color.Accent().Inherit(base).Render("› ")
	}
	return base.Render("  ")
}

// renderTodoLine is the one-line form of a task: state icon, name and
// daily marker on the left; due date and priority badge on the right.
func renderTodoLine(todo tasks.TodoView, selected bool, width int) string {
	base := rowBase(selected)

	icon, iconStyle := "○", color.Muted()
	nameStyle := lipgloss.NewStyle().Bold(selected)
	switch {
	case todo.Done():
		icon, iconStyle = "✔", color.Success()
		nameStyle = color.Muted().Strikethrough(true)
	case tasks.IsExpiringSoon(todo.ExpiresAt) && !todo.IsDaily:
		iconStyle = color.Warning()
	}

	left := cursorMark(base, selected) +
		iconStyle.Inherit(base).Render(icon) + base.Render(" ") +
		nameStyle.Inherit(base).Render(todo.Name)
	if todo.IsDaily {
		left += color.Muted().Inherit(base).Render(" ↻")
	}

	var right string
	if !todo.IsDaily {
		right = color.Muted().Inherit(base).Render(todo.ExpiresAt.Format("02 Jan 15:04"))
	}
	if badge := priorityBadge(todo.Priority); badge != "" {
		if right != "" {
			right += base.Render(" ")
		}
		right += badge
	}

	return line(base, left, right, width)
}

func renderDumpLine(d tasks.DumpEntry, selected bool, width int) string {
	base := rowBase(selected)
	left := cursorMark(base, selected) + lipgloss.NewStyle().Bold(selected).Inherit(base).Render(d.Name)

	var right string
	if !d.CreatedAt.IsZero() {
		right = color.Muted().Inherit(base).Render(d.CreatedAt.Format("02 Jan 15:04"))
	}
	return line(base, left, right, width)
}

// --- modals ------------------------------------------------------------

// renderModal returns the box floating over the screen for the current
// mode, or "" when there is none.
func (m Model) renderModal(width int) string {
	mw := modalWidth(width)
	box := modalStyle().Width(mw)

	switch m.mode {
	case modeNotes:
		if m.replacingNote {
			title := color.Accent().Render(fmt.Sprintf(i18n.T("tui.prompt.replaceNote"), m.target.Name))
			hint := color.Muted().Render(i18n.T("tui.prompt.replaceHint"))
			return box.Render(title + "\n\n" + themedInput(m.input).View() + "\n\n" + hint)
		}
		title := color.Accent().Render(fmt.Sprintf(i18n.T("tui.prompt.notes"), m.target.Name))
		return box.Render(title + "\n\n" + themedInput(m.input).View())

	case modeAddDump:
		title := color.Accent().Render(i18n.T("tui.prompt.dumpName"))
		return box.Render(title + "\n\n" + themedInput(m.input).View())

	case modeConfirmDelete:
		question := fmt.Sprintf(i18n.T("tui.prompt.delete"), m.target.Name)
		if m.targetDump != nil {
			question = fmt.Sprintf(i18n.T("tui.prompt.deleteDump"), m.targetDump.Name)
		}
		title := color.Warning().Bold(true).Render(question)
		return box.Render(title + "\n\n" + color.Muted().Render(m.promptHelp()))

	case modeAddTodo:
		return box.Render(m.renderForm())

	case modeNotesView:
		return m.renderNotesModal(box)

	case modeQuote:
		return m.renderQuoteModal(box, mw)
	}
	return ""
}

func (m Model) renderForm() string {
	title := i18n.T("tui.form.title")
	switch {
	case m.form.editing != nil:
		title = i18n.T("tui.form.editTitle")
	case m.form.fromDump != nil:
		title = i18n.T("tui.form.fromDumpTitle")
	}
	lines := []string{color.Accent().Render(title), ""}
	labelStyle := lipgloss.NewStyle().Width(labelWidth())
	for i := range m.form.fields {
		label := labelStyle.Render(fieldLabel(i))
		if i == m.form.focus {
			label = color.Accent().Render("› " + label)
		} else {
			label = color.Muted().Render("  " + label)
		}
		lines = append(lines, label+"  "+themedInput(m.form.fields[i]).View())
	}
	if m.form.err != nil {
		lines = append(lines, "", color.Error().Render(m.form.err.Error()))
	}
	return strings.Join(lines, "\n")
}

// --- footer ------------------------------------------------------------

func (m Model) renderFooter(width int) string {
	bar := m.renderStatusBar(width)

	if m.mode != modeNormal {
		hint := color.Muted().Render(" " + m.promptHelp())
		return bar + "\n" + lipgloss.NewStyle().MaxWidth(width).Render(hint)
	}

	// pad every line: the full help (?) spans several lines, and a leading
	// space on the string alone would only shift the first one
	h := m.help
	h.Styles = helpStyles()
	h.SetWidth(width - 1)
	return bar + "\n" + lipgloss.NewStyle().PaddingLeft(1).Render(h.View(m.keys))
}

// renderStatusBar is the full-width bar: mode pill and message on the left,
// last backup, keymap and theme on the right.
func (m Model) renderStatusBar(width int) string {
	base := color.Bar()

	left := pillStyle().Render(m.modeLabel()) + base.Render(" ")
	switch {
	case m.mode == modeSearch:
		left += color.Accent().Inherit(base).Render("/ ") + themedInput(m.input).View()
	case m.status != "" && m.statusErr:
		left += color.Error().Inherit(base).Render(m.status)
	case m.status != "":
		left += color.Success().Inherit(base).Render(m.status)
	}

	info := "◐ " + color.GetCurrentTheme().Name
	if CurrentKeymap() == KeymapVim {
		info = "vim · " + info
	}
	right := base.Render(info + " ")
	if m.backup != "" {
		note := base
		if !m.backupOn {
			note = color.Muted().Inherit(base)
		}
		right = note.Render(m.backup) + base.Render(" · ") + right
	}
	return line(base, left, right, width)
}

func (m Model) modeLabel() string {
	switch m.mode {
	case modeSearch:
		return i18n.T("tui.mode.search")
	case modeNotes:
		return i18n.T("tui.mode.notes")
	case modeConfirmDelete:
		return i18n.T("tui.mode.delete")
	case modeAddDump:
		return i18n.T("tui.mode.dump")
	case modeNotesView:
		return i18n.T("tui.mode.notes")
	case modeQuote:
		return i18n.T("tui.mode.quote")
	case modeAddTodo:
		if m.form.editing != nil {
			return i18n.T("tui.mode.edit")
		}
		return i18n.T("tui.mode.add")
	default:
		return i18n.T("tui.mode.normal")
	}
}

// promptHelp lists the keys available while a prompt or the form is open.
func (m Model) promptHelp() string {
	switch m.mode {
	case modeConfirmDelete:
		return "y " + i18n.T("tui.help.confirm") + " • n/esc " + i18n.T("tui.help.cancel")
	case modeNotesView:
		return "↑/↓ pgup/pgdn " + i18n.T("tui.help.scroll") + " • esc " + i18n.T("tui.help.close")
	case modeQuote:
		return "esc/enter " + i18n.T("tui.help.close")
	case modeAddTodo:
		return "tab/↓ " + i18n.T("tui.help.nextField") + " • enter " + i18n.T("tui.help.confirm") + " • esc " + i18n.T("tui.help.cancel")
	default:
		return "enter " + i18n.T("tui.help.confirm") + " • esc " + i18n.T("tui.help.cancel")
	}
}
