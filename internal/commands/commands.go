package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/luis-octavius/akrasia/internal/db"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/cron"
	"github.com/spf13/cobra"
)

var (
	name           string
	notes          string
	description    string
	priority       string
	date           []int
	daysBackfill   int
	todayOnly      string
	todayLimit     int
	todayJSON      bool
	filterPriority string
	deleteYes      bool
	focusLimit     int
)

// rootCmd is the CLI entrypoint that registers all Akrasia subcommands.
var rootCmd = &cobra.Command{
	Use:   "akrasia",
	Short: "An app that helps with fighting akrasia",
	Long: `Akrasia is a word in greek that means "Incontinence", which means a lack of self-control.
This app is constructed to simply help to fight akrasía rapidly in the terminal, by adding things to do
and to keep track of these things for you.`,
}

// Execute runs the root command and exits the process on fatal command errors.
func ExecuteWithContext(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// createDailyUpdate installs a system cron job that runs the daily update command.
var createDailyUpdate = &cobra.Command{
	Use:     "create-cron",
	Short:   "create the cron job to update daily tasks",
	Aliases: []string{"cc"},
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := user.Current()
		if err != nil {
			return fmt.Errorf("error getting user: %v", err)
		}

		schedule := "00 22 * * *"

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
			fmt.Printf("\nUser: %v\n", user)
		}

		return nil
	},
}

// updateDailyTodo triggers the daily reset flow used by cron.
var updateDailyTodo = &cobra.Command{
	Use:     "update-daily",
	Short:   "reset daily tasks for a new day and record completion history (runs via cron)",
	Aliases: []string{"upd"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.UpdateDailyTodo()
		if err != nil {
			return fmt.Errorf("error in update daily todo: %v", err)
		}

		return nil
	},
}

// add creates a new task with optional metadata such as priority and daily mode.
var add = &cobra.Command{
	Use:     "add --name <name> --desc [description] --date ['13,10']",
	Short:   "create a new task with optional description, priority, and expiration date",
	Aliases: []string{"a"},
	Example: "akrasia add --name \"Morning run\" --daily --priority high",
	Args:    cobra.MaximumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		expiresAt, err := parseDate(date)
		if err != nil {
			return err
		}

		isDaily, err := cmd.Flags().GetBool("daily")

		err = tkm.AddTodo(name, description, priority, isDaily, expiresAt)
		if err != nil {
			return err
		}

		return nil
	},
}

// getAll lists all tasks, optionally filtered by priority.
var getAll = &cobra.Command{
	Use:     "get-all",
	Short:   "returns all the todos saved in storage",
	Aliases: []string{"ga"},
	Example: "akrasia get-all --priority high",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.GetTodos(filterPriority)
		if err != nil {
			return err
		}

		return nil
	},
}

// today shows a categorized dashboard for what needs attention today.
var today = &cobra.Command{
	Use:     "today",
	Short:   "show a daily focus dashboard (overdue, due today, daily pending, expiring soon)",
	Aliases: []string{"td"},
	Example: "akrasia today --only overdue --limit 5 --priority high",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if filterPriority != "" && filterPriority != "high" && filterPriority != "medium" && filterPriority != "low" {
			return fmt.Errorf("invalid value for --priority: %q (use high|medium|low)", filterPriority)
		}

		if todayOnly != "" && todayOnly != "overdue" && todayOnly != "today" && todayOnly != "daily" && todayOnly != "soon" {
			return fmt.Errorf("invalid value for --only: %q (use overdue|today|daily|soon)", todayOnly)
		}

		if todayLimit < 0 {
			return fmt.Errorf("--limit cannot be negative")
		}

		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.GetTodayFocus(tasks.TodayOptions{
			Only:     todayOnly,
			Limit:    todayLimit,
			JSON:     todayJSON,
			Priority: filterPriority,
		})
		if err != nil {
			return err
		}

		return nil
	},
}

// focus shows the top actionable tasks to execute now.
var focus = &cobra.Command{
	Use:     "focus",
	Short:   "show top 1-3 priority items to focus on right now",
	Aliases: []string{"fc"},
	Example: "akrasia focus --limit 3 --priority high",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if focusLimit < 1 || focusLimit > 3 {
			return fmt.Errorf("--limit must be between 1 and 3")
		}

		if filterPriority != "" && filterPriority != "high" && filterPriority != "medium" && filterPriority != "low" {
			return fmt.Errorf("invalid value for --priority: %q (use high|medium|low)", filterPriority)
		}

		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}
		return tkm.GetFocus(focusLimit, filterPriority)
	},
}

// getTodoByName searches for a task using case-insensitive fuzzy matching.
var getTodoByName = &cobra.Command{
	Use:     "get-by-name <name>",
	Short:   "search for a task by name (case-insensitive, fuzzy match)",
	Aliases: []string{"gn", "name"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if name == "" {
			return errors.New("name cannot be empty")
		}

		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.GetTodoByName(name)
		if err != nil {
			return err
		}

		return nil
	},
}

