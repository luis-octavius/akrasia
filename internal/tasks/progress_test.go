package tasks

import (
	"context"
	"testing"
	"time"
)

func TestTodayProgress(t *testing.T) {
	now := time.Date(2026, time.October, 19, 10, 0, 0, 0, time.Local)

	todos := []TodoView{
		{Name: "daily-done", IsDaily: true, DoneToday: true},
		{Name: "daily-pending", IsDaily: true},
		{Name: "overdue", ExpiresAt: now.AddDate(0, 0, -2)},
		{Name: "due-today", ExpiresAt: now.Add(5 * time.Hour)},
		{Name: "done-today", Concluded: true, DoneToday: true, ExpiresAt: now.AddDate(0, 0, 3)},
		{Name: "done-yesterday", Concluded: true, ExpiresAt: now.AddDate(0, 0, -1)},
		{Name: "due-tomorrow", ExpiresAt: now.AddDate(0, 0, 1)},
	}

	got := todayProgress(todos, now)
	want := Progress{Done: 2, Total: 5}
	if got != want {
		t.Fatalf("todayProgress = %+v, want %+v", got, want)
	}
	if r := got.Ratio(); r != 0.4 {
		t.Fatalf("Ratio = %v, want 0.4", r)
	}
}

func TestProgressRatioWithNothingDue(t *testing.T) {
	if r := (Progress{}).Ratio(); r != 1 {
		t.Fatalf("an empty day should count as done, got %v", r)
	}
}

func TestTodayProgressFromDatabase(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	addTestTodo(t, tkm, "read", "medium", true)
	addTestTodo(t, tkm, "run", "high", true)

	if _, err := tkm.Complete(ctx, "read", ""); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	got, err := tkm.TodayProgress(ctx)
	if err != nil {
		t.Fatalf("TodayProgress() error = %v", err)
	}
	if got != (Progress{Done: 1, Total: 2}) {
		t.Fatalf("TodayProgress = %+v, want 1/2", got)
	}
}
