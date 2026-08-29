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
	NoExpiring    = i18n.T("taskNoExpiring")
	SuccessDelete = i18n.T("taskSuccessDelete")
)

// addTodo persists a new task record and prints success feedback.
func (tkm *TaskManager) AddTodo(name, description, priority string, isDaily bool, expiresAt time.Time) error {
	descriptionField := validateDescription(description)

	_, err := tkm.Queries.AddTodo(context.Background(), database.AddTodoParams{
		ID:          uuid.New(),
		Name:        name,
		Description: descriptionField,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Concluded:   false,
		ExpiresAt:   expiresAt,
		Priority:    priority,
		IsDaily:     isDaily,
	})
	if err != nil {
		return fmt.Errorf(i18n.T("errorCreateTask"), err)
	}

	color.MsgSuccess(fmt.Sprintf(i18n.T("createdTask"), name))
	generateRandomQuote()
	return nil
}

// getTodos lists all tasks, optionally filtered by priority.
func (tkm *TaskManager) GetTodos(priorityFilter string) error {
	todos, err := tkm.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("errorGetTaskDatabase"), err)
	}

	todos = filterTodosByPriority(todos, priorityFilter)

	fmt.Println(i18n.T("tasks"))

	for _, todo := range todos {
		printTodo(todo)
	}

	return nil
}

// getTodayFocus builds and prints the daily dashboard, or JSON when requested.
func (tkm *TaskManager) GetTodayFocus(opts TodayOptions) error {
	todos, err := tkm.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("errorGetTaskDatabase"), err)
	}

	todos = filterTodosByPriority(todos, opts.Priority)
	view := buildTodayView(todos, time.Now())
	view = applyTodayOptions(view, opts)

	if opts.JSON {
		payload, err := json.MarshalIndent(view, "", "  ")
		if err != nil {
			return fmt.Errorf(i18n.T("errorEncodeTodayJSON"), err)
		}

		fmt.Println(string(payload))
		return nil
	}

	if len(view.Overdue) == 0 && len(view.DueToday) == 0 && len(view.Daily) == 0 && len(view.ExpiringSoon) == 0 {
		color.MsgSuccess(i18n.T("noPendingTasks"))
		return nil
	}

	fmt.Println(i18n.T("todayFocus"))

	printTodaySection(i18n.T("overdue"), view.Overdue)
	printTodaySection(i18n.T("dueToday"), view.DueToday)
	printTodaySection(i18n.T("dailyPending"), view.Daily)
	printTodaySection(i18n.T("expiringSoon"), view.ExpiringSoon)

	return nil
}

// getFocus returns the top actionable items across today sections.
func (tkm *TaskManager) GetFocus(limit int, priorityFilter string) error {
	todos, err := tkm.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("errorGetTaskDatabase"), err)
	}

	todos = filterTodosByPriority(todos, priorityFilter)
	view := buildTodayView(todos, time.Now())

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
		color.MsgSuccess(i18n.T("noFocusTasks"))
		return nil
	}

	fmt.Printf(i18n.T("focus"), len(focus))
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
		return fmt.Errorf(i18n.T("errorGetTaskByName"), err)
	}

	s := fmt.Sprintf(i18n.T("tableGetTaskByName"), todo.Name, todo.Description.String, todo.ExpiresAt.Format(time.RFC1123))
	color.MsgSuccess(s)
	return nil
}

// updateToConcluded marks a task as done and writes an entry to history.
func (tkm *TaskManager) UpdateToConcluded(name, notes string) error {
	todo, err := tkm.Queries.UpdateTodoStatusByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf(i18n.T("errorUpdateTaskStatus"), name, err)
	}

	_, err = tkm.Queries.AddTodoHistory(context.Background(), database.AddTodoHistoryParams{
		ID:          uuid.New(),
		TodoID:      todo.ID,
		Completed:   sql.NullBool{Bool: true, Valid: true},
		CompletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Notes:       sql.NullString{String: notes, Valid: true},
	})
	if err != nil {
		return fmt.Errorf(i18n.T("errorCreateTaskHistory"))
	}

	fmt.Println(i18n.T("updatedSuccessfully"))
	generateRandomQuote()
	return nil
}

// deleteConcluded removes all tasks with concluded status.
func (tkm *TaskManager) DeleteConcluded() error {
	err := tkm.Queries.DeleteConcluded(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("errorDeleteConcludedTasks"), err)
	}

	color.MsgSuccess(SuccessDelete)
	generateRandomQuote()

	return nil
}

// getAllDailyTodos lists every task marked as daily.
func (tkm *TaskManager) GetAllDailyTodos() error {
	todos, err := tkm.Queries.GetDailyTodos(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("errorGetAllDailyTodos"), err)
	}

	for _, todo := range todos {
		printTodo(todo)
	}

	return nil
}

// checkExpired prints expired non-daily tasks.
func (tkm *TaskManager) CheckExpired() error {
	todos, err := tkm.Queries.CheckExpired(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("errorCheckExpiredTasks"), err)
	}

	if len(todos) == 0 {
		fmt.Println(NoExpiring)
		return nil
	}

	fmt.Println(i18n.T("expired"))
	for _, todo := range todos {
		printTodo(todo)
	}

	return nil
}

// checkExpiring prints tasks that expire within the configured warning window.
func (tkm *TaskManager) CheckExpiring() error {
	todos, err := tkm.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf(i18n.T("errorCheckExpiringTasks"), err)
	}

	var countExpiring int

	for _, todo := range todos {
		isTodoExpiring := checkIfTodoExpires(todo.ExpiresAt)
		if isTodoExpiring && !todo.Concluded {
			fmt.Println(i18n.T("expiring"))
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
		return fmt.Errorf(i18n.T("errorGetTodoByName"), err)
	}

	err = tkm.Queries.DeleteTodoByName(context.Background(), todo.Name)
	if err != nil {
		return fmt.Errorf(i18n.T("errorDeleteTaskByName"), err)
	}

	color.MsgSuccess(fmt.Sprintf(i18n.T("taskDeletedSuccessfully"), todo.Name))

	return nil
}

