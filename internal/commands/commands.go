package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/luis-octavius/akrasia/internal/db"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/internal/tui"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

	// edit flags; only the ones the user passes are applied
	editName        string
	editDescription string
	editPriority    string
	editDaily       bool
	editDate        []int
)

// rootCmd is the CLI entrypoint that registers all Akrasia subcommands.
// Run without arguments it opens the TUI; `akrasia help` shows the help.
var rootCmd = &cobra.Command{
	Use:   "akrasia",
	Short: i18n.T("commands.commands.rootCmdShort"),
	Long:  i18n.T("commands.commands.rootCmdLong"),
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// pipes and scripts get the help instead of a TUI they can't drive
		if !isInteractive() {
			return cmd.Help()
		}

		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		return tui.Run(cmd.Context(), tkm)
	},
}

// isInteractive reports whether both stdin and stdout are terminals. It is
// a variable so tests can force either path.
var isInteractive = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
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
	Args:    cobra.MaximumNArgs(2),
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

		expiresAt, err := tasks.ParseDate(date)
		if err != nil {
			return err
		}

		isDaily, err := cmd.Flags().GetBool("daily")
		if err != nil {
			return err
		}

		todo, err := tkm.AddTodo(cmd.Context(), tasks.NewTodo{
			Name:        taskName,
			Description: taskDesc,
			Priority:    priority,
			IsDaily:     isDaily,
			ExpiresAt:   expiresAt,
		})
		if err != nil {
			return err
		}

		color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.createdTask"), todo.Name))
		printQuote()
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

		todos, err := tkm.ListTodos(cmd.Context(), filterPriority)
		if err != nil {
			return err
		}

		fmt.Println(i18n.T("tasks.tasks.allTasks"))
		printTodos(todos)
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

		view, err := tkm.Today(cmd.Context(), tasks.TodayOptions{
			Only:     todayOnly,
			Limit:    todayLimit,
			Priority: filterPriority,
		})
		if err != nil {
			return err
		}

		progress, err := tkm.TodayProgress(cmd.Context())
		if err != nil {
			return err
		}

		return printTodayView(view, progress, todayJSON)
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
		todos, err := tkm.Focus(cmd.Context(), focusLimit, filterPriority)
		if err != nil {
			return err
		}

		progress, err := tkm.TodayProgress(cmd.Context())
		if err != nil {
			return err
		}

		printFocus(todos, progress)
		return nil
	},
}

// getTodoByName searches for a task using case-insensitive fuzzy matching.
var getTodoByName = &cobra.Command{
	Use:     i18n.T("commands.commands.getTodoByNameUse"),
	Short:   i18n.T("commands.commands.getTodoByNameShort"),
	Aliases: []string{"gn", "name"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskName := name

		if len(args) > 0 {
			taskName = args[0]
		}

		if taskName == "" {
			return errors.New(i18n.T("commands.error.emptyName"))
		}

		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		todo, err := tkm.FindTodo(cmd.Context(), taskName)
		if err != nil {
			return err
		}

		printFoundTodo(todo)
		return nil
	},
}

// updateStatusToConcluded marks a task as completed and records completion history.
var updateStatusToConcluded = &cobra.Command{
	Use:     i18n.T("commands.commands.updateStatusUse"),
	Short:   i18n.T("commands.commands.updateStatusShort"),
	Aliases: []string{"us"},
	Args:    cobra.MaximumNArgs(1),
	Example: i18n.T("commands.commands.updateStatusExample"),
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		taskName := name

		if len(args) > 0 {
			taskName = args[0]
		}

		if _, err := tkm.Complete(cmd.Context(), taskName, notes); err != nil {
			return err
		}

		fmt.Println(i18n.T("tasks.info.updatedSuccessfully"))
		printQuote()
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

		todos, err := tkm.Expired(cmd.Context())
		if err != nil {
			return err
		}

		printExpired(todos)
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

		todos, err := tkm.Expiring(cmd.Context())
		if err != nil {
			return err
		}

		printExpiring(todos)
		return nil
	},
}

// initCmd initializes the local database schema and storage.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: i18n.T("commands.commands.initCmdShort"),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := db.InitDB()
		if err != nil {
			return fmt.Errorf("%s%w", i18n.T("commands.error.openDatabase"), err)
		}

		fmt.Printf("%s", i18n.T("commands.info.initSuccessful"))
		return nil
	},
}

