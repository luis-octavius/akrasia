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
)

const (
	NoExpiring    = "Your tasks are not fleeing. You have time, yet your focus must remain steadfast."
	SuccessDelete = "Concluded Todos deleted successfully!"
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
		return fmt.Errorf("Error creating task: %w", err)
	}

	color.MsgSuccess(fmt.Sprintf("Task %v created successfully!\n", name))
	generateRandomQuote()
	return nil
}

// getTodos lists all tasks, optionally filtered by priority.
func (tkm *TaskManager) GetTodos(priorityFilter string) error {
	todos, err := tkm.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error getting tasks from database: %w", err)
	}

	todos = filterTodosByPriority(todos, priorityFilter)

	fmt.Println("Tasks: ")

	for _, todo := range todos {
		printTodo(todo)
	}

	return nil
}

// getTodayFocus builds and prints the daily dashboard, or JSON when requested.
func (tkm *TaskManager) GetTodayFocus(opts TodayOptions) error {
	todos, err := tkm.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error getting tasks from database: %w", err)
	}

	todos = filterTodosByPriority(todos, opts.Priority)
	view := buildTodayView(todos, time.Now())
	view = applyTodayOptions(view, opts)

	if opts.JSON {
		payload, err := json.MarshalIndent(view, "", "  ")
		if err != nil {
			return fmt.Errorf("error encoding today output to json: %w", err)
		}

		fmt.Println(string(payload))
		return nil
	}

	if len(view.Overdue) == 0 && len(view.DueToday) == 0 && len(view.Daily) == 0 && len(view.ExpiringSoon) == 0 {
		color.MsgSuccess("You are clear for today. No pending items need attention.")
		return nil
	}

	fmt.Println("TODAY FOCUS")

	printTodaySection("OVERDUE", view.Overdue)
	printTodaySection("DUE TODAY", view.DueToday)
	printTodaySection("DAILY PENDING", view.Daily)
	printTodaySection("EXPIRING SOON", view.ExpiringSoon)

	return nil
}

// getFocus returns the top actionable items across today sections.
func (tkm *TaskManager) GetFocus(limit int, priorityFilter string) error {
	todos, err := tkm.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error getting tasks from database: %w", err)
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
		color.MsgSuccess("No focus items for now. You are clear.")
		return nil
	}

	fmt.Printf("FOCUS (%d)\n", len(focus))
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
		return fmt.Errorf("error getting task from provided name: %w", err)
	}

	s := fmt.Sprintf("Name: %s | Description: %s\nExpires: %v\n", todo.Name, todo.Description.String, todo.ExpiresAt.Format(time.RFC1123))
	color.MsgSuccess(s)
	return nil
}

// updateToConcluded marks a task as done and writes an entry to history.
func (tkm *TaskManager) UpdateToConcluded(name, notes string) error {
	todo, err := tkm.Queries.UpdateTodoStatusByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("Error updating status of task '%v': %w", name, err)
	}

	_, err = tkm.Queries.AddTodoHistory(context.Background(), database.AddTodoHistoryParams{
		ID:          uuid.New(),
		TodoID:      todo.ID,
		Completed:   sql.NullBool{Bool: true, Valid: true},
		CompletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Notes:       sql.NullString{String: notes, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("Error creating todo history")
	}

	fmt.Println("Updated successfully")
	generateRandomQuote()
	return nil
}

// deleteConcluded removes all tasks with concluded status.
func (tkm *TaskManager) DeleteConcluded() error {
	err := tkm.Queries.DeleteConcluded(context.Background())
	if err != nil {
		return fmt.Errorf("Error deleting concluded tasks: %w", err)
	}

	color.MsgSuccess(SuccessDelete)
	generateRandomQuote()

	return nil
}

// getAllDailyTodos lists every task marked as daily.
func (tkm *TaskManager) GetAllDailyTodos() error {
	todos, err := tkm.Queries.GetDailyTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error in getAllDailyTodos: %v", err)
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
		return fmt.Errorf("Error checking expired tasks: %w", err)
	}

	if len(todos) == 0 {
		fmt.Println(NoExpiring)
		return nil
	}

	fmt.Println("EXPIRED: ")
	for _, todo := range todos {
		printTodo(todo)
	}

	return nil
}

// checkExpiring prints tasks that expire within the configured warning window.
func (tkm *TaskManager) CheckExpiring() error {
	todos, err := tkm.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting tasks: %w", err)
	}

	var countExpiring int

	for _, todo := range todos {
		isTodoExpiring := checkIfTodoExpires(todo.ExpiresAt)
		if isTodoExpiring && !todo.Concluded {
			fmt.Println("Expiring...")
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
		return fmt.Errorf("error getting todo by name: %w", err)
	}

	err = tkm.Queries.DeleteTodoByName(context.Background(), todo.Name)
	if err != nil {
		return fmt.Errorf("error deleting todo by name: %w", err)
	}

	color.MsgSuccess(fmt.Sprintf("Todo %s deleted successfully", todo.Name))

	return nil
}

