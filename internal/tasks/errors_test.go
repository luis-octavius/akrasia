package tasks

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	database "github.com/luis-octavius/akrasia/internal/db/out"
)

func TestGetTodoByNameNotFound(t *testing.T) {
	tkm, _ := newTestTaskManager(t)

	_, err := findTodoByName(context.Background(), tkm, "does-not-exist")
	if !errors.Is(err, ErrTodoNotFound) {
		t.Fatalf("expected ErrTodoNotFound, got %v", err)
	}
}

func TestBackfillNonDailyTask(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	now := time.Now()

	_, err := tkm.Queries.AddTodo(context.Background(), database.AddTodoParams{
		ID:           uuid.New(),
		Name:         "one-off",
		CreatedAt:    now,
		UpdatedAt:    now,
		ExpiresAt:    now.Add(24 * time.Hour),
		Priority:     "low",
		HistorySince: sql.NullString{String: now.Format(time.DateOnly), Valid: true},
	})
	if err != nil {
		t.Fatalf("AddTodo() error = %v", err)
	}

	_, err = tkm.BackfillDailyHistory(context.Background(), 3, "one-off")
	if !errors.Is(err, ErrNotDaily) {
		t.Fatalf("expected ErrNotDaily, got %v", err)
	}
}
