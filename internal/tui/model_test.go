package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/google/uuid"
	"github.com/luis-octavius/akrasia/internal/backup"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

type fakeService struct {
	today    tasks.TodayView
	all      []tasks.TodoView
	daily    []tasks.TodoView
	dumps    []tasks.DumpEntry
	streak   int64
	periods  []tasks.StreakPeriod
	progress tasks.Progress
	notes    []tasks.NoteEntry
	err      error

	calls        map[string]int
	edited       []editCall
	dumpsDeleted []uuid.UUID
	promoted     []tasks.NewTodo
	completed    []string // "name|notes"
	deleted      []string
	added        []tasks.NewTodo
	dumped       []string
}

func (f *fakeService) called(name string) {
	if f.calls == nil {
		f.calls = map[string]int{}
	}
	f.calls[name]++
}

func (f *fakeService) Today(context.Context, tasks.TodayOptions) (tasks.TodayView, error) {
	f.called("today")
	return f.today, f.err
}

func (f *fakeService) ListTodos(context.Context, string) ([]tasks.TodoView, error) {
	f.called("all")
	return f.all, f.err
}

func (f *fakeService) DailyTodos(context.Context) ([]tasks.TodoView, error) {
	f.called("daily")
	return f.daily, f.err
}

func (f *fakeService) Dumps(context.Context) ([]tasks.DumpEntry, error) {
	f.called("dumps")
	return f.dumps, f.err
}

func (f *fakeService) AddTodo(_ context.Context, in tasks.NewTodo) (tasks.TodoView, error) {
	f.added = append(f.added, in)
	return tasks.TodoView{Name: in.Name}, nil
}

func (f *fakeService) EditTodo(_ context.Context, name string, patch tasks.TodoPatch) (tasks.TodoView, error) {
	f.edited = append(f.edited, editCall{name, patch})
	return tasks.TodoView{Name: name}, nil
}

func (f *fakeService) Complete(_ context.Context, name, notes string) (tasks.TodoView, error) {
	f.completed = append(f.completed, name+"|"+notes)
	f.progress.Done++
	return tasks.TodoView{Name: name}, nil
}

func (f *fakeService) DeleteByName(_ context.Context, name string) (tasks.TodoView, error) {
	f.deleted = append(f.deleted, name)
	return tasks.TodoView{Name: name}, nil
}

func (f *fakeService) DeleteDump(_ context.Context, id uuid.UUID) error {
	f.dumpsDeleted = append(f.dumpsDeleted, id)
	return nil
}

func (f *fakeService) PromoteDump(_ context.Context, id uuid.UUID, in tasks.NewTodo) (tasks.TodoView, error) {
	f.promoted = append(f.promoted, in)
	f.dumpsDeleted = append(f.dumpsDeleted, id)
	return tasks.TodoView{Name: in.Name}, nil
}

func (f *fakeService) AddDump(_ context.Context, name string) (tasks.DumpEntry, error) {
	f.dumped = append(f.dumped, name)
	return tasks.DumpEntry{Name: name}, nil
}

func (f *fakeService) CurrentStreak(_ context.Context, name string) (tasks.TodoView, int64, error) {
	f.called("streak")
	return tasks.TodoView{Name: name}, f.streak, nil
}

func (f *fakeService) TodayProgress(context.Context) (tasks.Progress, error) {
	f.called("progress")
	return f.progress, f.err
}

func (f *fakeService) Notes(_ context.Context, name string, limit int) (tasks.TodoView, []tasks.NoteEntry, error) {
	f.called("notes")
	entries := f.notes
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}
	return tasks.TodoView{Name: name}, entries, nil
}

func (f *fakeService) StreakHistory(_ context.Context, name string) (tasks.TodoView, []tasks.StreakPeriod, error) {
	return tasks.TodoView{Name: name}, f.periods, nil
}

type editCall struct {
	name  string
	patch tasks.TodoPatch
}

// send runs each msg through Update and keeps feeding back the model's own
// messages (loads, writes, streaks, progress), including the ones inside a
// tea.Batch, the way the runtime would. Other commands, like cursor
// blinks, are not run.
func send(t *testing.T, m Model, msgs ...tea.Msg) Model {
	t.Helper()
	for _, msg := range msgs {
		queue := []tea.Msg{msg}
		for steps := 0; len(queue) > 0 && steps < 20; steps++ {
			next, cmd := m.Update(queue[0])
			m = next.(Model)
			queue = append(queue[1:], ownMessages(cmd)...)
		}
	}
	return m
}

// ownMessages runs cmd (and every command of a batch) and keeps only the
// messages the model produces itself.
func ownMessages(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	switch out := runQuick(cmd).(type) {
	case tea.BatchMsg:
		var msgs []tea.Msg
		for _, c := range out {
			msgs = append(msgs, ownMessages(c)...)
		}
		return msgs
	case loadedMsg, actionMsg, streakMsg, progressMsg, notesMsg, notePrefillMsg:
		return []tea.Msg{out}
	}
	return nil
}