// delByName deletes a single task by name.
var delByName = &cobra.Command{
	Use:     i18n.T("commands.commands.delByNameUse"),
	Short:   i18n.T("commands.commands.delByNameShort"),
	Aliases: []string{"deln", "dn"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteYes {
			return errors.New(i18n.T("commands.error.destructiveAction"))
		}

		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		taskName := name

		if len(args) > 0 {
			taskName = args[0]
		}

		todo, err := tkm.DeleteByName(cmd.Context(), taskName)
		if err != nil {
			return err
		}

		color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.taskDeletedSuccessfully"), todo.Name))
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

		todos, err := tkm.DailyTodos(cmd.Context())
		if err != nil {
			return err
		}

		printTodos(todos)
		return nil
	},
}

// getTodoCurrentStreak returns the current completion streak for a task.
var getTodoCurrentStreak = &cobra.Command{
	Use:     i18n.T("commands.commands.getTodoCurrentStreakUse"),
	Short:   i18n.T("commands.commands.getTodoCurrentStreakShort"),
	Aliases: []string{"curr", "cs"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		taskName := name

		if len(args) > 0 {
			taskName = args[0]
		}

		todo, streak, err := tkm.CurrentStreak(cmd.Context(), taskName)
		if err != nil {
			return err
		}

		fmt.Printf(i18n.T("tasks.info.currentStreak"), todo.Name, streak)
		return nil
	},
}

// getTodoStreakHistory returns the streak history timeline for a task.
var getTodoStreakHistory = &cobra.Command{
	Use:     i18n.T("commands.commands.getTodoStreakHistoryUse"),
	Short:   i18n.T("commands.commands.getTodoStreakHistoryShort"),
	Aliases: []string{"his", "sh"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		taskName := name

		if len(args) > 0 {
			taskName = args[0]
		}

		_, periods, err := tkm.StreakHistory(cmd.Context(), taskName)
		if err != nil {
			return err
		}

		printStreakHistory(periods)
		return nil
	},
}

// debugTask prints the raw history/streak data for a task (diagnostics).
var debugTask = &cobra.Command{
	Use:     i18n.T("commands.commands.debugUse"),
	Short:   i18n.T("commands.commands.debugShort"),
	Aliases: []string{"dbg"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		taskName := name
		if len(args) > 0 {
			taskName = args[0]
		}

		if taskName == "" {
			return errors.New(i18n.T("commands.error.emptyName"))
		}

		report, err := tkm.Debug(cmd.Context(), taskName)
		if err != nil {
			return err
		}

		printDebugReport(report)
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

		report, err := tkm.BackfillDailyHistory(cmd.Context(), daysBackfill, name)
		if err != nil {
			return err
		}

		printBackfillReport(report, name)
		return nil
	},
}

// dump stores a named entry in the dump table, or lists stored entries.
var dumpCmd = &cobra.Command{
	Use:     i18n.T("commands.commands.dumpUse"),
	Short:   i18n.T("commands.commands.dumpShort"),
	Aliases: []string{"dp"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		dumpName := name
		if len(args) > 0 {
			dumpName = args[0]
		}

		if dumpName == "" {
			entries, err := tkm.Dumps(cmd.Context())
			if err != nil {
				return err
			}

			printDumps(entries)
			return nil
		}

		entry, err := tkm.AddDump(cmd.Context(), dumpName)
		if err != nil {
			return err
		}

		color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.addedDump"), entry.Name))
		return nil
	},
}

// editCmd changes the fields of an existing task.
var editCmd = &cobra.Command{
	Use:     i18n.T("commands.commands.editUse"),
	Short:   i18n.T("commands.commands.editShort"),
	Aliases: []string{"e"},
	Example: i18n.T("commands.commands.editExample"),
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		patch, err := editPatch(cmd)
		if err != nil {
			return err
		}
		if patch.Empty() {
			return errors.New(i18n.T("commands.error.nothingToEdit"))
		}

		todo, err := tkm.EditTodo(cmd.Context(), args[0], patch)
		if err != nil {
			return err
		}

		color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.updatedTask"), todo.Name))
		return nil
	},
}

// editPatch builds the patch from the flags that were actually passed, so
// `--desc ""` clears the description while leaving it out keeps it.
func editPatch(cmd *cobra.Command) (tasks.TodoPatch, error) {
	var patch tasks.TodoPatch
	flags := cmd.Flags()

	if flags.Changed("name") {
		patch.Name = &editName
	}
	if flags.Changed("desc") {
		patch.Description = &editDescription
	}
	if flags.Changed("priority") {
		patch.Priority = &editPriority
	}
	if flags.Changed("daily") {
		patch.IsDaily = &editDaily
	}
	if flags.Changed("date") {
		expiresAt, err := tasks.ParseDate(editDate)
		if err != nil {
			return patch, err
		}
		patch.ExpiresAt = &expiresAt
	}
	return patch, nil
}

