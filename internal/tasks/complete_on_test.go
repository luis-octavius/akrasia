package tasks

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	database "github.com/luis-octavius/akrasia/internal/db/out"
)

func daysAgo(n int) time.Time { return dayOf(time.Now()).AddDate(0, 0, -n) }

// extendTo moves a task's streak window back to day, as backfill does.
func extendTo(day time.Time) database.ExtendHistorySinceParams {
	return database.ExtendHistorySinceParams{
		Earliest: sql.NullString{String: day.Format(time.DateOnly), Valid: true},
	}
}

func TestCompleteOnRecordsThatDay(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addTestTodo(t, tkm, "read", "medium", true)

	if _, err := tkm.CompleteOn(ctx, "read", "late night", false, daysAgo(1)); err != nil {
		t.Fatalf("CompleteOn() error = %v", err)
	}
	if got := find(t, tkm, "read"); got.DoneToday {
		t.Fatal("completing yesterday must not make today done")
	}

	report, err := tkm.Debug(ctx, "read")
	if err != nil || len(report.History) != 1 || report.History[0].Date != daysAgo(1).Format(time.DateOnly) || report.History[0].Notes != "late night" {
		t.Fatalf("history = %+v, %v", report.History, err)
	}

	// yesterday and today make a streak of two; the window starts before them
	if _, err := tkm.Complete(ctx, "read", ""); err != nil {
		t.Fatal(err)
	}
	arg := extendTo(daysAgo(1))
	arg.ID = find(t, tkm, "read").ID
	if err := tkm.Queries.ExtendHistorySince(ctx, arg); err != nil {
		t.Fatal(err)
	}
	if _, streak, err := tkm.CurrentStreak(ctx, "read"); err != nil || streak != 2 {
		t.Fatalf("streak = %d, %v; want 2", streak, err)
	}
}

func TestCompleteOnToday(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addTestTodo(t, tkm, "read", "medium", true)

	if _, err := tkm.CompleteOn(ctx, "read", "", false, time.Now()); err != nil {
		t.Fatal(err)
	}
	if !find(t, tkm, "read").Done() {
		t.Fatal("completing on today's date is completing today")
	}
}

func TestCompleteOnRejectsDaysOutOfRange(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addTestTodo(t, tkm, "read", "medium", true)

	for name, day := range map[string]time.Time{
		"tomorrow": daysAgo(-1),
		"too old":  daysAgo(8),
	} {
		if _, err := tkm.CompleteOn(ctx, "read", "", false, day); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%s: error = %v, want invalid input", name, err)
		}
	}
	if _, err := tkm.CompleteOn(ctx, "read", "", false, daysAgo(7)); err != nil {
		t.Errorf("7 days back is the limit and allowed, got %v", err)
	}
}

func TestCompleteOnLastWeekDoesNotFinishThisWeek(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "review", Recurrence: RecurWeekly, RepeatOn: 1}); err != nil {
		t.Fatal(err)
	}

	lastSunday := startOfWeek(time.Now()).AddDate(0, 0, -1)
	if _, err := tkm.CompleteOn(ctx, "review", "", false, lastSunday); err != nil {
		t.Fatal(err)
	}
	if find(t, tkm, "review").Done() {
		t.Fatal("a completion made last week must not finish this week")
	}
}

func TestCompleteOnPastDayLeavesARecurringParentAlone(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "routine", IsDaily: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "stretch", Parent: "routine"}); err != nil {
		t.Fatal(err)
	}

	if _, err := tkm.CompleteOn(ctx, "routine/stretch", "", false, daysAgo(1)); err != nil {
		t.Fatal(err)
	}
	if find(t, tkm, "routine").Done() {
		t.Fatal("yesterday's subtask must not complete today's routine")
	}
}

func TestQueueIDsAreRemembered(t *testing.T) {
	tkm, conn := newTestTaskManager(t)
	ctx := context.Background()
	const id = "11111111-1111-1111-1111-111111111111"

	if seen, err := tkm.QueueSeen(ctx, id); err != nil || seen {
		t.Fatalf("QueueSeen = %v, %v; want false", seen, err)
	}
	for range 2 {
		if err := tkm.QueueRemember(ctx, id); err != nil {
			t.Fatalf("QueueRemember() error = %v (remembering twice is fine)", err)
		}
	}
	if seen, _ := tkm.QueueSeen(ctx, id); !seen {
		t.Fatal("the id should be remembered")
	}

	// old ids are forgotten
	if _, err := conn.ExecContext(ctx, `UPDATE web_queue SET created_at = datetime('now', 'localtime', '-40 days')`); err != nil {
		t.Fatal(err)
	}
	if err := tkm.QueueRemember(ctx, "22222222-2222-2222-2222-222222222222"); err != nil {
		t.Fatal(err)
	}
	if seen, _ := tkm.QueueSeen(ctx, id); seen {
		t.Fatal("an id older than 30 days should be pruned")
	}
}