// runQuick runs cmd but gives up on slow ones: cursor blinks sleep for
// half a second and never produce a message the tests care about.
func runQuick(cmd tea.Cmd) tea.Msg {
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()
	select {
	case msg := <-ch:
		return msg
	case <-time.After(50 * time.Millisecond):
		return nil
	}
}

func start(t *testing.T, svc Service, layout string) Model {
	t.Helper()
	m := newModel(context.Background(), svc, layout)
	m.saveTheme = func(name string) error { color.SetTheme(name); return nil }
	return send(t, m, ownMessages(m.Init())...)
}

func runes(s string) tea.KeyPressMsg { return tea.KeyPressMsg{Code: []rune(s)[0], Text: s} }

var (
	keyDown  = tea.KeyPressMsg{Code: tea.KeyDown}
	keyUp    = tea.KeyPressMsg{Code: tea.KeyUp}
	keyTab   = tea.KeyPressMsg{Code: tea.KeyTab}
	keyEnter = tea.KeyPressMsg{Code: tea.KeyEnter}
	keyEsc   = tea.KeyPressMsg{Code: tea.KeyEscape}
	keySpace = tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
)

func todo(name string) tasks.TodoView {
	return tasks.TodoView{Name: name, ExpiresAt: time.Now().Add(48 * time.Hour)}
}

func TestTodayTabLoadsSectionsAndSkipsHeaders(t *testing.T) {
	svc := &fakeService{today: tasks.TodayView{
		Overdue:      []tasks.TodoView{todo("late")},
		ExpiringSoon: []tasks.TodoView{todo("soon")},
	}}

	m := start(t, svc, KeymapDefault)

	// header, late, header, soon
	if got := len(m.rows()); got != 4 {
		t.Fatalf("expected 4 rows, got %d", got)
	}
	if got := m.tabs[tabToday].cursor; got != 1 {
		t.Fatalf("cursor should start on the first task, got %d", got)
	}

	m = send(t, m, keyDown)
	if got := m.tabs[tabToday].cursor; got != 3 {
		t.Fatalf("down should skip the header and land on 3, got %d", got)
	}

	m = send(t, m, keyDown)
	if got := m.tabs[tabToday].cursor; got != 3 {
		t.Fatalf("down at the end should stay, got %d", got)
	}

	m = send(t, m, keyUp)
	if got := m.tabs[tabToday].cursor; got != 1 {
		t.Fatalf("up should skip the header back to 1, got %d", got)
	}

	view := m.View().Content
	for _, want := range []string{"late", "soon"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q:\n%s", want, view)
		}
	}
}

func TestVimKeysOnlyInVimLayout(t *testing.T) {
	svc := &fakeService{all: []tasks.TodoView{todo("a"), todo("b"), todo("c")}}

	def := send(t, start(t, svc, KeymapDefault), keyTab, runes("j"))
	if got := def.tabs[tabAll].cursor; got != 0 {
		t.Fatalf("j should do nothing in the default layout, cursor %d", got)
	}

	vim := send(t, start(t, svc, KeymapVim), keyTab, runes("j"), runes("j"))
	if got := vim.tabs[tabAll].cursor; got != 2 {
		t.Fatalf("j should move down in the vim layout, cursor %d", got)
	}

	vim = send(t, vim, runes("g"))
	if got := vim.tabs[tabAll].cursor; got != 0 {
		t.Fatalf("g should jump to the top, cursor %d", got)
	}

	// arrows keep working with vim on
	vim = send(t, vim, keyDown)
	if got := vim.tabs[tabAll].cursor; got != 1 {
		t.Fatalf("down arrow should still work, cursor %d", got)
	}
}

func TestTabSwitchLoadsOnce(t *testing.T) {
	svc := &fakeService{all: []tasks.TodoView{todo("a"), todo("b")}}

	m := start(t, svc, KeymapDefault)
	m = send(t, m, keyTab)

	if m.active != tabAll || len(m.rows()) != 2 {
		t.Fatalf("expected All tab loaded with 2 rows, got tab %v with %d rows", m.active, len(m.rows()))
	}

	m = send(t, m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}, keyTab)
	if svc.calls["all"] != 1 {
		t.Fatalf("All tab should load once, loaded %d times", svc.calls["all"])
	}

	send(t, m, runes("r"))
	if svc.calls["all"] != 2 {
		t.Fatalf("reload should fetch again, loaded %d times", svc.calls["all"])
	}
}

