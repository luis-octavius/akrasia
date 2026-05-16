package tasks

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/color"
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

// printTodo receives a Todo and create a readable output
func printTodo(todo database.Todo) {
	todoTime := todo.ExpiresAt.Format(time.RFC822)

	var status string

	if todo.Concluded == true {
		status = "Done"
	} else {
		status = "Not done"
	}

	s := fmt.Sprintf("%v | %v\n%v | %v\n\n", todo.Name, todo.Description.String, todoTime, status)

	if ok := checkIfTodoExpires(todo.ExpiresAt); !ok {
		color.MsgSuccess(s)
		return
	}

	color.MsgWarning(s)
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
func buildTodayView(todos []database.Todo, now time.Time) TodayView {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	view := TodayView{}

	for _, todo := range todos {
		if todo.Concluded {
			continue
		}

		if todo.IsDaily {
			view.Daily = append(view.Daily, todo)
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
func printTodaySection(title string, todos []database.Todo) {
	if len(todos) == 0 {
		return
	}

	fmt.Printf("\n%s (%d)\n", title, len(todos))
	for _, todo := range todos {
		printTodo(todo)
	}
}
