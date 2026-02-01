package main

import (
	"context"
	"errors"
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
		return fmt.Errorf("Error creating the todo: %v", err)
	}

	log.Printf("Todo %v created successfully!\n", name)
	log.Printf("\n%s", generateRandomQuote())
	return nil
}

func (cfg *Config) getTodos() error {
	todos, err := cfg.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error getting todos from database: %v", err)
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
		return fmt.Errorf("error getting todo from provided name: %v", err)
	}

	printTodo(todo, "blue")
	return nil
}

func (cfg *Config) updateToConcluded(name string) error {
	todo, err := cfg.Queries.UpdateTodoStatusByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("Error updating status of Todo '%v': %v", name, err)
	}

	printTodo(todo, "green")
	fmt.Printf("%s", generateRandomQuote())
	return nil
}

func (cfg *Config) deleteConcluded() error {
	err := cfg.Queries.DeleteConcluded(context.Background())
	if err != nil {
		return fmt.Errorf("Error deleting concluded todos: %v", err)
	}

	colorized, _ := color.ColorizeOutput("blue", SuccessDelete)
	fmt.Println(colorized)
	fmt.Printf("\n%s\n", generateRandomQuote())

	return nil
}

func (cfg *Config) checkExpired() error {
	todos, err := cfg.Queries.CheckExpired(context.Background())
	if err != nil {
		return fmt.Errorf("Error checking expired todos: %v", err)
	}

	if len(todos) == 0 {
		return errors.New("There are not expired todos")
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
		return fmt.Errorf("Error getting todos: %v", err)
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
		return fmt.Errorf("Error deleting todo: %v", err)
	}

	fmt.Println(SuccessDelete)

	return nil
}
