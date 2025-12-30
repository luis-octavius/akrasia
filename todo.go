package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/luis-octavius/akrasia/internal/database"
)

func (cfg *Config) addTodo(name, description string) error {
	descriptionField := sql.NullString{}

	if description == "" {
		descriptionField.String = ""
		descriptionField.Valid = false
	} else {
		descriptionField.String = description
		descriptionField.Valid = true
	}

	_, err := cfg.Queries.AddTodo(context.Background(), database.AddTodoParams{
		ID:          uuid.New(),
		Name:        name,
		Description: descriptionField,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Concluded:   false,
	})
	if err != nil {
		return fmt.Errorf("error creating the todo: %v", err)
	}

	log.Printf("Todo %v created successfully!\n", name)
	return nil
}

func (cfg *Config) getTodos() error {
	todos, err := cfg.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error getting todos from database: %v", err)
	}

	for _, todo := range todos {
		printTodo(todo)
	}

	return nil
}

func (cfg *Config) getTodoByName(name string) error {
	todo, err := cfg.Queries.GetTodoByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("error getting todo from provided name: %v", err)
	}

	printTodo(todo)
	return nil
}

func (cfg *Config) updateToConcluded(name string) error {
	todo, err := cfg.Queries.UpdateTodoStatusByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("Error updating status of Todo '%v': %v", name, err)
	}

	printTodo(todo)
	return nil
}

func (cfg *Config) deleteConcluded() error {
	err := cfg.Queries.DeleteConcluded(context.Background())
	if err != nil {
		return fmt.Errorf("Error deleting concluded todos: %v", err)
	}

	return nil
}

func printTodo(todo database.Todo) {
	expiresAt := todo.ExpiresAt.Time.UTC()
	var status string

	if todo.Concluded == true {
		status = "Done"
	} else {
		status = "Not done"
	}
	fmt.Printf("\nID: %v\nName: %v\nDescription: %v\nExpires At: %v\nStatus: %v\n", todo.ID, todo.Name, todo.Description.String, expiresAt, status)
}
