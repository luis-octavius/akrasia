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

// updateToConcluded marks a task as done and writes an entry to history.
func (tkm *TaskManager) UpdateToConcluded(name, notes string) error {
	todo, err := GetTodoByName(tkm, name)
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.updateTaskStatus"), name, err)
	}

	// only updates status to concluded if is not a daily task
	if !todo.IsDaily {
		_, err = tkm.Queries.UpdateTodoStatusById(context.Background(), todo.ID)
		if err != nil {
			return fmt.Errorf(i18n.T("tasks.error.updateTaskStatus"), name, err)
		}
	}

	_, err = tkm.Queries.AddTodoHistory(context.Background(), database.AddTodoHistoryParams{
		ID:          uuid.New(),
		TodoID:      todo.ID,
		Completed:   sql.NullBool{Bool: true, Valid: true},
		CompletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Notes:       sql.NullString{String: notes, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("tasks.error.createTaskHistory"), err)
	}

	fmt.Println(i18n.T("tasks.info.updatedSuccessfully"))
	generateRandomQuote()
	return nil
}
