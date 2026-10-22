package tasks

import (
	"context"
	"database/sql"
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
	// IsDaily true makes the task daily and false makes it a one-off task
	// (it clears any recurrence); Recurrence and RepeatOn are the general
	// form and win over it.
	IsDaily    *bool
	Recurrence *string
	RepeatOn   *int
	ExpiresAt  *time.Time
	// DueTime is the time of day as "HH:MM"; a pointer to "" removes it
	DueTime *string
}

// Empty reports whether the patch changes nothing.
func (p TodoPatch) Empty() bool {
	return p.Name == nil && p.Description == nil && p.Priority == nil &&
		p.IsDaily == nil && p.Recurrence == nil && p.RepeatOn == nil &&
		p.ExpiresAt == nil && p.DueTime == nil
}

// changesRecurrence reports whether the patch touches how the task recurs.
func (p TodoPatch) changesRecurrence() bool {
	return p.IsDaily != nil || p.Recurrence != nil || p.RepeatOn != nil
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
		Recurrence:  recurrenceOf(row.IsDaily, row.Recurrence),
		RepeatOn:    row.RepeatOn,
		ExpiresAt:   row.ExpiresAt,
		DueTime:     row.DueTime,
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
	if patch.changesRecurrence() {
		if err := applyRecurrence(&params, patch); err != nil {
			return TodoView{}, err
		}
	}
	if patch.ExpiresAt != nil {
		params.ExpiresAt = *patch.ExpiresAt
	}
	if patch.DueTime != nil {
		dueTime, err := ParseClock(*patch.DueTime)
		if err != nil {
			return TodoView{}, localizedError{msg: err.Error(), sentinel: ErrInvalidInput}
		}
		params.DueTime = sql.NullString{String: dueTime, Valid: dueTime != ""}
	}

	parentID, err := optionalUUID(row.ParentID)
	if err != nil {
		return TodoView{}, err
	}
	if parentID != uuid.Nil && (patch.changesRecurrence() || patch.ExpiresAt != nil) {
		parent, err := tkm.family(ctx, parentID)
		if err != nil {
			return TodoView{}, err
		}
		if err := subtaskRules(parent, params.Recurrence, params.RepeatOn, params.ExpiresAt); err != nil {
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

	view.DoneToday, view.DonePeriod, err = tkm.doneStatus(ctx, updated.ID)
	return view, err
}

// applyRecurrence sets the recurrence columns of params from the patch.
// Changing the kind without giving a day resets the day to today's.
func applyRecurrence(params *database.UpdateTodoParams, patch TodoPatch) error {
	recurrence, repeatOn := params.Recurrence, int(params.RepeatOn.Int64)

	switch {
	case patch.Recurrence != nil:
		recurrence = *patch.Recurrence
	case patch.IsDaily != nil && *patch.IsDaily:
		recurrence = RecurDaily
	case patch.IsDaily != nil:
		recurrence = RecurNone
	}
	if recurrence != params.Recurrence {
		repeatOn = 0
	}
	if patch.RepeatOn != nil {
		repeatOn = *patch.RepeatOn
	}

	var err error
	params.Recurrence, params.RepeatOn, err = normalizeRecurrence(false, recurrence, repeatOn, time.Now())
	if err != nil {
		return err
	}
	params.IsDaily = params.Recurrence != RecurNone
	return nil
}

// carrySubtasks applies a parent's new daily mode and date to its
// subtasks, then brings the parent's done state in line with them.
func (tkm *TaskManager) carrySubtasks(ctx context.Context, parent TodoView, patch TodoPatch) error {
	id := uuid.NullUUID{UUID: parent.ID, Valid: true}

	if patch.changesRecurrence() {
		err := tkm.Queries.SetSubtasksRecurrence(ctx, database.SetSubtasksRecurrenceParams{
			IsDaily:    parent.IsDaily,
			Recurrence: parent.Recurrence,
			RepeatOn:   sql.NullInt64{Int64: int64(parent.RepeatOn), Valid: parent.RepeatOn != 0},
			ParentID:   id,
		})
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
	if patch.changesRecurrence() {
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
