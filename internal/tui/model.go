package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// Service is the part of tasks.TaskManager the TUI reads from. It is an
// interface so the model can be tested without a database.
type Service interface {
	Today(ctx context.Context, opts tasks.TodayOptions) (tasks.TodayView, error)
	ListTodos(ctx context.Context, priorityFilter string) ([]tasks.TodoView, error)
	DailyTodos(ctx context.Context) ([]tasks.TodoView, error)
	Dumps(ctx context.Context) ([]tasks.DumpEntry, error)
}

type tab int

const (
	tabToday tab = iota
	tabAll
	tabDaily
	tabDump
	tabCount
)

func (t tab) title() string {
	switch t {
	case tabToday:
		return i18n.T("tui.tabs.today")
	case tabAll:
		return i18n.T("tui.tabs.all")
	case tabDaily:
		return i18n.T("tui.tabs.daily")
	default:
		return i18n.T("tui.tabs.dump")
	}
}

// row is one line of a tab: either a section header or an item.
type row struct {
	header string
	todo   *tasks.TodoView
	dump   *tasks.DumpEntry
}

func (r row) selectable() bool { return r.header == "" }

type tabState struct {
	rows    []row
	cursor  int
	loaded  bool
	loading bool
	err     error
}

// loadedMsg carries the result of fetching one tab.
type loadedMsg struct {
	tab  tab
	rows []row
	err  error
}

// Model is the root Bubble Tea model.
type Model struct {
	ctx    context.Context
	svc    Service
	keys   keyMap
	help   help.Model
	active tab
	tabs   [tabCount]tabState
	width  int
	height int
}

// New builds the root model. ctx is used for every query, so cancelling it
// aborts in-flight loads.
func New(ctx context.Context, svc Service) Model {
	return Model{
		ctx:  ctx,
		svc:  svc,
		keys: defaultKeyMap(),
		help: help.New(),
	}
}

// Init loads the first tab. Init cannot change the model, so a tab that is
// not loaded yet is shown as loading (see View).
func (m Model) Init() tea.Cmd {
	return m.load(m.active)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.Width = msg.Width
		return m, nil

	case loadedMsg:
		st := &m.tabs[msg.tab]
		st.loading = false
		st.loaded = true
		st.err = msg.err
		st.rows = msg.rows
		st.cursor = clampCursor(st.rows, st.cursor)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	st := &m.tabs[m.active]

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		st.cursor = moveCursor(st.rows, st.cursor, -1)

	case key.Matches(msg, m.keys.Down):
		st.cursor = moveCursor(st.rows, st.cursor, 1)

	case key.Matches(msg, m.keys.NextTab):
		return m.switchTab((m.active + 1) % tabCount)

	case key.Matches(msg, m.keys.PrevTab):
		return m.switchTab((m.active + tabCount - 1) % tabCount)

	case key.Matches(msg, m.keys.Reload):
		return m.reload(m.active)
	}

	return m, nil
}

// switchTab activates t and loads it the first time it is shown.
func (m Model) switchTab(t tab) (tea.Model, tea.Cmd) {
	m.active = t
	if m.tabs[t].loaded || m.tabs[t].loading {
		return m, nil
	}
	return m.reload(t)
}

func (m Model) reload(t tab) (tea.Model, tea.Cmd) {
	m.tabs[t].loading = true
	return m, m.load(t)
}

func (m Model) load(t tab) tea.Cmd {
	ctx, svc := m.ctx, m.svc
	return func() tea.Msg {
		rows, err := fetchRows(ctx, svc, t)
		return loadedMsg{tab: t, rows: rows, err: err}
	}
}

func fetchRows(ctx context.Context, svc Service, t tab) ([]row, error) {
	switch t {
	case tabToday:
		view, err := svc.Today(ctx, tasks.TodayOptions{})
		if err != nil {
			return nil, err
		}
		var rows []row
		rows = appendSection(rows, i18n.T("tasks.tasks.overdue"), view.Overdue)
		rows = appendSection(rows, i18n.T("tasks.tasks.dueToday"), view.DueToday)
		rows = appendSection(rows, i18n.T("tasks.tasks.dailyPending"), view.Daily)
		rows = appendSection(rows, i18n.T("tasks.tasks.expiringSoon"), view.ExpiringSoon)
		return rows, nil

	case tabAll:
		todos, err := svc.ListTodos(ctx, "")
		return todoRows(todos), err

	case tabDaily:
		todos, err := svc.DailyTodos(ctx)
		return todoRows(todos), err

	default:
		dumps, err := svc.Dumps(ctx)
		if err != nil {
			return nil, err
		}
		rows := make([]row, len(dumps))
		for i := range dumps {
			rows[i] = row{dump: &dumps[i]}
		}
		return rows, nil
	}
}

func appendSection(rows []row, title string, todos []tasks.TodoView) []row {
	if len(todos) == 0 {
		return rows
	}
	rows = append(rows, row{header: title})
	return append(rows, todoRows(todos)...)
}

func todoRows(todos []tasks.TodoView) []row {
	rows := make([]row, len(todos))
	for i := range todos {
		rows[i] = row{todo: &todos[i]}
	}
	return rows
}

// moveCursor steps by delta, skipping section headers, and stays put at
// the edges.
func moveCursor(rows []row, cursor, delta int) int {
	for i := cursor + delta; i >= 0 && i < len(rows); i += delta {
		if rows[i].selectable() {
			return i
		}
	}
	return cursor
}

// clampCursor keeps the cursor on a selectable row after a reload.
func clampCursor(rows []row, cursor int) int {
	if cursor >= len(rows) {
		cursor = len(rows) - 1
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor < len(rows) && !rows[cursor].selectable() {
		if next := moveCursor(rows, cursor, 1); next != cursor {
			return next
		}
		return moveCursor(rows, cursor, -1)
	}
	return cursor
}
