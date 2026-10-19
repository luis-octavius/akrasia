package tasks

import (
	"context"
	"fmt"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// CurrentStreak returns the task that best matches name and its current
// streak in days.
func (tkm *TaskManager) CurrentStreak(ctx context.Context, name string) (TodoView, int64, error) {
	todo, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, 0, err
	}

	streak, err := tkm.Queries.GetCurrentStreak(ctx, database.GetCurrentStreakParams{
		TodoID: todo.ID,
		ID:     todo.ID,
	})
	if err != nil {
		return TodoView{}, 0, fmt.Errorf("%s: %w", i18n.T("tasks.error.getCurrentStreak"), err)
	}

	view, err := viewFromNameRow(todo)
	return view, streak, err
}

// StreakHistory returns the task that best matches name and its completed
// streak intervals, longest first.
func (tkm *TaskManager) StreakHistory(ctx context.Context, name string) (TodoView, []StreakPeriod, error) {
	todo, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, nil, err
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
