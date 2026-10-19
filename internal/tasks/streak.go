package tasks

import (
	"context"
	"fmt"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// getCurrentStreak prints the current streak count for a named task.
func (tkm *TaskManager) GetCurrentStreak(name string) error {
	todo, err := GetTodoByName(tkm, name)
	if err != nil {
		return err
	}

	streak, err := tkm.Queries.GetCurrentStreak(context.Background(), database.GetCurrentStreakParams{
		TodoID: todo.ID,
		ID:     todo.ID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("tasks.error.getCurrentStreak"), err)
	}

	fmt.Printf(i18n.T("tasks.info.currentStreak"), todo.Name, streak)
	return nil
}

// getStreakHistory prints historical streak intervals for a named task.
func (tkm *TaskManager) GetStreakHistory(name string) error {
	todo, err := GetTodoByName(tkm, name)
	if err != nil {
		return err
	}

	streak_history, err := tkm.Queries.GetStreakHistory(context.Background(), database.GetStreakHistoryParams{
		TodoID: todo.ID,
		ID:     todo.ID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("tasks.error.getStreakHistory"), err)
	}

	for i, streak := range streak_history {
		fmt.Printf(i18n.T("tasks.tasks.getStreakHistory"), i+1, streak.StartDate, streak.EndDate, streak.StreakLength)
	}

	return nil
}
