package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/luis-octavius/akrasia/internal/tasks"
)

func TestWrapTextHasNoLeadingSpaces(t *testing.T) {
	text := strings.Repeat("word ", 30)

	for _, line := range strings.Split(wrapText(text, 44), "\n") {
		if strings.HasPrefix(line, " ") {
			t.Fatalf("line starts with a space: %q", line)
		}
		if len(line) > 40 {
			t.Fatalf("line longer than 40 chars: %q", line)
		}
	}
}

func TestRenderTodoShowsTheTime(t *testing.T) {
	due := time.Date(2026, time.December, 10, 0, 0, 0, 0, time.UTC)

	timed := ansi.Strip(renderTodo(tasks.TodoView{Name: "call", ExpiresAt: due, DueTime: "14:30"}))
	if !strings.Contains(timed, "10 Dec 26 14:30") {
		t.Errorf("a timed task should show its date and time:\n%s", timed)
	}

	untimed := ansi.Strip(renderTodo(tasks.TodoView{Name: "call", ExpiresAt: due}))
	if !strings.Contains(untimed, "10 Dec 26") || strings.Contains(untimed, "00:00") {
		t.Errorf("an untimed task shows only its date:\n%s", untimed)
	}

	daily := ansi.Strip(renderTodo(tasks.TodoView{Name: "run", IsDaily: true, DueTime: "07:00"}))
	if !strings.Contains(daily, "07:00") {
		t.Errorf("a timed daily task should show its time:\n%s", daily)
	}

	sub := ansi.Strip(renderSubtask(tasks.TodoView{Name: "bake", ExpiresAt: due, DueTime: "16:00"}))
	if !strings.Contains(sub, "10 Dec 16:00") {
		t.Errorf("a timed subtask should show its time:\n%s", sub)
	}
}