// updateStatusToConcluded marks a task as completed and records completion history.
var updateStatusToConcluded = &cobra.Command{
	Use:     "update-status <name>",
	Short:   "mark a task as completed and record it in history with optional notes",
	Aliases: []string{"us"},
	Example: "akrasia update-status --name \"Morning run\" --notes \"Done after lunch\"",
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.UpdateToConcluded(name, notes)
		if err != nil {
			return err
		}

		return nil
	},
}

// deleteConcluded removes all concluded tasks after explicit confirmation.
var deleteConcluded = &cobra.Command{
	Use:     "delete-concluded",
	Short:   "delete all concluded todos",
	Aliases: []string{"dc", "delc"},
	Example: "akrasia delete-concluded --yes",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteYes {
			return errors.New("destructive action blocked: use --yes to confirm")
		}

		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.DeleteConcluded()
		if err != nil {
			return err
		}

		return nil
	},
}

// checkExpired lists expired non-daily tasks.
var checkExpired = &cobra.Command{
	Use:     "check-expired",
	Short:   "show tasks that have passed their expiration date",
	Aliases: []string{"ce"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.CheckExpired()
		if err != nil {
			return err
		}

		return nil
	},
}

// checkExpiring lists tasks that are approaching expiration.
var checkExpiring = &cobra.Command{
	Use:     "check-expiring",
	Short:   "check todos that are expiring in 5 days",
	Aliases: []string{"cx", "chex"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.CheckExpiring()
		if err != nil {
			return err
		}

		return nil
	},
}

// initCmd initializes the local database schema and storage.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize akrasia app",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := db.InitDB()
		if err != nil {
			log.Fatal("Error opening database: ", err)
		}

		fmt.Printf("Akrasia App initialized successfully!")
	},
}

// delByName deletes a single task by name.
var delByName = &cobra.Command{
	Use:     "delete-by-name",
	Short:   "delete a todo by name",
	Aliases: []string{"deln", "dn"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteYes {
			return errors.New("destructive action blocked: use --yes to confirm")
		}

		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.DeleteByName(name)
		if err != nil {
			return err
		}

		return nil
	},
}

// getAllDaily shows all tasks marked as daily.
var getAllDaily = &cobra.Command{
	Use:     "get-daily",
	Short:   "show all daily tasks",
	Aliases: []string{"gd"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.GetAllDailyTodos()
		if err != nil {
			return err
		}

		return nil
	},
}

// getTodoCurrentStreak returns the current completion streak for a task.
var getTodoCurrentStreak = &cobra.Command{
	Use:     "streak",
	Short:   "get the current streak of provided todo",
	Aliases: []string{"curr", "cs"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.GetCurrentStreak(name)
		if err != nil {
			return err
		}
		return nil
	},
}

// getTodoStreakHistory returns the streak history timeline for a task.
var getTodoStreakHistory = &cobra.Command{
	Use:     "history",
	Short:   "get the streak history of provided todo",
	Aliases: []string{"his", "sh"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.GetStreakHistory(name)
		if err != nil {
			return err
		}

		return nil
	},
}

// backfillHistory inserts missing history snapshots for daily tasks.
var backfillHistory = &cobra.Command{
	Use:     "backfill-history",
	Short:   "backfill missing daily task history entries",
	Aliases: []string{"bf"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.BackfillDailyHistory(daysBackfill, name)
		if err != nil {
			return err
		}

		return nil
	},
}

// init wires subcommands, flags, and required arguments into the root command.
func init() {
	commands := map[string]*cobra.Command{
		"add":                     add,
		"getAll":                  getAll,
		"today":                   today,
		"focus":                   focus,
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
	getAll.Flags().StringVar(&filterPriority, "priority", "", "filter tasks by priority: high|medium|low")
	today.Flags().StringVar(&filterPriority, "priority", "", "filter tasks by priority: high|medium|low")
	focus.Flags().IntVar(&focusLimit, "limit", 3, "number of focus items (1-3)")
	focus.Flags().StringVar(&filterPriority, "priority", "", "filter tasks by priority: high|medium|low")
	getTodoByName.Flags().StringVar(&name, "name", "", "task name")
	delByName.Flags().StringVar(&name, "name", "", "todo name")
	delByName.Flags().BoolVar(&deleteYes, "yes", false, "confirm destructive deletion of concluded tasks")
	updateStatusToConcluded.Flags().StringVar(&name, "name", "", "task name")
	updateStatusToConcluded.Flags().StringVar(&notes, "notes", "", "daily notes")
	deleteConcluded.Flags().BoolVar(&deleteYes, "yes", false, "confirm destructive deletion of concluded tasks")
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
