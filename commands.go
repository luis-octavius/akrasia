package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
	"time"
)

type cliCommand struct {
	name        string
	description string
	usage       string 
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
		fmt.Printf("%s: %s\n", cmd.name, cmd.usage)
	}
	return nil
}

func commandAdd(t *TodoManager, args []string) error {
	taskDescription := strings.Join(args, " ")
	fmt.Printf("Task description: %v\n", taskDescription)

	err := t.Add(taskDescription); if err != nil {
		return fmt.Errorf("error creating todo: %w\n", err)
	}
	
	return nil
}

func commandDelete(t *TodoManager, args []string) error {
	taskId, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("error in conversion: %w\n", err)
	}
	fmt.Printf("Task Id: %v\n", taskId)

	err = t.Delete(taskId); if err != nil {
		return fmt.Errorf("error deleting todo: %w\n", err)
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

	fmt.Printf("%v. %v\n%v\n", todo.id, todo.description, todo.createdAt.Format(time.RFC1123))

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
			usage: "add study to algorithms proof (add a todo with the name 'study to algorithms proof')",
			callback: func(args []string) error {
				return commandAdd(t, args)
			},
		},
		"help": {
			name:        "help",
			description: "shows how to use the commands",
			usage: "",
			callback: func(args []string) error {
				return commandHelp(t, args)
			},
		},
		"get-all": {
			name:        "get all",
			description: "show all todos",
			usage: "get-all (shows all todos in the database)",
			callback: func(args []string) error {
				return commandGetAll(t, args)
			},
		},
		"get": {
			name:        "get",
			description: "get a todo by id",
			usage: "get 1 (return the todo with id of 1)",
			callback: func(args []string) error {
				return commandGet(t, args)
			},
		},
		"delete": {
			name:        "delete",
			description: "delete a todo with the id",
			usage: "delete 1 (delete a todo with id of 1)",
			callback: func(args []string) error {
				return commandDelete(t, args)
			},
		},
		"exit": {
			name:        "exit",
			description: "close the program",
			usage: "exit (closes the program)",
			callback: func(args []string) error {
				return commandExit(t, args)
			},
		},
	}
}
