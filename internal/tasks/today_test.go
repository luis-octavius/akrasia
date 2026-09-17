package tasks

import (
	"testing"
	"time"

	database "github.com/luis-octavius/akrasia/internal/db/out"
)

func TestBuildTodayView(t *testing.T) {
	now := time.Date(2026, time.March, 8, 10, 0, 0, 0, time.UTC)

	todos := []database.Todo{
		{ID: "overdue-medium-id", Name: "overdue-medium", Priority: "medium", ExpiresAt: now.Add(-24 * time.Hour)},
		{ID: "due-today-high-id", Name: "due-today-high", Priority: "high", ExpiresAt: now.Add(2 * time.Hour)},
		{ID: "due-today-low-id", Name: "due-today-low", Priority: "low", ExpiresAt: now.Add(3 * time.Hour)},
		{ID: "expiring-soon-id", Name: "expiring-soon", Priority: "medium", ExpiresAt: now.Add(48 * time.Hour)},
		{ID: "daily-pending-id", Name: "daily-pending", IsDaily: true, Priority: "high", ExpiresAt: now.Add(24 * time.Hour)},
		{ID: "daily-done-id", Name: "daily-done", IsDaily: true, Priority: "high", ExpiresAt: now.Add(24 * time.Hour)},
		{ID: "done-ignored-id", Name: "done-ignored", Concluded: true, ExpiresAt: now.Add(24 * time.Hour)},
		{ID: "far-future-id", Name: "far-future", ExpiresAt: now.Add(10 * 24 * time.Hour)},
	}

	doneToday := map[string]bool{"daily-done-id": true}

	view := buildTodayView(todos, now, doneToday)

	if len(view.Overdue) != 1 || view.Overdue[0].Name != "overdue-medium" {
		t.Fatalf("unexpected overdue view: %#v", view.Overdue)
	}

	if len(view.DueToday) != 2 {
		t.Fatalf("expected 2 due-today tasks, got %d", len(view.DueToday))
	}

	if view.DueToday[0].Name != "due-today-high" {
		t.Fatalf("expected high priority first in due-today, got %s", view.DueToday[0].Name)
	}

	// Only the daily task NOT in doneToday should show up as pending —
	// a daily task marked done today must disappear even though
	// todo.Concluded was never set (see buildTodayView doc comment).
	if len(view.Daily) != 1 || view.Daily[0].Name != "daily-pending" {
		t.Fatalf("unexpected daily view: %#v", view.Daily)
	}

	if len(view.ExpiringSoon) != 1 || view.ExpiringSoon[0].Name != "expiring-soon" {
		t.Fatalf("unexpected expiring-soon view: %#v", view.ExpiringSoon)
	}
}

func TestApplyTodayOptionsOnly(t *testing.T) {
	view := TodayView{
		Overdue:      []database.Todo{{Name: "o1"}},
		DueToday:     []database.Todo{{Name: "t1"}},
		Daily:        []database.Todo{{Name: "d1"}},
		ExpiringSoon: []database.Todo{{Name: "s1"}},
	}

	filtered := applyTodayOptions(view, TodayOptions{Only: "daily"})

	if len(filtered.Daily) != 1 {
		t.Fatalf("expected daily section to remain")
	}

	if len(filtered.Overdue) != 0 || len(filtered.DueToday) != 0 || len(filtered.ExpiringSoon) != 0 {
		t.Fatalf("expected non-daily sections to be empty")
	}
}

func TestApplyTodayOptionsLimit(t *testing.T) {
	view := TodayView{
		DueToday: []database.Todo{
			{Name: "a"},
			{Name: "b"},
			{Name: "c"},
		},
	}

	limited := applyTodayOptions(view, TodayOptions{Limit: 2})

	if len(limited.DueToday) != 2 {
		t.Fatalf("expected 2 due-today tasks after limit, got %d", len(limited.DueToday))
	}
}