// updateDailyTodo snapshots daily completion state, then resets daily tasks.
func (tkm *TaskManager) UpdateDailyTodo() error {
	now := time.Now()
	fmt.Printf("Executing daily todo update at: %s\n", now.Format(time.DateTime))

	ctx := context.Background()

	// First, get all daily tasks BEFORE resetting them
	dailyTasks, err := tkm.Queries.GetDailyTodos(ctx)
	if err != nil {
		return fmt.Errorf("error getting daily tasks: %w", err)
	}

	if len(dailyTasks) == 0 {
		color.MsgError("No daily tasks to be updated")
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
			fmt.Printf("Warning: Could not record history for task '%s': %v\n", task.Name, err)
			continue
		}
		recordedCount++
	}

	fmt.Printf("Recorded history for %d daily tasks\n", recordedCount)

	// Now reset all daily tasks for the new day
	_, err = tkm.Queries.UpdateDailyTodo(ctx)
	if err != nil {
		return fmt.Errorf("Error updating daily tasks: %w", err)
	}

	color.MsgSuccess(fmt.Sprintf("Tasks updated successfully (%d history entries recorded)", recordedCount))

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
		return fmt.Errorf("Error getting todo by name")
	}

	streak, err := tkm.Queries.GetCurrentStreak(context.Background(), todo.ID)
	if err != nil {
		return fmt.Errorf("Error getting current todo streak")
	}

	fmt.Printf("\nYour current streak with %s is: %d\n", todo.Name, streak)
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
		return fmt.Errorf("Error getting todo by name")
	}

	streak_history, err := tkm.Queries.GetStreakHistory(context.Background(), todo.ID)
	if err != nil {
		return fmt.Errorf("Error getting todo streak history")
	}

	for i, streak := range streak_history {
		fmt.Printf("\n%d. Started Date: %v | End Date: %v | Total Days: %d\n", i+1, streak.StartDate, streak.EndDate, streak.StreakLength)
	}

	return nil
}

// backfillDailyHistory inserts missing daily history rows for a date interval.
// It excludes today (handled by update-daily) and marks entries with notes="backfilled".
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
			return fmt.Errorf("Error finding task '%s': %w", taskName, err)
		}

		if !todo.IsDaily {
			return fmt.Errorf("Task '%s' is not a daily task", taskName)
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
			return fmt.Errorf("Error getting daily tasks: %w", err)
		}
	}

	if len(dailyTasks) == 0 {
		color.MsgError("No daily tasks found to backfill")
		return nil
	}

	totalBackfilled := 0
	totalSkipped := 0

	for _, task := range dailyTasks {
		// Start from the later of: task creation date or startDate
		backfillStart := task.CreatedAt
		if startDate.After(backfillStart) {
			backfillStart = startDate
		}

		fmt.Printf("  %s: ", task.Name)

		taskInserted := 0
		taskSkipped := 0

		// Create a history entry for each day from backfillStart to yesterday (exclude today)
		currentDate := backfillStart
		dayCount := 0
		for currentDate.Before(today) {
			dateOnly := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, now.Location())
			dateStr := dateOnly.Format(time.DateOnly)

			// Try to insert; ON CONFLICT DO NOTHING silently skips duplicates
			res, err := tkm.Queries.AddDailyTaskHistoryForDate(ctx, database.AddDailyTaskHistoryForDateParams{
				ID:          uuid.New(),
				TodoID:      task.ID,
				Date:        dateStr,
				Completed:   sql.NullBool{Bool: false, Valid: true},
				CompletedAt: sql.NullTime{},
				Notes:       sql.NullString{String: "backfilled", Valid: true},
			})

			if err != nil {
				fmt.Printf("\n  Error backfilling %s for %s: %v\n", task.Name, dateStr, err)
			} else {
				affected, _ := res.RowsAffected()
				if affected > 0 {
					taskInserted++
				} else {
					taskSkipped++
				}
			}

			dayCount++
			if dayCount%10 == 0 {
				fmt.Print(".")
			}

			currentDate = currentDate.AddDate(0, 0, 1)
		}

		totalBackfilled += taskInserted
		totalSkipped += taskSkipped

		if taskInserted > 0 || taskSkipped > 0 {
			fmt.Printf(" %d inserted, %d already existed\n", taskInserted, taskSkipped)
		} else {
			fmt.Println(" no days to backfill")
		}
	}

	if taskName != "" {
		color.MsgSuccess(fmt.Sprintf("Backfilled %d history entries for task '%s'", totalBackfilled, taskName))
	} else {
		color.MsgSuccess(fmt.Sprintf("Backfilled %d history entries across %d daily task(s) (%d already existed)", totalBackfilled, len(dailyTasks), totalSkipped))
	}

	return nil
}
