package tasks

import (
	"context"
	"errors"
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }

func TestEditTodoRenameKeepsHistory(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	addTestTodo(t, tkm, "read", "medium", true)
	if _, err := tkm.Complete(ctx, "read", "chapter 1"); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	got, err := tkm.EditTodo(ctx, "read", TodoPatch{Name: ptr("read 30min")})
	if err != nil {
		t.Fatalf("EditTodo() error = %v", err)
	}
	if got.Name != "read 30min" || !got.DoneToday {
		t.Fatalf("renamed task should keep today's completion, got %+v", got)
	}

	report, err := tkm.Debug(ctx, "read 30min")
	if err != nil {
		t.Fatalf("Debug() error = %v", err)
	}
	if len(report.History) != 1 || report.History[0].Notes != "chapter 1" {
		t.Fatalf("history should follow the rename, got %+v", report.History)
	}
}

func TestEditTodoOnlyChangesPatchedFields(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	original, err := tkm.AddTodo(ctx, NewTodo{
		Name:        "taxes",
		Description: "pay before friday",
		Priority:    "high",
		ExpiresAt:   time.Now().Add(72 * time.Hour),
	})
	if err != nil {
		t.Fatalf("AddTodo() error = %v", err)
	}

	got, err := tkm.EditTodo(ctx, "taxes", TodoPatch{Priority: ptr("LOW"), Description: ptr("")})
	if err != nil {
		t.Fatalf("EditTodo() error = %v", err)
	}

	if got.Priority != "low" || got.Description.Valid {
		t.Fatalf("priority should be low and description cleared, got %+v", got)
	}
	if got.Name != "taxes" || !got.ExpiresAt.Equal(original.ExpiresAt) || got.IsDaily {
		t.Fatalf("untouched fields changed: %+v", got)
	}
}

func TestEditTodoOneOffBecomesDaily(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "stretch", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatalf("AddTodo() error = %v", err)
	}

	got, err := tkm.EditTodo(ctx, "stretch", TodoPatch{IsDaily: ptr(true)})
	if err != nil {
		t.Fatalf("EditTodo() error = %v", err)
	}
	if !got.IsDaily {
		t.Fatalf("task should be daily now")
	}

	report, err := tkm.Debug(ctx, "stretch")
	if err != nil {
		t.Fatalf("Debug() error = %v", err)
	}
	if report.HistorySince == "" {
		t.Fatalf("a task turned daily needs a streak window")
	}

	daily, err := tkm.DailyTodos(ctx)
	if err != nil || len(daily) != 1 || daily[0].Name != "stretch" {
		t.Fatalf("stretch should be listed as daily, got %v (%v)", daily, err)
	}
}

func TestEditTodoErrors(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	addTestTodo(t, tkm, "taxes", "high", false)
	addTestTodo(t, tkm, "groceries", "low", false)

	cases := []struct {
		name  string
		patch TodoPatch
		want  error
	}{
		{"taxes", TodoPatch{Name: ptr("groceries")}, ErrNameTaken},
		{"taxes", TodoPatch{Name: ptr("  ")}, ErrInvalidInput},
		{"taxes", TodoPatch{Priority: ptr("urgent")}, ErrInvalidInput},
		{"nope", TodoPatch{Priority: ptr("low")}, ErrTodoNotFound},
	}
	for _, c := range cases {
		if _, err := tkm.EditTodo(ctx, c.name, c.patch); !errors.Is(err, c.want) {
			t.Errorf("EditTodo(%q, %+v) error = %v, want %v", c.name, c.patch, err, c.want)
		}
	}
}

func TestTodoPatchEmpty(t *testing.T) {
	if !(TodoPatch{}).Empty() {
		t.Error("zero patch should be empty")
	}
	if (TodoPatch{IsDaily: ptr(false)}).Empty() {
		t.Error("a patch setting daily=false is not empty")
	}
}
