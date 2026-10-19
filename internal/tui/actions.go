package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// priorityCycle is the order the priority filter steps through.
var priorityCycle = []string{"", "high", "medium", "low"}

func (m Model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	case msg.Type == tea.KeyEsc:
		m.search, m.priority, m.details, m.status = "", "", false, ""
		m.clampCursor(m.active)
	}

	cmd := m.loadStreak()
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
	if r.todo.Done() {
		m.setStatus(fmt.Sprintf(i18n.T("tui.prompt.alreadyDone"), r.todo.Name), nil)
		return m, nil
	}

	m.mode, m.target = modeNotes, *r.todo
	m.status = ""
	m.openInput("")
	return m, textinput.Blink
}

// startDelete asks for confirmation before deleting the selected task.
func (m Model) startDelete() (tea.Model, tea.Cmd) {
	r, ok := m.selected()
	if !ok || r.todo == nil {
		m.setStatus(i18n.T("tui.prompt.noSelection"), nil)
		return m, nil
	}

	m.mode, m.target = modeConfirmDelete, *r.todo
	m.status = ""
	return m, nil
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
	m.form = newTodoForm(m.width)
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
func (m Model) handlePromptKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		if m.mode == modeSearch {
			m.search = ""
			m.clampCursor(m.active)
		}
		m.closePrompt()
		return m, nil

	case tea.KeyEnter:
		value, current := m.input.Value(), m.mode
		m.closePrompt()

		switch current {
		case modeNotes:
			return m, m.completeCmd(m.target.Name, value)
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
func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "s", "S":
		m.mode = modeNormal
		return m, m.deleteCmd(m.target.Name)
	case "n", "N", "esc":
		m.mode = modeNormal
	}
	return m, nil
}

func (m Model) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = modeNormal
		return m, nil

	case tea.KeyEnter:
		if !m.form.onLastField() {
			m.form.next()
			return m, nil
		}
		in, err := m.form.value()
		if err != nil {
			m.form.err = err
			return m, nil
		}
		m.mode = modeNormal
		return m, m.addTodoCmd(in)

	case tea.KeyTab, tea.KeyDown:
		m.form.next()
		return m, nil

	case tea.KeyShiftTab, tea.KeyUp:
		m.form.prev()
		return m, nil
	}

	m.form.err = nil // typing clears the last validation error
	cmd := m.form.update(msg)
	return m, cmd
}

func (m Model) completeCmd(name, notes string) tea.Cmd {
	ctx, svc := m.ctx, m.svc
	return func() tea.Msg {
		todo, err := svc.Complete(ctx, name, notes)
		return actionMsg{status: fmt.Sprintf(i18n.T("tui.status.completed"), todo.Name), err: err}
	}
}

func (m Model) deleteCmd(name string) tea.Cmd {
	ctx, svc := m.ctx, m.svc
	return func() tea.Msg {
		todo, err := svc.DeleteByName(ctx, name)
		return actionMsg{status: fmt.Sprintf(i18n.T("tui.status.deleted"), todo.Name), err: err}
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
