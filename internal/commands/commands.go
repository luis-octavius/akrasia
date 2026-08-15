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
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/cron"
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
	Short: i18n.T("rootCmdShort"),
	Long: i18n.T("rootCmdLong"),
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
	Short:   i18n.T("createDailyUpdateShort"),
	Aliases: []string{"cc"},
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := user.Current()
		if err != nil {
			return fmt.Errorf(i18n.T("errorGetUser"), err)
		}

		schedule := "00 22 * * *"

		executablePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf(i18n.T("errorGetExecutable"), err)
		}

		akrasiaCommand := fmt.Sprintf("%s upd", executablePath)

		name := "createUpdateDaily"

		comment := "update daily todos in akrasia app"

		manager := cron.NewManager()

		if err := manager.ValidateSchedule(schedule); err != nil {
			return fmt.Errorf(i18n.T("errorInvalidSchedule"), err)
		}

		err = manager.AddJob(name, schedule, akrasiaCommand, user.Username, comment)
		if err != nil {
			return fmt.Errorf(i18n.T("errorInvalidJob"), err)
		}

		fmt.Printf(i18n.T("cronjobCreated"), name)
		if user.Username != "" {
			fmt.Printf(i18n.T("cronjobUser"), user)
		}

		return nil
	},
}

// updateDailyTodo triggers the daily reset flow used by cron.
var updateDailyTodo = &cobra.Command{
	Use:     "update-daily",
	Short:   i18n.T("updateDailyTodoShort"),
	Aliases: []string{"upd"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		err = tkm.UpdateDailyTodo()
		if err != nil {
			return fmt.Errorf(i18n.T("errorUpdateDailyTodo"), err)
		}

		return nil
	},
}

