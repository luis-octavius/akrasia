package tasks

import (
	"context"
	"errors"
	"testing"
	"time"
)

// family adds a parent with subtasks and returns the parent's name.
func addFamily(t *testing.T, tkm *TaskManager, parent NewTodo, subtasks ...string) {
	t.Helper()
	ctx := context.Background()

	if _, err := tkm.AddTodo(ctx, parent); err != nil {
		t.Fatalf("AddTodo(%q) error = %v", parent.Name, err)
	}
	for _, name := range subtasks {
		if _, err := tkm.AddTodo(ctx, NewTodo{Name: name, Parent: parent.Name}); err != nil {
			t.Fatalf("AddTodo(%q under %q) error = %v", name, parent.Name, err)
		}
	}
}

func inDays(n int) time.Time {
	return dayOf(time.Now()).AddDate(0, 0, n)
}

func find(t *testing.T, tkm *TaskManager, name string) TodoView {
	t.Helper()
	v, err := tkm.FindTodo(context.Background(), name)
	if err != nil {
		t.Fatalf("FindTodo(%q) error = %v", name, err)
	}
	return v
}

func TestSubtasksAreListedUnderTheirParent(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	addFamily(t, tkm, NewTodo{Name: "Party", ExpiresAt: inDays(5)}, "Bake cake", "Invite friends")

	todos, err := tkm.ListTodos(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 1 || len(todos[0].Subtasks) != 2 {
		t.Fatalf("ListTodos() = %+v, want one parent with two subtasks", todos)
	}

	sub := todos[0].Subtasks[0]
	if sub.Parent != "Party" || sub.Path() != "Party/"+sub.Name {
		t.Errorf("subtask parent = %q, path = %q", sub.Parent, sub.Path())
	}
	if !sub.ExpiresAt.Equal(inDays(5)) {
		t.Errorf("subtask due %v, want the parent's date by default", sub.ExpiresAt)
	}
}

func TestSubtaskNamesAreUniquePerParent(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addFamily(t, tkm, NewTodo{Name: "Party", ExpiresAt: inDays(5)}, "Buy drinks")
	addFamily(t, tkm, NewTodo{Name: "Picnic", ExpiresAt: inDays(5)}, "Buy drinks")

	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "Buy drinks", Parent: "Party"}); !errors.Is(err, ErrNameTaken) {
		t.Errorf("same subtask twice under a parent: error = %v, want ErrNameTaken", err)
	}
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "Party", ExpiresAt: inDays(1)}); !errors.Is(err, ErrNameTaken) {
		t.Errorf("same top-level name twice: error = %v, want ErrNameTaken", err)
	}

	// two exact matches: the name alone is ambiguous, the full name is not
	if _, err := tkm.FindTodo(ctx, "buy drinks"); !errors.Is(err, ErrAmbiguousName) {
		t.Errorf("FindTodo(ambiguous) error = %v, want ErrAmbiguousName", err)
	}
	if v := find(t, tkm, "picnic/buy drinks"); v.Parent != "Picnic" {
		t.Errorf("FindTodo(path) = %+v, want the Picnic subtask", v)
	}
}

func TestPathWithSlashInName(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	addFamily(t, tkm, NewTodo{Name: "Taxes 2026/27", ExpiresAt: inDays(5)}, "Receipts")

	if v := find(t, tkm, "Taxes 2026/27/Receipts"); v.Name != "Receipts" {
		t.Errorf("FindTodo() = %q, want the subtask", v.Name)
	}
	if v := find(t, tkm, "Taxes 2026/27"); v.Name != "Taxes 2026/27" || len(v.Subtasks) != 1 {
		t.Errorf("FindTodo() = %+v, want the parent with its subtask", v)
	}
}

func TestSubtasksCannotNest(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	addFamily(t, tkm, NewTodo{Name: "Party", ExpiresAt: inDays(5)}, "Bake cake")

	_, err := tkm.AddTodo(context.Background(), NewTodo{Name: "Buy flour", Parent: "Party/Bake cake"})
	if !errors.Is(err, ErrNestedSubtask) {
		t.Fatalf("AddTodo(under a subtask) error = %v, want ErrNestedSubtask", err)
	}
}

