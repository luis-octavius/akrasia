package tasks

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestParseClock(t *testing.T) {
	cases := []struct {
		in, want string
		wantErr  bool
	}{
		{"14:30", "14:30", false},
		{"9:05", "09:05", false},
		{" 07:00 ", "07:00", false},
		{"00:00", "00:00", false},
		{"", "", false},
		{"24:00", "", true},
		{"12:60", "", true},
		{"noon", "", true},
		{"1430", "", true},
	}
	for _, c := range cases {
		got, err := ParseClock(c.in)
		if (err != nil) != c.wantErr || got != c.want {
			t.Errorf("ParseClock(%q) = %q, %v; want %q, error %v", c.in, got, err, c.want, c.wantErr)
		}
	}
}

func TestAddTodoWithTime(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	got, err := tkm.AddTodo(ctx, NewTodo{Name: "call", ExpiresAt: time.Now().AddDate(0, 0, 1), DueTime: "9:30"})
	if err != nil || got.DueTime != "09:30" {
		t.Fatalf("AddTodo() = %+v, %v; want time 09:30", got, err)
	}

	found, err := tkm.FindTodo(ctx, "call")
	if err != nil || found.DueTime != "09:30" {
		t.Fatalf("the time should be stored, got %+v, %v", found, err)
	}

	_, err = tkm.AddTodo(ctx, NewTodo{Name: "bad", ExpiresAt: time.Now().AddDate(0, 0, 1), DueTime: "25:00"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("an invalid time should be rejected as invalid input, got %v", err)
	}

	plain, err := tkm.AddTodo(ctx, NewTodo{Name: "plain", ExpiresAt: time.Now().AddDate(0, 0, 1)})
	if err != nil || plain.DueTime != "" {
		t.Fatalf("a task without time should have none, got %+v, %v", plain, err)
	}
}

func TestEditTodoTime(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "call", Priority: "high", ExpiresAt: time.Now().AddDate(0, 0, 1)}); err != nil {
		t.Fatal(err)
	}

	got, err := tkm.EditTodo(ctx, "call", TodoPatch{DueTime: ptr("15:00")})
	if err != nil || got.DueTime != "15:00" || got.Priority != "high" {
		t.Fatalf("EditTodo(time) = %+v, %v; want 15:00 and the priority kept", got, err)
	}

	// another field keeps the time
	got, err = tkm.EditTodo(ctx, "call", TodoPatch{Priority: ptr("low")})
	if err != nil || got.DueTime != "15:00" {
		t.Fatalf("editing the priority should keep the time, got %+v, %v", got, err)
	}

	got, err = tkm.EditTodo(ctx, "call", TodoPatch{DueTime: ptr("")})
	if err != nil || got.DueTime != "" {
		t.Fatalf("an empty time should remove it, got %+v, %v", got, err)
	}

	if _, err := tkm.EditTodo(ctx, "call", TodoPatch{DueTime: ptr("soon")}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("an invalid time should be rejected, got %v", err)
	}
	if (TodoPatch{DueTime: ptr("")}).Empty() {
		t.Fatal("a patch that only removes the time is not empty")
	}
}

func TestSortTodayTodosByTime(t *testing.T) {
	day := time.Date(2026, time.December, 10, 0, 0, 0, 0, time.UTC)
	todos := []TodoView{
		{Name: "untimed", Priority: "medium", ExpiresAt: day},
		{Name: "late", Priority: "medium", ExpiresAt: day, DueTime: "18:00"},
		{Name: "early", Priority: "medium", ExpiresAt: day, DueTime: "08:15"},
		{Name: "urgent", Priority: "high", ExpiresAt: day},
	}
	sortTodayTodos(todos)

	var got []string
	for _, todo := range todos {
		got = append(got, todo.Name)
	}
	want := []string{"urgent", "early", "late", "untimed"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

func TestSortDailyTodosByTime(t *testing.T) {
	// a daily task's date means nothing: the time orders them
	todos := []TodoView{
		{Name: "evening", IsDaily: true, Priority: "medium", ExpiresAt: time.Now().Add(-48 * time.Hour), DueTime: "21:00"},
		{Name: "morning", IsDaily: true, Priority: "medium", ExpiresAt: time.Now(), DueTime: "07:00"},
	}
	sortTodayTodos(todos)
	if todos[0].Name != "morning" {
		t.Fatalf("morning should come first, got %v then %v", todos[0].Name, todos[1].Name)
	}
}
