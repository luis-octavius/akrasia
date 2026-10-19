package tui

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// Service is the part of tasks.TaskManager the TUI uses. It is an
// interface so the model can be tested without a database.
type Service interface {
	Today(ctx context.Context, opts tasks.TodayOptions) (tasks.TodayView, error)
	ListTodos(ctx context.Context, priorityFilter string) ([]tasks.TodoView, error)
	DailyTodos(ctx context.Context) ([]tasks.TodoView, error)
	Dumps(ctx context.Context) ([]tasks.DumpEntry, error)
	AddTodo(ctx context.Context, in tasks.NewTodo) (tasks.TodoView, error)
	Complete(ctx context.Context, name, notes string) (tasks.TodoView, error)
	DeleteByName(ctx context.Context, name string) (tasks.TodoView, error)
	AddDump(ctx context.Context, name string) (tasks.DumpEntry, error)
	CurrentStreak(ctx context.Context, name string) (tasks.TodoView, int64, error)
	TodayProgress(ctx context.Context) (tasks.Progress, error)
	StreakHistory(ctx context.Context, name string) (tasks.TodoView, []tasks.StreakPeriod, error)
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

// mode decides where key presses go: the list, a one-line prompt, or the
// new-task form.
type mode int

const (
	modeNormal mode = iota
	modeNotes
	modeConfirmDelete
	modeSearch
	modeAddDump
	modeAddTodo
)

// row is one line of a tab: either a section header or an item.
type row struct {
	header string
	todo   *tasks.TodoView
	dump   *tasks.DumpEntry
}

func (r row) selectable() bool { return r.header == "" }

type tabState struct {
	rows    []row
	cursor  int // index into the filtered rows
	loaded  bool
	loading bool
	stale   bool // data changed since the last load
	err     error
}

// loadedMsg carries the result of fetching one tab.
type loadedMsg struct {
	tab  tab
	rows []row
	err  error
}

// actionMsg is the result of a write; on success every tab is reloaded.
type actionMsg struct {
	status string
	err    error
}

// progressMsg carries today's progress, refreshed with every load.
type progressMsg struct {
	progress tasks.Progress
	err      error
}

// streakMsg answers the streak lookup made when details open.
type streakMsg struct {
	name    string
	days    int64
	periods []tasks.StreakPeriod
	err     error
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

	mode   mode
	input  textinput.Model // notes, search and dump name prompts
	form   todoForm
	target tasks.TodoView // task the current prompt acts on

	search   string
	priority string

	progress      tasks.Progress
	progressKnown bool

	details       bool
	streakName    string
	streakDays    int64
	streakHistory []tasks.StreakPeriod
	streakKnown   bool

	// saveTheme persists the theme chosen with `t`; tests swap it so they
	// never write the real config.
	saveTheme func(name string) error

	status    string
	statusErr bool
}

// New builds the root model with the saved keymap. ctx is used for every
// query, so cancelling it aborts in-flight loads.
func New(ctx context.Context, svc Service) Model {
	return newModel(ctx, svc, CurrentKeymap())
}

func newModel(ctx context.Context, svc Service, layout string) Model {
	input := textinput.New()
	input.Prompt = ""

	return Model{
		ctx:   ctx,
		svc:   svc,
		keys:  newKeyMap(layout),
		help:  help.New(),
		input: input,

		saveTheme: color.SaveTheme,
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
		m.help.SetWidth(msg.Width)
		m.input.SetWidth(modalWidth(msg.Width) - 7)
		m.form.setWidth(modalWidth(msg.Width))
		return m, nil

	case loadedMsg:
		st := &m.tabs[msg.tab]
		st.loading = false
		st.loaded = true
		st.stale = false
		st.err = msg.err
		st.rows = msg.rows
		m.clampCursor(msg.tab)
		cmd := m.loadStreak()
		return m, cmd

	case actionMsg:
		m.setStatus(msg.status, msg.err)
		if msg.err != nil {
			return m, nil
		}
		for t := range m.tabs {
			m.tabs[t].stale = true
		}
		return m.reload(m.active)

	case progressMsg:
		if msg.err == nil {
			m.progress, m.progressKnown = msg.progress, true
		}
		return m, nil

	case streakMsg:
		if msg.err == nil && msg.name == m.streakName {
			m.streakDays, m.streakHistory, m.streakKnown = msg.days, msg.periods, true
		}
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.mode {
		case modeNormal:
			return m.handleNormalKey(msg)
		case modeAddTodo:
			return m.handleFormKey(msg)
		case modeConfirmDelete:
			return m.handleConfirmKey(msg)
		default:
			return m.handlePromptKey(msg)
		}
	}

	// anything else (cursor blink) belongs to whichever input is open
	var cmd tea.Cmd
	switch m.mode {
	case modeNormal, modeConfirmDelete:
	case modeAddTodo:
		cmd = m.form.update(msg)
	default:
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m *Model) setStatus(status string, err error) {
	if err != nil {
		m.status, m.statusErr = err.Error(), true
		return
	}
	m.status, m.statusErr = status, false
}

// switchTab activates t and loads it when it was never loaded or changed.
func (m Model) switchTab(t tab) (tea.Model, tea.Cmd) {
	m.active = t
	m.clampCursor(t)
	st := m.tabs[t]
	if st.loading || (st.loaded && !st.stale) {
		cmd := m.loadStreak()
		return m, cmd
	}
	return m.reload(t)
}

func (m Model) reload(t tab) (tea.Model, tea.Cmd) {
	m.tabs[t].loading = true
	return m, m.load(t)
}

// load fetches tab t and, alongside it, today's progress, so the bar
// follows every load, reload and write.
func (m Model) load(t tab) tea.Cmd {
	ctx, svc := m.ctx, m.svc
	return tea.Batch(
		func() tea.Msg {
			rows, err := fetchRows(ctx, svc, t)
			return loadedMsg{tab: t, rows: rows, err: err}
		},
		func() tea.Msg {
			progress, err := svc.TodayProgress(ctx)
			return progressMsg{progress: progress, err: err}
		},
	)
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

// rows returns the active tab's rows after search and priority filters.
func (m Model) rows() []row {
	return m.filteredRows(m.active)
}

func (m Model) filteredRows(t tab) []row {
	all := m.tabs[t].rows
	if m.search == "" && m.priority == "" {
		return all
	}

	search := strings.ToLower(m.search)
	keep := func(r row) bool {
		switch {
		case r.todo != nil:
			return strings.Contains(strings.ToLower(r.todo.Name), search) &&
				(m.priority == "" || r.todo.Priority == m.priority)
		case r.dump != nil:
			return strings.Contains(strings.ToLower(r.dump.Name), search)
		}
		return false
	}

	// keep a header only if at least one of its items survives
	out := make([]row, 0, len(all))
	pendingHeader := -1
	for i, r := range all {
		if !r.selectable() {
			pendingHeader = i
			continue
		}
		if !keep(r) {
			continue
		}
		if pendingHeader >= 0 {
			out = append(out, all[pendingHeader])
			pendingHeader = -1
		}
		out = append(out, r)
	}
	return out
}

// selected returns the row under the cursor, if any.
func (m Model) selected() (row, bool) {
	rows := m.rows()
	cursor := m.tabs[m.active].cursor
	if cursor < 0 || cursor >= len(rows) || !rows[cursor].selectable() {
		return row{}, false
	}
	return rows[cursor], true
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

// clampCursor keeps t's cursor on a selectable row after its rows changed.
func (m *Model) clampCursor(t tab) {
	rows := m.filteredRows(t)
	cursor := min(m.tabs[t].cursor, len(rows)-1)
	cursor = max(cursor, 0)
	if cursor < len(rows) && !rows[cursor].selectable() {
		if next := moveCursor(rows, cursor, 1); next != cursor {
			cursor = next
		} else {
			cursor = moveCursor(rows, cursor, -1)
		}
	}
	m.tabs[t].cursor = cursor
}

// loadStreak fetches the streak for the selected daily task while the
// details pane is open.
func (m *Model) loadStreak() tea.Cmd {
	r, ok := m.selected()
	if !m.details || !ok || r.todo == nil || !r.todo.IsDaily {
		m.streakName, m.streakKnown = "", false
		return nil
	}
	if r.todo.Name == m.streakName {
		return nil
	}

	m.streakName, m.streakKnown = r.todo.Name, false
	ctx, svc, name := m.ctx, m.svc, r.todo.Name
	return func() tea.Msg {
		_, days, err := svc.CurrentStreak(ctx, name)
		if err != nil {
			return streakMsg{name: name, err: err}
		}
		_, periods, err := svc.StreakHistory(ctx, name)
		return streakMsg{name: name, days: days, periods: periods, err: err}
	}
}