func TestSubtaskDateAndDailyRules(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addFamily(t, tkm, NewTodo{Name: "Party", ExpiresAt: inDays(5)})
	addFamily(t, tkm, NewTodo{Name: "Morning routine", IsDaily: true, ExpiresAt: inDays(1)})

	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "Late", Parent: "Party", ExpiresAt: inDays(6)}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("subtask due after the parent: error = %v, want ErrInvalidInput", err)
	}
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "Early", Parent: "Party", ExpiresAt: inDays(2)}); err != nil {
		t.Errorf("subtask due before the parent: error = %v", err)
	}
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "Daily", Parent: "Party", IsDaily: true}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("daily subtask of a one-off task: error = %v, want ErrInvalidInput", err)
	}

	// a daily parent makes its subtasks daily
	sub, err := tkm.AddTodo(ctx, NewTodo{Name: "Stretch", Parent: "Morning routine"})
	if err != nil || !sub.IsDaily {
		t.Errorf("subtask of a daily task = %+v, %v; want daily", sub, err)
	}
}

func TestParentCompletesWithItsSubtasks(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addFamily(t, tkm, NewTodo{Name: "Party", ExpiresAt: inDays(5)}, "Bake cake", "Invite friends")

	if _, err := tkm.Complete(ctx, "Bake cake", ""); err != nil {
		t.Fatal(err)
	}
	if find(t, tkm, "Party").Done() {
		t.Fatal("parent done with a subtask still pending")
	}

	if _, err := tkm.Complete(ctx, "Invite friends", ""); err != nil {
		t.Fatal(err)
	}
	party := find(t, tkm, "Party")
	if !party.Done() || party.SubtaskProgress() != (Progress{Done: 2, Total: 2}) {
		t.Fatalf("parent after its last subtask: done %v, progress %+v", party.Done(), party.SubtaskProgress())
	}

	// a new pending subtask reopens it
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "Buy drinks", Parent: "Party"}); err != nil {
		t.Fatal(err)
	}
	if find(t, tkm, "Party").Done() {
		t.Error("parent still done after a new pending subtask")
	}
}

func TestCompletingParentNeedsConfirmation(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addFamily(t, tkm, NewTodo{Name: "Party", ExpiresAt: inDays(5)}, "Bake cake", "Invite friends")

	if _, err := tkm.Complete(ctx, "Party", ""); !errors.Is(err, ErrPendingSubtasks) {
		t.Fatalf("Complete(parent) error = %v, want ErrPendingSubtasks", err)
	}
	if find(t, tkm, "Party").Done() {
		t.Fatal("parent was completed without confirmation")
	}

	if _, err := tkm.CompleteAll(ctx, "Party", "all set"); err != nil {
		t.Fatalf("CompleteAll() error = %v", err)
	}
	party := find(t, tkm, "Party")
	if !party.Done() || party.SubtaskProgress().Done != 2 {
		t.Errorf("after CompleteAll: done %v, progress %+v", party.Done(), party.SubtaskProgress())
	}

	_, notes, err := tkm.Notes(ctx, "Party", 0)
	if err != nil || len(notes) != 1 || notes[0].Notes != "all set" {
		t.Errorf("parent notes = %+v, %v; want the note given to CompleteAll", notes, err)
	}
}

func TestDailyParentClosesTheDay(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addFamily(t, tkm, NewTodo{Name: "Morning routine", IsDaily: true, ExpiresAt: inDays(1)}, "Stretch", "Meditate")

	for _, name := range []string{"Stretch", "Meditate"} {
		if _, err := tkm.Complete(ctx, name, ""); err != nil {
			t.Fatal(err)
		}
	}

	routine := find(t, tkm, "Morning routine")
	if !routine.DoneToday {
		t.Fatal("daily parent not done today after all its subtasks")
	}
	if _, streak, err := tkm.CurrentStreak(ctx, "Morning routine"); err != nil || streak != 1 {
		t.Errorf("parent streak = %d, %v; want 1", streak, err)
	}

	// a new subtask added the same day reopens today
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "Read", Parent: "Morning routine"}); err != nil {
		t.Fatal(err)
	}
	if find(t, tkm, "Morning routine").DoneToday {
		t.Error("daily parent still done today after a new subtask")
	}
}

