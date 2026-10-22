package tasks

import (
	"context"
	"fmt"

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
	return tkm.complete(ctx, row, notes, false)
}

// complete does Complete; with all set it first completes the pending
// subtasks instead of refusing.
func (tkm *TaskManager) complete(ctx context.Context, row database.GetTodoByNameRow, notes string, all bool) (TodoView, error) {
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
				if err := tkm.markDone(ctx, s.ID, s.IsDaily, ""); err != nil {
					return TodoView{}, err
				}
			}
		}
	}

	// only one-off tasks are concluded; daily ones live on their history
	if err := tkm.markDone(ctx, view.ID, view.IsDaily, notes); err != nil {
		return TodoView{}, err
	}
	if !view.IsDaily {
		view.Concluded = true
	}
	view.DoneToday, view.DonePeriod = true, true

	if view.IsSubtask() {
		if err := tkm.syncParent(ctx, view.ParentID); err != nil {
			return view, err
		}
	}
	return view, nil
}