func TestCompleteAsksForNotes(t *testing.T) {
	svc := &fakeService{all: []tasks.TodoView{todo("taxes")}}
	m := send(t, start(t, svc, KeymapDefault), keyTab)

	m = send(t, m, keySpace)
	if m.mode != modeNotes {
		t.Fatalf("space should open the notes prompt, mode %v", m.mode)
	}

	m = send(t, m, runes("paid"), keyEnter)
	if len(svc.completed) != 1 || svc.completed[0] != "taxes|paid" {
		t.Fatalf("expected taxes completed with notes, got %v", svc.completed)
	}
	if m.mode != modeQuote || !strings.Contains(m.status, "taxes") {
		t.Fatalf("expected the quote pop-up and a status, got mode %v status %q", m.mode, m.status)
	}
	if svc.calls["all"] != 2 {
		t.Fatalf("a write should reload the active tab, loaded %d times", svc.calls["all"])
	}
}

func TestCompleteCanBeCancelled(t *testing.T) {
	svc := &fakeService{all: []tasks.TodoView{todo("taxes")}}
	m := send(t, start(t, svc, KeymapDefault), keyTab, keySpace, keyEsc)

	if m.mode != modeNormal || len(svc.completed) != 0 {
		t.Fatalf("esc should cancel, mode %v completed %v", m.mode, svc.completed)
	}
}

func TestCompleteSkipsDoneTasks(t *testing.T) {
	done := todo("taxes")
	done.Concluded = true
	svc := &fakeService{all: []tasks.TodoView{done}}

	m := send(t, start(t, svc, KeymapDefault), keyTab, keySpace)
	if m.mode != modeNormal || !strings.Contains(m.status, "taxes") {
		t.Fatalf("done task should only show a status, mode %v status %q", m.mode, m.status)
	}
}

func TestDeleteNeedsConfirmation(t *testing.T) {
	svc := &fakeService{all: []tasks.TodoView{todo("taxes")}}
	m := send(t, start(t, svc, KeymapDefault), keyTab)

	m = send(t, m, runes("x"), runes("n"))
	if len(svc.deleted) != 0 || m.mode != modeNormal {
		t.Fatalf("n should cancel the delete, deleted %v", svc.deleted)
	}

	m = send(t, m, runes("x"))
	if !strings.Contains(m.View().Content, "taxes") || m.mode != modeConfirmDelete {
		t.Fatalf("x should ask for confirmation")
	}

	send(t, m, runes("y"))
	if len(svc.deleted) != 1 || svc.deleted[0] != "taxes" {
		t.Fatalf("y should delete taxes, deleted %v", svc.deleted)
	}
}

func TestAddTodoForm(t *testing.T) {
	svc := &fakeService{}
	m := send(t, start(t, svc, KeymapDefault), runes("a"))
	if m.mode != modeAddTodo {
		t.Fatalf("a should open the form, mode %v", m.mode)
	}

	// name, description, priority (invalid), daily, date
	m = send(t, m, runes("write"), keyEnter, keyEnter, runes("urgent"), keyEnter, runes("y"), keyEnter, keyEnter)
	if m.form.err == nil || len(svc.added) != 0 {
		t.Fatalf("invalid priority should keep the form open with an error")
	}

	// fix the priority: back up three fields, clear it, type a valid one
	m = send(t, m, keyUp, keyUp)
	for range "urgent" {
		m = send(t, m, tea.KeyPressMsg{Code: tea.KeyBackspace})
	}
	m = send(t, m, runes("high"), keyDown, keyDown, keyEnter)

	if len(svc.added) != 1 {
		t.Fatalf("expected one task added, got %v (form err %v)", svc.added, m.form.err)
	}
	got := svc.added[0]
	if got.Name != "write" || got.Priority != "high" || !got.IsDaily || got.ExpiresAt.Before(time.Now()) {
		t.Fatalf("unexpected task: %+v", got)
	}
	if m.mode != modeNormal {
		t.Fatalf("form should close after submit, mode %v", m.mode)
	}
}

