package tui

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

// todoForm collects the fields of a task, one text input per field. It
// creates a task, or edits one when editing is set.
type todoForm struct {
	fields [fieldCount]textinput.Model
	focus  int
	err    error

	// editing is the task being edited; nil when the form creates a task.
	editing *tasks.TodoView
	// initial holds the prefilled values, to tell which fields changed.
	initial [fieldCount]string
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

// newEditForm opens the form prefilled with todo. Daily tasks have no due
// date, so their date field starts empty.
func newEditForm(width int, todo tasks.TodoView) todoForm {
	f := newTodoForm(width)
	f.editing = &todo

	daily, date := "n", todo.ExpiresAt.Format("02/01")
	if todo.IsDaily {
		daily, date = "y", ""
	}
	f.initial = [fieldCount]string{todo.Name, todo.Description.String, todo.Priority, daily, date}
	for i, v := range f.initial {
		f.fields[i].SetValue(v)
		f.fields[i].CursorEnd()
	}
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

// setWidth sizes the inputs to fit a modal of the given width: border and
// padding take 6 columns, the focus mark and gap 4, the cursor 1.
func (f *todoForm) setWidth(width int) {
	for i := range f.fields {
		f.fields[i].SetWidth(max(width-labelWidth()-11, 10))
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
	}

	if in.Name == "" {
		return in, errors.New(i18n.T("tui.form.nameRequired"))
	}

	var err error
	if in.Priority, err = parsePriority(f.fields[fieldPriority].Value()); err != nil {
		return in, err
	}
	if in.IsDaily, err = parseDaily(f.fields[fieldDaily].Value()); err != nil {
		return in, err
	}
	in.ExpiresAt, err = parseDueDate(f.fields[fieldDate].Value())
	return in, err
}

// patch returns only the fields that differ from the prefilled values, so
// an untouched past due date is not re-validated (and rejected) on save.
func (f todoForm) patch() (tasks.TodoPatch, error) {
	var p tasks.TodoPatch
	changed := func(i int) (string, bool) {
		v := strings.TrimSpace(f.fields[i].Value())
		return v, v != strings.TrimSpace(f.initial[i])
	}

	if v, ok := changed(fieldName); ok {
		if v == "" {
			return p, errors.New(i18n.T("tui.form.nameRequired"))
		}
		p.Name = &v
	}
	if v, ok := changed(fieldDescription); ok {
		p.Description = &v
	}
	if v, ok := changed(fieldPriority); ok {
		priority, err := parsePriority(v)
		if err != nil {
			return p, err
		}
		p.Priority = &priority
	}
	daily, err := parseDaily(f.fields[fieldDaily].Value())
	if err != nil {
		return p, err
	}
	if daily != f.editing.IsDaily {
		p.IsDaily = &daily
	}
	if v, ok := changed(fieldDate); ok {
		date, err := parseDueDate(v)
		if err != nil {
			return p, err
		}
		p.ExpiresAt = &date
	}
	return p, nil
}

func parsePriority(s string) (string, error) {
	p := strings.ToLower(strings.TrimSpace(s))
	switch p {
	case "", "high", "medium", "low":
		return p, nil
	}
	return "", errors.New(i18n.T("tui.form.invalidPriority"))
}

// parseDaily accepts y/s ("sim") or n; empty means no.
func parseDaily(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "n":
		return false, nil
	case "y", "s":
		return true, nil
	}
	return false, errors.New(i18n.T("tui.form.invalidDaily"))
}

func parseDueDate(s string) (time.Time, error) {
	parts, err := dateParts(s)
	if err != nil {
		return time.Time{}, err
	}
	return tasks.ParseDate(parts)
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
