package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
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
// A subtask can't leave its parent's daily mode or be due after it; a
// parent takes its subtasks along when its daily mode or date changes.
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

	parentID, err := optionalUUID(row.ParentID)
	if err != nil {
		return TodoView{}, err
	}
	if parentID != uuid.Nil && (patch.IsDaily != nil || patch.ExpiresAt != nil) {
		parent, err := tkm.family(ctx, parentID)
		if err != nil {
			return TodoView{}, err
		}
		if err := subtaskRules(parent, params.IsDaily, params.ExpiresAt); err != nil {
			return TodoView{}, err
		}
	}

	updated, err := tkm.Queries.UpdateTodo(ctx, params)
	if isUniqueViolation(err) {
		return TodoView{}, localizedError{
			msg:      fmt.Sprintf(i18n.T("tasks.error.nameTaken"), path(row.ParentName, params.Name)),
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
	view.Parent = row.ParentName

	if parentID == uuid.Nil {
		if err := tkm.carrySubtasks(ctx, view, patch); err != nil {
			return view, err
		}
	}

	view.DoneToday, err = tkm.isDoneToday(ctx, updated.ID)
	return view, err
}

// carrySubtasks applies a parent's new daily mode and date to its
// subtasks, then brings the parent's done state in line with them.
func (tkm *TaskManager) carrySubtasks(ctx context.Context, parent TodoView, patch TodoPatch) error {
	id := uuid.NullUUID{UUID: parent.ID, Valid: true}

	if patch.IsDaily != nil {
		err := tkm.Queries.SetSubtasksDaily(ctx, database.SetSubtasksDailyParams{IsDaily: parent.IsDaily, ParentID: id})
		if err != nil {
			return fmt.Errorf(i18n.T("tasks.error.updateTask"), err)
		}
	}
	if patch.ExpiresAt != nil {
		err := tkm.Queries.ClampSubtaskDates(ctx, database.ClampSubtaskDatesParams{ExpiresAt: parent.ExpiresAt, ParentID: id})
		if err != nil {
			return fmt.Errorf(i18n.T("tasks.error.updateTask"), err)
		}
	}
	if patch.IsDaily != nil {
		return tkm.syncParent(ctx, parent.ID)
	}
	return nil
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
