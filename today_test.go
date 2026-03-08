package main

import (
	"testing"
	"time"

	"github.com/luis-octavius/akrasia/internal/database"
)

func TestBuildTodayView(t *testing.T) {
	now := time.Date(2026, time.March, 8, 10, 0, 0, 0, time.UTC)

	todos := []database.Todo{
		{Name: "overdue-medium", Priority: "medium", ExpiresAt: now.Add(-24 * time.Hour)},
		{Name: "due-today-high", Priority: "high", ExpiresAt: now.Add(2 * time.Hour)},
		{Name: "due-today-low", Priority: "low", ExpiresAt: now.Add(3 * time.Hour)},
		{Name: "expiring-soon", Priority: "medium", ExpiresAt: now.Add(48 * time.Hour)},
		{Name: "daily", IsDaily: true, Priority: "high", ExpiresAt: now.Add(24 * time.Hour)},
		{Name: "done-ignored", Concluded: true, ExpiresAt: now.Add(24 * time.Hour)},
		{Name: "far-future", ExpiresAt: now.Add(10 * 24 * time.Hour)},
	}

	view := buildTodayView(todos, now)

	if len(view.Overdue) != 1 || view.Overdue[0].Name != "overdue-medium" {
		t.Fatalf("unexpected overdue view: %#v", view.Overdue)
	}

	if len(view.DueToday) != 2 {
		t.Fatalf("expected 2 due-today tasks, got %d", len(view.DueToday))
	}

	if view.DueToday[0].Name != "due-today-high" {
		t.Fatalf("expected high priority first in due-today, got %s", view.DueToday[0].Name)
	}

	if len(view.Daily) != 1 || view.Daily[0].Name != "daily" {
		t.Fatalf("unexpected daily view: %#v", view.Daily)
	}

	if len(view.ExpiringSoon) != 1 || view.ExpiringSoon[0].Name != "expiring-soon" {
		t.Fatalf("unexpected expiring-soon view: %#v", view.ExpiringSoon)
	}
}
