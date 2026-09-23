package commands

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/luis-octavius/akrasia/internal/db"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/spf13/cobra"
	"github.com/luis-octavius/akrasia/pkg/i18n"
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
	Short: i18n.T("commands.commands.rootCmdShort"),
	Long: i18n.T("commands.commands.rootCmdLong"),
}

// Execute runs the root command and exits the process on fatal command errors.
func ExecuteWithContext(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// NOTE: create-cron / update-daily were removed. Daily tasks no longer
// carry mutable "concluded" state that needs a scheduled reset — see
// migration 006_add_history_since.sql and MarkDaily-style writes in
// internal/tasks/task.go. A future reminder feature (e.g. "you haven't
// done X today") can reuse pkg/cron, but it would only ever notify —
// never write app state — so it doesn't belong in this command file
// until it exists.

// add creates a new task with optional metadata such as priority and daily mode.
var add = &cobra.Command{
	Use:     i18n.T("commands.commands.addUse"),
	Short:   i18n.T("commands.commands.addShort"),
	Aliases: []string{"a"},
	Example: i18n.T("commands.commands.addExample"),
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		// Handle positional arguments and flags
		// Priority: positional args > flags (for backward compatibility)
		taskName := name
		taskDesc := description

		if len(args) > 0 {
			taskName = args[0]
		}
		if len(args) > 1 {
			taskDesc = args[1]
		}

		// Validate that name was provided
		if taskName == "" {
			return errors.New(i18n.T("commands.error.taskName"))
		}

		expiresAt, err := parseDate(date)
		if err != nil {
			return err
		}

		isDaily, err := cmd.Flags().GetBool("daily")

		err = tkm.AddTodo(taskName, taskDesc, priority, isDaily, expiresAt)
		if err != nil {
			return err
		}

		return nil
	},
}