func TestTodayProgressCountsAParentOnce(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addFamily(t, tkm, NewTodo{Name: "Morning routine", IsDaily: true, ExpiresAt: inDays(1)}, "Stretch", "Meditate", "Read")

	if _, err := tkm.Complete(ctx, "Stretch", ""); err != nil {
		t.Fatal(err)
	}
	p, err := tkm.TodayProgress(ctx)
	if err != nil || p != (Progress{Done: 0, Total: 1}) {
		t.Errorf("TodayProgress() = %+v, %v; want the parent as one pending item", p, err)
	}
}

func TestDeletingParentDeletesSubtasks(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addFamily(t, tkm, NewTodo{Name: "Party", ExpiresAt: inDays(5)}, "Bake cake", "Invite friends")

	deleted, err := tkm.DeleteByName(ctx, "Party")
	if err != nil || len(deleted.Subtasks) != 2 {
		t.Fatalf("DeleteByName() = %+v, %v; want it with its 2 subtasks", deleted, err)
	}
	if _, err := tkm.FindTodo(ctx, "Bake cake"); !errors.Is(err, ErrTodoNotFound) {
		t.Errorf("subtask still there: %v", err)
	}
}

func TestDeletingLastPendingSubtaskCompletesParent(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addFamily(t, tkm, NewTodo{Name: "Party", ExpiresAt: inDays(5)}, "Bake cake", "Invite friends")

	if _, err := tkm.Complete(ctx, "Bake cake", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := tkm.DeleteByName(ctx, "Invite friends"); err != nil {
		t.Fatal(err)
	}
	if !find(t, tkm, "Party").Done() {
		t.Error("parent not done once its only pending subtask was deleted")
	}
}

func TestEditingParentCarriesSubtasks(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addFamily(t, tkm, NewTodo{Name: "Party", ExpiresAt: inDays(5)}, "Bake cake")

	earlier := inDays(3)
	if _, err := tkm.EditTodo(ctx, "Party", TodoPatch{ExpiresAt: &earlier}); err != nil {
		t.Fatal(err)
	}
	if got := find(t, tkm, "Bake cake").ExpiresAt; !got.Equal(earlier) {
		t.Errorf("subtask due %v after the parent moved to %v", got, earlier)
	}

	daily := true
	if _, err := tkm.EditTodo(ctx, "Party", TodoPatch{IsDaily: &daily}); err != nil {
		t.Fatal(err)
	}
	if !find(t, tkm, "Bake cake").IsDaily {
		t.Error("subtask not daily after its parent became daily")
	}

	notDaily := false
	if _, err := tkm.EditTodo(ctx, "Party/Bake cake", TodoPatch{IsDaily: &notDaily}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("subtask leaving its parent's daily mode: error = %v, want ErrInvalidInput", err)
	}
}

func TestSubtaskRenameKeepsNamesUniquePerParent(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()
	addFamily(t, tkm, NewTodo{Name: "Party", ExpiresAt: inDays(5)}, "Bake cake", "Invite friends")

	name := "Bake cake"
	if _, err := tkm.EditTodo(ctx, "Party/Invite friends", TodoPatch{Name: &name}); !errors.Is(err, ErrNameTaken) {
		t.Errorf("renaming to a sibling's name: error = %v, want ErrNameTaken", err)
	}
	other := "Party"
	if _, err := tkm.EditTodo(ctx, "Party/Invite friends", TodoPatch{Name: &other}); err != nil {
		t.Errorf("a subtask may share its parent's name: %v", err)
	}

	// the bare name still reaches the parent; the subtask has its full name
	if v := find(t, tkm, "Party"); v.IsSubtask() {
		t.Errorf("FindTodo(Party) = the subtask, want the parent")
	}
	if v := find(t, tkm, "Party/Party"); !v.IsSubtask() {
		t.Errorf("FindTodo(Party/Party) = the parent, want the subtask")
	}
}
