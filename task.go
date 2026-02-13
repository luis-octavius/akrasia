package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
		return fmt.Errorf("Error creating the todo: %w", err)
	}

	log.Printf("Todo %v created successfully!\n", name)
	log.Printf("\n%s\n", generateRandomQuote())
	return nil
}

func (cfg *Config) getTodos() error {
	todos, err := cfg.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error getting todos from database: %w", err)
	}

	fmt.Println("Todos: ")

	for _, todo := range todos {
		printTodo(todo, "blue")
	}

	return nil
}

func (cfg *Config) getTodoByName(name string) error {
	todo, err := cfg.Queries.GetTodoByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("error getting todo from provided name: %w", err)
	}

	printTodo(todo, "blue")
	return nil
}

func (cfg *Config) updateToConcluded(name string) error {
	todo, err := cfg.Queries.UpdateTodoStatusByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("Error updating status of Todo '%v': %w", name, err)
	}

	printTodo(todo, "green")
	fmt.Printf("\n%s\n", generateRandomQuote())
	return nil
}

func (cfg *Config) deleteConcluded() error {
	err := cfg.Queries.DeleteConcluded(context.Background())
	if err != nil {
		return fmt.Errorf("Error deleting concluded todos: %w", err)
	}

	colorized, _ := color.ColorizeOutput("blue", SuccessDelete)
	fmt.Println(colorized)
	fmt.Printf("\n%s\n", generateRandomQuote())

	return nil
}

func (cfg *Config) checkExpired() error {
	todos, err := cfg.Queries.CheckExpired(context.Background())
	if err != nil {
		return fmt.Errorf("Error checking expired todos: %w", err)
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
		return fmt.Errorf("Error getting todos: %w", err)
	}

	var countExpiring int

	for _, todo := range todos {
		isTodoExpiring := checkIfTodoExpires(todo.ExpiresAt.Time)
		if isTodoExpiring {
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
		return fmt.Errorf("Error deleting todo: %w", err)
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
		return fmt.Errorf("error updating daily todos: %w", err)
	}

	// just for debugging
	for _, todo := range todos {
		printTodo(todo, "blue")
	}

	return nil
}
