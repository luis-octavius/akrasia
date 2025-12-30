package main

import (
	"fmt"
	"log"

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
	Args:    cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		var description string

		lenArgs := len(args)

		if lenArgs == 0 {
			return fmt.Errorf("Not enough arguments provided")
		}

		if lenArgs == 1 {
			name = args[0]
		}

		if lenArgs == 2 {
			name = args[0]
			description = args[1]
		}

		err := cfg.addTodo(name, description)
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
	Short:   "delete all concluded todos in storage",
	Aliases: []string{"dc"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cfg.deleteConcluded()
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
}
