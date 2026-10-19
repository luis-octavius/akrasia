package tui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// priorityCycle is the order the priority filter steps through.
var priorityCycle = []string{"", "high", "medium", "low"}

func (m Model) handleNormalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	st := &m.tabs[m.active]
	rows := m.rows()

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		st.cursor = moveCursor(rows, st.cursor, -1)
	case key.Matches(msg, m.keys.Down):
		st.cursor = moveCursor(rows, st.cursor, 1)
	case key.Matches(msg, m.keys.Top):
		st.cursor = moveCursor(rows, -1, 1)
	case key.Matches(msg, m.keys.Bottom):
		st.cursor = moveCursor(rows, len(rows), -1)

	case key.Matches(msg, m.keys.NextTab):
		return m.switchTab((m.active + 1) % tabCount)
	case key.Matches(msg, m.keys.PrevTab):
		return m.switchTab((m.active + tabCount - 1) % tabCount)
	case key.Matches(msg, m.keys.Reload):
		return m.reload(m.active)

	case key.Matches(msg, m.keys.Complete):
		return m.startComplete()
	case key.Matches(msg, m.keys.Delete):
		return m.startDelete()
	case key.Matches(msg, m.keys.Add):
		return m.startAdd()
	case key.Matches(msg, m.keys.Edit):
		return m.startEdit()
	case key.Matches(msg, m.keys.Notes):
		return m.startNotesView()
	case key.Matches(msg, m.keys.Convert):
		return m.startConvert()
	case key.Matches(msg, m.keys.Quote):
		m.openQuote("")
		return m, nil

	case key.Matches(msg, m.keys.Search):
		m.mode = modeSearch
		m.status = ""
		m.openInput(m.search)
		return m, textinput.Blink

	case key.Matches(msg, m.keys.Priority):
		m.priority = nextPriority(m.priority)
		m.clampCursor(m.active)

	case key.Matches(msg, m.keys.Details):
		m.details = !m.details

	case key.Matches(msg, m.keys.Help):
		m.help.ShowAll = !m.help.ShowAll

	case key.Matches(msg, m.keys.Theme):
		name := color.NextTheme(color.GetCurrentTheme().Name)
		err := m.saveTheme(name)
		m.setStatus(fmt.Sprintf(i18n.T("tui.status.theme"), name), err)

	case msg.String() == "esc":
		m.search, m.priority, m.details, m.status = "", "", false, ""
		m.clampCursor(m.active)
	}

	cmd := m.loadDetails()
	return m, cmd
}

func nextPriority(current string) string {
	for i, p := range priorityCycle {
		if p == current {
			return priorityCycle[(i+1)%len(priorityCycle)]
		}
	}
	return ""
}

// startComplete asks for optional notes before marking the selected task
// done.
func (m Model) startComplete() (tea.Model, tea.Cmd) {
	r, ok := m.selected()
	if !ok || r.todo == nil {
		m.setStatus(i18n.T("tui.prompt.noSelection"), nil)
		return m, nil
	}
	// done today: completing again replaces today's note, so offer that
	if r.todo.DoneToday {
		m.mode, m.target, m.replacingNote = modeNotes, *r.todo, true
		m.status = ""
		m.openInput("")
		return m, tea.Batch(textinput.Blink, m.todayNoteCmd(r.todo.Name))
	}
	if r.todo.Done() {
		m.setStatus(fmt.Sprintf(i18n.T("tui.prompt.alreadyDone"), r.todo.Name), nil)
		return m, nil
	}

	m.mode, m.target, m.replacingNote = modeNotes, *r.todo, false
	m.status = ""
	m.openInput("")
	return m, textinput.Blink
}

// todayNoteCmd fetches the note written today, if any, to prefill the
// prompt that replaces it.
func (m Model) todayNoteCmd(name string) tea.Cmd {
	ctx, svc := m.ctx, m.svc
	return func() tea.Msg {
		_, entries, err := svc.Notes(ctx, name, 1)
		if err != nil || len(entries) == 0 || entries[0].Date != time.Now().Format(time.DateOnly) {
			return nil
		}
		return notePrefillMsg{name: name, text: entries[0].Notes}
	}
}

// startDelete asks for confirmation before deleting the selected task.
func (m Model) startDelete() (tea.Model, tea.Cmd) {
	r, ok := m.selected()
	switch {
	case ok && r.dump != nil:
		entry := *r.dump
		m.mode, m.targetDump = modeConfirmDelete, &entry
	case ok && r.todo != nil:
		m.mode, m.target, m.targetDump = modeConfirmDelete, *r.todo, nil
	default:
		m.setStatus(i18n.T("tui.prompt.noSelection"), nil)
		return m, nil
	}
	m.status = ""
	return m, nil
}

// startConvert opens the new-task form for the selected dump entry.
func (m Model) startConvert() (tea.Model, tea.Cmd) {
	r, ok := m.selected()
	if !ok || r.dump == nil {
		return m, nil
	}

	m.mode, m.status = modeAddTodo, ""
	m.form = newDumpForm(modalWidth(m.width), *r.dump)
	return m, textinput.Blink
}

// startAdd opens the dump prompt on the Dump tab and the task form
// everywhere else.
func (m Model) startAdd() (tea.Model, tea.Cmd) {
	if m.active == tabDump {
		m.mode = modeAddDump
		m.status = ""
		m.openInput("")
		return m, textinput.Blink
	}

	m.mode = modeAddTodo
	m.status = ""
	m.form = newTodoForm(modalWidth(m.width))
	return m, textinput.Blink
}

