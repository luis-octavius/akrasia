package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// BackfillDailyHistory records real completions for days the user forgot to
// log. It excludes today — use `done` for that — and tags each row with
// notes="backfilled" purely as a note to the user, not a semantic marker
// the streak queries treat specially. Days that fail to insert are reported
// in the result instead of aborting the run.
func (tkm *TaskManager) BackfillDailyHistory(ctx context.Context, daysBack int, taskName string) (BackfillReport, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startDate := today.AddDate(0, 0, -daysBack)

	// Get all daily tasks, or a specific one if taskName is provided
	var dailyTasks []database.Todo
	var err error

	if taskName != "" {
		todo, err := findTodoByName(ctx, tkm, taskName)
		if err != nil {
			return BackfillReport{}, fmt.Errorf(i18n.T("tasks.error.findTask"), taskName, err)
		}

		if !todo.IsDaily {
			return BackfillReport{}, localizedError{
				msg:      fmt.Sprintf(i18n.T("tasks.error.taskNotDaily"), taskName),
				sentinel: ErrNotDaily,
			}
		}

		dailyTasks = []database.Todo{{ID: todo.ID, Name: todo.Name, IsDaily: todo.IsDaily}}
	} else {
		dailyTasks, err = tkm.Queries.GetDailyTodos(ctx)
		if err != nil {
			return BackfillReport{}, fmt.Errorf(i18n.T("tasks.error.getDaily"), err)
		}
	}

	var report BackfillReport

	for _, task := range dailyTasks {
		// Start at startDate (today - daysBack). We deliberately do NOT
		// clamp to task.CreatedAt: users can legitimately record days from
		// before they added the task to Akrasia.
		result := BackfillTaskResult{Name: task.Name}

		// Insert a real completion for each day from startDate to
		// yesterday (exclude today — that's what `done` is for). There is
		// no "neutral" state anymore: dates before task.history_since are
		// simply never looked at by the streak queries, so nothing here
		// can accidentally overwrite a genuine miss.
		for day := startDate; day.Before(today); day = day.AddDate(0, 0, 1) {
			dateStr := day.Format(time.DateOnly)

			_, err := tkm.Queries.BackfillDailyDone(ctx, database.BackfillDailyDoneParams{
				ID:     uuid.New(),
				TodoID: task.ID,
				Date:   dateStr,
			})
			if err != nil {
				result.Failures = append(result.Failures, BackfillFailure{Date: dateStr, Err: err})
				continue
			}
			result.Inserted++
		}

		if result.Inserted > 0 {
			// Pull the streak lower bound back to the earliest day we just
			// recorded, so the streak queries actually count these days.
			if err := tkm.Queries.ExtendHistorySince(ctx, database.ExtendHistorySinceParams{
				Earliest: sql.NullString{String: startDate.Format(time.DateOnly), Valid: true},
				ID:       task.ID,
			}); err != nil {
				return report, fmt.Errorf("%s: %w", i18n.T("tasks.error.extendHistorySince"), err)
			}
		}

		report.Tasks = append(report.Tasks, result)
	}

	return report, nil
}
