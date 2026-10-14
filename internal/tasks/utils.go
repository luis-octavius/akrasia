package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

func validateDescription(description string) sql.NullString {
	descriptionField := sql.NullString{}

	if description == "" {
		descriptionField.String = ""
		descriptionField.Valid = false
	} else {
		descriptionField.String = description
		descriptionField.Valid = true
	}

	return descriptionField
}

// checkIfTodoExpires checks if a task will expire by 5 days
func checkIfTodoExpires(expiresAt time.Time) bool {
	actualDay := time.Now()

	diff := expiresAt.Sub(actualDay).String()
	hour, _, _ := strings.Cut(diff, "h")
	hourToInt, _ := strconv.Atoi(hour)
	if hourToInt <= (24 * 5) {
		return true
	}
	return false
}

// todoKey turns a todo's ID (stored generically as interface{}) into a
// comparable map key, since it can't be used directly as one.
func todoKey(id interface{}) string {
	return fmt.Sprintf("%v", id)
}

// printTodo receives a Todo and create a readable output
func printTodo(todo database.Todo) {
	fmt.Println(renderTodo(todo))
}

// renderTodo mounts the string to beautifully output the task
func renderTodo(todo database.Todo) string {
	icon, status := "○", i18n.T("tasks.tasks.notDone")

	if todo.Concluded {
		icon, status = "✔", i18n.T("tasks.tasks.done")
	}

	state := color.Muted()
	switch {
	case todo.Concluded:
		state = color.Success()
	case checkIfTodoExpires(todo.ExpiresAt):
		state = color.Warning()
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n", state.Render(icon), color.Title().Render(todo.Name))

	if desc := todo.Description.String; desc != "" {
		fmt.Fprintf(&b, "- %s\n", desc)
	}

	expires := fmt.Sprintf("%s · %s\n", todo.ExpiresAt.Format("02 Jan 06 15:04"), status)
	fmt.Fprintf(&b, "  %s\n", color.Muted().Render(expires))

	if priority := todo.Priority; priority != "" {
		fmt.Fprintf(&b, "  %s\n", color.Muted().Render("Priority: " + todo.Priority))
	}

	return b.String()
}

// filterTodosByPriority keeps only tasks matching the provided priority.
func filterTodosByPriority(todos []database.Todo, priorityFilter string) []database.Todo {
	if priorityFilter == "" {
		return todos
	}

	filtered := make([]database.Todo, 0, len(todos))
	for _, todo := range todos {
		if todo.Priority == priorityFilter {
			filtered = append(filtered, todo)
		}
	}

	return filtered
}

// buildTodayView classifies pending tasks into today-oriented sections.
// doneToday holds the keys (see todoKey) of tasks already completed today —
// daily tasks are judged against it instead of todo.Concluded, because that
// flag is never reset back to false once a daily task is marked done.
func buildTodayView(todos []database.Todo, now time.Time, doneToday map[string]bool) TodayView {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	view := TodayView{}

	for _, todo := range todos {
		if todo.IsDaily {
			if doneToday[todoKey(todo.ID)] {
				continue
			}
			view.Daily = append(view.Daily, todo)
			continue
		}

		if todo.Concluded {
			continue
		}

		if todo.ExpiresAt.Before(dayStart) {
			view.Overdue = append(view.Overdue, todo)
			continue
		}

		if !todo.ExpiresAt.Before(dayStart) && todo.ExpiresAt.Before(dayEnd) {
			view.DueToday = append(view.DueToday, todo)
			continue
		}

		if todo.ExpiresAt.Before(now.Add(5 * 24 * time.Hour)) {
			view.ExpiringSoon = append(view.ExpiringSoon, todo)
		}
	}

	sortTodayTodos(view.Overdue)
	sortTodayTodos(view.DueToday)
	sortTodayTodos(view.Daily)
	sortTodayTodos(view.ExpiringSoon)

	return view
}

// applyTodayOptions applies limit and section filters to a today view.
func applyTodayOptions(view TodayView, opts TodayOptions) TodayView {
	if opts.Limit > 0 {
		view.Overdue = limitTodos(view.Overdue, opts.Limit)
		view.DueToday = limitTodos(view.DueToday, opts.Limit)
		view.Daily = limitTodos(view.Daily, opts.Limit)
		view.ExpiringSoon = limitTodos(view.ExpiringSoon, opts.Limit)
	}

	switch opts.Only {
	case "overdue":
		view.DueToday = nil
		view.Daily = nil
		view.ExpiringSoon = nil
	case "today":
		view.Overdue = nil
		view.Daily = nil
		view.ExpiringSoon = nil
	case "daily":
		view.Overdue = nil
		view.DueToday = nil
		view.ExpiringSoon = nil
	case "soon":
		view.Overdue = nil
		view.DueToday = nil
		view.Daily = nil
	}

	return view
}

// limitTodos truncates a slice to the requested size when needed.
func limitTodos(todos []database.Todo, limit int) []database.Todo {
	if limit <= 0 || len(todos) <= limit {
		return todos
	}

	return todos[:limit]
}

// sortTodayTodos orders tasks by priority first, then expiration date.
func sortTodayTodos(todos []database.Todo) {
	sort.Slice(todos, func(i, j int) bool {
		leftPriority := priorityRank(todos[i].Priority)
		rightPriority := priorityRank(todos[j].Priority)

		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}

		return todos[i].ExpiresAt.Before(todos[j].ExpiresAt)
	})
}

// priorityRank maps textual priority to a sortable numeric value.
func priorityRank(priority string) int {
	switch priority {
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

// printTodaySection renders one titled section from the today dashboard.
func printTodaySection(tkm *TaskManager, title string, todos []database.Todo) {
	if len(todos) == 0 {
		return
	}

	fmt.Printf("\n%s (%d)\n", title, len(todos))
	for _, todo := range todos {
		printTodo(todo)
	}
}

func GetTodoByName(tkm *TaskManager, name string) (database.GetTodoByNameRow, error) {
	loweredName := strings.ToLower(name)

	todo, err := tkm.Queries.GetTodoByName(context.Background(), database.GetTodoByNameParams{
		Name:    loweredName,
		Column2: sql.NullString{String: loweredName, Valid: true},
		Column3: sql.NullString{String: loweredName, Valid: true},
		Column4: sql.NullString{String: loweredName, Valid: true},
	})
	if err != nil {
		return database.GetTodoByNameRow{}, fmt.Errorf(i18n.T("tasks.error.getTodoByName"), err)
	}

	return todo, nil
}

func transformListTodosToTodo(todoStatus database.ListTodosWithTodayStatusRow) database.Todo {
	var todo database.Todo
	todo = database.Todo{
		ID:           todoStatus.ID,
		Name:         todoStatus.Name,
		Description:  todoStatus.Description,
		CreatedAt:    todoStatus.CreatedAt,
		UpdatedAt:    todoStatus.UpdatedAt,
		ExpiresAt:    todoStatus.ExpiresAt,
		HistorySince: todoStatus.HistorySince,
		IsDaily:      todoStatus.IsDaily,
		Concluded:    todoStatus.DoneToday,
		Priority:     todoStatus.Priority,
	}

	return todo
}