// progressCmd shows how many of today's tasks are done.
var progressCmd = &cobra.Command{
	Use:     "progress",
	Short:   i18n.T("commands.commands.progressShort"),
	Aliases: []string{"pg"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		progress, err := tkm.TodayProgress(cmd.Context())
		if err != nil {
			return err
		}

		printProgress(progress)
		return nil
	},
}

// tuiCmd opens the interactive terminal interface.
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: i18n.T("commands.commands.tuiShort"),
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		return tui.Run(cmd.Context(), tkm)
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

			found := slices.Contains(availableThemes, action)

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
			found := slices.Contains(availableLanguages, action)

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

// configKeymap selects the TUI key layout.
var configKeymap = &cobra.Command{
	Use:     i18n.T("commands.commands.configKeymapUse"),
	Short:   i18n.T("commands.commands.configKeymapShort"),
	Aliases: []string{"k"},
	Example: "akrasia config keymap vim\nakrasia config keymap list\nakrasia config keymap show",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		switch action := args[0]; action {
		case "list":
			fmt.Println(i18n.T("commands.info.availableKeymaps"))
			for _, keymap := range tui.Keymaps() {
				fmt.Printf(" - %s\n", keymap)
			}
			return nil

		case "show":
			fmt.Printf(i18n.T("commands.info.currentKeymap"), tui.CurrentKeymap())
			return nil

		default:
			if !slices.Contains(tui.Keymaps(), action) {
				return fmt.Errorf(i18n.T("commands.error.unknownKeymap"), action, tui.Keymaps())
			}

			if err := tui.SetKeymap(action); err != nil {
				return fmt.Errorf(i18n.T("commands.error.setKeymap"), err)
			}
			fmt.Printf(i18n.T("commands.info.keymapSet"), action)
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
		"checkExpiring":           checkExpiring,
		"checkExpired":            checkExpired,
		"initCmd":                 initCmd,
		"delByName":               delByName,
		"getAllDaily":             getAllDaily,
		"getTodoCurrentStreak":    getTodoCurrentStreak,
		"getTodoStreakHistory":    getTodoStreakHistory,
		"backfillHistory":         backfillHistory,
		"debugTask":               debugTask,
		"dumpCmd":                 dumpCmd,
		"tuiCmd":                  tuiCmd,
		"progressCmd":             progressCmd,
		"editCmd":                 editCmd,
		"config":                  config,
	}

	for _, cmd := range commands {
		rootCmd.AddCommand(cmd)
	}

	// Set usage func for localised help
	rootCmd.SetUsageFunc(UsageFunc)

	// Add subcommands to config
	config.AddCommand(configTheme, configLanguage, configKeymap)

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
	getTodoCurrentStreak.Flags().StringVar(&name, "name", "", i18n.T("commands.flags.getTodoCurrentStreakName"))
	getTodoStreakHistory.Flags().StringVar(&name, "name", "", i18n.T("commands.flags.getTodoStreakHistoryName"))
	backfillHistory.Flags().IntVar(&daysBackfill, "days", 30, i18n.T("commands.flags.backfillHistoryDays"))
	backfillHistory.Flags().StringVar(&name, "task", "", i18n.T("commands.flags.backfillHistoryName"))
	debugTask.Flags().StringVar(&name, "task", "", i18n.T("commands.flags.debugName"))
	dumpCmd.Flags().StringVar(&name, "name", "", i18n.T("commands.flags.dumpName"))
	editCmd.Flags().StringVar(&editName, "name", "", i18n.T("commands.flags.editName"))
	editCmd.Flags().StringVar(&editDescription, "desc", "", i18n.T("commands.flags.editDescription"))
	editCmd.Flags().StringVar(&editPriority, "priority", "", i18n.T("commands.flags.editPriority"))
	editCmd.Flags().BoolVar(&editDaily, "daily", false, i18n.T("commands.flags.editDaily"))
	editCmd.Flags().IntSliceVar(&editDate, "date", nil, i18n.T("commands.flags.editDate"))
	today.Flags().StringVar(&todayOnly, "only", "", i18n.T("commands.flags.todayOnly"))
	today.Flags().IntVar(&todayLimit, "limit", 0, i18n.T("commands.flags.todayLimit"))
	today.Flags().BoolVar(&todayJSON, "json", false, i18n.T("commands.flags.todayJSON"))
}
