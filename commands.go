package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
)

type cliCommand struct {
	name        string
	description string
	callback    func(args []string) error
}

func commandExit(t *TodoManager, args []string) error {
	fmt.Println("Closing Akrasía... Bye bye!")
	os.Exit(0)
	return nil
}

func commandHelp(t *TodoManager, args []string) error {
	commands := getCommands(t)
	for _, cmd := range commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func commandAdd(t *TodoManager, args []string) error {
	taskDescription := strings.Join(args, " ")
	fmt.Printf("Task description: %v\n", taskDescription)

	err := t.Add(taskDescription); if err != nil {
		fmt.Printf("error creating todo: %w", err)
	}
	
	return nil
}

func commandDelete(t *TodoManager, args []string) error {
	taskId, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("error in conversion: %w", err)
	}
	fmt.Printf("Task Id: %v\n", taskId)

	err = t.Delete(taskId); if err != nil {
		return fmt.Errorf("error deleting todo: %w", err)
	}
	return nil
}

func commandGet(t *TodoManager, args []string) error {
	taskId, err := strconv.Atoi(args[0]); if err != nil {
		fmt.Printf("error converting id: %w\n", err)
	}

	todo, err := t.Get(taskId); if err != nil {
		fmt.Printf("error getting todo: %w\n", err)
	}

	fmt.Printf("%v. %v\n%v\n", todo.id, todo.description, todo.createdAt)

	return nil
}

func commandGetAll(t *TodoManager, args []string) error {
	err := t.GetAll(); if err != nil {
		fmt.Printf("error generating todo list: %w", err)
	}

	return nil
}

func getCommands(t *TodoManager) map[string]cliCommand {
	return map[string]cliCommand{
		"add": {
			name:        "add",
			description: "add a todo",
			callback: func(args []string) error {
				return commandAdd(t, args)
			},
		},
		"get-all": {
			name:        "get all",
			description: "show all todos",
			callback: func(args []string) error {
				return commandGetAll(t, args)
			},
		},
		"get": {
			name:        "get",
			description: "get a todo by id",
			callback: func(args []string) error {
				return commandGet(t, args)
			},
		},
		"delete": {
			name:        "delete",
			description: "delete a todo with the id",
			callback: func(args []string) error {
				return commandDelete(t, args)
			},
		},
		"exit": {
			name:        "exit",
			description: "close the program",
			callback: func(args []string) error {
				return commandExit(t, args)
			},
		},
	}
}
