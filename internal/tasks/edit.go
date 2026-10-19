package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// TodoPatch lists the fields to change on a task; a nil field keeps the
// current value. An empty Description or Priority clears it.
type TodoPatch struct {
	Name        *string
	Description *string
	Priority    *string
	IsDaily     *bool
	ExpiresAt   *time.Time
}

// Empty reports whether the patch changes nothing.
func (p TodoPatch) Empty() bool {
	return p.Name == nil && p.Description == nil && p.Priority == nil &&
		p.IsDaily == nil && p.ExpiresAt == nil
}

// EditTodo applies patch to the task that best matches name and returns
// the updated task. The history is kept: it is linked by ID, not by name.
func (tkm *TaskManager) EditTodo(ctx context.Context, name string, patch TodoPatch) (TodoView, error) {
	row, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, err
	}

	params := database.UpdateTodoParams{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Priority:    row.Priority,
		IsDaily:     row.IsDaily,
		ExpiresAt:   row.ExpiresAt,
		UpdatedAt:   time.Now(),
	}

	if patch.Name != nil {
		params.Name = strings.TrimSpace(*patch.Name)
		if params.Name == "" {
			return TodoView{}, localizedError{msg: i18n.T("tasks.error.emptyName"), sentinel: ErrInvalidInput}
		}
	}
	if patch.Description != nil {
		params.Description = validateDescription(strings.TrimSpace(*patch.Description))
	}
	if patch.Priority != nil {
		params.Priority = strings.ToLower(strings.TrimSpace(*patch.Priority))
		if !validPriority(params.Priority) {
			return TodoView{}, localizedError{
				msg:      fmt.Sprintf(i18n.T("tasks.error.invalidPriority"), *patch.Priority),
				sentinel: ErrInvalidInput,
			}
		}
	}
	if patch.IsDaily != nil {
		params.IsDaily = *patch.IsDaily
	}
	if patch.ExpiresAt != nil {
		params.ExpiresAt = *patch.ExpiresAt
	}

	updated, err := tkm.Queries.UpdateTodo(ctx, params)
	if isUniqueViolation(err) {
		return TodoView{}, localizedError{
			msg:      fmt.Sprintf(i18n.T("tasks.error.nameTaken"), params.Name),
			sentinel: ErrNameTaken,
		}
	}
	if err != nil {
		return TodoView{}, fmt.Errorf(i18n.T("tasks.error.updateTask"), err)
	}

	view, err := viewFromTodo(updated)
	if err != nil {
		return TodoView{}, err
	}
	view.DoneToday, err = tkm.isDoneToday(ctx, updated.ID)
	return view, err
}

func validPriority(p string) bool {
	switch p {
	case "", "high", "medium", "low":
		return true
	}
	return false
}

// isUniqueViolation recognizes SQLite's UNIQUE constraint error (task names
// are unique).
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