func TestAddDumpOnDumpTab(t *testing.T) {
	svc := &fakeService{}
	m := send(t, start(t, svc, KeymapDefault), tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.active != tabDump {
		t.Fatalf("expected Dump tab, got %v", m.active)
	}

	send(t, m, runes("a"), runes("idea"), keyEnter)
	if len(svc.dumped) != 1 || svc.dumped[0] != "idea" {
		t.Fatalf("expected dump added, got %v", svc.dumped)
	}
}

func TestSearchAndPriorityFilter(t *testing.T) {
	high := todo("taxes")
	high.Priority = "high"
	low := todo("groceries")
	low.Priority = "low"
	svc := &fakeService{today: tasks.TodayView{
		Overdue:      []tasks.TodoView{high},
		ExpiringSoon: []tasks.TodoView{low},
	}}
	m := start(t, svc, KeymapDefault)

	m = send(t, m, runes("/"), runes("groc"))
	rows := m.rows()
	if len(rows) != 2 || rows[1].todo.Name != "groceries" {
		t.Fatalf("search should keep groceries and its header only, got %d rows", len(rows))
	}
	if got := m.tabs[tabToday].cursor; got != 1 {
		t.Fatalf("cursor should move onto the match, got %d", got)
	}

	m = send(t, m, keyEnter, keyEsc)
	if m.search != "" || len(m.rows()) != 4 {
		t.Fatalf("esc should clear the search, search %q rows %d", m.search, len(m.rows()))
	}

	m = send(t, m, runes("p"))
	if m.priority != "high" || len(m.rows()) != 2 || m.rows()[1].todo.Name != "taxes" {
		t.Fatalf("p should filter by high priority, got %q with %d rows", m.priority, len(m.rows()))
	}
}

func TestDetailsShowStreakForDaily(t *testing.T) {
	daily := todo("read")
	daily.IsDaily = true
	svc := &fakeService{daily: []tasks.TodoView{daily}, streak: 7}

	m := send(t, start(t, svc, KeymapDefault), keyTab, keyTab, keyEnter)
	m.width = 100

	if svc.calls["streak"] != 1 {
		t.Fatalf("opening details should fetch the streak once, got %d", svc.calls["streak"])
	}
	if !strings.Contains(m.View().Content, "7") {
		t.Fatalf("details should show the streak:\n%s", m.View().Content)
	}
}

func TestLoadErrorIsShown(t *testing.T) {
	m := start(t, &fakeService{err: errors.New("database is locked")}, KeymapDefault)

	if !strings.Contains(m.View().Content, "database is locked") {
		t.Fatalf("expected error in view:\n%s", m.View().Content)
	}
}

func TestQuitOnlyInNormalMode(t *testing.T) {
	m := start(t, &fakeService{}, KeymapDefault)

	_, cmd := m.Update(runes("q"))
	if cmd == nil {
		t.Fatal("q should return a command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("q should quit")
	}

	// inside the search prompt q is just a letter
	m = send(t, m, runes("/"), runes("q"))
	if m.search != "q" {
		t.Fatalf("q should be typed into the search, got %q", m.search)
	}
}

func TestVisibleRangeKeepsCursorOnScreen(t *testing.T) {
	cases := []struct{ total, cursor, height, start, end int }{
		{10, 0, 0, 0, 10},
		{3, 2, 5, 0, 3},
		{10, 4, 5, 0, 5},
		{10, 7, 5, 3, 8},
		{10, 9, 5, 5, 10},
	}
	for _, c := range cases {
		start, end := visibleRange(c.total, c.cursor, c.height)
		if start != c.start || end != c.end {
			t.Errorf("visibleRange(%d, %d, %d) = %d,%d, want %d,%d",
				c.total, c.cursor, c.height, start, end, c.start, c.end)
		}
	}
}

func TestDateParts(t *testing.T) {
	cases := map[string][]int{"": nil, "5": {5}, "5/11": {5, 11}, " 7 / 12 ": {7, 12}, "5/1/2027": {5, 1, 2027}}
	for in, want := range cases {
		got, err := dateParts(in)
		if err != nil || len(got) != len(want) {
			t.Fatalf("dateParts(%q) = %v, %v; want %v", in, got, err, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("dateParts(%q) = %v, want %v", in, got, want)
			}
		}
	}

	for _, bad := range []string{"x", "1/2/3/4", "5/"} {
		if _, err := dateParts(bad); err == nil {
			t.Errorf("dateParts(%q) should fail", bad)
		}
	}
}

func TestThemeKeyCyclesAndReportsTheme(t *testing.T) {
	orig := color.GetCurrentTheme().Name
	t.Cleanup(func() { color.SetTheme(orig) })
	color.SetTheme("default")

	m := send(t, start(t, &fakeService{}, KeymapDefault), runes("t"))

	want := color.NextTheme("default")
	if got := color.GetCurrentTheme().Name; got != want {
		t.Fatalf("t should switch to %q, got %q", want, got)
	}
	if !strings.Contains(m.status, want) {
		t.Fatalf("status should name the new theme, got %q", m.status)
	}
}

// Every theme must render without panics, and no line may be wider than
// the terminal, even on narrow screens and with every overlay open.
func TestViewFitsEveryThemeAndWidth(t *testing.T) {
	orig := color.GetCurrentTheme().Name
	t.Cleanup(func() { color.SetTheme(orig) })

	daily := todo("a daily task with a fairly long name to force truncation")
	daily.IsDaily = true
	high := todo("taxes")
	high.Priority = "high"
	svc := &fakeService{
		today:   tasks.TodayView{Overdue: []tasks.TodoView{high}, Daily: []tasks.TodoView{daily}},
		streak:  3,
		periods: []tasks.StreakPeriod{{StartDate: "2026-10-01", EndDate: "2026-10-03", Days: 3}},
	}

	for _, theme := range color.GetAvailableThemes() {
		color.SetTheme(theme)
		for _, width := range []int{40, 60, 100} {
			m := start(t, svc, KeymapDefault)
			m = send(t, m, tea.WindowSizeMsg{Width: width, Height: 20})

			screens := map[string]Model{
				"list":    m,
				"details": send(t, m, keyDown, keyEnter),
				"delete":  send(t, m, runes("x")),
				"form":    send(t, m, runes("a")),
				"help":    send(t, m, runes("?")),
				"notes":   send(t, m, runes("n")),
				"quote":   send(t, m, runes("Q")),
			}
			for name, s := range screens {
				for i, l := range strings.Split(s.View().Content, "\n") {
					if w := lipgloss.Width(l); w > width {
						t.Fatalf("%s/%d/%s: line %d is %d wide:\n%s", theme, width, name, i, w, l)
					}
				}
			}
		}
	}
}

func TestDetailsListStreakHistory(t *testing.T) {
	daily := todo("read")
	daily.IsDaily = true
	svc := &fakeService{
		daily:   []tasks.TodoView{daily},
		streak:  4,
		periods: []tasks.StreakPeriod{{StartDate: "2026-09-01", EndDate: "2026-09-10", Days: 10}},
	}

	m := send(t, start(t, svc, KeymapDefault), tea.WindowSizeMsg{Width: 100, Height: 24}, keyTab, keyTab, keyEnter)

	view := m.View().Content
	for _, want := range []string{"2026-09-01", "2026-09-10", "10"} {
		if !strings.Contains(view, want) {
			t.Fatalf("details should list the streak history (%q missing):\n%s", want, view)
		}
	}
}

func TestModalFloatsOverList(t *testing.T) {
	svc := &fakeService{all: []tasks.TodoView{todo("taxes"), todo("groceries")}}
	m := send(t, start(t, svc, KeymapDefault), tea.WindowSizeMsg{Width: 100, Height: 24}, keyTab, runes("x"))

	view := m.View().Content
	if !strings.Contains(view, "groceries") {
		t.Fatalf("the list should stay visible behind the modal:\n%s", view)
	}
	if !strings.Contains(view, `"taxes"`) {
		t.Fatalf("the delete modal should name the task:\n%s", view)
	}
	if got := lipgloss.Height(view); got != 24 {
		t.Fatalf("screen should fill the terminal height, got %d lines", got)
	}
}

// The full help spans several lines; all of them must share the same
// one-column left margin, not just the first.
func TestFullHelpLinesShareMargin(t *testing.T) {
	m := send(t, start(t, &fakeService{}, KeymapDefault), tea.WindowSizeMsg{Width: 120, Height: 30}, runes("?"))

	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	var help []string
	for i, l := range lines {
		if strings.Contains(l, "NORMAL") {
			help = lines[i+1:] // everything below the status bar
			break
		}
	}
	if len(help) < 2 {
		t.Fatalf("full help should span several rows:\n%s", strings.Join(lines, "\n"))
	}
	for i, l := range help {
		if margin := len(l) - len(strings.TrimLeft(l, " ")); margin != 1 {
			t.Fatalf("help row %d starts at column %d, want 1:\n%s", i, margin, strings.Join(help, "\n"))
		}
	}
}

func TestProgressLineFollowsCompletions(t *testing.T) {
	svc := &fakeService{
		all:      []tasks.TodoView{todo("taxes")},
		progress: tasks.Progress{Done: 1, Total: 3},
	}
	m := send(t, start(t, svc, KeymapDefault), tea.WindowSizeMsg{Width: 100, Height: 24})

	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "Today 1/3") || !strings.Contains(view, "33%") {
		t.Fatalf("expected the progress line with 1/3 and 33%%:\n%s", view)
	}

	m = send(t, m, keyTab, keySpace, keyEnter)
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "Today 2/3") {
		t.Fatalf("progress should refresh after completing a task:\n%s", view)
	}
}

