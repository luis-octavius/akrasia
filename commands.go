package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	name        string
	description string
	// todoTime    []int
	date []int
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
	Args:    cobra.MaximumNArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		expiresAt, err := parseDate(date)
		if err != nil {
			return err
		}

		fmt.Printf("Slice is: %v\n", date)

		err = cfg.addTodo(name, description, expiresAt)
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
		if name == "" {
			return errors.New("name cannot be empty")
		}

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

var checkExpiring = &cobra.Command{
	Use:     "check-expiring",
	Short:   "check todos that are expiring in 5 days",
	Aliases: []string{"cx", "chex"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cfg.checkExpiring()
		if err != nil {
			return err
		}

		return nil
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize akrasia app",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := initDB()
		if err != nil {
			log.Fatal("Error opening database: ", err)
		}

		fmt.Printf("Akrasia App initialized successfully!")
	},
}

func init() {
	// add flags
	add.Flags().IntSliceVar(&date, "date", []int{}, "add date to todo")
	add.Flags().StringVar(&name, "name", "", "todo name")

	// mark as required
	if err := add.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	add.Flags().StringVar(&description, "desc", "", "todo description")
	// add.Flags().IntSliceVar(&todoTime, "t", []int{}, "add a specific time")
	rootCmd.AddCommand(add)

	rootCmd.AddCommand(getAll)

	// getTodoByName flag
	getTodoByName.Flags().StringVar(&name, "name", "", "todo name")
	rootCmd.AddCommand(getTodoByName)

	updateStatusToConcluded.Flags().StringVar(&name, "name", "", "todo name")
	rootCmd.AddCommand(updateStatusToConcluded)
	rootCmd.AddCommand(deleteConcluded)
	rootCmd.AddCommand(checkExpiring)
	rootCmd.AddCommand(checkExpired)
	rootCmd.AddCommand(initCmd)
}
