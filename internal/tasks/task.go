package tasks

import (
	"context"
	"database/sql"
	"encoding/json"
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
	todos, err := tkm.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTaskDatabase"), err)
	}

	todos = filterTodosByPriority(todos, priorityFilter)

	fmt.Println(i18n.T("tasks.tasks.tasks"))

	for _, todo := range todos {
		printTodo(todo)
	}

	return nil
}

// getTodayFocus builds and prints the daily dashboard, or JSON when requested.
func (tkm *TaskManager) GetTodayFocus(opts TodayOptions) error {
	todos, err := tkm.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTaskDatabase"), err)
	}

	todos = filterTodosByPriority(todos, opts.Priority)
	doneToday, err := tkm.doneTodaySet()
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTaskDatabase"), err)
	}
	view := buildTodayView(todos, time.Now(), doneToday)
	view = applyTodayOptions(view, opts)

	if opts.JSON {
		payload, err := json.MarshalIndent(view, "", "  ")
		if err != nil {
			return fmt.Errorf(i18n.T("tasks.error.encodeTodayJSON"), err)
		}

		fmt.Println(string(payload))
		return nil
	}

	if len(view.Overdue) == 0 && len(view.DueToday) == 0 && len(view.Daily) == 0 && len(view.ExpiringSoon) == 0 {
		color.MsgSuccess(i18n.T("tasks.info.noPendingTasks"))
		return nil
	}

	fmt.Println(i18n.T("tasks.tasks.todayFocus"))

	printTodaySection(i18n.T("tasks.tasks.overdue"), view.Overdue)
	printTodaySection(i18n.T("tasks.tasks.dueToday"), view.DueToday)
	printTodaySection(i18n.T("tasks.tasks.dailyPending"), view.Daily)
	printTodaySection(i18n.T("tasks.tasks.expiringSoon"), view.ExpiringSoon)

	return nil
}

// getFocus returns the top actionable items across today sections.
func (tkm *TaskManager) GetFocus(limit int, priorityFilter string) error {
	todos, err := tkm.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTaskDatabase"), err)
	}

	todos = filterTodosByPriority(todos, priorityFilter)
	doneToday, err := tkm.doneTodaySet()
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTaskDatabase"), err)
	}
	view := buildTodayView(todos, time.Now(), doneToday)

	focus := make([]database.Todo, 0, limit)
	for _, section := range [][]database.Todo{view.Overdue, view.DueToday, view.Daily, view.ExpiringSoon} {
		for _, todo := range section {
			if len(focus) >= limit {
				break
			}
			focus = append(focus, todo)
		}
		if len(focus) >= limit {
			break
		}
	}

	if len(focus) == 0 {
		color.MsgSuccess(i18n.T("tasks.info.noFocusTasks"))
		return nil
	}

	fmt.Printf(i18n.T("tasks.tasks.focus"), len(focus))
	for _, todo := range focus {
		printTodo(todo)
	}

	return nil
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

// updateToConcluded marks a task as done and writes an entry to history.
func (tkm *TaskManager) UpdateToConcluded(name, notes string) error {
	todo, err := tkm.Queries.UpdateTodoStatusByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.updateTaskStatus"), name, err)
	}

	_, err = tkm.Queries.AddTodoHistory(context.Background(), database.AddTodoHistoryParams{
		ID:          uuid.New(),
		TodoID:      todo.ID,
		Completed:   sql.NullBool{Bool: true, Valid: true},
		CompletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Notes:       sql.NullString{String: notes, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("%s", i18n.T("tasks.error.createTaskHistory"))
	}

	fmt.Println(i18n.T("tasks.info.updatedSuccessfully"))
	generateRandomQuote()
	return nil
}

// deleteConcluded removes all tasks with concluded status.
func (tkm *TaskManager) DeleteConcluded() error {
	err := tkm.Queries.DeleteConcluded(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.deleteConcludedTasks"), err)
	}

	color.MsgSuccess(SuccessDelete)
	generateRandomQuote()

	return nil
}

// getAllDailyTodos lists every task marked as daily, with today's status.
func (tkm *TaskManager) GetAllDailyTodos() error {
	todos, err := tkm.Queries.GetDailyTodos(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getAllDailyTodos"), err)
	}

	doneToday, err := tkm.doneTodaySet()
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getAllDailyTodos"), err)
	}

	for _, todo := range todos {
		printDailyTodo(todo, doneToday[todoKey(todo.ID)])
	}

	return nil
}

// deleteByName removes one task identified by name.
func (tkm *TaskManager) DeleteByName(name string) error {
	todo, err := GetTodoByName(tkm, name)
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTodoByName"), err)
	}

	err = tkm.Queries.DeleteTodoByName(context.Background(), todo.Name)
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.deleteTaskByName"), err)
	}

	color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.taskDeletedSuccessfully"), todo.Name))

	return nil
}


