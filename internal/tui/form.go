package tui

import (
	"errors"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

const (
	fieldName = iota
	fieldDescription
	fieldPriority
	fieldDaily
	fieldDate
	fieldCount
)

// todoForm collects the fields of a new task, one text input per field.
type todoForm struct {
	fields [fieldCount]textinput.Model
	focus  int
	err    error
}

func newTodoForm(width int) todoForm {
	var f todoForm
	for i := range f.fields {
		in := textinput.New()
		in.Prompt = ""
		f.fields[i] = in
	}
	f.fields[fieldDaily].CharLimit = 1
	f.fields[fieldPriority].Placeholder = i18n.T("tui.form.priorityHint")
	f.fields[fieldDaily].Placeholder = i18n.T("tui.form.dailyHint")
	f.fields[fieldDate].Placeholder = i18n.T("tui.form.dateHint")
	f.setWidth(width)
	f.fields[fieldName].Focus()
	return f
}

func fieldLabel(i int) string {
	switch i {
	case fieldName:
		return i18n.T("tui.form.name")
	case fieldDescription:
		return i18n.T("tui.form.desc")
	case fieldPriority:
		return i18n.T("tui.form.priority")
	case fieldDaily:
		return i18n.T("tui.form.daily")
	default:
		return i18n.T("tui.form.date")
	}
}

// labelWidth is the widest field label, so inputs line up in one column.
func labelWidth() int {
	w := 0
	for i := 0; i < fieldCount; i++ {
		w = max(w, lipgloss.Width(fieldLabel(i)))
	}
	return w
}

func (f *todoForm) setWidth(width int) {
	for i := range f.fields {
		f.fields[i].Width = max(width-labelWidth()-8, 10)
	}
}

func (f todoForm) onLastField() bool { return f.focus == fieldCount-1 }

func (f *todoForm) next() { f.focusField((f.focus + 1) % fieldCount) }

func (f *todoForm) prev() { f.focusField((f.focus + fieldCount - 1) % fieldCount) }

func (f *todoForm) focusField(i int) {
	f.fields[f.focus].Blur()
	f.focus = i
	f.fields[f.focus].Focus()
}

func (f *todoForm) update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.fields[f.focus], cmd = f.fields[f.focus].Update(msg)
	return cmd
}

// value validates the fields and builds the task; the due date goes
// through the same rules as `akrasia add --date`.
func (f todoForm) value() (tasks.NewTodo, error) {
	in := tasks.NewTodo{
		Name:        strings.TrimSpace(f.fields[fieldName].Value()),
		Description: strings.TrimSpace(f.fields[fieldDescription].Value()),
		Priority:    strings.ToLower(strings.TrimSpace(f.fields[fieldPriority].Value())),
	}

	if in.Name == "" {
		return in, errors.New(i18n.T("tui.form.nameRequired"))
	}

	switch in.Priority {
	case "", "high", "medium", "low":
	default:
		return in, errors.New(i18n.T("tui.form.invalidPriority"))
	}

	switch strings.ToLower(f.fields[fieldDaily].Value()) {
	case "", "n":
	case "y", "s":
		in.IsDaily = true
	default:
		return in, errors.New(i18n.T("tui.form.invalidDaily"))
	}

	parts, err := dateParts(f.fields[fieldDate].Value())
	if err != nil {
		return in, err
	}
	in.ExpiresAt, err = tasks.ParseDate(parts)
	return in, err
}

// dateParts turns "dd" or "dd/mm" into the parts tasks.ParseDate expects.
func dateParts(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	fields := strings.Split(s, "/")
	if len(fields) > 2 {
		return nil, errors.New(i18n.T("tui.form.invalidDate"))
	}

	parts := make([]int, 0, len(fields))
	for _, field := range fields {
		n, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil {
			return nil, errors.New(i18n.T("tui.form.invalidDate"))
		}
		parts = append(parts, n)
	}
	return parts, nil
}