// startEdit opens the form prefilled with the selected task.
func (m Model) startEdit() (tea.Model, tea.Cmd) {
	r, ok := m.selected()
	if !ok || r.todo == nil {
		m.setStatus(i18n.T("tui.prompt.noSelection"), nil)
		return m, nil
	}

	m.mode = modeAddTodo
	m.status = ""
	m.form = newEditForm(modalWidth(m.width), *r.todo)
	return m, textinput.Blink
}

func (m *Model) openInput(value string) {
	m.input.SetValue(value)
	m.input.CursorEnd()
	m.input.Focus()
}

func (m *Model) closePrompt() {
	m.mode = modeNormal
	m.input.Blur()
	m.input.SetValue("")
}

// handlePromptKey drives the one-line prompts: notes, search and dump.
func (m Model) handlePromptKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.mode == modeSearch {
			m.search = ""
			m.clampCursor(m.active)
		}
		m.closePrompt()
		return m, nil

	case "enter":
		value, current := m.input.Value(), m.mode
		m.closePrompt()

		switch current {
		case modeNotes:
			return m, m.completeCmd(m.target.Name, value, m.replacingNote)
		case modeAddDump:
			if value == "" {
				return m, nil
			}
			return m, m.addDumpCmd(value)
		}
		return m, nil // search stays applied
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.mode == modeSearch {
		m.search = m.input.Value()
		m.clampCursor(m.active)
	}
	return m, cmd
}

// handleConfirmKey answers the delete confirmation. "s" is accepted too,
// for "sim".
func (m Model) handleConfirmKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "s", "S":
		m.mode = modeNormal
		if m.targetDump != nil {
			return m, m.deleteDumpCmd(*m.targetDump)
		}
		return m, m.deleteCmd(m.target.Name)
	case "n", "N", "esc":
		m.mode = modeNormal
	}
	return m, nil
}

func (m Model) handleFormKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeNormal
		return m, nil

	case "enter":
		if !m.form.onLastField() {
			m.form.next()
			return m, nil
		}
		if m.form.editing != nil {
			return m.submitEdit()
		}
		in, err := m.form.value()
		if err != nil {
			m.form.err = err
			return m, nil
		}
		m.mode = modeNormal
		if m.form.fromDump != nil {
			return m, m.promoteCmd(*m.form.fromDump, in)
		}
		return m, m.addTodoCmd(in)

	case "tab", "down":
		m.form.next()
		return m, nil

	case "shift+tab", "up":
		m.form.prev()
		return m, nil
	}

	m.form.err = nil // typing clears the last validation error
	cmd := m.form.update(msg)
	return m, cmd
}

// submitEdit saves only the fields that changed; with no change it just
// closes the form.
func (m Model) submitEdit() (tea.Model, tea.Cmd) {
	patch, err := m.form.patch()
	if err != nil {
		m.form.err = err
		return m, nil
	}
	m.mode = modeNormal
	if patch.Empty() {
		return m, nil
	}
	return m, m.editCmd(m.form.editing.Name, patch)
}

func (m Model) editCmd(name string, patch tasks.TodoPatch) tea.Cmd {
	ctx, svc := m.ctx, m.svc
	return func() tea.Msg {
		todo, err := svc.EditTodo(ctx, name, patch)
		return actionMsg{status: fmt.Sprintf(i18n.T("tui.status.updated"), todo.Name), err: err}
	}
}

// completeCmd completes the task with notes. When replacing today's note
// it only reports the update; a real completion celebrates with a quote.
func (m Model) completeCmd(name, notes string, replacing bool) tea.Cmd {
	ctx, svc := m.ctx, m.svc
	return func() tea.Msg {
		todo, err := svc.Complete(ctx, name, notes)
		if replacing {
			return actionMsg{status: fmt.Sprintf(i18n.T("tui.status.noteUpdated"), todo.Name), err: err}
		}
		return actionMsg{status: fmt.Sprintf(i18n.T("tui.status.completed"), todo.Name), completed: true, err: err}
	}
}

func (m Model) deleteCmd(name string) tea.Cmd {
	ctx, svc := m.ctx, m.svc
	return func() tea.Msg {
		todo, err := svc.DeleteByName(ctx, name)
		return actionMsg{status: fmt.Sprintf(i18n.T("tui.status.deleted"), todo.Name), err: err}
	}
}

func (m Model) deleteDumpCmd(entry tasks.DumpEntry) tea.Cmd {
	ctx, svc := m.ctx, m.svc
	return func() tea.Msg {
		err := svc.DeleteDump(ctx, entry.ID)
		return actionMsg{status: fmt.Sprintf(i18n.T("tui.status.dumpDeleted"), entry.Name), err: err}
	}
}

func (m Model) promoteCmd(entry tasks.DumpEntry, in tasks.NewTodo) tea.Cmd {
	ctx, svc := m.ctx, m.svc
	return func() tea.Msg {
		todo, err := svc.PromoteDump(ctx, entry.ID, in)
		return actionMsg{status: fmt.Sprintf(i18n.T("tui.status.promoted"), todo.Name), err: err}
	}
}

func (m Model) addDumpCmd(name string) tea.Cmd {
	ctx, svc := m.ctx, m.svc
	return func() tea.Msg {
		entry, err := svc.AddDump(ctx, name)
		return actionMsg{status: fmt.Sprintf(i18n.T("tui.status.dumpAdded"), entry.Name), err: err}
	}
}

func (m Model) addTodoCmd(in tasks.NewTodo) tea.Cmd {
	ctx, svc := m.ctx, m.svc
	return func() tea.Msg {
		todo, err := svc.AddTodo(ctx, in)
		return actionMsg{status: fmt.Sprintf(i18n.T("tui.status.created"), todo.Name), err: err}
	}
}
