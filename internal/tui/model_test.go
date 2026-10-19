package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/luis-octavius/akrasia/internal/tasks"
)

type fakeService struct {
	today tasks.TodayView
	all   []tasks.TodoView
	daily []tasks.TodoView
	dumps []tasks.DumpEntry
	err   error
	calls map[string]int
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

// send runs msg through Update and then any command it returns, the way
// the Bubble Tea runtime would for a single step.
func send(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	next, cmd := m.Update(msg)
	m = next.(Model)
	if cmd != nil {
		if out := cmd(); out != nil {
			if _, quit := out.(tea.QuitMsg); !quit {
				next, _ = m.Update(out)
				m = next.(Model)
			}
		}
	}
	return m
}

func start(t *testing.T, svc Service) Model {
	t.Helper()
	m := New(context.Background(), svc)
	return send(t, m, m.Init()())
}

var (
	keyDown = tea.KeyMsg{Type: tea.KeyDown}
	keyUp   = tea.KeyMsg{Type: tea.KeyUp}
	keyTab  = tea.KeyMsg{Type: tea.KeyTab}
)

func todo(name string) tasks.TodoView {
	return tasks.TodoView{Name: name, ExpiresAt: time.Now().Add(48 * time.Hour)}
}

func TestTodayTabLoadsSectionsAndSkipsHeaders(t *testing.T) {
	svc := &fakeService{today: tasks.TodayView{
		Overdue:      []tasks.TodoView{todo("late")},
		ExpiringSoon: []tasks.TodoView{todo("soon")},
	}}

	m := start(t, svc)
	st := m.tabs[tabToday]

	// header, late, header, soon
	if len(st.rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(st.rows))
	}
	if st.cursor != 1 {
		t.Fatalf("cursor should start on the first task, got %d", st.cursor)
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

	view := m.View()
	for _, want := range []string{"late", "soon"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q:\n%s", want, view)
		}
	}
}

func TestTabSwitchLoadsOnce(t *testing.T) {
	svc := &fakeService{all: []tasks.TodoView{todo("a"), todo("b")}}

	m := start(t, svc)
	m = send(t, m, keyTab)

	if m.active != tabAll {
		t.Fatalf("expected All tab, got %v", m.active)
	}
	if len(m.tabs[tabAll].rows) != 2 {
		t.Fatalf("expected All tab loaded with 2 rows, got %d", len(m.tabs[tabAll].rows))
	}

	// back to Today and forward again: All is already loaded
	m = send(t, m, tea.KeyMsg{Type: tea.KeyShiftTab})
	m = send(t, m, keyTab)
	if svc.calls["all"] != 1 {
		t.Fatalf("All tab should load once, loaded %d times", svc.calls["all"])
	}

	m = send(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if svc.calls["all"] != 2 {
		t.Fatalf("reload should fetch again, loaded %d times", svc.calls["all"])
	}
}

func TestLoadErrorIsShown(t *testing.T) {
	m := start(t, &fakeService{err: errors.New("database is locked")})

	if !strings.Contains(m.View(), "database is locked") {
		t.Fatalf("expected error in view:\n%s", m.View())
	}
}

func TestQuit(t *testing.T) {
	m := start(t, &fakeService{})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("q should return a command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("q should quit")
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