// add creates a new task with optional metadata such as priority and daily mode.
var add = &cobra.Command{
	Use:     i18n.T("addUse"),
	Short:   i18n.T("addShort"),
	Aliases: []string{"a"},
	Example: i18n.T("addExample"),
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
			return errors.New(i18n.T("errorTaskName"))
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
	Short:   i18n.T("getAllShort"),
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
	Short:   i18n.T("todayShort"),
	Aliases: []string{"td"},
	Example: "akrasia today --only overdue --limit 5 --priority high",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if filterPriority != "" && filterPriority != "high" && filterPriority != "medium" && filterPriority != "low" {
			return fmt.Errorf(i18n.T("errorInvalidPriority"), filterPriority)
		}

		if todayOnly != "" && todayOnly != "overdue" && todayOnly != "today" && todayOnly != "daily" && todayOnly != "soon" {
			return fmt.Errorf(i18n.T("errorInvalidTodayOnly"), todayOnly)
		}

		if todayLimit < 0 {
			return fmt.Errorf(i18n.T("errorInvalidTodayLimit"))
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
	Short:   i18n.T("focusShort"),
	Aliases: []string{"fc"},
	Example: "akrasia focus --limit 3 --priority high",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if focusLimit < 1 || focusLimit > 3 {
			return fmt.Errorf(i18n.T("errorFocusLimit"))
		}

		if filterPriority != "" && filterPriority != "high" && filterPriority != "medium" && filterPriority != "low" {
			return fmt.Errorf(i18n.T("errorInvalidFocusPriority"), filterPriority)
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
	Short:   i18n.T("getTodoByNameShort"),
	Aliases: []string{"gn", "name"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if name == "" {
			return errors.New(i18n.T("errorEmptyName"))
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
	Short:   i18n.T("updateStatusShort"),
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
	Short:   i18n.T("deleteConcludedShort"),
	Aliases: []string{"dc", "delc"},
	Example: "akrasia delete-concluded --yes",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteYes {
			return errors.New(i18n.T("errorDestructiveAction"))
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
	Short:   i18n.T("checkExpiredShort"),
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
	Short:   i18n.T("checkExpiringShort"),
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
	Short: i18n.T("initCmdShort"),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := db.InitDB()
		if err != nil {
			log.Fatal(i18n.T("errorOpenDatabase"), err)
		}

		fmt.Printf(i18n.T("initSuccessful"))
	},
}

// delByName deletes a single task by name.
var delByName = &cobra.Command{
	Use:     "delete-by-name",
	Short:   i18n.T("delByNameShort"),
	Aliases: []string{"deln", "dn"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteYes {
			return errors.New(i18n.T("errorDestructiveAction"))
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
	Short:   i18n.T("getAllDailyShort"),
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
	Short:   i18n.T("getTodoCurrentStreakShort"),
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
	Short:   i18n.T("getTodoStreakHistoryShort"),
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
	Short:   i18n.T("backfillHistoryShort"),
	Aliases: []string{"bf"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		// decrease by one to not generate errors with today
		daysBackfill = daysBackfill - 1

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
	Short:   i18n.T("configShort"),
	Aliases: []string{"cfg"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// configTheme manages theme selection.
var configTheme = &cobra.Command{
	Use:     "theme [list|show|<theme-name>]",
	Short:   i18n.T("configThemeShort"),
	Aliases: []string{"t"},
	Example: "akrasia config theme high-contrast\nakrasia config theme list\nakrasia config theme show",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		action := args[0]

		switch action {
		case "list":
			fmt.Println(i18n.T("availableThemes"))
			for _, theme := range color.GetAvailableThemes() {
				fmt.Printf("  - %s\n", theme)
			}
			return nil

		case "show":
			currentTheme := color.GetCurrentTheme()
			fmt.Printf(i18n.T("currentTheme"), currentTheme.Name)
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
				return fmt.Errorf(i18n.T("errorUnknownTheme"), action, availableThemes)
			}

			if err := color.SaveTheme(action); err != nil {
				return fmt.Errorf(i18n.T("errorSaveTheme"), err)
			}

			fmt.Printf(i18n.T("themeSet"), action)
			return nil
		}
	},
}

var configLanguage = &cobra.Command{
	Use:     "language [list|show|<language-code>]",
	Short:   i18n.T("configLanguageShort"),
	Aliases: []string{"l"},
	Example: "akrasia config language pt\nakrasia config language list\nakrasia config language show",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		action := args[0]

		switch action {
		case "list":
			fmt.Println(i18n.T("availableLanguages"))
			for _, language := range i18n.GetAvailableLanguages() {
				fmt.Printf(" - %s\n", language)
			}
			return nil

		case "show":
			fmt.Printf(i18n.T("currentLanguage"), i18n.GetCurrentLanguage())
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
				return fmt.Errorf(i18n.T("errorUnknownLanguage"), action, availableLanguages)
			}

			if err := i18n.SetLanguage(action); err != nil {
				return fmt.Errorf(i18n.T("errorSetLanguage"), err)
			}
			fmt.Printf(i18n.T("languageSet"), action)
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
		"updateDailyTodo":         updateDailyTodo,
		"createDailyUpdate":       createDailyUpdate,
		"getAllDaily":             getAllDaily,
		"getTodoCurrentStreak":    getTodoCurrentStreak,
		"getTodoStreakHistory":    getTodoStreakHistory,
		"backfillHistory":         backfillHistory,
		"config":                  config,
	}

	for _, cmd := range commands {
		rootCmd.AddCommand(cmd)
	}

	// Add subcommands to config
	config.AddCommand(configTheme, configLanguage)

	// flags for commands - add; getTodoByName; delByName; updateStatusToConcluded
	add.Flags().IntSliceVar(&date, "date", []int{}, i18n.T("addFlagDate"))
	add.Flags().StringVar(&name, "name", "", i18n.T("addFlagName"))
	add.Flags().StringVar(&priority, "priority", "", i18n.T("addFlagPriority"))
	add.Flags().Bool("daily", false, i18n.T("addFlagDaily"))
	add.Flags().StringVar(&description, "desc", "", i18n.T("addFlagDescription"))
	getAll.Flags().StringVar(&filterPriority, "priority", "", i18n.T("getAllFlagFilterPriority"))
	today.Flags().StringVar(&filterPriority, "priority", "", i18n.T("todayFlagFilterPriority"))
	focus.Flags().IntVar(&focusLimit, "limit", 3, i18n.T("focusFlagFocusLimit"))
	focus.Flags().StringVar(&filterPriority, "priority", "", i18n.T("focusFlagPriority"))
	getTodoByName.Flags().StringVar(&name, "name", "", i18n.T("getTodoByNameFlagName"))
	delByName.Flags().StringVar(&name, "name", "", i18n.T("delByNameFlagName"))
	delByName.Flags().BoolVar(&deleteYes, "yes", false, i18n.T("delByNameFlagDeleteYes"))
	updateStatusToConcluded.Flags().StringVar(&name, "name", "", i18n.T("updateStatusToConcludedFlagName"))
	updateStatusToConcluded.Flags().StringVar(&notes, "notes", "", i18n.T("updateStatusToConcludedFlagNotes"))
	deleteConcluded.Flags().BoolVar(&deleteYes, "yes", false, i18n.T("deleteConcludedFlagDeleteYes"))
	getTodoCurrentStreak.Flags().StringVar(&name, "name", "", i18n.T("getTodoCurrentStreakFlagName"))
	getTodoStreakHistory.Flags().StringVar(&name, "name", "", i18n.T("getTodoStreakHistoryFlagName"))
	backfillHistory.Flags().IntVar(&daysBackfill, "days", 30, i18n.T("backfillHistoryFlagDays"))
	backfillHistory.Flags().StringVar(&name, "task", "", i18n.T("backfillHistoryFlagName"))
	today.Flags().StringVar(&todayOnly, "only", "", i18n.T("todayFlagTodayOnly"))
	today.Flags().IntVar(&todayLimit, "limit", 0, i18n.T("todayFlagLimit"))
	today.Flags().BoolVar(&todayJSON, "json", false, i18n.T("todayFlagTodayJSON"))
}
