package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// AddTodo persists a new task and returns it.
func (tkm *TaskManager) AddTodo(ctx context.Context, in NewTodo) (TodoView, error) {
	now := time.Now()

	todo, err := tkm.Queries.AddTodo(ctx, database.AddTodoParams{
		ID:           uuid.New(),
		Name:         in.Name,
		Description:  validateDescription(in.Description),
		CreatedAt:    now,
		UpdatedAt:    now,
		Concluded:    false,
		ExpiresAt:    in.ExpiresAt,
		Priority:     in.Priority,
		IsDaily:      in.IsDaily,
		HistorySince: sql.NullString{String: now.Format(time.DateOnly), Valid: true},
	})
	if isUniqueViolation(err) {
		return TodoView{}, localizedError{
			msg:      fmt.Sprintf(i18n.T("tasks.error.nameTaken"), in.Name),
			sentinel: ErrNameTaken,
		}
	}
	if err != nil {
		return TodoView{}, fmt.Errorf(i18n.T("tasks.error.createTask"), err)
	}

	return viewFromTodo(todo)
}

// ListTodos returns every task with today's status, optionally filtered by
// priority.
func (tkm *TaskManager) ListTodos(ctx context.Context, priorityFilter string) ([]TodoView, error) {
	return tkm.listTodoViews(ctx, func(v TodoView) bool {
		return priorityFilter == "" || v.Priority == priorityFilter
	})
}

// listTodoViews loads every task with today's status and keeps the ones
// accepted by keep.
func (tkm *TaskManager) listTodoViews(ctx context.Context, keep func(TodoView) bool) ([]TodoView, error) {
	rows, err := tkm.Queries.ListTodosWithTodayStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf(i18n.T("tasks.error.getTaskDatabase"), err)
	}

	views := make([]TodoView, 0, len(rows))
	for _, row := range rows {
		view, err := newTodoView(row)
		if err != nil {
			return nil, err
		}
		if keep(view) {
			views = append(views, view)
		}
	}

	return views, nil
}

// FindTodo returns the task that best matches name (case-insensitive,
// fuzzy), with today's status.
func (tkm *TaskManager) FindTodo(ctx context.Context, name string) (TodoView, error) {
	row, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, err
	}

	view, err := viewFromNameRow(row)
	if err != nil {
		return TodoView{}, err
	}

	view.DoneToday, err = tkm.isDoneToday(ctx, row.ID)
	if err != nil {
		return TodoView{}, err
	}

	return view, nil
}

// DailyTodos returns every daily task with whether it was done today.
func (tkm *TaskManager) DailyTodos(ctx context.Context) ([]TodoView, error) {
	todos, err := tkm.listTodoViews(ctx, func(v TodoView) bool { return v.IsDaily })
	if err != nil {
		return nil, fmt.Errorf(i18n.T("tasks.error.getAllDailyTodos"), err)
	}

	return todos, nil
}

// DeleteByName removes the task that best matches name and returns it, so
// callers can show which task the fuzzy search actually picked.
func (tkm *TaskManager) DeleteByName(ctx context.Context, name string) (TodoView, error) {
	row, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, err
	}

	if err := tkm.Queries.DeleteTodoByName(ctx, row.Name); err != nil {
		return TodoView{}, fmt.Errorf(i18n.T("tasks.error.deleteTaskByName"), err)
	}

	return viewFromNameRow(row)
}
