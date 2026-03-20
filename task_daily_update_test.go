package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luis-octavius/akrasia/internal/database"
)

func TestUpdateDailyTodoResetsAndPushesExpiration(t *testing.T) {
	oldDBPath := dbPath
	dbPath = filepath.Join(t.TempDir(), "akrasia-test.db")
	t.Cleanup(func() {
		dbPath = oldDBPath
	})

	db, err := initDB()
	if err != nil {
		t.Fatalf("initDB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	queries := database.New(db)
	cfg := Config{Queries: queries}
	ctx := context.Background()

	// Create a completed daily task
	completedTask, err := queries.AddTodo(ctx, database.AddTodoParams{
		ID:        uuid.New(),
		Name:      "daily-completed",
		CreatedAt: time.Now().Add(-48 * time.Hour),
		UpdatedAt: time.Now().Add(-48 * time.Hour),
		Concluded: true, // This task was completed
		ExpiresAt: time.Now().Add(-24 * time.Hour),
		Priority:  "medium",
		IsDaily:   true,
	})
	if err != nil {
		t.Fatalf("AddTodo() error = %v", err)
	}

	// Create an incomplete daily task
	incompleteTask, err := queries.AddTodo(ctx, database.AddTodoParams{
		ID:        uuid.New(),
		Name:      "daily-incomplete",
		CreatedAt: time.Now().Add(-48 * time.Hour),
		UpdatedAt: time.Now().Add(-48 * time.Hour),
		Concluded: false, // This task was NOT completed
		ExpiresAt: time.Now().Add(-24 * time.Hour),
		Priority:  "medium",
		IsDaily:   true,
	})
	if err != nil {
		t.Fatalf("AddTodo() error = %v", err)
	}

	// Run daily update
	err = cfg.updateDailyTodo()
	if err != nil {
		t.Fatalf("updateDailyTodo() error = %v", err)
	}

	// Verify tasks were reset
	dailyTodos, err := queries.GetDailyTodos(ctx)
	if err != nil {
		t.Fatalf("GetDailyTodos() error = %v", err)
	}

	if len(dailyTodos) != 2 {
		t.Fatalf("expected 2 daily todos, got %d", len(dailyTodos))
	}

	var expectedNextDay string
	err = db.QueryRowContext(ctx, "SELECT date('now', 'localtime', '+1 day')").Scan(&expectedNextDay)
	if err != nil {
		t.Fatalf("failed to calculate expected next day: %v", err)
	}

	for _, todo := range dailyTodos {
		if todo.Concluded {
			t.Errorf("task %s: expected concluded=false after daily update", todo.Name)
		}

		if todo.ExpiresAt.Format(time.DateOnly) != expectedNextDay {
			t.Errorf("task %s: expected expires_at date %s, got %v", todo.Name, expectedNextDay, todo.ExpiresAt)
		}
	}

	// Verify history snapshot rows were recorded.
	rows, err := db.QueryContext(ctx, "SELECT todo_id, completed FROM todos_history")
	if err != nil {
		t.Fatalf("Query history error = %v", err)
	}
	defer rows.Close()

	historyByTodo := map[string]bool{}
	for rows.Next() {
		var todoID string
		var completed bool
		if err := rows.Scan(&todoID, &completed); err != nil {
			t.Fatalf("Scan error = %v", err)
		}
		historyByTodo[todoID] = completed
	}

	if len(historyByTodo) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(historyByTodo))
	}

	if done, ok := historyByTodo[completedTask.ID.(string)]; !ok || !done {
		t.Fatalf("expected completed task history row with completed=true")
	}

	if done, ok := historyByTodo[incompleteTask.ID.(string)]; !ok || done {
		t.Fatalf("expected incomplete task history row with completed=false")
	}
}
