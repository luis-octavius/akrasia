package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "akrasia",
	Short: "An app that helps with fighting akrasia",
	Long: `Akrasia is a word in greek that means "Incontinence", which means a lack of self-control. 
This app is constructed to simply help to fight akrasía rapidly in the terminal, by adding things to do 
and to keep track of these things for you.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

var add = &cobra.Command{
	Use:     "add <name> [description]",
	Short:   "adds a todo in storage, description is optional",
	Aliases: []string{"a"},
	Args:    cobra.MaximumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		var description string
		var expiresAt time.Time

		lenArgs := len(args)

		switch lenArgs {
		case 0:
			return fmt.Errorf("Not enough arguments provided")
		case 1:
			name = args[0]
		case 2:
			name = args[0]
			description = args[1]
		case 3:
			name = args[0]
			description = args[1]

			if args[2] == "" {
				expiresAt, _ = parseTime([]string{})
			} else {
				splitTime := strings.Split(args[2], " ")
				expiresAt, _ = parseTime(splitTime)
			}
		}

		err := cfg.addTodo(name, description, expiresAt)
		if err != nil {
			return err
		}

		return nil
	},
}

var getAll = &cobra.Command{
	Use:     "get-all",
	Short:   "returns all the todos saved in storage",
	Aliases: []string{"ga"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cfg.getTodos()
		if err != nil {
			return err
		}

		return nil
	},
}

var getTodoByName = &cobra.Command{
	Use:     "get-by-name <name>",
	Short:   "gets a todo by name",
	Aliases: []string{"gn", "name"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		err := cfg.getTodoByName(name)
		if err != nil {
			return err
		}

		return nil
	},
}

var updateStatusToConcluded = &cobra.Command{
	Use:     "update-status <name>",
	Short:   "update concluded status to true",
	Aliases: []string{"us"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		err := cfg.updateToConcluded(name)
		if err != nil {
			return err
		}

		return nil
	},
}

var deleteConcluded = &cobra.Command{
	Use:     "delete-concluded",
	Short:   "delete all concluded todos",
	Aliases: []string{"dc", "delc"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cfg.deleteConcluded()
		if err != nil {
			return err
		}

		return nil
	},
}

var checkExpired = &cobra.Command{
	Use:     "check-expired",
	Short:   "check expired todos",
	Aliases: []string{"ce"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cfg.checkExpired()
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(add)
	rootCmd.AddCommand(getAll)
	rootCmd.AddCommand(getTodoByName)
	rootCmd.AddCommand(updateStatusToConcluded)
	rootCmd.AddCommand(deleteConcluded)
	rootCmd.AddCommand(checkExpired)
}
