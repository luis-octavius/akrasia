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

// AddTodo persists a new task and returns it. With Parent set, it becomes
// a subtask of that task: it follows the parent's daily mode, is due no
// later than it (its date by default), and reopens a parent that was done.
func (tkm *TaskManager) AddTodo(ctx context.Context, in NewTodo) (TodoView, error) {
	now := time.Now()

	var parent TodoView
	if in.Parent != "" {
		var err error
		if parent, err = tkm.parentFor(ctx, in.Parent); err != nil {
			return TodoView{}, err
		}
		if parent.IsDaily {
			in.IsDaily = true
		}
		if in.ExpiresAt.IsZero() {
			in.ExpiresAt = parent.ExpiresAt
		}
		if err := subtaskRules(parent, in.IsDaily, in.ExpiresAt); err != nil {
			return TodoView{}, err
		}
	}

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
		ParentID:     uuid.NullUUID{UUID: parent.ID, Valid: in.Parent != ""},
	})
	if isUniqueViolation(err) {
		return TodoView{}, localizedError{
			msg:      fmt.Sprintf(i18n.T("tasks.error.nameTaken"), path(parent.Name, in.Name)),
			sentinel: ErrNameTaken,
		}
	}
	if err != nil {
		return TodoView{}, fmt.Errorf(i18n.T("tasks.error.createTask"), err)
	}

	view, err := viewFromTodo(todo)
	if err != nil {
		return TodoView{}, err
	}
	if in.Parent != "" {
		view.Parent = parent.Name
		// a new pending subtask means the parent isn't done any more
		if err := tkm.syncParent(ctx, parent.ID); err != nil {
			return view, err
		}
	}
	return view, nil
}

// parentFor finds the task a subtask goes under; it must be top-level.
func (tkm *TaskManager) parentFor(ctx context.Context, name string) (TodoView, error) {
	row, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, err
	}
	if row.ParentName != "" {
		return TodoView{}, localizedError{
			msg:      fmt.Sprintf(i18n.T("tasks.error.nestedSubtask"), path(row.ParentName, row.Name)),
			sentinel: ErrNestedSubtask,
		}
	}
	id, err := toUUID(row.ID)
	if err != nil {
		return TodoView{}, err
	}
	return tkm.family(ctx, id)
}

// ListTodos returns every top-level task with today's status and its
// subtasks attached, optionally filtered by priority (of the top-level
// task).
func (tkm *TaskManager) ListTodos(ctx context.Context, priorityFilter string) ([]TodoView, error) {
	return tkm.listTodoViews(ctx, func(v TodoView) bool {
		return priorityFilter == "" || v.Priority == priorityFilter
	})
}

// listTodoViews loads every top-level task with today's status and its
// subtasks, and keeps the ones accepted by keep.
func (tkm *TaskManager) listTodoViews(ctx context.Context, keep func(TodoView) bool) ([]TodoView, error) {
	flat, err := tkm.allTodoViews(ctx)
	if err != nil {
		return nil, err
	}

	var views []TodoView
	for _, view := range nest(flat) {
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

	if !view.IsSubtask() {
		family, err := tkm.family(ctx, view.ID)
		if err != nil {
			return TodoView{}, err
		}
		view.Subtasks = family.Subtasks
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
// callers can show which task the fuzzy search actually picked. A parent
// goes with its subtasks (returned in Subtasks); removing a pending
// subtask can leave the parent with every subtask done, which completes it.
func (tkm *TaskManager) DeleteByName(ctx context.Context, name string) (TodoView, error) {
	todo, err := tkm.FindTodo(ctx, name)
	if err != nil {
		return TodoView{}, err
	}

	if len(todo.Subtasks) > 0 {
		if err := tkm.Queries.DeleteSubtasks(ctx, uuid.NullUUID{UUID: todo.ID, Valid: true}); err != nil {
			return TodoView{}, fmt.Errorf(i18n.T("tasks.error.deleteTaskByName"), err)
		}
	}
	if err := tkm.Queries.DeleteTodoByID(ctx, todo.ID); err != nil {
		return TodoView{}, fmt.Errorf(i18n.T("tasks.error.deleteTaskByName"), err)
	}

	if todo.IsSubtask() {
		if err := tkm.syncParent(ctx, todo.ParentID); err != nil {
			return todo, err
		}
	}
	return todo, nil
}
