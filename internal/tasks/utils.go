package tasks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// ExpiringWindow is how close to its deadline a task must be to count as
// expiring soon.
const ExpiringWindow = 5 * 24 * time.Hour

func validateDescription(description string) sql.NullString {
	return sql.NullString{String: description, Valid: description != ""}
}

// IsExpiringSoon reports whether expiresAt falls within ExpiringWindow from
// now (already expired counts too).
func IsExpiringSoon(expiresAt time.Time) bool {
	return time.Until(expiresAt) <= ExpiringWindow
}

// buildTodayView classifies pending tasks into today-oriented sections.
// Daily tasks are judged by DoneToday instead of Concluded, because that
// flag is never set for daily tasks.
func buildTodayView(todos []TodoView, now time.Time) TodayView {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	view := TodayView{}

	for _, todo := range todos {
		if todo.IsDaily {
			if todo.DoneToday {
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

		if todo.ExpiresAt.Before(dayEnd) {
			view.DueToday = append(view.DueToday, todo)
			continue
		}

		if todo.ExpiresAt.Before(now.Add(ExpiringWindow)) {
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
func limitTodos(todos []TodoView, limit int) []TodoView {
	if limit <= 0 || len(todos) <= limit {
		return todos
	}

	return todos[:limit]
}

// sortTodayTodos orders tasks by priority first, then expiration date.
func sortTodayTodos(todos []TodoView) {
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

// findTodoByName runs the fuzzy name search and maps "no rows" to
// ErrTodoNotFound.
func findTodoByName(ctx context.Context, tkm *TaskManager, name string) (database.GetTodoByNameRow, error) {
	loweredName := strings.ToLower(name)

	todo, err := tkm.Queries.GetTodoByName(ctx, database.GetTodoByNameParams{
		Name:    loweredName,
		Column2: sql.NullString{String: loweredName, Valid: true},
		Column3: sql.NullString{String: loweredName, Valid: true},
		Column4: sql.NullString{String: loweredName, Valid: true},
	})
	if errors.Is(err, sql.ErrNoRows) {
		return database.GetTodoByNameRow{}, localizedError{
			msg:      fmt.Sprintf(i18n.T("tasks.error.todoNotFound"), name),
			sentinel: ErrTodoNotFound,
		}
	}
	if err != nil {
		return database.GetTodoByNameRow{}, fmt.Errorf(i18n.T("tasks.error.getTodoByName"), err)
	}

	return todo, nil
}

func newTodoView(todoStatus database.ListTodosWithTodayStatusRow) (TodoView, error) {
	id, err := toUUID(todoStatus.ID)
	if err != nil {
		return TodoView{}, err
	}

	return TodoView{
		ID:          id,
		Name:        todoStatus.Name,
		Description: todoStatus.Description,
		Concluded:   todoStatus.Concluded,
		ExpiresAt:   todoStatus.ExpiresAt,
		Priority:    todoStatus.Priority,
		IsDaily:     todoStatus.IsDaily,
		DoneToday:   todoStatus.DoneToday,
	}, nil
}

// viewFromTodo converts a plain todos row; DoneToday is left false because
// the row does not carry it.
func viewFromTodo(todo database.Todo) (TodoView, error) {
	id, err := toUUID(todo.ID)
	if err != nil {
		return TodoView{}, err
	}

	return TodoView{
		ID:          id,
		Name:        todo.Name,
		Description: todo.Description,
		Concluded:   todo.Concluded,
		ExpiresAt:   todo.ExpiresAt,
		Priority:    todo.Priority,
		IsDaily:     todo.IsDaily,
	}, nil
}

func viewFromNameRow(todo database.GetTodoByNameRow) (TodoView, error) {
	return viewFromTodo(database.Todo{
		ID:          todo.ID,
		Name:        todo.Name,
		Description: todo.Description,
		Concluded:   todo.Concluded,
		ExpiresAt:   todo.ExpiresAt,
		Priority:    todo.Priority,
		IsDaily:     todo.IsDaily,
	})
}

func viewsFromTodos(todos []database.Todo) ([]TodoView, error) {
	views := make([]TodoView, 0, len(todos))
	for _, todo := range todos {
		view, err := viewFromTodo(todo)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}

	return views, nil
}

func toUUID(v interface{}) (uuid.UUID, error) {
	switch id := v.(type) {
	case uuid.UUID:
		return id, nil
	case string:
		return uuid.Parse(id)
	case []byte:
		return uuid.ParseBytes(id)
	default:
		return uuid.Nil, fmt.Errorf("unexpected id type %T", v)
	}
}