func TestProgressLineWithNothingDue(t *testing.T) {
	m := send(t, start(t, &fakeService{}, KeymapDefault), tea.WindowSizeMsg{Width: 100, Height: 24})

	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "Today: nothing due") || !strings.Contains(view, "100%") {
		t.Fatalf("an empty day should read as done:\n%s", view)
	}
}

func keyBackspace() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeyBackspace} }

func TestEditSendsOnlyChangedFields(t *testing.T) {
	taxes := todo("taxes")
	taxes.Priority = "high"
	svc := &fakeService{all: []tasks.TodoView{taxes}}
	m := send(t, start(t, svc, KeymapDefault), keyTab, runes("e"))

	if m.mode != modeAddTodo || m.form.editing == nil {
		t.Fatalf("e should open the edit form")
	}
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "Edit task") || !strings.Contains(view, "taxes") {
		t.Fatalf("form should be titled Edit task and prefilled:\n%s", view)
	}

	// priority field: clear "high", type "low", then enter through to save
	m = send(t, m, keyDown, keyDown)
	for range "high" {
		m = send(t, m, keyBackspace())
	}
	m = send(t, m, runes("low"), keyEnter, keyEnter, keyEnter)

	if len(svc.edited) != 1 {
		t.Fatalf("expected one edit, got %d (form err %v)", len(svc.edited), m.form.err)
	}
	got := svc.edited[0]
	if got.name != "taxes" || got.patch.Priority == nil || *got.patch.Priority != "low" {
		t.Fatalf("expected priority low for taxes, got %+v", got)
	}
	if got.patch.Name != nil || got.patch.Description != nil || got.patch.IsDaily != nil || got.patch.ExpiresAt != nil {
		t.Fatalf("only the priority should change, got %+v", got.patch)
	}
	if !strings.Contains(m.status, "taxes") {
		t.Fatalf("status should confirm the update, got %q", m.status)
	}
}

