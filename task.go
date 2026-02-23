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
	expiresField := validateTime(expiresAt)

	_, err := cfg.Queries.AddTodo(context.Background(), database.AddTodoParams{
		ID:          uuid.New(),
		Name:        name,
		Description: descriptionField,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Concluded:   false,
		ExpiresAt:   expiresField,
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
	todo, err := cfg.Queries.GetTodoByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("error getting task from provided name: %w", err)
	}

	printTodo(todo, "blue")
	return nil
}

func (cfg *Config) updateToConcluded(name string) error {
	todo, err := cfg.Queries.UpdateTodoStatusByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("Error updating status of task '%v': %w", name, err)
	}

	printTodo(todo, "green")
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
		isTodoExpiring := checkIfTodoExpires(todo.ExpiresAt.Time)
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

	expiresAt := sql.NullTime{
		Time:  now.Add(24 * time.Hour),
		Valid: true,
	}

	fmt.Printf("Setting new expires_at to: %s\n", expiresAt.Time.Format(time.DateTime))

	todos, err := cfg.Queries.UpdateDailyTodo(context.Background(), expiresAt)
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
		fmt.Printf(" Old expires: %v -> New expires: %v\n", todo.ExpiresAt, expiresAt.Time)
		fmt.Printf(" Concluded: %v -> false\n\n", todo.Concluded)
	}

	return nil
}
