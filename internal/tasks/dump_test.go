package tasks

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDeleteAndPromoteDump(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	for _, name := range []string{"app idea", "learn sql", "buy plants"} {
		if _, err := tkm.AddDump(ctx, name); err != nil {
			t.Fatalf("AddDump(%q) error = %v", name, err)
		}
	}

	second, err := tkm.DumpAt(ctx, 2)
	if err != nil || second.Name != "learn sql" {
		t.Fatalf("DumpAt(2) = %+v, %v; want learn sql", second, err)
	}
	if _, err := tkm.DumpAt(ctx, 4); !errors.Is(err, ErrDumpNotFound) {
		t.Fatalf("DumpAt(4) should be ErrDumpNotFound, got %v", err)
	}

	if err := tkm.DeleteDump(ctx, second.ID); err != nil {
		t.Fatalf("DeleteDump() error = %v", err)
	}

	first, _ := tkm.DumpAt(ctx, 1)
	todo, err := tkm.PromoteDump(ctx, first.ID, NewTodo{Name: first.Name, Priority: "high", ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatalf("PromoteDump() error = %v", err)
	}
	if todo.Name != "app idea" || todo.Priority != "high" {
		t.Fatalf("unexpected task from dump: %+v", todo)
	}

	left, err := tkm.Dumps(ctx)
	if err != nil || len(left) != 1 || left[0].Name != "buy plants" {
		t.Fatalf("only buy plants should be left in the dump, got %v (%v)", left, err)
	}
}

func TestPromoteDumpKeepsEntryWhenNameIsTaken(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	addTestTodo(t, tkm, "taxes", "high", false)
	entry, err := tkm.AddDump(ctx, "taxes")
	if err != nil {
		t.Fatalf("AddDump() error = %v", err)
	}

	if _, err := tkm.PromoteDump(ctx, entry.ID, NewTodo{Name: "taxes", ExpiresAt: time.Now().Add(time.Hour)}); !errors.Is(err, ErrNameTaken) {
		t.Fatalf("expected ErrNameTaken, got %v", err)
	}
	if left, _ := tkm.Dumps(ctx); len(left) != 1 {
		t.Fatalf("the entry must stay in the dump when the task can't be created, got %v", left)
	}
}