func TestEditWithoutChangesSavesNothing(t *testing.T) {
	// an overdue task: its prefilled date is in the past and must not be
	// re-validated when left untouched
	late := tasks.TodoView{Name: "late", ExpiresAt: time.Now().AddDate(0, 0, -3)}
	svc := &fakeService{all: []tasks.TodoView{late}}

	m := send(t, start(t, svc, KeymapDefault), keyTab, runes("e"), keyEnter, keyEnter, keyEnter, keyEnter, keyEnter)

	if m.form.err != nil {
		t.Fatalf("untouched form should not fail validation: %v", m.form.err)
	}
	if m.mode != modeNormal || len(svc.edited) != 0 {
		t.Fatalf("no change should just close the form, mode %v edits %v", m.mode, svc.edited)
	}
}

func TestEditDailyToggle(t *testing.T) {
	daily := todo("read")
	daily.IsDaily = true
	svc := &fakeService{all: []tasks.TodoView{daily}}

	// daily field: y -> n
	m := send(t, start(t, svc, KeymapDefault), keyTab, runes("e"), keyDown, keyDown, keyDown, keyBackspace(), runes("n"), keyEnter, keyEnter)

	if len(svc.edited) != 1 || svc.edited[0].patch.IsDaily == nil || *svc.edited[0].patch.IsDaily {
		t.Fatalf("expected daily=false, got %+v (form err %v)", svc.edited, m.form.err)
	}
}

func rustNotes(n int) []tasks.NoteEntry {
	entries := make([]tasks.NoteEntry, n)
	for i := range entries {
		entries[i] = tasks.NoteEntry{Date: fmt.Sprintf("2026-10-%02d", 30-i), Notes: fmt.Sprintf("lesson %d", n-i)}
	}
	entries[n-1] = tasks.NoteEntry{Date: entries[n-1].Date, Backfilled: true}
	return entries
}

func TestDetailsShowRecentNotes(t *testing.T) {
	rust := todo("rust")
	rust.IsDaily = true
	svc := &fakeService{daily: []tasks.TodoView{rust}, notes: rustNotes(6)}

	m := send(t, start(t, svc, KeymapDefault), tea.WindowSizeMsg{Width: 110, Height: 30}, keyTab, keyTab, keyEnter)

	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "Recent notes") || !strings.Contains(view, "lesson 6") || !strings.Contains(view, "lesson 4") {
		t.Fatalf("details should list the last notes:\n%s", view)
	}
	if strings.Contains(view, "lesson 3") {
		t.Fatalf("details should stop at %d notes:\n%s", detailsNotesLimit, view)
	}
}

func TestNotesWindowOpensScrollsAndCloses(t *testing.T) {
	rust := todo("rust")
	svc := &fakeService{all: []tasks.TodoView{rust}, notes: rustNotes(30)}

	m := send(t, start(t, svc, KeymapDefault), tea.WindowSizeMsg{Width: 100, Height: 20}, keyTab, runes("n"))
	if m.mode != modeNotesView {
		t.Fatalf("n should open the notes window, mode %v", m.mode)
	}

	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "Notes · rust") || !strings.Contains(view, "lesson 30") || strings.Contains(view, "backfilled") {
		t.Fatalf("window should show the newest notes first and fit the screen:\n%s", view)
	}
	if got := lipgloss.Height(m.View().Content); got != 20 {
		t.Fatalf("screen should keep the terminal height, got %d", got)
	}

	for range 40 {
		m = send(t, m, keyDown)
	}
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "backfilled") {
		t.Fatalf("scrolling down should reach the oldest (backfilled) day:\n%s", view)
	}

	m = send(t, m, keyEsc)
	if m.mode != modeNormal {
		t.Fatalf("esc should close the window, mode %v", m.mode)
	}
}

