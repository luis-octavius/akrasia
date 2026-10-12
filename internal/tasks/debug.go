package tasks

import (
	"context"
	"fmt"
	"time"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// DebugTask prints the exact data the streak queries operate on: the task's
// streak window (history_since), every history row flagged by whether it
// falls inside that window, and the resulting streak. It exists so that
// "why is my streak N?" can be answered by looking at the database instead
// of guessing.
func (tkm *TaskManager) DebugTask(name string) error {
	ctx := context.Background()

	todo, err := GetTodoByName(tkm, name)
	if err != nil {
		return fmt.Errorf("%s", i18n.T("tasks.error.getTodoByNameStreak"))
	}

	rows, err := tkm.Queries.GetHistoryByTodo(ctx, todo.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("tasks.error.getStreakHistory"), err)
	}

	streak, err := tkm.Queries.GetCurrentStreak(ctx, database.GetCurrentStreakParams{
		TodoID: todo.ID,
		ID:     todo.ID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("tasks.error.getCurrentStreak"), err)
	}

	historySince := i18n.T("tasks.debug.noWindow")
	if todo.HistorySince.Valid {
		historySince = todo.HistorySince.Time.Format(time.DateOnly)
	}

	fmt.Printf(i18n.T("tasks.debug.title"), todo.Name)
	fmt.Printf(i18n.T("tasks.debug.meta"),
		todo.ID, todo.IsDaily,
		todo.CreatedAt.Format(time.DateOnly), historySince)

	fmt.Printf(i18n.T("tasks.debug.historyHeader"), len(rows))
	for _, r := range rows {
		inStreak := todo.HistorySince.Valid && r.Date >= historySince
		notes := r.Notes.String
		fmt.Printf(i18n.T("tasks.debug.historyRow"), r.Date, r.Completed.Bool, notes, inStreak)
	}

	fmt.Printf(i18n.T("tasks.debug.streak"), streak, historySince)
	return nil
}
