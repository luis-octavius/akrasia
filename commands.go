package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/luis-octavius/akrasia/pkg/cron"
	"github.com/spf13/cobra"
)

var (
	name        string
	description string
	priority    string
	date        []int
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

var createDailyUpdate = &cobra.Command{
	Use:     "create-cron",
	Short:   "create the cron job to update daily tasks",
	Aliases: []string{"cc"},
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := user.Current()
		if err != nil {
			return fmt.Errorf("error getting user: %v", err)
		}

		schedule := "00 13 * * *"

		homePath, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home dir: %v", err)
		}

		akrasiaCommand := path.Join(homePath, "go/bin/akrasia") + " upd"

		name := "createUpdateDaily"

		comment := "update daily todos in akrasia app"

		manager := cron.NewManager()

		if err := manager.ValidateSchedule(schedule); err != nil {
			return fmt.Errorf("invalid schedule format: %v", err)
		}

		err = manager.AddJob(name, schedule, akrasiaCommand, user.Username, comment)
		if err != nil {
			return fmt.Errorf("invalid job format: %v", err)
		}

		fmt.Printf("cron job '%s' created successfully\n", name)
		if user.Username != "" {
			fmt.Printf("\nUser: %s\n", user)
		}

		return nil
	},
}

var updateDailyTodo = &cobra.Command{
	Use:     "update-daily",
	Short:   "update daily todos",
	Aliases: []string{"upd"},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cfg.updateDailyTodo()
		if err != nil {
			return fmt.Errorf("error in update daily todo: %v", err)
		}

		return nil
	},
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

		isDaily, err := cmd.Flags().GetBool("daily")

		err = cfg.addTodo(name, description, priority, isDaily, expiresAt)
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

var delByName = &cobra.Command{
	Use:     "delete-by-name",
	Short:   "delete a todo by name",
	Aliases: []string{"deln", "dn"},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cfg.deleteByName(name)
		if err != nil {
			return err
		}

		return nil
	},
}

var getAllDaily = &cobra.Command{
	Use:     "get-daily",
	Short:   "show all daily tasks",
	Aliases: []string{"gd"},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cfg.getAllDailyTodos()
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	// add flags

	commands := map[string]*cobra.Command{
		"add":                     add,
		"getAll":                  getAll,
		"getTodoByName":           getTodoByName,
		"updateStatusToConcluded": updateStatusToConcluded,
		"deleteConcluded":         deleteConcluded,
		"checkExpiring":           checkExpiring,
		"checkExpired":            checkExpired,
		"initCmd":                 initCmd,
		"delByName":               delByName,
		"updateDailyTodo":         updateDailyTodo,
		"createDailyUpdate":       createDailyUpdate,
		"getAllDaily":             getAllDaily,
	}

	for _, cmd := range commands {
		rootCmd.AddCommand(cmd)
	}

	// flags for commands - add; getTodoByName; delByName; updateStatusToConcluded
	add.Flags().IntSliceVar(&date, "date", []int{}, "add date to task")
	add.Flags().StringVar(&name, "name", "", "task name")
	add.Flags().StringVar(&priority, "priority", "", "task priority")
	add.Flags().Bool("daily", false, "daily task")
	add.Flags().StringVar(&description, "desc", "", "task description")
	getTodoByName.Flags().StringVar(&name, "name", "", "task name")
	delByName.Flags().StringVar(&name, "name", "", "todo name")
	updateStatusToConcluded.Flags().StringVar(&name, "name", "", "task name")

	// mark as required
	if err := add.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
}
