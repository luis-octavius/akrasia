package tasks

import (
	"context"
	"testing"
	"time"
)

func TestTimedDue(t *testing.T) {
	// Thursday 2026-10-22, 15:00
	now := time.Date(2026, time.October, 22, 15, 0, 0, 0, time.UTC)
	today := time.Date(2026, time.October, 22, 0, 0, 0, 0, time.UTC)

	weekly := func(name string, day int, clock string) TodoView {
		return TodoView{Name: name, IsDaily: true, Recurrence: RecurWeekly, RepeatOn: day, DueTime: clock}
	}
	todos := []TodoView{
		{Name: "due-today", ExpiresAt: today, DueTime: "14:30"},
		{Name: "later-today", ExpiresAt: today, DueTime: "18:00"},
		{Name: "no-time", ExpiresAt: today},
		{Name: "tomorrow", ExpiresAt: today.AddDate(0, 0, 1), DueTime: "09:00"},
		{Name: "yesterday", ExpiresAt: today.AddDate(0, 0, -1), DueTime: "09:00"},
		{Name: "concluded", ExpiresAt: today, DueTime: "09:00", Concluded: true},
		{Name: "daily", IsDaily: true, Recurrence: RecurDaily, DueTime: "07:00"},
		{Name: "daily-done", IsDaily: true, Recurrence: RecurDaily, DueTime: "07:00", DonePeriod: true},
		weekly("weekly-thursday", 4, "08:00"),
		// due since Monday, but reminded on its day only
		weekly("weekly-monday", 1, "08:00"),
		{Name: "monthly-22", IsDaily: true, Recurrence: RecurMonthly, RepeatOn: 22, DueTime: "06:00"},
		{
			Name: "parent", ExpiresAt: today, DueTime: "05:00",
			Subtasks: []TodoView{{Name: "child", ExpiresAt: today, DueTime: "13:00"}},
		},
	}

	var got []string
	for _, todo := range timedDue(todos, now) {
		got = append(got, todo.Name)
	}
	want := []string{"parent", "monthly-22", "daily", "weekly-thursday", "child", "due-today"}
	if len(got) != len(want) {
		t.Fatalf("timedDue = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("timedDue = %v, want %v", got, want)
		}
	}
}

func TestIsOnDay(t *testing.T) {
	thu := time.Date(2026, time.October, 22, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		todo TodoView
		want bool
	}{
		{TodoView{IsDaily: true, Recurrence: RecurDaily}, true},
		{TodoView{IsDaily: true, Recurrence: RecurWeekly, RepeatOn: 4}, true},
		{TodoView{IsDaily: true, Recurrence: RecurWeekly, RepeatOn: 1}, false},
		{TodoView{IsDaily: true, Recurrence: RecurMonthly, RepeatOn: 22}, true},
		{TodoView{IsDaily: true, Recurrence: RecurMonthly, RepeatOn: 5}, false},
		{TodoView{IsDaily: true, Recurrence: RecurMonthly, RepeatOn: 31}, false},
	}
	for _, c := range cases {
		if got := IsOnDay(c.todo, thu); got != c.want {
			t.Errorf("IsOnDay(%+v) = %v, want %v", c.todo, got, c.want)
		}
	}
	// 31 on the last day of a 28-day month
	feb := time.Date(2027, time.February, 28, 0, 0, 0, 0, time.UTC)
	if !IsOnDay(TodoView{IsDaily: true, Recurrence: RecurMonthly, RepeatOn: 31}, feb) {
		t.Error("a monthly task on the 31st falls on February 28th")
	}
}

func TestTimedDueFromDatabase(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	now := time.Now()

	// 00:00 has always passed; 23:59 may not have
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "early", ExpiresAt: dayOf(now), DueTime: "00:00"}); err != nil {
		t.Fatal(err)
	}
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "other day", ExpiresAt: dayOf(now).AddDate(0, 0, 3), DueTime: "00:00"}); err != nil {
		t.Fatal(err)
	}
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "habit", IsDaily: true, DueTime: "00:00"}); err != nil {
		t.Fatal(err)
	}

	due, err := tkm.TimedDue(ctx, now)
	if err != nil || len(due) != 2 {
		t.Fatalf("TimedDue = %+v, %v; want early and habit", due, err)
	}

	if _, err := tkm.Complete(ctx, "habit", ""); err != nil {
		t.Fatal(err)
	}
	if due, _ = tkm.TimedDue(ctx, now); len(due) != 1 || due[0].Name != "early" {
		t.Fatalf("a done task is not reminded, got %+v", due)
	}
}
