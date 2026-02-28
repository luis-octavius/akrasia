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

	fmt.Printf("Task %v created successfully!\n", name)
	fmt.Printf("\n%s\n", generateRandomQuote())
	return nil
}

func (cfg *Config) getTodos() error {
	todos, err := cfg.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error getting tasks from database: %w", err)
	}

	fmt.Println("Tasks: ")

	for _, todo := range todos {
		printTodo(todo, "blue")
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

	fmt.Printf("Name: %s | Description: %s\nExpires: %v\n", todo.Name, todo.Description.String, todo.ExpiresAt.Format(time.RFC1123))

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
	fmt.Printf("\n%s\n", generateRandomQuote())
	return nil
}

func (cfg *Config) deleteConcluded() error {
	err := cfg.Queries.DeleteConcluded(context.Background())
	if err != nil {
		return fmt.Errorf("Error deleting concluded tasks: %w", err)
	}

	colorized, _ := color.ColorizeOutput("blue", SuccessDelete)
	fmt.Println(colorized)
	fmt.Printf("\n%s\n", generateRandomQuote())

	return nil
}

func (cfg *Config) getAllDailyTodos() error {
	todos, err := cfg.Queries.GetDailyTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error in getAllDailyTodos: %v", err)
	}

	for _, todo := range todos {
		printTodo(todo, "blue")
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
		printTodo(todo, "red")
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
			printTodo(todo, "red")
			countExpiring++
		}
	}

	if countExpiring == 0 {
		colored, _ := color.ColorizeOutput("red", NoExpiring)
		fmt.Printf("%s\n\n", colored)
	}

	return nil
}

func (cfg *Config) deleteByName(name string) error {
	err := cfg.Queries.DeleteTodoByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("Error deleting task: %w", err)
	}

	fmt.Println(SuccessDelete)

	return nil
}

// updateDailyTodo execute the query to update the daily todos
// it handles the logic for the detached command
func (cfg *Config) updateDailyTodo() error {
	now := time.Now()
	fmt.Printf("Executing daily todo update at: %s\n", now.Format(time.DateTime))

	todos, err := cfg.Queries.UpdateDailyTodo(context.Background(), time.Now().Add(24*time.Hour))
	if err != nil {
		return fmt.Errorf("Error updating daily tasks: %w", err)
	}

	message := ""

	if len(todos) == 0 {
		message, _ = color.ColorizeOutput("red", "No daily tasks to be updated")
	} else {
		message, _ = color.ColorizeOutput("blue", "Tasks updated successfully!")
	}

	fmt.Println(message)

	// just for debugging
	for i, todo := range todos {
		fmt.Printf(" %d. Task #%d: %s\n", i+1, todo.ID, todo.Name)
		fmt.Printf(" Old expires: %v -> New expires: \n", todo.ExpiresAt)
		fmt.Printf(" Concluded: %v -> false\n\n", todo.Concluded)
	}

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
