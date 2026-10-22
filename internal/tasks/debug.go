package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// Debug returns the exact data the streak queries operate on: the task's
// streak window (history_since), every history row flagged by whether it
// falls inside that window, and the resulting streak. It exists so that
// "why is my streak N?" can be answered by looking at the database instead
// of guessing.
func (tkm *TaskManager) Debug(ctx context.Context, name string) (DebugReport, error) {
	todo, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return DebugReport{}, err
	}

	rows, err := tkm.Queries.GetHistoryByTodo(ctx, todo.ID)
	if err != nil {
		return DebugReport{}, fmt.Errorf("%s: %w", i18n.T("tasks.error.getStreakHistory"), err)
	}

	streak, err := tkm.currentStreak(ctx, todo, time.Now())
	if err != nil {
		return DebugReport{}, err
	}

	view, err := viewFromNameRow(todo)
	if err != nil {
		return DebugReport{}, err
	}

	report := DebugReport{
		Todo:      view,
		CreatedAt: todo.CreatedAt,
		Streak:    streak,
	}
	if todo.HistorySince.Valid {
		report.HistorySince = todo.HistorySince.Time.Format(time.DateOnly)
	}

	for _, r := range rows {
		report.History = append(report.History, HistoryEntry{
			Date:      r.Date,
			Completed: r.Completed.Bool,
			Notes:     r.Notes.String,
			InStreak:  report.HistorySince != "" && r.Date >= report.HistorySince,
		})
	}

	return report, nil
}
