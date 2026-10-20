package tui

import (
	"context"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
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
	EditTodo(ctx context.Context, name string, patch tasks.TodoPatch) (tasks.TodoView, error)
	Complete(ctx context.Context, name, notes string) (tasks.TodoView, error)
	CompleteAll(ctx context.Context, name, notes string) (tasks.TodoView, error)
	DeleteByName(ctx context.Context, name string) (tasks.TodoView, error)
	AddDump(ctx context.Context, name string) (tasks.DumpEntry, error)
	DeleteDump(ctx context.Context, id uuid.UUID) error
	PromoteDump(ctx context.Context, id uuid.UUID, in tasks.NewTodo) (tasks.TodoView, error)
	CurrentStreak(ctx context.Context, name string) (tasks.TodoView, int64, error)
	TodayProgress(ctx context.Context) (tasks.Progress, error)
	StreakHistory(ctx context.Context, name string) (tasks.TodoView, []tasks.StreakPeriod, error)
	Notes(ctx context.Context, name string, limit int) (tasks.TodoView, []tasks.NoteEntry, error)
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
	modeNotesView
	modeQuote
	modeConfirmComplete
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
	status    string
	completed bool // a task was completed: celebrate with a quote
	err       error
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

// notePrefillMsg carries today's note of a task, to prefill the prompt
// that replaces it.
type notePrefillMsg struct {
	name string
	text string
}

// notesMsg carries a task's notes: the last few for the details panel, or
// all of them (full) for the notes window.
type notesMsg struct {
	name    string
	entries []tasks.NoteEntry
	full    bool
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
	// targetDump is the dump entry the delete confirmation acts on; when
	// set, x on the Dump tab deletes it instead of a task
	targetDump *tasks.DumpEntry
	// replacingNote: the notes prompt rewrites today's note of a task that
	// is already done today
	replacingNote bool
	// completeAll: the notes prompt completes a parent and its pending
	// subtasks, as confirmed in modeConfirmComplete
	completeAll bool

	// folded holds the parents whose subtasks are hidden (z)
	folded map[uuid.UUID]bool

	search   string
	priority string

	progress      tasks.Progress
	progressKnown bool

	// details is the panel opened with enter; detailsName is the task its
	// extra data (streak, notes) was loaded for
	details       bool
	detailsName   string
	streakDays    int64
	streakHistory []tasks.StreakPeriod
	streakKnown   bool
	detailNotes   []tasks.NoteEntry
	notesKnown    bool

	// notesView is the scrollable window opened with n
	notesView viewport.Model

	// quote is shown in a pop-up after completing a task (quoteTitle says
	// which) or on demand with Q (no title)
	quote      string
	quoteTitle string

	// saveTheme persists the theme chosen with `t`; tests swap it so they
	// never write the real config.
	saveTheme func(name string) error

	status    string
	statusErr bool

	// backup is the status bar's note on the last backup, read once at
	// start; backupOn is false when no backup folder is set
	backup   string
	backupOn bool
}

// New builds the root model with the saved keymap. ctx is used for every
// query, so cancelling it aborts in-flight loads.
func New(ctx context.Context, svc Service) Model {
	m := newModel(ctx, svc, CurrentKeymap())
	m.backup, m.backupOn = lastBackup(time.Now())
	return m
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
		if m.mode == modeNotesView {
			m.notesView.SetWidth(m.notesWidth())
			m.notesView.SetHeight(m.notesHeight(lipgloss.Height(m.notesView.GetContent())))
		}
		return m, nil

	case loadedMsg:
		st := &m.tabs[msg.tab]
		st.loading = false
		st.loaded = true
		st.stale = false
		st.err = msg.err
		st.rows = msg.rows
		m.clampCursor(msg.tab)
		cmd := m.loadDetails()
		return m, cmd

	case actionMsg:
		m.setStatus(msg.status, msg.err)
		if msg.err != nil {
			return m, nil
		}
		for t := range m.tabs {
			m.tabs[t].stale = true
		}
		m.detailsName = "" // the write may have changed the streak or notes
		if msg.completed {
			m.openQuote(msg.status)
		}
		return m.reload(m.active)

	case notePrefillMsg:
		if m.mode == modeNotes && m.replacingNote && msg.name == m.target.Path() && m.input.Value() == "" {
			m.input.SetValue(msg.text)
			m.input.CursorEnd()
		}
		return m, nil

	case notesMsg:
		switch {
		case msg.full && m.mode == modeNotesView && msg.name == m.target.Path():
			m.setNotesContent(msg.entries, msg.err)
		case !msg.full && msg.err == nil && msg.name == m.detailsName:
			m.detailNotes, m.notesKnown = msg.entries, true
		}
		return m, nil

	case progressMsg:
		if msg.err == nil {
			m.progress, m.progressKnown = msg.progress, true
		}
		return m, nil

	case streakMsg:
		if msg.err == nil && msg.name == m.detailsName {
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
		case modeConfirmComplete:
			return m.handleConfirmCompleteKey(msg)
		case modeNotesView:
			return m.handleNotesViewKey(msg)
		case modeQuote:
			return m.handleQuoteKey(msg)
		default:
			return m.handlePromptKey(msg)
		}
	}

	// anything else (cursor blink) belongs to whichever input is open
	var cmd tea.Cmd
	switch m.mode {
	case modeNormal, modeConfirmDelete, modeConfirmComplete, modeQuote:
	case modeNotesView:
		m.notesView, cmd = m.notesView.Update(msg)
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
		cmd := m.loadDetails()
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

// todoRows lists each task followed by its subtasks.
func todoRows(todos []tasks.TodoView) []row {
	rows := make([]row, 0, len(todos))
	for i := range todos {
		rows = append(rows, row{todo: &todos[i]})
		for j := range todos[i].Subtasks {
			rows = append(rows, row{todo: &todos[i].Subtasks[j]})
		}
	}
	return rows
}

// rows returns the active tab's rows after search and priority filters.
func (m Model) rows() []row {
	return m.filteredRows(m.active)
}

func (m Model) filteredRows(t tab) []row {
	all := m.tabs[t].rows
	if m.search == "" && m.priority == "" && len(m.folded) == 0 {
		return all
	}

	search := strings.ToLower(m.search)
	keep := func(r row) bool {
		switch {
		case r.todo != nil:
			if r.todo.IsSubtask() && m.folded[r.todo.ParentID] {
				return false
			}
			// the full name, so searching a parent shows its subtasks too
			return strings.Contains(strings.ToLower(r.todo.Path()), search) &&
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

// detailsNotesLimit is how many notes the details panel shows.
const detailsNotesLimit = 3

// loadDetails fetches the extra data of the details panel for the selected
// task: its last notes, plus the streak and streak history when it is
// daily. It does nothing while the panel is closed or already loaded.
func (m *Model) loadDetails() tea.Cmd {
	r, ok := m.selected()
	if !m.details || !ok || r.todo == nil {
		m.detailsName, m.streakKnown, m.notesKnown = "", false, false
		return nil
	}
	if r.todo.Path() == m.detailsName {
		return nil
	}

	m.detailsName, m.streakKnown, m.notesKnown = r.todo.Path(), false, false
	ctx, svc, name := m.ctx, m.svc, r.todo.Path()

	cmds := []tea.Cmd{func() tea.Msg {
		_, entries, err := svc.Notes(ctx, name, detailsNotesLimit)
		return notesMsg{name: name, entries: entries, err: err}
	}}
	if r.todo.IsDaily {
		cmds = append(cmds, func() tea.Msg {
			_, days, err := svc.CurrentStreak(ctx, name)
			if err != nil {
				return streakMsg{name: name, err: err}
			}
			_, periods, err := svc.StreakHistory(ctx, name)
			return streakMsg{name: name, days: days, periods: periods, err: err}
		})
	}
	return tea.Batch(cmds...)
}
