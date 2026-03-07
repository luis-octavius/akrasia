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

	_, err = queries.AddTodo(ctx, database.AddTodoParams{
		ID:        uuid.New(),
		Name:      "daily-test-task",
		CreatedAt: time.Now().Add(-48 * time.Hour),
		UpdatedAt: time.Now().Add(-48 * time.Hour),
		Concluded: true,
		ExpiresAt: time.Now().Add(-24 * time.Hour),
		Priority:  "medium",
		IsDaily:   true,
	})
	if err != nil {
		t.Fatalf("AddTodo() error = %v", err)
	}

	err = cfg.updateDailyTodo()
	if err != nil {
		t.Fatalf("updateDailyTodo() error = %v", err)
	}

	dailyTodos, err := queries.GetDailyTodos(ctx)
	if err != nil {
		t.Fatalf("GetDailyTodos() error = %v", err)
	}

	if len(dailyTodos) != 1 {
		t.Fatalf("expected 1 daily todo, got %d", len(dailyTodos))
	}

	updated := dailyTodos[0]

	if updated.Concluded {
		t.Fatalf("expected concluded to be false after daily update")
	}

	if !updated.ExpiresAt.After(time.Now()) {
		t.Fatalf("expected expires_at to be in the future, got %v", updated.ExpiresAt)
	}

	if !updated.UpdatedAt.After(time.Now().Add(-10 * time.Second)) {
		t.Fatalf("expected updated_at to be refreshed, got %v", updated.UpdatedAt)
	}
}
