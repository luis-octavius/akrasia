package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/luis-octavius/akrasia/internal/database"
	"github.com/luis-octavius/akrasia/pkg/color"
)

const (
	NoExpiring    = "Your tasks are not fleeing. You have time, yet your focus must remain steadfast."
	SuccessDelete = "Concluded Todos deleted successfully!"
)

// todayView groups tasks into sections used by today/focus outputs.
type todayView struct {
	Overdue      []database.Todo
	DueToday     []database.Todo
	Daily        []database.Todo
	ExpiringSoon []database.Todo
}

// todayOptions configures section filtering, limits, and output format.
type todayOptions struct {
	Only     string
	Limit    int
	JSON     bool
	Priority string
}

// addTodo persists a new task record and prints success feedback.
func (cfg *Config) addTodo(name, description, priority string, isDaily bool, expiresAt time.Time) error {
	descriptionField := validateDescription(description)

	_, err := cfg.Queries.AddTodo(context.Background(), database.AddTodoParams{
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
func (cfg *Config) getTodos(priorityFilter string) error {
	todos, err := cfg.Queries.GetTodos(context.Background())
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
func (cfg *Config) getTodayFocus(opts todayOptions) error {
	todos, err := cfg.Queries.GetTodos(context.Background())
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
func (cfg *Config) getFocus(limit int, priorityFilter string) error {
	todos, err := cfg.Queries.GetTodos(context.Background())
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
func (cfg *Config) getTodoByName(name string) error {
	todo, err := cfg.Queries.GetTodoByName(context.Background(), database.GetTodoByNameParams{
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
func (cfg *Config) updateToConcluded(name, notes string) error {
	todo, err := cfg.Queries.UpdateTodoStatusByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("Error updating status of task '%v': %w", name, err)
	}

	_, err = cfg.Queries.AddTodoHistory(context.Background(), database.AddTodoHistoryParams{
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
func (cfg *Config) deleteConcluded() error {
	err := cfg.Queries.DeleteConcluded(context.Background())
	if err != nil {
		return fmt.Errorf("Error deleting concluded tasks: %w", err)
	}

	color.MsgSuccess(SuccessDelete)
	generateRandomQuote()

	return nil
}

// getAllDailyTodos lists every task marked as daily.
func (cfg *Config) getAllDailyTodos() error {
	todos, err := cfg.Queries.GetDailyTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error in getAllDailyTodos: %v", err)
	}

	for _, todo := range todos {
		printTodo(todo)
	}

	return nil
}

// checkExpired prints expired non-daily tasks.
func (cfg *Config) checkExpired() error {
	todos, err := cfg.Queries.CheckExpired(context.Background())
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
func (cfg *Config) checkExpiring() error {
	todos, err := cfg.Queries.GetTodos(context.Background())
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
func (cfg *Config) deleteByName(name string) error {
	todo, err := cfg.Queries.GetTodoByName(context.Background(), database.GetTodoByNameParams{
		LOWER:   name,
		LOWER_2: name,
		LOWER_3: name,
		LOWER_4: name,
	})
	if err != nil {
		return fmt.Errorf("error getting todo by name: %w", err)
	}

	err = cfg.Queries.DeleteTodoByName(context.Background(), todo.Name)
	if err != nil {
		return fmt.Errorf("error deleting todo by name: %w", err)
	}

	color.MsgSuccess(fmt.Sprintf("Todo %s deleted successfully", todo.Name))

	return nil
}

// updateDailyTodo snapshots daily completion state, then resets daily tasks.
func (cfg *Config) updateDailyTodo() error {
	now := time.Now()
	fmt.Printf("Executing daily todo update at: %s\n", now.Format(time.DateTime))

	ctx := context.Background()

	// First, get all daily tasks BEFORE resetting them
	dailyTasks, err := cfg.Queries.GetDailyTodos(ctx)
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

		_, err := cfg.Queries.AddDailyTaskHistory(ctx, database.AddDailyTaskHistoryParams{
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
	_, err = cfg.Queries.UpdateDailyTodo(ctx)
	if err != nil {
		return fmt.Errorf("Error updating daily tasks: %w", err)
	}

	color.MsgSuccess(fmt.Sprintf("Tasks updated successfully (%d history entries recorded)", recordedCount))

	return nil
}

// getCurrentStreak prints the current streak count for a named task.
func (cfg *Config) getCurrentStreak(name string) error {
	todo, err := cfg.Queries.GetTodoByName(context.Background(), database.GetTodoByNameParams{
		LOWER:   name,
		LOWER_2: name,
		LOWER_3: name,
		LOWER_4: name,
	})
	if err != nil {
		return fmt.Errorf("Error getting todo by name")
	}

	streak, err := cfg.Queries.GetCurrentStreak(context.Background(), todo.ID)
	if err != nil {
		return fmt.Errorf("Error getting current todo streak")
	}

	fmt.Printf("\nYour current streak with %s is: %d\n", todo.Name, streak)
	return nil
}

// getStreakHistory prints historical streak intervals for a named task.
func (cfg *Config) getStreakHistory(name string) error {
	todo, err := cfg.Queries.GetTodoByName(context.Background(), database.GetTodoByNameParams{
		LOWER:   name,
		LOWER_2: name,
		LOWER_3: name,
		LOWER_4: name,
	})
	if err != nil {
		return fmt.Errorf("Error getting todo by name")
	}

	streak_history, err := cfg.Queries.GetStreakHistory(context.Background(), todo.ID)
	if err != nil {
		return fmt.Errorf("Error getting todo streak history")
	}

	for i, streak := range streak_history {
		fmt.Printf("\n%d. Started Date: %v | End Date: %v | Total Days: %d\n", i+1, streak.StartDate, streak.EndDate, streak.StreakLength)
	}

	return nil
}

// backfillDailyHistory inserts missing daily history rows for a date interval.
func (cfg *Config) backfillDailyHistory(daysBack int, taskName string) error {
	ctx := context.Background()
	now := time.Now()
	startDate := now.AddDate(0, 0, -daysBack)

	// Get all daily tasks, or a specific one if taskName is provided
	var dailyTasks []database.Todo
	var err error

	if taskName != "" {
		// Get specific task by name
		todo, err := cfg.Queries.GetTodoByName(ctx, database.GetTodoByNameParams{
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
		dailyTasks, err = cfg.Queries.GetDailyTodos(ctx)
		if err != nil {
			return fmt.Errorf("Error getting daily tasks: %w", err)
		}
	}

	if len(dailyTasks) == 0 {
		color.MsgError("No daily tasks found to backfill")
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
			_, err := cfg.Queries.AddDailyTaskHistoryForDate(ctx, database.AddDailyTaskHistoryForDateParams{
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
				fmt.Printf("Warning: Could not backfill %s for date %s: %v\n", task.Name, dateOnly.Format(time.DateOnly), err)
			} else {
				totalBackfilled++
			}

			currentDate = currentDate.AddDate(0, 0, 1)
		}
	}

	if taskName != "" {
		color.MsgSuccess(fmt.Sprintf("Backfilled %d history entries for task '%s'", totalBackfilled, taskName))
	} else {
		color.MsgSuccess(fmt.Sprintf("Backfilled %d history entries for %d daily task(s)", totalBackfilled, len(dailyTasks)))
	}

	return nil
}