func TestDetailsNotesRefreshAfterCompleting(t *testing.T) {
	rust := todo("rust")
	svc := &fakeService{all: []tasks.TodoView{rust}}
	m := send(t, start(t, svc, KeymapDefault), tea.WindowSizeMsg{Width: 110, Height: 30}, keyTab, keyEnter)

	svc.notes = []tasks.NoteEntry{{Date: "2026-10-19", Notes: "ownership"}}
	m = send(t, m, keySpace, runes("ownership"), keyEnter, keyEsc) // esc closes the quote

	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "ownership") {
		t.Fatalf("the new note should show in the details right away:\n%s", view)
	}
}

// A details panel with more lines than fit must be cut, not push the
// status bar and help off the screen.
func TestTallDetailsKeepScreenHeight(t *testing.T) {
	rust := todo("rust")
	rust.IsDaily = true
	periods := make([]tasks.StreakPeriod, 5)
	for i := range periods {
		periods[i] = tasks.StreakPeriod{StartDate: "2026-09-01", EndDate: "2026-09-09", Days: 9}
	}
	svc := &fakeService{daily: []tasks.TodoView{rust}, notes: rustNotes(3), streak: 9, periods: periods}

	m := send(t, start(t, svc, KeymapDefault), tea.WindowSizeMsg{Width: 110, Height: 16}, keyTab, keyTab, keyEnter)

	view := m.View().Content
	if got := lipgloss.Height(view); got != 16 {
		t.Fatalf("screen should be exactly 16 lines, got %d:\n%s", got, ansi.Strip(view))
	}
	if !strings.Contains(ansi.Strip(view), "NORMAL") {
		t.Fatalf("status bar should stay visible:\n%s", ansi.Strip(view))
	}
}

func TestCompletingOpensQuote(t *testing.T) {
	svc := &fakeService{all: []tasks.TodoView{todo("taxes")}}
	m := send(t, start(t, svc, KeymapDefault), tea.WindowSizeMsg{Width: 100, Height: 24}, keyTab, keySpace, keyEnter)

	if m.mode != modeQuote || m.quote == "" {
		t.Fatalf("completing should open a quote, mode %v quote %q", m.mode, m.quote)
	}
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "taxes completed") {
		t.Fatalf("the pop-up should say what was completed:\n%s", view)
	}
	if got := lipgloss.Height(m.View().Content); got != 24 {
		t.Fatalf("screen should keep the terminal height, got %d", got)
	}

	// other keys are ignored while it is open
	m = send(t, m, runes("x"), keyDown)
	if m.mode != modeQuote || len(svc.deleted) != 0 {
		t.Fatalf("keys behind the pop-up should be ignored, mode %v deleted %v", m.mode, svc.deleted)
	}

	m = send(t, m, keyEsc)
	if m.mode != modeNormal {
		t.Fatalf("esc should close the quote, mode %v", m.mode)
	}
}

func TestQuoteOnDemandAndOnlyForCompletions(t *testing.T) {
	svc := &fakeService{all: []tasks.TodoView{todo("taxes"), todo("groceries")}}
	m := send(t, start(t, svc, KeymapDefault), keyTab)

	m = send(t, m, runes("Q"))
	if m.mode != modeQuote || m.quoteTitle != "" {
		t.Fatalf("Q should open an untitled quote, mode %v title %q", m.mode, m.quoteTitle)
	}
	m = send(t, m, runes("q"))
	if m.mode != modeNormal {
		t.Fatalf("q should close the quote, not quit; mode %v", m.mode)
	}

	m = send(t, m, runes("x"), runes("y"))
	if m.mode == modeQuote {
		t.Fatal("deleting should not open a quote")
	}
	m = send(t, m, runes("a"), runes("new"), keyEnter, keyEnter, keyEnter, keyEnter, keyEnter)
	if m.mode == modeQuote {
		t.Fatal("adding should not open a quote")
	}
}

func TestSpaceOnTaskDoneTodayReplacesNote(t *testing.T) {
	read := todo("read")
	read.IsDaily, read.DoneToday = true, true
	today := time.Now().Format(time.DateOnly)
	svc := &fakeService{
		daily: []tasks.TodoView{read},
		notes: []tasks.NoteEntry{{Date: today, Notes: "chapter 1"}},
	}

	m := send(t, start(t, svc, KeymapDefault), tea.WindowSizeMsg{Width: 100, Height: 24}, keyTab, keyTab, keySpace)

	if m.mode != modeNotes || !m.replacingNote {
		t.Fatalf("space on a task done today should open the replace prompt, mode %v", m.mode)
	}
	if got := m.input.Value(); got != "chapter 1" {
		t.Fatalf("prompt should be prefilled with today's note, got %q", got)
	}
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "Replace today's note") {
		t.Fatalf("prompt should say it replaces the note:\n%s", view)
	}

	m = send(t, m, runes(" and 2"), keyEnter)
	if len(svc.completed) != 1 || svc.completed[0] != "read|chapter 1 and 2" {
		t.Fatalf("expected the note replaced, got %v", svc.completed)
	}
	if m.mode != modeNormal || !strings.Contains(m.status, "note") {
		t.Fatalf("replacing a note should not open a quote; mode %v status %q", m.mode, m.status)
	}
}

