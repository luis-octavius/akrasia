package tasks

import (
	"context"
	"fmt"
	"sort"
	"time"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// CurrentStreak returns the task that best matches name and its current
// streak: days for a daily task, weeks for a weekly one, months for a
// monthly one.
func (tkm *TaskManager) CurrentStreak(ctx context.Context, name string) (TodoView, int64, error) {
	todo, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, 0, err
	}

	streak, err := tkm.currentStreak(ctx, todo, time.Now())
	if err != nil {
		return TodoView{}, 0, err
	}

	view, err := viewFromNameRow(todo)
	return view, streak, err
}

func (tkm *TaskManager) currentStreak(ctx context.Context, todo database.GetTodoByNameRow, now time.Time) (int64, error) {
	if rec := recurrenceOf(todo.IsDaily, todo.Recurrence); rec == RecurWeekly || rec == RecurMonthly {
		runs, err := tkm.periodRuns(ctx, todo, rec)
		if err != nil {
			return 0, err
		}
		return currentPeriodStreak(runs, rec, now), nil
	}

	streak, err := tkm.Queries.GetCurrentStreak(ctx, database.GetCurrentStreakParams{
		TodoID: todo.ID,
		ID:     todo.ID,
	})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", i18n.T("tasks.error.getCurrentStreak"), err)
	}
	return streak, nil
}

// StreakHistory returns the task that best matches name and its completed
// streak intervals, longest first.
func (tkm *TaskManager) StreakHistory(ctx context.Context, name string) (TodoView, []StreakPeriod, error) {
	todo, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, nil, err
	}

	if rec := recurrenceOf(todo.IsDaily, todo.Recurrence); rec == RecurWeekly || rec == RecurMonthly {
		runs, err := tkm.periodRuns(ctx, todo, rec)
		if err != nil {
			return TodoView{}, nil, err
		}
		view, err := viewFromNameRow(todo)
		return view, periodStreakHistory(runs), err
	}

	rows, err := tkm.Queries.GetStreakHistory(ctx, database.GetStreakHistoryParams{
		TodoID: todo.ID,
		ID:     todo.ID,
	})
	if err != nil {
		return TodoView{}, nil, fmt.Errorf("%s: %w", i18n.T("tasks.error.getStreakHistory"), err)
	}

	periods := make([]StreakPeriod, 0, len(rows))
	for _, r := range rows {
		periods = append(periods, StreakPeriod{
			StartDate: fmt.Sprint(r.StartDate),
			EndDate:   fmt.Sprint(r.EndDate),
			Days:      r.StreakLength,
		})
	}

	view, err := viewFromNameRow(todo)
	return view, periods, err
}

// periodStart is the first day of the week (Monday) or month the date is in.
func periodStart(t time.Time, recurrence string) time.Time {
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	if recurrence == RecurMonthly {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	return day.AddDate(0, 0, 1-ISOWeekday(day))
}

// nextPeriod is the start of the period after the one starting at start.
func nextPeriod(start time.Time, recurrence string) time.Time {
	if recurrence == RecurMonthly {
		return start.AddDate(0, 1, 0)
	}
	return start.AddDate(0, 0, 7)
}

// periodRun is a run of consecutive completed periods; first and last are
// the start dates of its first and last period.
type periodRun struct {
	first, last time.Time
	length      int64
}

// periodRuns groups the completed history of a weekly or monthly task into
// runs of consecutive weeks or months, oldest first. Like the daily
// streak, it ignores dates before the task's history_since.
func (tkm *TaskManager) periodRuns(ctx context.Context, todo database.GetTodoByNameRow, recurrence string) ([]periodRun, error) {
	rows, err := tkm.Queries.GetHistoryByTodo(ctx, todo.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("tasks.error.getStreakHistory"), err)
	}

	seen := make(map[time.Time]bool)
	for _, r := range rows {
		if !r.Completed.Bool {
			continue
		}
		day, err := time.Parse(time.DateOnly, r.Date)
		if err != nil || (todo.HistorySince.Valid && r.Date < todo.HistorySince.Time.Format(time.DateOnly)) {
			continue
		}
		seen[periodStart(day, recurrence)] = true
	}

	starts := make([]time.Time, 0, len(seen))
	for s := range seen {
		starts = append(starts, s)
	}
	sort.Slice(starts, func(i, j int) bool { return starts[i].Before(starts[j]) })

	var runs []periodRun
	for _, s := range starts {
		if n := len(runs); n > 0 && nextPeriod(runs[n-1].last, recurrence).Equal(s) {
			runs[n-1].last = s
			runs[n-1].length++
			continue
		}
		runs = append(runs, periodRun{first: s, last: s, length: 1})
	}
	return runs, nil
}

// currentPeriodStreak is the length of the run that reaches this period or
// the one before it: the period in progress can still be done, so not
// having done it yet does not break the streak.
func currentPeriodStreak(runs []periodRun, recurrence string, now time.Time) int64 {
	if len(runs) == 0 {
		return 0
	}
	current := periodStart(now, recurrence)
	last := runs[len(runs)-1]
	if last.last.Equal(current) || nextPeriod(last.last, recurrence).Equal(current) {
		return last.length
	}
	return 0
}

// periodStreakHistory lists the runs longest first, the latest first on a
// tie, like the daily streak history.
func periodStreakHistory(runs []periodRun) []StreakPeriod {
	sorted := append([]periodRun(nil), runs...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].length != sorted[j].length {
			return sorted[i].length > sorted[j].length
		}
		return sorted[i].first.After(sorted[j].first)
	})

	periods := make([]StreakPeriod, 0, len(sorted))
	for _, r := range sorted {
		periods = append(periods, StreakPeriod{
			StartDate: r.first.Format(time.DateOnly),
			EndDate:   r.last.Format(time.DateOnly),
			Days:      r.length,
		})
	}
	return periods
}
