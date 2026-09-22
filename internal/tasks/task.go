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
	todo, err := tkm.Queries.GetTodoByName(context.Background(), database.GetTodoByNameParams{
		LOWER:   name,
		LOWER_2: name,
		LOWER_3: name,
		LOWER_4: name,
	})
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
		return fmt.Errorf(i18n.T("tasks.error.createTaskHistory"))
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

// checkExpired prints expired non-daily tasks.
func (tkm *TaskManager) CheckExpired() error {
	todos, err := tkm.Queries.CheckExpired(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.checkExpiredTasks"), err)
	}

	if len(todos) == 0 {
		fmt.Println(NoExpiring)
		return nil
	}

	fmt.Println(i18n.T("tasks.tasks.expired"))
	for _, todo := range todos {
		printTodo(todo)
	}

	return nil
}

// checkExpiring prints tasks that expire within the configured warning window.
func (tkm *TaskManager) CheckExpiring() error {
	todos, err := tkm.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.checkExpiringTasks"), err)
	}

	var countExpiring int

	for _, todo := range todos {
		isTodoExpiring := checkIfTodoExpires(todo.ExpiresAt)
		if isTodoExpiring && !todo.Concluded {
			fmt.Println(i18n.T("tasks.tasks.expiring"))
			printTodo(todo)
			countExpiring++
		}
	}

	if countExpiring == 0 {
		color.MsgError(NoExpiring)
	}

	return nil
}

// deleteByName removes one task identified by name.
func (tkm *TaskManager) DeleteByName(name string) error {
	todo, err := tkm.Queries.GetTodoByName(context.Background(), database.GetTodoByNameParams{
		LOWER:   name,
		LOWER_2: name,
		LOWER_3: name,
		LOWER_4: name,
	})
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

// getCurrentStreak prints the current streak count for a named task.
func (tkm *TaskManager) GetCurrentStreak(name string) error {
	todo, err := tkm.Queries.GetTodoByName(context.Background(), database.GetTodoByNameParams{
		LOWER:   name,
		LOWER_2: name,
		LOWER_3: name,
		LOWER_4: name,
	})
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTodoByNameStreak"))
	}

	streak, err := tkm.Queries.GetCurrentStreak(context.Background(), database.GetCurrentStreakParams{
		TodoID:   todo.ID,
		TodoID_2: todo.ID,
	})
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getCurrentStreak"))
	}

	fmt.Printf(i18n.T("tasks.info.currentStreak"), todo.Name, streak)
	return nil
}

// getStreakHistory prints historical streak intervals for a named task.
func (tkm *TaskManager) GetStreakHistory(name string) error {
	todo, err := tkm.Queries.GetTodoByName(context.Background(), database.GetTodoByNameParams{
		LOWER:   name,
		LOWER_2: name,
		LOWER_3: name,
		LOWER_4: name,
	})
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTaskByName"))
	}

	streak_history, err := tkm.Queries.GetStreakHistory(context.Background(), database.GetStreakHistoryParams{
		TodoID:   todo.ID,
		TodoID_2: todo.ID,
	})
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getStreakHistory"))
	}

	for i, streak := range streak_history {
		fmt.Printf(i18n.T("tasks.tasks.getStreakHistory"), i+1, streak.StartDate, streak.EndDate, streak.StreakLength)
	}

	return nil
}

// backfillDailyHistory records real completions for days the user forgot to
// log. It excludes today — use `done` for that — and tags each row with
// notes="backfilled" purely as a note to the user, not a semantic marker
// the streak queries treat specially.
func (tkm *TaskManager) BackfillDailyHistory(daysBack int, taskName string) error {
	ctx := context.Background()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startDate := today.AddDate(0, 0, -daysBack)

	// Get all daily tasks, or a specific one if taskName is provided
	var dailyTasks []database.Todo
	var err error

	if taskName != "" {
		// Get specific task by name
		todo, err := tkm.Queries.GetTodoByName(ctx, database.GetTodoByNameParams{
			LOWER:   taskName,
			LOWER_2: taskName,
			LOWER_3: taskName,
			LOWER_4: taskName,
		})
		if err != nil {
			return fmt.Errorf(i18n.T("tasks.error.findTask"), taskName, err)
		}

		if !todo.IsDaily {
			return fmt.Errorf(i18n.T("tasks.error.taskNotDaily"), taskName)
		}

		dailyTasks = []database.Todo{
			{
				ID:          todo.ID,
				Name:        todo.Name,
				Description: todo.Description,
				CreatedAt:   todo.CreatedAt,
				UpdatedAt:   todo.UpdatedAt,
				Concluded:   todo.Concluded,
				ExpiresAt:   todo.ExpiresAt,
				Priority:    todo.Priority,
				IsDaily:     todo.IsDaily,
			},
		}
	} else {
		// Get all daily tasks
		dailyTasks, err = tkm.Queries.GetDailyTodos(ctx)
		if err != nil {
			return fmt.Errorf(i18n.T("tasks.error.getDaily"), err)
		}
	}

	if len(dailyTasks) == 0 {
		color.MsgError(i18n.T("tasks.info.noDailyToBackfill"))
		return nil
	}

	totalBackfilled := 0

	for _, task := range dailyTasks {
		// Start from the later of: task creation date or startDate
		backfillStart := task.CreatedAt
		if startDate.After(backfillStart) {
			backfillStart = startDate
		}

		fmt.Printf("  %s: ", task.Name)

		taskInserted := 0

		// Insert a real completion for each day from backfillStart to
		// yesterday (exclude today — that's what `done` is for). There is
		// no "neutral" state anymore: dates before task.history_since are
		// simply never looked at by the streak queries, so nothing here
		// can accidentally overwrite a genuine miss.
		currentDate := backfillStart
		dayCount := 0
		for currentDate.Before(today) {
			dateOnly := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, now.Location())
			dateStr := dateOnly.Format(time.DateOnly)

			_, err := tkm.Queries.BackfillDailyDone(ctx, database.BackfillDailyDoneParams{
				ID:     uuid.New(),
				TodoID: task.ID,
				Date:   dateStr,
			})

			if err != nil {
				fmt.Printf(i18n.T("tasks.info.cantBackfillTask"), task.Name, dateOnly.Format(time.DateOnly), err)
			} else {
				taskInserted++
			}

			dayCount++
			if dayCount%10 == 0 {
				fmt.Print(".")
			}

			currentDate = currentDate.AddDate(0, 0, 1)
		}

		totalBackfilled += taskInserted

		if taskInserted > 0 {
			fmt.Printf(" %d days marked as completed\n", taskInserted)
		} else {
			fmt.Println(" no days to backfill")
		}
	}

	if taskName != "" {
		color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.backfilledTask"), totalBackfilled, taskName))
	} else {
		color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.backfilledDailyTask"), totalBackfilled, len(dailyTasks)))
	}

	return nil
}