// updateDailyTodo snapshots daily completion state, then resets daily tasks.
func (tkm *TaskManager) UpdateDailyTodo() error {
	now := time.Now()
	fmt.Printf(i18n.T("executeDailyTaskUpdate"), now.Format(time.DateTime))

	ctx := context.Background()

	// First, get all daily tasks BEFORE resetting them
	dailyTasks, err := tkm.Queries.GetDailyTodos(ctx)
	if err != nil {
		return fmt.Errorf(i18n.T("errorGetDailyTasks"), err)
	}

	if len(dailyTasks) == 0 {
		color.MsgError(i18n.T("noDailyTasksToUpdate"))
		return nil
	}

	// Record one history snapshot per task/day before resetting statuses.
	recordedCount := 0

	for _, task := range dailyTasks {
		// Skip if we already have a history entry for this date (prevents duplicates)
		// Create history entry: completed = task.Concluded (true if done, false if not)
		completedAt := sql.NullTime{}
		if task.Concluded {
			completedAt = sql.NullTime{Time: task.UpdatedAt, Valid: true}
		}

		_, err := tkm.Queries.AddDailyTaskHistory(ctx, database.AddDailyTaskHistoryParams{
			ID:          uuid.New(),
			TodoID:      task.ID,
			Completed:   sql.NullBool{Bool: task.Concluded, Valid: true},
			CompletedAt: completedAt,
			Notes:       sql.NullString{},
		})
		if err != nil {
			fmt.Printf(i18n.T("cantRecordTaskHistory"), task.Name, err)
			continue
		}
		recordedCount++
	}

	fmt.Printf(i18n.T("recordTaskHistorySuccessful"), recordedCount)

	// Now reset all daily tasks for the new day
	_, err = tkm.Queries.UpdateDailyTodo(ctx)
	if err != nil {
		return fmt.Errorf(i18n.T("errorUpdateDailyTasks"), err)
	}

	color.MsgSuccess(fmt.Sprintf(i18n.T("tasksUpdatedSuccessfully"), recordedCount))

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
		return fmt.Errorf(i18n.T("errorGetTodoByNameStreak"))
	}

	streak, err := tkm.Queries.GetCurrentStreak(context.Background(), todo.ID)
	if err != nil {
		return fmt.Errorf(i18n.T("errorGetCurrentStreak"))
	}

	fmt.Printf(i18n.T("currentStreak"), todo.Name, streak)
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
		return fmt.Errorf(i18n.T("errorGetTaskByName"))
	}

	streak_history, err := tkm.Queries.GetStreakHistory(context.Background(), todo.ID)
	if err != nil {
		return fmt.Errorf(i18n.T("errorGetStreakHistory"))
	}

	for i, streak := range streak_history {
		fmt.Printf(i18n.T("tableGetStreakHistory"), i+1, streak.StartDate, streak.EndDate, streak.StreakLength)
	}

	return nil
}

// backfillDailyHistory inserts missing daily history rows for a date interval.
func (tkm *TaskManager) BackfillDailyHistory(daysBack int, taskName string) error {
	ctx := context.Background()
	now := time.Now()
	startDate := now.AddDate(0, 0, -daysBack)

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
			return fmt.Errorf(i18n.T("errorFindTask"), taskName, err)
		}

		if !todo.IsDaily {
			return fmt.Errorf(i18n.T("errorTaskNotDaily"), taskName)
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
			return fmt.Errorf(i18n.T("errorGetDaily"), err)
		}
	}

	if len(dailyTasks) == 0 {
		color.MsgError(i18n.T("noDailyToBackfill"))
		return nil
	}

	totalBackfilled := 0

	for _, task := range dailyTasks {
		// Start from the later of: task creation date or (now - daysBack)
		backfillStart := task.CreatedAt
		if startDate.After(backfillStart) {
			backfillStart = startDate
		}

		// Create a history entry for each day from backfillStart to now
		currentDate := backfillStart
		for currentDate.Before(now) || currentDate.Equal(now) {
			dateOnly := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, now.Location())

			// Try to insert; ON CONFLICT DO NOTHING will silently skip if already exists
			_, err := tkm.Queries.AddDailyTaskHistoryForDate(ctx, database.AddDailyTaskHistoryForDateParams{
				ID:          uuid.New(),
				TodoID:      task.ID,
				Date:        dateOnly.Format(time.DateOnly),
				Completed:   sql.NullBool{Bool: false, Valid: true}, // conservative default: not completed
				CompletedAt: sql.NullTime{},
				Notes:       sql.NullString{String: "backfilled", Valid: true},
			})

			if err != nil {
				// Might be a conflict (already exists) or real error
				// For now, just log warnings and continue
				fmt.Printf(i18n.T("cantBackfillTask"), task.Name, dateOnly.Format(time.DateOnly), err)
			} else {
				totalBackfilled++
			}

			currentDate = currentDate.AddDate(0, 0, 1)
		}
	}

	if taskName != "" {
		color.MsgSuccess(fmt.Sprintf(i18n.T("backfilledTask"), totalBackfilled, taskName))
	} else {
		color.MsgSuccess(fmt.Sprintf(i18n.T("backfilledDailyTask"), totalBackfilled, len(dailyTasks)))
	}

	return nil
}
