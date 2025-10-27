package main

import (
	"fmt"
	"time"
)

type Todo struct {
	id          int
	description string
	createdAt   time.Time
	Status      bool
}

type TodoManager struct {
	Todos map[int]Todo
	Id    int
}

func CreateTodoManager() *TodoManager {
	todoManager := TodoManager{
		Todos: map[int]Todo{},
		Id:    1,
	}
	return &todoManager
}

func (t *TodoManager) Add(description string) error {
	newTodo := Todo{
		id:          t.Id,
		description: description,
		createdAt:   time.Now(),
		Status:      false,
	}

	t.Todos[t.Id] = newTodo
	t.Id++
	fmt.Println("Id: ", t.Id)
	fmt.Println("Todo added succesfully!")
	return nil
}

func (t *TodoManager) Get(id int) (Todo, error) {
	todo, ok := t.Todos[id]
	if !ok {
		return Todo{}, fmt.Errorf("Todo with the ID provided doesn't exist")
	}

	return todo, nil
}

func (t *TodoManager) GetAll() error {
	for _, todo := range t.Todos {
		fmt.Printf("%v. %v\n%v\n", todo.id, todo.description, todo.createdAt.Format(time.RFC1123))
	}

	return nil
}

func (t *TodoManager) Delete(id int) error {
	_, err := t.Get(id)
	if err != nil {
		return fmt.Errorf("Todo with ID %w doesn't exists in Todo Manager", id)
	}

	delete(t.Todos, id)
	fmt.Printf("\nTodo with ID %w deleted succesfully", id)
	return nil
}
