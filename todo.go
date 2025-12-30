package main

import (
	"context"
	"database/sql"
	"errors"
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

func (cfg *Config) addTodo(name, description string, expiresAt time.Time) error {
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
	})
	if err != nil {
		return fmt.Errorf("%v Error creating the todo: %v", emoji.Success, err)
	}

	log.Printf("%v Todo %v created successfully!\n", emoji.Success, name)
	return nil
}

func validateDescription(description string) sql.NullString {
	descriptionField := sql.NullString{}

	if description == "" {
		descriptionField.String = ""
		descriptionField.Valid = false
	} else {
		descriptionField.String = description
		descriptionField.Valid = true
	}

	return descriptionField
}

func validateTime(expiresAt time.Time) sql.NullTime {
	t := sql.NullTime{}

	if expiresAt.IsZero() {
		t.Valid = false
		return t
	}

	t.Time = expiresAt
	t.Valid = true

	return t
}

func (cfg *Config) getTodos() error {
	todos, err := cfg.Queries.GetTodos(context.Background())
	if err != nil {
		return fmt.Errorf("error getting todos from database: %v", err)
	}

	emojiText := emoji.AddEmoji(emoji.Bell, "Todos: ")
	fmt.Printf("%v\n", emojiText)

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
	return nil
}

func (cfg *Config) deleteConcluded() error {
	err := cfg.Queries.DeleteConcluded(context.Background())
	if err != nil {
		return fmt.Errorf("Error deleting concluded todos: %v", err)
	}

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

	fmt.Println(emoji.AddEmoji(emoji.StatusExpired, "EXPIRED: "))
	for _, todo := range todos {
		printTodo(todo, "red")
	}

	return nil
}

// printTodo receives a Todo and create a readable output
func printTodo(todo database.Todo, colorName string) {
	expiresAt := todo.ExpiresAt.Time.UTC()
	var status string

	if todo.Concluded == true {
		status = "Done"
	} else {
		status = "Not done"
	}

	s := fmt.Sprintf("%v %v | %v | %v | %v\n", emoji.Todo, todo.Name, todo.Description.String, expiresAt, status)

	colorized, _ := color.ColorizeOutput(colorName, s)

	io.WriteString(os.Stdout, colorized)
}
