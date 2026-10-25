package tasks

import (
	"context"
	"fmt"
	"time"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// Complete marks the task that best matches name as done, writes an entry
// to history and returns the updated task. A parent with pending subtasks
// is not completed: the error (ErrPendingSubtasks) says how many, and
// CompleteAll completes them too. Completing a subtask completes its
// parent once every subtask is done.
func (tkm *TaskManager) Complete(ctx context.Context, name, notes string) (TodoView, error) {
	row, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, fmt.Errorf(i18n.T("tasks.error.updateTaskStatus"), name, err)
	}
	return tkm.complete(ctx, row, notes, false, time.Time{})
}

// MaxBackdate is how far back CompleteOn accepts a day: a completion made
// offline on the phone and synced later.
const MaxBackdate = 7 * 24 * time.Hour

// CompleteOn is Complete (or CompleteAll, with all set) for a completion
// made on day instead of today, as the phone does when it syncs a task it
// completed offline. day must not be in the future nor older than
// MaxBackdate. The history row, the streak and the week or month of a
// recurring task follow day. A subtask of a recurring parent completed on
// a past day doesn't complete its parent, which only follows today.
func (tkm *TaskManager) CompleteOn(ctx context.Context, name, notes string, all bool, day time.Time) (TodoView, error) {
	if err := checkBackdate(day, time.Now()); err != nil {
		return TodoView{}, err
	}
	row, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, fmt.Errorf(i18n.T("tasks.error.updateTaskStatus"), name, err)
	}
	return tkm.complete(ctx, row, notes, all, dayOf(day))
}

func checkBackdate(day, now time.Time) error {
	today := dayOf(now)
	day = dayOf(day.In(now.Location()))
	if day.After(today) || today.Sub(day) > MaxBackdate {
		return localizedError{
			msg:      fmt.Sprintf(i18n.T("tasks.error.invalidCompletionDay"), day.Format(time.DateOnly)),
			sentinel: ErrInvalidInput,
		}
	}
	return nil
}

// complete does Complete; with all set it first completes the pending
// subtasks instead of refusing. day is the day of the completion (zero:
// today).
func (tkm *TaskManager) complete(ctx context.Context, row database.GetTodoByNameRow, notes string, all bool, day time.Time) (TodoView, error) {
	view, err := viewFromNameRow(row)
	if err != nil {
		return TodoView{}, err
	}

	if !view.IsSubtask() {
		parent, err := tkm.family(ctx, view.ID)
		if err != nil {
			return TodoView{}, err
		}
		if p := parent.SubtaskProgress(); p.Done < p.Total {
			if !all {
				return TodoView{}, pendingSubtasksError(parent)
			}
			for _, s := range parent.Subtasks {
				if s.Done() {
					continue
				}
				if err := tkm.markDoneOn(ctx, s.ID, s.IsDaily, "", day); err != nil {
					return TodoView{}, err
				}
			}
		}
	}

	// only one-off tasks are concluded; daily ones live on their history
	if err := tkm.markDoneOn(ctx, view.ID, view.IsDaily, notes, day); err != nil {
		return TodoView{}, err
	}
	if !view.IsDaily {
		view.Concluded = true
	}
	view.DoneToday, view.DonePeriod = true, true

	// a recurring parent only follows today: for a past day, its subtask
	// is recorded and the parent left alone
	pastRecurring := view.IsDaily && !day.IsZero() && !day.Equal(dayOf(time.Now()))
	if view.IsSubtask() && !pastRecurring {
		if err := tkm.syncParentOn(ctx, view.ParentID, day); err != nil {
			return view, err
		}
	}
	return view, nil
}
