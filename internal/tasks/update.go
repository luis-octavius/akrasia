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

// Complete marks the task that best matches name as done, writes an entry
// to history and returns the updated task.
func (tkm *TaskManager) Complete(ctx context.Context, name, notes string) (TodoView, error) {
	todo, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, fmt.Errorf(i18n.T("tasks.error.updateTaskStatus"), name, err)
	}

	// only updates status to concluded if is not a daily task
	if !todo.IsDaily {
		_, err = tkm.Queries.UpdateTodoStatusById(ctx, todo.ID)
		if err != nil {
			return TodoView{}, fmt.Errorf(i18n.T("tasks.error.updateTaskStatus"), name, err)
		}
		todo.Concluded = true
	}

	_, err = tkm.Queries.AddTodoHistory(ctx, database.AddTodoHistoryParams{
		ID:          uuid.New(),
		TodoID:      todo.ID,
		Completed:   sql.NullBool{Bool: true, Valid: true},
		CompletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Notes:       sql.NullString{String: notes, Valid: true},
	})
	if err != nil {
		return TodoView{}, fmt.Errorf("%s: %w", i18n.T("tasks.error.createTaskHistory"), err)
	}

	view, err := viewFromNameRow(todo)
	if err != nil {
		return TodoView{}, err
	}
	view.DoneToday = true

	return view, nil
}
