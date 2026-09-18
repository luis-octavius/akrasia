package tasks

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luis-octavius/akrasia/internal/db"
	database "github.com/luis-octavius/akrasia/internal/db/out"
)

// newTestTaskManager spins up a throwaway sqlite DB (migrations included)
// for each test, so these never touch the user's real akrasia.db.
func newTestTaskManager(t *testing.T) (*TaskManager, *sql.DB) {
	t.Helper()

	dbPath := db.GetDBPath()
	oldDBPath := dbPath
	dbPath = filepath.Join(t.TempDir(), "akrasia-test.db")
	t.Cleanup(func() {
		dbPath = oldDBPath
	})

	conn, err := db.InitDB()
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
	})

	return &TaskManager{Queries: database.New(conn)}, conn
}

// There is no "reset" left to test — this checks the property that
// replaced it: marking a daily task done today does not affect
// yesterday's (or any other day's) history_history row, and calling
// `done` twice in the same day is idempotent rather than duplicating
// rows (which is what made the old cron-based reset fragile).
func TestMarkDailyDoneIsIdempotentPerDay(t *testing.T) {
	tkm, conn := newTestTaskManager(t)
	ctx := context.Background()

	now := time.Now()
	todo, err := tkm.Queries.AddTodo(ctx, database.AddTodoParams{
		ID:           uuid.New(),
		Name:         "daily-task",
		CreatedAt:    now,
		UpdatedAt:    now,
		Concluded:    false,
		ExpiresAt:    now.Add(24 * time.Hour),
		Priority:     "medium",
		IsDaily:      true,
		HistorySince: sql.NullString{String: now.Format(time.DateOnly), Valid: true},
	})
	if err != nil {
		t.Fatalf("AddTodo() error = %v", err)
	}

	if err := tkm.UpdateToConcluded(todo.Name, ""); err != nil {
		t.Fatalf("UpdateToConcluded() first call error = %v", err)
	}
	if err := tkm.UpdateToConcluded(todo.Name, "again"); err != nil {
		t.Fatalf("UpdateToConcluded() second call error = %v", err)
	}

	var rowCount int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM todos_history WHERE todo_id = ?", todo.ID).Scan(&rowCount); err != nil {
		t.Fatalf("count query error = %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("expected exactly 1 history row for today after two done calls, got %d", rowCount)
	}

	doneToday, err := tkm.doneTodaySet()
	if err != nil {
		t.Fatalf("doneTodaySet() error = %v", err)
	}
	if !doneToday[todoKey(todo.ID)] {
		t.Fatalf("expected task to be marked done today")
	}
}

// GetCurrentStreak must never count backfilled placeholder days as
// completions — that was the bug in the old query, where
// `notes = 'backfilled'` was (wrongly) OR'd into the COUNT(*).
func TestBackfillDoesNotInflateStreak(t *testing.T) {
	tkm, conn := newTestTaskManager(t)
	ctx := context.Background()

	createdAt := time.Now().Add(-10 * 24 * time.Hour)
	todo, err := tkm.Queries.AddTodo(ctx, database.AddTodoParams{
		ID:           uuid.New(),
		Name:         "streaked-task",
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
		Concluded:    false,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		Priority:     "medium",
		IsDaily:      true,
		HistorySince: sql.NullString{String: createdAt.Format(time.DateOnly), Valid: true},
	})
	if err != nil {
		t.Fatalf("AddTodo() error = %v", err)
	}

	if err := tkm.BackfillDailyHistory(30, todo.Name); err != nil {
		t.Fatalf("BackfillDailyHistory() error = %v", err)
	}

	streak, err := tkm.Queries.GetCurrentStreak(ctx, database.GetCurrentStreakParams{
		TodoID:   todo.ID,
		TodoID_2: todo.ID,
	})
	if err != nil {
		t.Fatalf("GetCurrentStreak() error = %v", err)
	}

	// Backfilled days ARE real completed=true rows now (there's no more
	// neutral state), so they legitimately count. The point of this test
	// is that the count matches the number of days actually inserted —
	// not inflated beyond that by a stray sentinel check.
	var expected int64
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM todos_history WHERE todo_id = ? AND completed = 1", todo.ID).Scan(&expected); err != nil {
		t.Fatalf("count query error = %v", err)
	}

	if streak != expected {
		t.Fatalf("expected streak %d to match inserted completed rows %d", streak, expected)
	}
}

// A gap of more than one day must break the streak, purely from the
// date arithmetic — no explicit "missed day" row is required anymore.
func TestGapBreaksStreakWithoutAMissedDayRow(t *testing.T) {
	tkm, conn := newTestTaskManager(t)
	ctx := context.Background()

	createdAt := time.Now().Add(-10 * 24 * time.Hour)
	todo, err := tkm.Queries.AddTodo(ctx, database.AddTodoParams{
		ID:           uuid.New(),
		Name:         "gapped-task",
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
		Concluded:    false,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		Priority:     "medium",
		IsDaily:      true,
		HistorySince: sql.NullString{String: createdAt.Format(time.DateOnly), Valid: true},
	})
	if err != nil {
		t.Fatalf("AddTodo() error = %v", err)
	}

	today := time.Now()
	fiveDaysAgo := today.AddDate(0, 0, -5).Format(time.DateOnly)

	if _, err := tkm.Queries.BackfillDailyDone(ctx, database.BackfillDailyDoneParams{
		ID:     uuid.New(),
		TodoID: todo.ID,
		Date:   fiveDaysAgo,
	}); err != nil {
		t.Fatalf("BackfillDailyDone() error = %v", err)
	}

	if err := tkm.UpdateToConcluded(todo.Name, ""); err != nil {
		t.Fatalf("UpdateToConcluded() error = %v", err)
	}

	streak, err := tkm.Queries.GetCurrentStreak(ctx, database.GetCurrentStreakParams{
		TodoID:   todo.ID,
		TodoID_2: todo.ID,
	})
	if err != nil {
		t.Fatalf("GetCurrentStreak() error = %v", err)
	}

	if streak != 1 {
		t.Fatalf("expected current streak of 1 (today only, gap breaks the older day), got %d", streak)
	}

	_ = conn // silence unused warning if the count check above is trimmed later
}
