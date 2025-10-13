package main

import (
	"todo-terminal/todo"
	"fmt"
)

func main() {

	manager := todo.NewTodoManager()

	manager.AddTodo("Teste")

	todo, exists := manager.GetTodo(1)
	if exists {
		fmt.Println(todo)
	}

}
