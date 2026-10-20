package tasks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/luis-octavius/akrasia/pkg/i18n"
)

func TestNotesNewestFirstWithBackfilledFlag(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	addTestTodo(t, tkm, "rust", "high", true)
	if _, err := tkm.BackfillDailyHistory(ctx, 2, "rust"); err != nil {
		t.Fatalf("BackfillDailyHistory() error = %v", err)
	}
	if _, err := tkm.Complete(ctx, "rust", "functions and return types"); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	todo, notes, err := tkm.Notes(ctx, "rust", 0)
	if err != nil {
		t.Fatalf("Notes() error = %v", err)
	}
	if todo.Name != "rust" || len(notes) != 3 {
		t.Fatalf("expected 3 days for rust, got %d for %q", len(notes), todo.Name)
	}

	if notes[0].Notes != "functions and return types" || notes[0].Backfilled {
		t.Errorf("today should come first with its note, got %+v", notes[0])
	}
	for _, n := range notes[1:] {
		if !n.Backfilled || n.Notes != "" {
			t.Errorf("backfilled day should be flagged with no text, got %+v", n)
		}
	}
	if notes[1].Date <= notes[2].Date {
		t.Errorf("days should be newest first, got %s then %s", notes[1].Date, notes[2].Date)
	}
}

func TestNotesLimitAndEmpty(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	addTestTodo(t, tkm, "rust", "high", true)
	addTestTodo(t, tkm, "fresh", "low", true)
	if _, err := tkm.BackfillDailyHistory(ctx, 5, "rust"); err != nil {
		t.Fatalf("BackfillDailyHistory() error = %v", err)
	}

	if _, notes, err := tkm.Notes(ctx, "rust", 2); err != nil || len(notes) != 2 {
		t.Fatalf("limit 2 should return 2 days, got %d (%v)", len(notes), err)
	}
	if _, notes, err := tkm.Notes(ctx, "fresh", 0); err != nil || len(notes) != 0 {
		t.Fatalf("a task never done has no notes, got %v (%v)", notes, err)
	}
	if _, _, err := tkm.Notes(ctx, "missing", 0); !errors.Is(err, ErrTodoNotFound) {
		t.Fatalf("unknown task should be ErrTodoNotFound, got %v", err)
	}
}

// quoteCount must match the quotes in the locales: every one of them is
// reachable and there is none past the end.
func TestQuoteCountMatchesLocales(t *testing.T) {
	for i := 1; i <= quoteCount; i++ {
		key := fmt.Sprintf("quotes.quote%d", i)
		if i18n.T(key) == key {
			t.Errorf("%s is missing from the locales", key)
		}
	}
	if next := fmt.Sprintf("quotes.quote%d", quoteCount+1); i18n.T(next) != next {
		t.Errorf("%s exists: raise quoteCount", next)
	}
	if q := RandomQuote(); strings.HasPrefix(q, "quotes.") {
		t.Errorf("RandomQuote returned an id: %q", q)
	}
}
