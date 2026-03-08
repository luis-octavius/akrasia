package main

import (
	"context"
	"database/sql"
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

func (cfg *Config) getTodos() error {
	todos, err := cfg.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error getting tasks from database: %w", err)
	}

	fmt.Println("Tasks: ")

	for _, todo := range todos {
		printTodo(todo)
	}

	return nil
}

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

func (cfg *Config) updateToConcluded(name, notes string) error {
	todo, err := cfg.Queries.UpdateTodoStatusByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("Error updating status of task '%v': %w", name, err)
	}

	_, err = cfg.Queries.AddTodoHistory(context.Background(), database.AddTodoHistoryParams{
		ID:          uuid.New(),
		TodoID:      todo.ID,
		Date:        sql.NullTime{Time: time.Now(), Valid: true},
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

func (cfg *Config) deleteConcluded() error {
	err := cfg.Queries.DeleteConcluded(context.Background())
	if err != nil {
		return fmt.Errorf("Error deleting concluded tasks: %w", err)
	}

	color.MsgSuccess(SuccessDelete)
	generateRandomQuote()

	return nil
}

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

func (cfg *Config) deleteByName(name string) error {
	err := cfg.Queries.DeleteTodoByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("Error deleting task: %w", err)
	}

	color.MsgSuccess(SuccessDelete)

	return nil
}

// updateDailyTodo execute the query to update the daily todos
// it handles the logic for the detached command
func (cfg *Config) updateDailyTodo() error {
	now := time.Now()
	fmt.Printf("Executing daily todo update at: %s\n", now.Format(time.DateTime))

	ctx := context.Background()

	// First, get all daily tasks BEFORE resetting them
	dailyTasks, err := cfg.Queries.GetDailyTodos(ctx)
	if err != nil {
		return fmt.Errorf("Error getting daily tasks: %w", err)
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
