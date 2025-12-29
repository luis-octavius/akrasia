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
		fmt.Printf("ID: %v\nName: %v\nDescription: %v\nDate Added: %v\n", todo.ID, todo.Name, todo.Description, todo.CreatedAt)
	}

	return nil
}
