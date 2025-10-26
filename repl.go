package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func startRepl() {
	scanner := bufio.NewScanner(os.Stdin)

	todoManager := CreateTodoManager()

	commands := getCommands(todoManager)

	for {
		fmt.Printf("Akrasía >> ")

		scanner.Scan()
		words := cleanWords(scanner.Text())

		commandWord := words[0]
		args := words[1:]
		fmt.Println("command word: ", commandWord)

		cmd, ok := commands[commandWord]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		if err := cmd.callback(args); err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func cleanWords(words string) []string {
	lowered := strings.ToLower(words)
	return strings.Fields(lowered)
}

func menu() {
	fmt.Println("Welcome to Akrasía")
	fmt.Println("Your note app in the CLI that helps you to fight against akrasía (incontinence)")
	const OPTIONS string = `
1. Add a Todo 
2. Get All Todos 
3. Get One Todo by ID 
4. Delete a Todo by ID 
5. Exit 
`
	fmt.Println(OPTIONS)
}
