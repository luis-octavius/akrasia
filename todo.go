package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/luis-octavius/akrasia/internal/database"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/emoji"
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
		return fmt.Errorf("%v Error creating the todo: %v", emoji.Success, err)
	}

	log.Printf("%v Todo %v created successfully!\n", emoji.Success, name)
	return nil
}

func (cfg *Config) getTodos() error {
	todos, err := cfg.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error getting todos from database: %v", err)
	}

	emojiText := emoji.AddEmoji(emoji.Bell, "Todos: ")
	fmt.Printf("%v\n", emojiText)

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

// printTodo receives a Todo and create a readable output
func printTodo(todo database.Todo) {
	expiresAt := todo.ExpiresAt.Time.UTC()
	var status string

	if todo.Concluded == true {
		status = "Done"
	} else {
		status = "Not done"
	}

	s := fmt.Sprintf("%v %v | %v | %v | %v\n", emoji.Todo, todo.Name, todo.Description.String, expiresAt, status)

	colorized, _ := color.ColorizeOutput("red", s)

	io.WriteString(os.Stdout, colorized)
}
