package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/luis-octavius/akrasia/internal/tasks"
)

type fakeService struct {
	today  tasks.TodayView
	all    []tasks.TodoView
	daily  []tasks.TodoView
	dumps  []tasks.DumpEntry
	streak int64
	err    error

	calls     map[string]int
	completed []string // "name|notes"
	deleted   []string
	added     []tasks.NewTodo
	dumped    []string
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

func (f *fakeService) Complete(_ context.Context, name, notes string) (tasks.TodoView, error) {
	f.completed = append(f.completed, name+"|"+notes)
	return tasks.TodoView{Name: name}, nil
}

func (f *fakeService) DeleteByName(_ context.Context, name string) (tasks.TodoView, error) {
	f.deleted = append(f.deleted, name)
	return tasks.TodoView{Name: name}, nil
}

func (f *fakeService) AddDump(_ context.Context, name string) (tasks.DumpEntry, error) {
	f.dumped = append(f.dumped, name)
	return tasks.DumpEntry{Name: name}, nil
}

func (f *fakeService) CurrentStreak(_ context.Context, name string) (tasks.TodoView, int64, error) {
	f.called("streak")
	return tasks.TodoView{Name: name}, f.streak, nil
}

// send runs msg through Update and keeps feeding back the model's own
// messages (loads, writes, streaks), the way the runtime would. Other
// commands, like cursor blinks, are not run.
func send(t *testing.T, m Model, msgs ...tea.Msg) Model {
	t.Helper()
	for _, msg := range msgs {
		for i := 0; msg != nil && i < 5; i++ {
			next, cmd := m.Update(msg)
			m = next.(Model)
			msg = nil
			if cmd == nil {
				break
			}
			switch out := runQuick(cmd).(type) {
			case loadedMsg, actionMsg, streakMsg:
				msg = out
			}
		}
	}
	return m
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
	return send(t, m, m.Init()())
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
	if m.mode != modeNormal || !strings.Contains(m.status, "taxes") {
		t.Fatalf("expected normal mode and a status, got mode %v status %q", m.mode, m.status)
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
	cases := map[string][]int{"": nil, "5": {5}, "5/11": {5, 11}, " 7 / 12 ": {7, 12}}
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

	for _, bad := range []string{"x", "1/2/3", "5/"} {
		if _, err := dateParts(bad); err == nil {
			t.Errorf("dateParts(%q) should fail", bad)
		}
	}
}