func TestReplacePromptStartsEmptyWithoutNoteToday(t *testing.T) {
	read := todo("read")
	read.IsDaily, read.DoneToday = true, true
	svc := &fakeService{
		daily: []tasks.TodoView{read},
		notes: []tasks.NoteEntry{{Date: "2026-01-01", Notes: "old"}},
	}

	m := send(t, start(t, svc, KeymapDefault), keyTab, keyTab, keySpace)
	if m.input.Value() != "" {
		t.Fatalf("a note from another day must not be prefilled, got %q", m.input.Value())
	}
}

func dumpTab(t *testing.T, svc *fakeService) Model {
	t.Helper()
	m := send(t, start(t, svc, KeymapDefault), tea.WindowSizeMsg{Width: 100, Height: 24}, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.active != tabDump {
		t.Fatalf("expected the Dump tab, got %v", m.active)
	}
	return m
}

func TestDeleteDumpNeedsConfirmation(t *testing.T) {
	idea := tasks.DumpEntry{ID: uuid.New(), Name: "app idea"}
	svc := &fakeService{dumps: []tasks.DumpEntry{idea}}
	m := dumpTab(t, svc)

	m = send(t, m, runes("x"))
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, `Delete dump "app idea"?`) {
		t.Fatalf("x on a dump should ask to delete it:\n%s", view)
	}

	send(t, m, runes("y"))
	if len(svc.dumpsDeleted) != 1 || svc.dumpsDeleted[0] != idea.ID {
		t.Fatalf("y should delete the dump entry, got %v", svc.dumpsDeleted)
	}
}

func TestConvertDumpToTask(t *testing.T) {
	idea := tasks.DumpEntry{ID: uuid.New(), Name: "learn sql"}
	svc := &fakeService{dumps: []tasks.DumpEntry{idea}}
	m := dumpTab(t, svc)

	m = send(t, m, runes("c"))
	if m.mode != modeAddTodo || m.form.fromDump == nil {
		t.Fatalf("c should open the new-task form for the dump entry")
	}
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "New task from dump") || !strings.Contains(view, "learn sql") {
		t.Fatalf("form should be titled and prefilled from the dump:\n%s", view)
	}

	// name, description, priority "high", daily, date
	m = send(t, m, keyEnter, keyEnter, runes("high"), keyEnter, keyEnter, keyEnter)
	if len(svc.promoted) != 1 || svc.promoted[0].Name != "learn sql" || svc.promoted[0].Priority != "high" {
		t.Fatalf("expected learn sql promoted with high priority, got %+v (form err %v)", svc.promoted, m.form.err)
	}
	if len(svc.dumpsDeleted) != 1 || svc.dumpsDeleted[0] != idea.ID || !strings.Contains(m.status, "learn sql") {
		t.Fatalf("the entry should leave the dump; deleted %v status %q", svc.dumpsDeleted, m.status)
	}
}

func TestConvertOnlyWorksOnDumps(t *testing.T) {
	svc := &fakeService{all: []tasks.TodoView{todo("taxes")}}
	m := send(t, start(t, svc, KeymapDefault), keyTab, runes("c"))
	if m.mode != modeNormal {
		t.Fatalf("c on a task should do nothing, mode %v", m.mode)
	}
}

func TestStatusBarShowsLastBackup(t *testing.T) {
	m := start(t, &fakeService{}, KeymapDefault)

	m.backup, m.backupOn = "backup today", true
	if bar := ansi.Strip(m.renderStatusBar(120)); !strings.Contains(bar, "backup today") {
		t.Errorf("status bar misses the last backup:\n%s", bar)
	}

	m.backup = ""
	if bar := ansi.Strip(m.renderStatusBar(120)); strings.Contains(bar, "backup") {
		t.Errorf("status bar shows a backup note it wasn't given:\n%s", bar)
	}
}

func TestLastBackup(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	now := time.Date(2026, 10, 19, 12, 0, 0, 0, time.Local)

	if label, on := lastBackup(now); on || label != i18n.T("tui.status.noBackup") {
		t.Errorf("without config: %q, %v", label, on)
	}

	dir := t.TempDir()
	if err := backup.SaveConfig(backup.Config{Dir: dir, Profile: "home"}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "akrasia-home-2026-10-18_090000.db"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	if label, on := lastBackup(now); !on || !strings.Contains(label, "2026-10-18") {
		t.Errorf("with a copy from yesterday: %q, %v", label, on)
	}

	if err := os.WriteFile(filepath.Join(dir, "akrasia-home-2026-10-19_090000.db"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	if label, _ := lastBackup(now); label != i18n.T("tui.status.backupToday") {
		t.Errorf("with a copy from today: %q", label)
	}
}