// getAll lists all tasks, optionally filtered by priority.
var getAll = &cobra.Command{
	Use:     "get-all",
	Short:   i18n.T("commands.commands.getAllShort"),
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
	Short:   i18n.T("commands.commands.todayShort"),
	Aliases: []string{"td"},
	Example: "akrasia today --only overdue --limit 5 --priority high",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if filterPriority != "" && filterPriority != "high" && filterPriority != "medium" && filterPriority != "low" {
			return fmt.Errorf(i18n.T("commands.error.invalidPriority"), filterPriority)
		}

		if todayOnly != "" && todayOnly != "overdue" && todayOnly != "today" && todayOnly != "daily" && todayOnly != "soon" {
			return fmt.Errorf(i18n.T("commands.error.invalidTodayOnly"), todayOnly)
		}

		if todayLimit < 0 {
			return fmt.Errorf("%s", i18n.T("commands.error.invalidTodayLimit"))
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
	Short:   i18n.T("commands.commands.focusShort"),
	Aliases: []string{"fc"},
	Example: "akrasia focus --limit 3 --priority high",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if focusLimit < 1 || focusLimit > 3 {
			return fmt.Errorf("%s", i18n.T("commands.error.focusLimit"))
		}

		if filterPriority != "" && filterPriority != "high" && filterPriority != "medium" && filterPriority != "low" {
			return fmt.Errorf(i18n.T("commands.error.invalidFocusPriority"), filterPriority)
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
	Use:     i18n.T("commands.commands.getTodoByNameUse"),
	Short:   i18n.T("commands.commands.getTodoByNameShort"),
	Aliases: []string{"gn", "name"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if name == "" {
			return errors.New(i18n.T("commands.error.emptyName"))
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
	Use:     i18n.T("commands.commands.updateStatusUse"),
	Short:   i18n.T("commands.commands.updateStatusShort"),
	Aliases: []string{"us"},
	Example: i18n.T("commands.commands.updateStatusExample"),
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
	Short:   i18n.T("commands.commands.deleteConcludedShort"),
	Aliases: []string{"dc", "delc"},
	Example: "akrasia delete-concluded --yes",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteYes {
			return errors.New(i18n.T("commands.error.destructiveAction"))
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
	Short:   i18n.T("commands.commands.checkExpiredShort"),
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
	Short:   i18n.T("commands.commands.checkExpiringShort"),
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
	Short: i18n.T("commands.commands.initCmdShort"),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := db.InitDB()
		if err != nil {
			log.Fatal(i18n.T("commands.error.openDatabase"), err)
		}

		fmt.Printf("%s", i18n.T("commands.info.initSuccessful"))
	},
}

// delByName deletes a single task by name.
var delByName = &cobra.Command{
	Use:     "delete-by-name",
	Short:   i18n.T("commands.commands.delByNameShort"),
	Aliases: []string{"deln", "dn"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteYes {
			return errors.New(i18n.T("commands.error.destructiveAction"))
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
	Short:   i18n.T("commands.commands.getAllDailyShort"),
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
	Short:   i18n.T("commands.commands.getTodoCurrentStreakShort"),
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
	Short:   i18n.T("commands.commands.getTodoStreakHistoryShort"),
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

// backfillHistory inserts real completions for days the user forgot to log.
// Excludes today — use `done` for that.
var backfillHistory = &cobra.Command{
	Use:     "backfill-history",
	Short:   i18n.T("commands.commands.backfillHistoryShort"),
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

// config manages application settings and themes.
var config = &cobra.Command{
	Use:     "config",
	Short:   i18n.T("commands.commands.configShort"),
	Aliases: []string{"cfg"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// configTheme manages theme selection.
var configTheme = &cobra.Command{
	Use:     i18n.T("commands.commands.configThemeUse"),
	Short:   i18n.T("commands.commands.configThemeShort"),
	Aliases: []string{"t"},
	Example: "akrasia config theme high-contrast\nakrasia config theme list\nakrasia config theme show",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		action := args[0]

		switch action {
		case "list":
			fmt.Println(i18n.T("commands.info.availableThemes"))
			for _, theme := range color.GetAvailableThemes() {
				fmt.Printf("  - %s\n", theme)
			}
			return nil

		case "show":
			currentTheme := color.GetCurrentTheme()
			fmt.Printf(i18n.T("commands.info.currentTheme"), currentTheme.Name)
			return nil

		default:
			// Try to set the theme
			availableThemes := color.GetAvailableThemes()
			found := false
			for _, t := range availableThemes {
				if t == action {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf(i18n.T("commands.error.unknownTheme"), action, availableThemes)
			}

			if err := color.SaveTheme(action); err != nil {
				return fmt.Errorf(i18n.T("commands.error.saveTheme"), err)
			}

			fmt.Printf(i18n.T("commands.info.themeSet"), action)
			return nil
		}
	},
}

var configLanguage = &cobra.Command{
	Use:     i18n.T("commands.commands.configLanguageUse"),
	Short:   i18n.T("commands.commands.configLanguageShort"),
	Aliases: []string{"l"},
	Example: "akrasia config language pt\nakrasia config language list\nakrasia config language show",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		action := args[0]

		switch action {
		case "list":
			fmt.Println(i18n.T("commands.info.availableLanguages"))
			for _, language := range i18n.GetAvailableLanguages() {
				fmt.Printf(" - %s\n", language)
			}
			return nil

		case "show":
			fmt.Printf(i18n.T("commands.info.currentLanguage"), i18n.GetCurrentLanguage())
			return nil

		default:
			availableLanguages := i18n.GetAvailableLanguages()
			found := false
			for _, l := range availableLanguages {
				if l == action {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf(i18n.T("commands.error.unknownLanguage"), action, availableLanguages)
			}

			if err := i18n.SetLanguage(action); err != nil {
				return fmt.Errorf(i18n.T("commands.error.setLanguage"), err)
			}
			fmt.Printf(i18n.T("commands.info.languageSet"), action)
			return nil
		}
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
		"getAllDaily":             getAllDaily,
		"getTodoCurrentStreak":    getTodoCurrentStreak,
		"getTodoStreakHistory":    getTodoStreakHistory,
		"backfillHistory":         backfillHistory,
		"config":                  config,
	}

	for _, cmd := range commands {
		rootCmd.AddCommand(cmd)
	}

	// Set usage func for localised help
	rootCmd.SetUsageFunc(UsageFunc)

	// Add subcommands to config
	config.AddCommand(configTheme, configLanguage)

	// flags for commands - add; getTodoByName; delByName; updateStatusToConcluded
	add.Flags().IntSliceVar(&date, "date", []int{}, i18n.T("commands.flags.addDate"))
	add.Flags().StringVar(&name, "name", "", i18n.T("commands.flags.addName"))
	add.Flags().StringVar(&priority, "priority", "", i18n.T("commands.flags.addPriority"))
	add.Flags().Bool("daily", false, i18n.T("commands.flags.addDaily"))
	add.Flags().StringVar(&description, "desc", "", i18n.T("commands.flags.addDescription"))
	getAll.Flags().StringVar(&filterPriority, "priority", "", i18n.T("commands.flags.getAllPriority"))
	today.Flags().StringVar(&filterPriority, "priority", "", i18n.T("commands.flags.todayPriority"))
	focus.Flags().IntVar(&focusLimit, "limit", 3, i18n.T("commands.flags.focusLimit"))
	focus.Flags().StringVar(&filterPriority, "priority", "", i18n.T("commands.flags.focusPriority"))
	getTodoByName.Flags().StringVar(&name, "name", "", i18n.T("commands.flags.getTodoByNameName"))
	delByName.Flags().StringVar(&name, "name", "", i18n.T("commands.flags.delByNameName"))
	delByName.Flags().BoolVar(&deleteYes, "yes", false, i18n.T("commands.flags.delByNameDeleteYes"))
	updateStatusToConcluded.Flags().StringVar(&name, "name", "", i18n.T("commands.flags.updateStatusToConcludedName"))
	updateStatusToConcluded.Flags().StringVar(&notes, "notes", "", i18n.T("commands.flags.updateStatusToConcludedNotes"))
	deleteConcluded.Flags().BoolVar(&deleteYes, "yes", false, i18n.T("commands.flags.deleteConcludedDeleteYes"))
	getTodoCurrentStreak.Flags().StringVar(&name, "name", "", i18n.T("commands.flags.getTodoCurrentStreakName"))
	getTodoStreakHistory.Flags().StringVar(&name, "name", "", i18n.T("commands.flags.getTodoStreakHistoryName"))
	backfillHistory.Flags().IntVar(&daysBackfill, "days", 30, i18n.T("commands.flags.backfillHistoryDays"))
	backfillHistory.Flags().StringVar(&name, "task", "", i18n.T("commands.flags.backfillHistoryName"))
	today.Flags().StringVar(&todayOnly, "only", "", i18n.T("commands.flags.todayOnly"))
	today.Flags().IntVar(&todayLimit, "limit", 0, i18n.T("commands.flags.todayLimit"))
	today.Flags().BoolVar(&todayJSON, "json", false, i18n.T("commands.flags.todayJSON"))
}
