package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/luis-octavius/akrasia/internal/tasks"
)

func TestWrapTextHasNoLeadingSpaces(t *testing.T) {
	text := strings.Repeat("word ", 30)

	for _, line := range strings.Split(wrapText(text, 44), "\n") {
		if strings.HasPrefix(line, " ") {
			t.Fatalf("line starts with a space: %q", line)
		}
		if len(line) > 40 {
			t.Fatalf("line longer than 40 chars: %q", line)
		}
	}
}

func TestRenderTodoShowsTheTime(t *testing.T) {
	due := time.Date(2026, time.December, 10, 0, 0, 0, 0, time.UTC)

	timed := ansi.Strip(renderTodo(tasks.TodoView{Name: "call", ExpiresAt: due, DueTime: "14:30"}))
	if !strings.Contains(timed, "10 Dec 26 14:30") {
		t.Errorf("a timed task should show its date and time:\n%s", timed)
	}

	untimed := ansi.Strip(renderTodo(tasks.TodoView{Name: "call", ExpiresAt: due}))
	if !strings.Contains(untimed, "10 Dec 26") || strings.Contains(untimed, "00:00") {
		t.Errorf("an untimed task shows only its date:\n%s", untimed)
	}

	daily := ansi.Strip(renderTodo(tasks.TodoView{Name: "run", IsDaily: true, DueTime: "07:00"}))
	if !strings.Contains(daily, "07:00") {
		t.Errorf("a timed daily task should show its time:\n%s", daily)
	}

	sub := ansi.Strip(renderSubtask(tasks.TodoView{Name: "bake", ExpiresAt: due, DueTime: "16:00"}))
	if !strings.Contains(sub, "10 Dec 16:00") {
		t.Errorf("a timed subtask should show its time:\n%s", sub)
	}
}

func TestRenderRecurringTaskShowsItsSchedule(t *testing.T) {
	weekly := ansi.Strip(renderTodo(tasks.TodoView{Name: "review", IsDaily: true, Recurrence: tasks.RecurWeekly, RepeatOn: 5, DueTime: "18:00"}))
	if !strings.Contains(weekly, "weekly on Fri") || !strings.Contains(weekly, "at 18:00") {
		t.Errorf("a weekly task should show its day and time:\n%s", weekly)
	}

	monthly := ansi.Strip(renderTodo(tasks.TodoView{Name: "rent", IsDaily: true, Recurrence: tasks.RecurMonthly, RepeatOn: 5}))
	if !strings.Contains(monthly, "monthly on day 5") {
		t.Errorf("a monthly task should show its day:\n%s", monthly)
	}

	daily := ansi.Strip(renderTodo(tasks.TodoView{Name: "run", IsDaily: true, Recurrence: tasks.RecurDaily}))
	if strings.Contains(daily, "weekly") || strings.Contains(daily, "monthly") {
		t.Errorf("a plain daily task shows no schedule:\n%s", daily)
	}
}

func TestRepeatFromFlags(t *testing.T) {
	friday := time.Date(2026, time.October, 23, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name                   string
		daily, weekly, monthly bool
		on                     string
		dateGiven              bool
		wantRec                string
		wantOn                 int
		wantErr                bool
	}{
		{name: "nothing"},
		{name: "daily", daily: true, wantRec: tasks.RecurDaily},
		{name: "weekly, no day", weekly: true, wantRec: tasks.RecurWeekly},
		{name: "weekly on", weekly: true, on: "fri", wantRec: tasks.RecurWeekly, wantOn: 5},
		{name: "weekly on in Portuguese", weekly: true, on: "seg", wantRec: tasks.RecurWeekly, wantOn: 1},
		{name: "weekly from the date", weekly: true, dateGiven: true, wantRec: tasks.RecurWeekly, wantOn: 5},
		{name: "monthly from the date", monthly: true, dateGiven: true, wantRec: tasks.RecurMonthly, wantOn: 23},
		{name: "monthly on", monthly: true, on: "31", wantRec: tasks.RecurMonthly, wantOn: 31},
		{name: "on wins over the date", weekly: true, on: "mon", dateGiven: true, wantRec: tasks.RecurWeekly, wantOn: 1},
		{name: "weekdays don't fit a monthly task", monthly: true, on: "fri", wantErr: true},
		{name: "day 15 doesn't fit a weekly task", weekly: true, on: "15", wantErr: true},
		{name: "on without a kind", on: "fri", wantErr: true},
		{name: "on with daily", daily: true, on: "fri", wantErr: true},
		{name: "daily and weekly", daily: true, weekly: true, wantErr: true},
		{name: "weekly and monthly", weekly: true, monthly: true, wantErr: true},
	}
	for _, c := range cases {
		rec, on, err := repeatFromFlags(c.daily, c.weekly, c.monthly, c.on, c.dateGiven, friday)
		if (err != nil) != c.wantErr || rec != c.wantRec || on != c.wantOn {
			t.Errorf("%s: got %q, %d, %v; want %q, %d, error %v", c.name, rec, on, err, c.wantRec, c.wantOn, c.wantErr)
		}
	}
}
