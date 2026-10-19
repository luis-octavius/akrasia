package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAddTodoReturnsCreatedTask(t *testing.T) {
	tkm, _ := newTestTaskManager(t)

	todo, err := tkm.AddTodo(context.Background(), NewTodo{
		Name:        "write",
		Description: "chapter 1",
		Priority:    "high",
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("AddTodo() error = %v", err)
	}

	if todo.ID == uuid.Nil || todo.Name != "write" || todo.Description.String != "chapter 1" || todo.Done() {
		t.Fatalf("unexpected created task: %+v", todo)
	}
}

func TestTodayAndFocusSkipDoneTasks(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	addTestTodo(t, tkm, "read", "medium", true)
	addTestTodo(t, tkm, "run", "high", true)
	addTestTodo(t, tkm, "taxes", "high", false)

	if _, err := tkm.Complete(ctx, "read", ""); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	view, err := tkm.Today(ctx, TodayOptions{})
	if err != nil {
		t.Fatalf("Today() error = %v", err)
	}
	if len(view.Daily) != 1 || view.Daily[0].Name != "run" {
		t.Fatalf("expected only run pending in daily, got %+v", view.Daily)
	}
	if len(view.ExpiringSoon) != 1 || view.ExpiringSoon[0].Name != "taxes" {
		t.Fatalf("expected taxes expiring soon, got %+v", view.ExpiringSoon)
	}

	focus, err := tkm.Focus(ctx, 1, "")
	if err != nil {
		t.Fatalf("Focus() error = %v", err)
	}
	if len(focus) != 1 || focus[0].Name != "run" {
		t.Fatalf("expected run as the single focus item, got %+v", focus)
	}
}

func TestExpiringExcludesConcluded(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	addTestTodo(t, tkm, "taxes", "high", false)
	addTestTodo(t, tkm, "groceries", "low", false)

	if _, err := tkm.Complete(ctx, "taxes", ""); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	todos, err := tkm.Expiring(ctx)
	if err != nil {
		t.Fatalf("Expiring() error = %v", err)
	}
	if len(todos) != 1 || todos[0].Name != "groceries" {
		t.Fatalf("expected only groceries, got %+v", todos)
	}
}

func TestDeleteByNameReturnsMatchedTask(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	addTestTodo(t, tkm, "groceries", "low", false)

	deleted, err := tkm.DeleteByName(ctx, "groc")
	if err != nil {
		t.Fatalf("DeleteByName() error = %v", err)
	}
	if deleted.Name != "groceries" {
		t.Fatalf("expected fuzzy match to delete groceries, got %q", deleted.Name)
	}

	if _, err := tkm.FindTodo(ctx, "groceries"); !errors.Is(err, ErrTodoNotFound) {
		t.Fatalf("expected task to be gone, got %v", err)
	}
}

func TestDumpsRoundTrip(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	if _, err := tkm.AddDump(ctx, "idea"); err != nil {
		t.Fatalf("AddDump() error = %v", err)
	}

	entries, err := tkm.Dumps(ctx)
	if err != nil {
		t.Fatalf("Dumps() error = %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "idea" || entries[0].ID == uuid.Nil {
		t.Fatalf("unexpected dumps: %+v", entries)
	}
}

func TestBackfillReportCountsDays(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	addTestTodo(t, tkm, "read", "medium", true)

	report, err := tkm.BackfillDailyHistory(ctx, 3, "read")
	if err != nil {
		t.Fatalf("BackfillDailyHistory() error = %v", err)
	}
	if len(report.Tasks) != 1 || report.Tasks[0].Inserted != 3 || report.Total() != 3 {
		t.Fatalf("expected 3 backfilled days, got %+v", report)
	}

	debug, err := tkm.Debug(ctx, "read")
	if err != nil {
		t.Fatalf("Debug() error = %v", err)
	}
	if len(debug.History) != 3 || debug.Streak != 3 {
		t.Fatalf("expected 3 history rows and streak 3, got %+v", debug)
	}
}
