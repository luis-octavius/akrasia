package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/luis-octavius/akrasia/internal/backup"
	"github.com/luis-octavius/akrasia/internal/notify"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
	"golang.org/x/term"
)

// printTodo prints one task in the standard list format.
func printTodo(todo tasks.TodoView) {
	lipgloss.Println(renderTodo(todo))
}

func printTodos(todos []tasks.TodoView) {
	for _, todo := range todos {
		printTodo(todo)
	}
}

// renderTodo mounts the string to beautifully output the task
func renderTodo(todo tasks.TodoView) string {
	icon, status := "○", i18n.T("tasks.tasks.notDone")

	if todo.Done() {
		icon, status = "✔", i18n.T("tasks.tasks.done")
	}

	state := color.Muted()
	switch {
	case todo.Done():
		state = color.Success()
	case tasks.IsExpiringSoon(todo.ExpiresAt):
		state = color.Warning()
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s %s", state.Render(icon), color.Title().Render(todo.Name))
	if len(todo.Subtasks) > 0 {
		p := todo.SubtaskProgress()
		b.WriteString(color.Muted().Render(fmt.Sprintf("  %d/%d", p.Done, p.Total)))
	}
	b.WriteString("\n")

	if desc := todo.Description.String; desc != "" {
		fmt.Fprintf(&b, "- %s\n", desc)
	}

	if !todo.IsDaily {
		expires := fmt.Sprintf("%s · %s\n", dueLabel(todo), status)
		fmt.Fprintf(&b, "  %s\n", color.Muted().Render(expires))
	} else if schedule := scheduleLabel(todo); schedule != "" {
		fmt.Fprintf(&b, "  %s\n", color.Muted().Render(schedule))
	}

	if priority := todo.Priority; priority != "" {
		fmt.Fprintf(&b, "  %s\n", color.Muted().Render("Priority: "+todo.Priority))
	}

	for _, sub := range todo.Subtasks {
		b.WriteString(renderSubtask(sub))
	}

	return b.String()
}

// renderSubtask is the one-line form of a subtask, indented under its
// parent: state, name and, for a one-off one, its due date.
func renderSubtask(sub tasks.TodoView) string {
	icon, state := "○", color.Muted()
	if sub.Done() {
		icon, state = "✔", color.Success()
	}

	line := "    " + state.Render(icon) + " " + sub.Name
	if !sub.IsDaily {
		line += color.Muted().Render(" · " + sub.ExpiresAt.Format("02 Jan") + timeSuffix(sub))
	} else if sub.DueTime != "" {
		line += color.Muted().Render(" · " + sub.DueTime)
	}
	return line + "\n"
}

// scheduleLabel is how a recurring task repeats and its time of day, as
// "weekly on Fri · at 18:00". A plain daily task without a time has none:
// it is already under the daily heading.
func scheduleLabel(todo tasks.TodoView) string {
	var parts []string
	if todo.Recurrence == tasks.RecurWeekly || todo.Recurrence == tasks.RecurMonthly {
		parts = append(parts, todo.Repeat())
	}
	if todo.DueTime != "" {
		parts = append(parts, fmt.Sprintf(i18n.T("tasks.tasks.atTime"), todo.DueTime))
	}
	return strings.Join(parts, " · ")
}

// streakUnit names what a streak counts for a weekly or monthly task; it
// is empty for a daily one, whose streaks print as before.
func streakUnit(todo tasks.TodoView) string {
	switch todo.Recurrence {
	case tasks.RecurWeekly:
		return i18n.T("tasks.tasks.streakUnitWeekly")
	case tasks.RecurMonthly:
		return i18n.T("tasks.tasks.streakUnitMonthly")
	}
	return ""
}

// dueLabel is a one-off task's due date, with its time of day when it has one.
func dueLabel(todo tasks.TodoView) string {
	return todo.ExpiresAt.Format("02 Jan 06") + timeSuffix(todo)
}

func timeSuffix(todo tasks.TodoView) string {
	if todo.DueTime == "" {
		return ""
	}
	return " " + todo.DueTime
}

// progressBarWidth is the width of the CLI progress bar, percentage included.
const progressBarWidth = 30

// progressLine is "Today 3/5  ▌▌▌▌▌░░░  60%" in the theme's colors.
func progressLine(p tasks.Progress) string {
	label := fmt.Sprintf(i18n.T("tasks.tasks.progress"), p.Done, p.Total)
	if p.Total == 0 {
		label = i18n.T("tasks.tasks.progressClear")
	}
	return color.Title().Render(label) + "  " + color.ProgressBar(p.Ratio(), progressBarWidth)
}

func printProgress(p tasks.Progress) {
	lipgloss.Println(progressLine(p))
}

func printTodayView(view tasks.TodayView, progress tasks.Progress, asJSON bool) error {
	if asJSON {
		payload, err := json.MarshalIndent(struct {
			tasks.TodayView
			Progress tasks.Progress
		}{view, progress}, "", "  ")
		if err != nil {
			return fmt.Errorf(i18n.T("tasks.error.encodeTodayJSON"), err)
		}

		fmt.Println(string(payload))
		return nil
	}

	printProgress(progress)
	fmt.Println()

	if view.Empty() {
		color.MsgSuccess(i18n.T("tasks.info.noPendingTasks"))
		return nil
	}

	fmt.Println(i18n.T("tasks.tasks.todayFocus"))

	printTodaySection(i18n.T("tasks.tasks.overdue"), view.Overdue)
	printTodaySection(i18n.T("tasks.tasks.dueToday"), view.DueToday)
	printTodaySection(tasks.RecurringTitle(view.Daily), view.Daily)
	printTodaySection(i18n.T("tasks.tasks.expiringSoon"), view.ExpiringSoon)

	return nil
}

// printTodaySection renders one titled section from the today dashboard.
func printTodaySection(title string, todos []tasks.TodoView) {
	if len(todos) == 0 {
		return
	}

	fmt.Printf("\n%s (%d)\n", title, len(todos))
	printTodos(todos)
}

func printFocus(todos []tasks.TodoView, progress tasks.Progress) {
	printProgress(progress)
	fmt.Println()

	if len(todos) == 0 {
		color.MsgSuccess(i18n.T("tasks.info.noFocusTasks"))
		return
	}

	fmt.Printf(i18n.T("tasks.tasks.focus"), len(todos))
	printTodos(todos)
}

func printFoundTodo(todo tasks.TodoView) {
	if todo.IsSubtask() {
		color.MsgMuted(fmt.Sprintf(i18n.T("tasks.info.subtaskOf"), todo.Parent))
	}
	s := fmt.Sprintf(i18n.T("tasks.tasks.getTaskByName"), todo.Name, todo.Description.String, todo.ExpiresAt.Format(time.RFC1123))
	color.MsgSuccess(s)
	for _, sub := range todo.Subtasks {
		lipgloss.Print(renderSubtask(sub))
	}
}

func printExpired(todos []tasks.TodoView) {
	if len(todos) == 0 {
		fmt.Println(i18n.T("tasks.info.NoExpiring"))
		return
	}

	fmt.Println(i18n.T("tasks.tasks.expired"))
	printTodos(todos)
}

func printExpiring(todos []tasks.TodoView) {
	if len(todos) == 0 {
		color.MsgError(i18n.T("tasks.info.NoExpiring"))
		return
	}

	fmt.Println(i18n.T("tasks.tasks.expiring"))
	printTodos(todos)
}

func printStreakHistory(todo tasks.TodoView, periods []tasks.StreakPeriod) {
	unit := streakUnit(todo)
	for i, p := range periods {
		if unit != "" {
			fmt.Printf(i18n.T("tasks.tasks.getStreakHistoryPeriods"), i+1, p.StartDate, p.EndDate, unit, p.Days)
			continue
		}
		fmt.Printf(i18n.T("tasks.tasks.getStreakHistory"), i+1, p.StartDate, p.EndDate, p.Days)
	}
}

func printDebugReport(r tasks.DebugReport) {
	historySince := r.HistorySince
	if historySince == "" {
		historySince = i18n.T("tasks.debug.noWindow")
	}

	fmt.Printf(i18n.T("tasks.debug.title"), r.Todo.Name)
	fmt.Printf(i18n.T("tasks.debug.meta"),
		r.Todo.ID, r.Todo.IsDaily,
		r.CreatedAt.Format(time.DateOnly), historySince)

	fmt.Printf(i18n.T("tasks.debug.historyHeader"), len(r.History))
	for _, h := range r.History {
		fmt.Printf(i18n.T("tasks.debug.historyRow"), h.Date, h.Completed, h.Notes, h.InStreak)
	}

	fmt.Printf(i18n.T("tasks.debug.streak"), r.Streak, historySince)
}

func printBackfillReport(r tasks.BackfillReport, taskName string) {
	if len(r.Tasks) == 0 {
		color.MsgError(i18n.T("tasks.info.noDailyToBackfill"))
		return
	}

	for _, t := range r.Tasks {
		fmt.Printf("  %s: ", t.Name)

		for _, f := range t.Failures {
			fmt.Printf(i18n.T("tasks.info.cantBackfillTask"), t.Name, f.Date, f.Err)
		}

		if t.Inserted > 0 {
			fmt.Printf(" %d days marked as completed\n", t.Inserted)
		} else {
			fmt.Println(" no days to backfill")
		}
	}

	if taskName != "" {
		color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.backfilledTask"), r.Total(), taskName))
	} else {
		color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.backfilledDailyTask"), r.Total(), len(r.Tasks)))
	}
}

// printNotes lists a task's completed days, newest first, with what was
// written on each. streak is shown only for daily tasks.
func printNotes(todo tasks.TodoView, streak int64, entries []tasks.NoteEntry) {
	title := fmt.Sprintf(i18n.T("tasks.tasks.notesTitle"), todo.Name)
	if todo.IsDaily {
		label := fmt.Sprintf(i18n.T("tasks.tasks.notesStreak"), streak)
		if unit := streakUnit(todo); unit != "" {
			label = fmt.Sprintf(i18n.T("tasks.tasks.notesStreakPeriods"), streak, unit)
		}
		title += color.Muted().Render(" (" + label + ")")
	}
	lipgloss.Println(color.Title().Render(title))

	if len(entries) == 0 {
		color.MsgMuted(i18n.T("tasks.tasks.noNotesYet"))
		return
	}

	for _, e := range entries {
		text := e.Notes
		switch {
		case e.Backfilled:
			text = color.Muted().Render(i18n.T("tasks.tasks.backfilled"))
		case text == "":
			text = color.Muted().Render(i18n.T("tasks.tasks.noNotes"))
		}
		lipgloss.Println(color.Accent().UnsetBold().Render(e.Date) + "  " + text)
	}
}

func printDumps(entries []tasks.DumpEntry) {
	if len(entries) == 0 {
		fmt.Println(i18n.T("tasks.info.noDump"))
		return
	}

	// the numbers are what --delete and --to-task take
	for i, d := range entries {
		createdAt := ""
		if !d.CreatedAt.IsZero() {
			// stored in local time, so no zone is shown
			createdAt = d.CreatedAt.Format("02 Jan 06 15:04")
		}
		fmt.Printf("%d. "+i18n.T("tasks.dump.row"), i+1, d.Name, createdAt)
	}
}

func printBackups(dir, profile string, copies []backup.Copy) {
	color.MsgTitle(fmt.Sprintf(i18n.T("commands.info.backupList"), profile, dir))
	if len(copies) == 0 {
		color.MsgMuted(i18n.T("commands.info.noBackups"))
		return
	}

	for _, c := range copies {
		fmt.Printf("  %s  %8s  %s\n", c.Time.Format("2006-01-02 15:04:05"), formatSize(c.Size), filepath.Base(c.Path))
	}
}

func printBackupConfig(cfg backup.Config) {
	dir := color.Muted().Render(i18n.T("commands.info.backupNotSet"))
	last := color.Muted().Render(i18n.T("commands.info.backupNever"))
	if cfg.Enabled() {
		dir = cfg.Dir
		if latest, ok, err := backup.Latest(cfg.Dir, cfg.ProfileName()); err == nil && ok {
			last = latest.Time.Format("2006-01-02 15:04")
		}
	}

	lipgloss.Printf(i18n.T("commands.info.backupConfig"), dir, cfg.KeepN(), cfg.ProfileName(), last)
}

func printNotifyConfig(cfg notify.Config) {
	webhook := color.Muted().Render(i18n.T("commands.info.notifyNotSet"))
	if cfg.Enabled() {
		// the end of the URL is a secret; the start is enough to recognise it
		webhook = cfg.Webhook[:strings.LastIndex(cfg.Webhook, "/")+1] + "…"
	}
	mention := color.Muted().Render(i18n.T("commands.info.notifyNoMention"))
	if cfg.Mention != "" {
		mention = cfg.Mention
	}

	lipgloss.Printf(i18n.T("commands.info.notifyConfig"),
		webhook, mention, cfg.DaysAhead(), strings.Join(cfg.Slots(), ", "), cfg.DailySlot())
}

// formatSize shows a file size in B, KB or MB.
func formatSize(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// printQuote prints a random quote wrapped to the terminal width.
func printQuote() {
	wrapped := wrapText(tasks.RandomQuote(), getTerminalWidth())
	color.MsgQuote(fmt.Sprintf("\n%s\n", wrapped))
}

// getTerminalWidth detects the current terminal width, with a safe minimum
func getTerminalWidth() int {
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width
	}
	// Default fallback width
	return 80
}

// wrapText wraps text to fit within terminal width with proper word boundaries
func wrapText(text string, maxWidth int) string {
	// Leave some margin for padding/formatting
	effectiveWidth := maxWidth - 4
	if effectiveWidth < 40 {
		effectiveWidth = 40
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		wordLen := len(word)
		currentLen := currentLine.Len()

		// If adding word would exceed width, start new line
		if currentLen > 0 && currentLen+1+wordLen > effectiveWidth {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}

		// Add space if not first word on line
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}

		currentLine.WriteString(word)
	}

	// Add remaining line
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}
