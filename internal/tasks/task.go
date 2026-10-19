package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

var (
	NoExpiring    = i18n.T("tasks.info.NoExpiring")
	SuccessDelete = i18n.T("tasks.info.SuccessDelete")
)

// doneTodaySet returns the IDs (as comparable keys) of every task already
// completed today, daily or not. This replaces reading todos.concluded for
// daily tasks, since that flag is no longer reset by anything — the only
// source of truth for "did I do this today" is todos_history itself.
func (tkm *TaskManager) doneTodaySet() (map[string]bool, error) {
	ids, err := tkm.Queries.GetDoneTodayIDs(context.Background())
	if err != nil {
		return nil, err
	}

	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[todoKey(id)] = true
	}

	return set, nil
}

// addTodo persists a new task record and prints success feedback.
func (tkm *TaskManager) AddTodo(name, description, priority string, isDaily bool, expiresAt time.Time) error {
	descriptionField := validateDescription(description)
	now := time.Now()

	_, err := tkm.Queries.AddTodo(context.Background(), database.AddTodoParams{
		ID:           uuid.New(),
		Name:         name,
		Description:  descriptionField,
		CreatedAt:    now,
		UpdatedAt:    now,
		Concluded:    false,
		ExpiresAt:    expiresAt,
		Priority:     priority,
		IsDaily:      isDaily,
		HistorySince: sql.NullString{String: now.Format(time.DateOnly), Valid: true},
	})
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.createTask"), err)
	}

	color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.createdTask"), name))
	generateRandomQuote()
	return nil
}

// getTodos lists all tasks, optionally filtered by priority.
func (tkm *TaskManager) GetTodos(priorityFilter string) error {
	todos, err := tkm.ListTodos(context.Background(), priorityFilter)
	if err != nil {
		return err
	}

	fmt.Println(i18n.T("tasks.tasks.allTasks"))

	for _, todo := range todos {
		printTodoView(todo)
	}

	return nil
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

// getTodoByName retrieves and prints a single task by fuzzy name search.
func (tkm *TaskManager) GetTodoByName(name string) error {
	todo, err := GetTodoByName(tkm, name)	

	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTaskByName"), err)
	}

	s := fmt.Sprintf(i18n.T("tasks.tasks.getTaskByName"), todo.Name, todo.Description.String, todo.ExpiresAt.Format(time.RFC1123))
	color.MsgSuccess(s)
	return nil
}

// getAllDailyTodos lists every task marked as daily, with today's status.
func (tkm *TaskManager) GetAllDailyTodos() error {
	todos, err := tkm.DailyTodos(context.Background())
	if err != nil {
		return err
	}

	for _, todo := range todos {
		printTodoView(todo)
	}

	return nil
}

// DailyTodos returns every daily task with whether it was done today.
func (tkm *TaskManager) DailyTodos(ctx context.Context) ([]TodoView, error) {
	todos, err := tkm.listTodoViews(ctx, func(v TodoView) bool { return v.IsDaily })
	if err != nil {
		return nil, fmt.Errorf(i18n.T("tasks.error.getAllDailyTodos"), err)
	}

	return todos, nil
}

// deleteByName removes one task identified by name.
func (tkm *TaskManager) DeleteByName(ctx context.Context, name string) error {
	todo, err := GetTodoByName(tkm, name)
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTodoByName"), err)
	}

	err = tkm.Queries.DeleteTodoByName(ctx, todo.Name)
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.deleteTaskByName"), err)
	}

	color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.taskDeletedSuccessfully"), todo.Name))

	return nil
}
