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

func (cfg *Config) checkExpired() error {
	todos, err := cfg.Queries.CheckExpired(context.Background())
	if err != nil {
		return fmt.Errorf("Error checking expired tasks: %w", err)
	}

	if len(todos) == 0 {
		return fmt.Errorf(NoExpiring)
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
	fmt.Println("executing update for daily todos: ", time.Now().Format(time.DateTime))

	expiresAt := sql.NullTime{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	todos, err := cfg.Queries.UpdateDailyTodo(context.Background(), expiresAt)
	if err != nil {
		return fmt.Errorf("error updating daily tasks: %w", err)
	}

	if len(todos) == 0 {
		fmt.Println("no tasks to be updated")
	} else {
		fmt.Println("tasks updated successfully!")
	}

	// just for debugging
	for _, todo := range todos {
		printTodo(todo, "blue")
	}

	return nil
}
