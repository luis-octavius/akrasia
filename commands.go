package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/luis-octavius/akrasia/pkg/cron"
	"github.com/spf13/cobra"
)

var (
	name         string
	notes        string
	description  string
	priority     string
	date         []int
	daysBackfill int
	todayOnly    string
	todayLimit   int
	todayJSON    bool
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

		schedule := "00 21 * * *"

		executablePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("error getting executable path: %v", err)
		}

		akrasiaCommand := fmt.Sprintf("%s upd", executablePath)

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
	Short:   "reset daily tasks for a new day and record completion history (runs via cron)",
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
	Short:   "create a new task with optional description, priority, and expiration date",
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

var today = &cobra.Command{
	Use:     "today",
	Short:   "show a daily focus dashboard (overdue, due today, daily pending, expiring soon)",
	Aliases: []string{"td"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if todayOnly != "" && todayOnly != "overdue" && todayOnly != "today" && todayOnly != "daily" && todayOnly != "soon" {
			return fmt.Errorf("invalid value for --only: %q (use overdue|today|daily|soon)", todayOnly)
		}

		if todayLimit < 0 {
			return fmt.Errorf("--limit cannot be negative")
		}

		err := cfg.getTodayFocus(todayOptions{
			Only:  todayOnly,
			Limit: todayLimit,
			JSON:  todayJSON,
		})
		if err != nil {
			return err
		}

		return nil
	},
}

var getTodoByName = &cobra.Command{
	Use:     "get-by-name <name>",
	Short:   "search for a task by name (case-insensitive, fuzzy match)",
	Aliases: []string{"gn", "name"},
	Args:    cobra.NoArgs,
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
	Short:   "mark a task as completed and record it in history with optional notes",
	Aliases: []string{"us"},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cfg.updateToConcluded(name, notes)
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
	Short:   "show tasks that have passed their expiration date",
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

var getTodoCurrentStreak = &cobra.Command{
	Use:     "streak",
	Short:   "get the current streak of provided todo",
	Aliases: []string{"curr", "cs"},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cfg.getCurrentStreak(name)
		if err != nil {
			return err
		}
		return nil
	},
}

var getTodoStreakHistory = &cobra.Command{
	Use:     "history",
	Short:   "get the streak history of provided todo",
	Aliases: []string{"his", "sh"},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cfg.getStreakHistory(name)
		if err != nil {
			return err
		}

		return nil
	},
}

var backfillHistory = &cobra.Command{
	Use:     "backfill-history",
	Short:   "backfill missing daily task history entries",
	Aliases: []string{"bf"},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cfg.backfillDailyHistory(daysBackfill, name)
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
		"today":                   today,
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
		"getTodoCurrentStreak":    getTodoCurrentStreak,
		"getTodoStreakHistory":    getTodoStreakHistory,
		"backfillHistory":         backfillHistory,
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
	updateStatusToConcluded.Flags().StringVar(&notes, "notes", "", "daily notes")
	getTodoCurrentStreak.Flags().StringVar(&name, "name", "", "current streak")
	getTodoStreakHistory.Flags().StringVar(&name, "name", "", "todo streak history")
	backfillHistory.Flags().IntVar(&daysBackfill, "days", 30, "number of days back to backfill (default 30)")
	backfillHistory.Flags().StringVar(&name, "task", "", "specific task name to backfill (optional, fills all daily tasks if not set)")
	today.Flags().StringVar(&todayOnly, "only", "", "show only one section: overdue|today|daily|soon")
	today.Flags().IntVar(&todayLimit, "limit", 0, "limit items per section (0 = no limit)")
	today.Flags().BoolVar(&todayJSON, "json", false, "print today output as JSON")

	// mark as required
	if err := add.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
}
