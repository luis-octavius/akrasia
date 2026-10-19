package tasks

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	database "github.com/luis-octavius/akrasia/internal/db/out"
)

func addTestTodo(t *testing.T, tkm *TaskManager, name, priority string, isDaily bool) database.Todo {
	t.Helper()
	now := time.Now()

	todo, err := tkm.Queries.AddTodo(context.Background(), database.AddTodoParams{
		ID:           uuid.New(),
		Name:         name,
		CreatedAt:    now,
		UpdatedAt:    now,
		ExpiresAt:    now.Add(48 * time.Hour),
		Priority:     priority,
		IsDaily:      isDaily,
		HistorySince: sql.NullString{String: now.Format(time.DateOnly), Valid: true},
	})
	if err != nil {
		t.Fatalf("AddTodo(%q) error = %v", name, err)
	}

	return todo
}

func TestDailyTodosReflectsDoneToday(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	addTestTodo(t, tkm, "read", "medium", true)
	addTestTodo(t, tkm, "run", "medium", true)
	addTestTodo(t, tkm, "taxes", "high", false)

	if _, err := tkm.Complete(context.Background(), "read", ""); err != nil {
		t.Fatalf("UpdateToConcluded() error = %v", err)
	}

	daily, err := tkm.DailyTodos(ctx)
	if err != nil {
		t.Fatalf("DailyTodos() error = %v", err)
	}
	if len(daily) != 2 {
		t.Fatalf("expected 2 daily tasks, got %d", len(daily))
	}

	for _, v := range daily {
		if v.ID == uuid.Nil {
			t.Errorf("%q has no ID", v.Name)
		}
		if want := v.Name == "read"; v.Done() != want {
			t.Errorf("%q Done() = %v, want %v", v.Name, v.Done(), want)
		}
	}
}

func TestListTodosFiltersAndKeepsConcluded(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	addTestTodo(t, tkm, "taxes", "high", false)
	addTestTodo(t, tkm, "groceries", "low", false)

	if _, err := tkm.Complete(context.Background(), "taxes", ""); err != nil {
		t.Fatalf("UpdateToConcluded() error = %v", err)
	}

	high, err := tkm.ListTodos(ctx, "high")
	if err != nil {
		t.Fatalf("ListTodos() error = %v", err)
	}
	if len(high) != 1 || high[0].Name != "taxes" {
		t.Fatalf("expected only taxes, got %+v", high)
	}
	if !high[0].Done() {
		t.Errorf("concluded non-daily task should be done")
	}

	all, err := tkm.ListTodos(ctx, "")
	if err != nil {
		t.Fatalf("ListTodos() error = %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(all))
	}
}
