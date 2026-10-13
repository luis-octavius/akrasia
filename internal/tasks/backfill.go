package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// backfillDailyHistory records real completions for days the user forgot to
// log. It excludes today — use `done` for that — and tags each row with
// notes="backfilled" purely as a note to the user, not a semantic marker
// the streak queries treat specially.
func (tkm *TaskManager) BackfillDailyHistory(daysBack int, taskName string) error {
	ctx := context.Background()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startDate := today.AddDate(0, 0, -daysBack)

	// Get all daily tasks, or a specific one if taskName is provided
	var dailyTasks []database.Todo
	var err error

	if taskName != "" {
		todo, err := GetTodoByName(tkm, taskName)
		if err != nil {
			return fmt.Errorf(i18n.T("tasks.error.findTask"), taskName, err)
		}

		if !todo.IsDaily {
			return fmt.Errorf(i18n.T("tasks.error.taskNotDaily"), taskName)
		}

		dailyTasks = []database.Todo{
			{
				ID:          todo.ID,
				Name:        todo.Name,
				Description: todo.Description,
				CreatedAt:   todo.CreatedAt,
				UpdatedAt:   todo.UpdatedAt,
				Concluded:   todo.Concluded,
				ExpiresAt:   todo.ExpiresAt,
				Priority:    todo.Priority,
				IsDaily:     todo.IsDaily,
			},
		}
	} else {
		// Get all daily tasks
		dailyTasks, err = tkm.Queries.GetDailyTodos(ctx)
		if err != nil {
			return fmt.Errorf(i18n.T("tasks.error.getDaily"), err)
		}
	}

	if len(dailyTasks) == 0 {
		color.MsgError(i18n.T("tasks.info.noDailyToBackfill"))
		return nil
	}

	totalBackfilled := 0

	for _, task := range dailyTasks {
		// Start at startDate (today - daysBack). We deliberately do NOT
		// clamp to task.CreatedAt: users can legitimately record days from
		// before they added the task to Akrasia.
		backfillStart := startDate

		fmt.Printf("  %s: ", task.Name)

		taskInserted := 0

		// Insert a real completion for each day from backfillStart to
		// yesterday (exclude today — that's what `done` is for). There is
		// no "neutral" state anymore: dates before task.history_since are
		// simply never looked at by the streak queries, so nothing here
		// can accidentally overwrite a genuine miss.
		currentDate := backfillStart
		dayCount := 0
		for currentDate.Before(today) {
			dateOnly := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, now.Location())
			dateStr := dateOnly.Format(time.DateOnly)

			_, err := tkm.Queries.BackfillDailyDone(ctx, database.BackfillDailyDoneParams{
				ID:     uuid.New(),
				TodoID: task.ID,
				Date:   dateStr,
			})

			if err != nil {
				fmt.Printf(i18n.T("tasks.info.cantBackfillTask"), task.Name, dateOnly.Format(time.DateOnly), err)
			} else {
				taskInserted++
			}

			dayCount++
			if dayCount%10 == 0 {
				fmt.Print(".")
			}

			currentDate = currentDate.AddDate(0, 0, 1)
		}

		totalBackfilled += taskInserted

		if taskInserted > 0 {
			// Pull the streak lower bound back to the earliest day we just
			// recorded, so the streak queries actually count these days.
			if err := tkm.Queries.ExtendHistorySince(ctx, database.ExtendHistorySinceParams{
				Earliest: sql.NullString{String: backfillStart.Format(time.DateOnly), Valid: true},
				ID:       task.ID,
			}); err != nil {
				return fmt.Errorf("%s: %w", i18n.T("tasks.error.extendHistorySince"), err)
			}
			fmt.Printf(" %d days marked as completed\n", taskInserted)
		} else {
			fmt.Println(" no days to backfill")
		}
	}

	if taskName != "" {
		color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.backfilledTask"), totalBackfilled, taskName))
	} else {
		color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.backfilledDailyTask"), totalBackfilled, len(dailyTasks)))
	}

	return nil
}
