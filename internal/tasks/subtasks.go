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

// Subtasks are one level deep: a top-level task can have subtasks, a
// subtask can't. The rules that keep a parent in step with them:
//   - a subtask follows its parent's recurrence and is due no later than it;
//   - when every subtask is done, the parent is done too (for a daily
//     parent: each day, once that day's subtasks are all done);
//   - a parent with a pending subtask is not done: adding one reopens it;
//   - completing a parent with pending subtasks completes them all, after
//     the caller confirmed it (CompleteAll);
//   - deleting a parent deletes its subtasks.

// allTodoViews loads every task, subtasks included, flat and in due date
// order.
func (tkm *TaskManager) allTodoViews(ctx context.Context) ([]TodoView, error) {
	rows, err := tkm.Queries.ListTodosWithTodayStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf(i18n.T("tasks.error.getTaskDatabase"), err)
	}

	views := make([]TodoView, 0, len(rows))
	for _, row := range rows {
		view, err := newTodoView(row)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

// nest returns the top-level tasks with their subtasks attached, keeping
// the order of flat. Subtasks get their parent's name.
func nest(flat []TodoView) []TodoView {
	index := make(map[uuid.UUID]int)
	var top []TodoView
	for _, v := range flat {
		if !v.IsSubtask() {
			index[v.ID] = len(top)
			top = append(top, v)
		}
	}

	for _, v := range flat {
		if !v.IsSubtask() {
			continue
		}
		i, ok := index[v.ParentID]
		if !ok {
			continue // its parent is gone
		}
		v.Parent = top[i].Name
		top[i].Subtasks = append(top[i].Subtasks, v)
	}
	return top
}

// family returns the task with id, with its subtasks attached.
func (tkm *TaskManager) family(ctx context.Context, id uuid.UUID) (TodoView, error) {
	flat, err := tkm.allTodoViews(ctx)
	if err != nil {
		return TodoView{}, err
	}
	for _, v := range nest(flat) {
		if v.ID == id {
			return v, nil
		}
	}
	return TodoView{}, localizedError{
		msg:      fmt.Sprintf(i18n.T("tasks.error.todoNotFound"), id),
		sentinel: ErrTodoNotFound,
	}
}

// syncParent marks the parent done when all its subtasks are done, and
// reopens it when one of them is pending. A parent without subtasks is
// left alone.
func (tkm *TaskManager) syncParent(ctx context.Context, parentID uuid.UUID) error {
	return tkm.syncParentOn(ctx, parentID, time.Time{})
}

// syncParentOn is syncParent for a completion made on day (zero: today).
func (tkm *TaskManager) syncParentOn(ctx context.Context, parentID uuid.UUID, day time.Time) error {
	parent, err := tkm.family(ctx, parentID)
	if err != nil || len(parent.Subtasks) == 0 {
		return err
	}

	p := parent.SubtaskProgress()
	switch {
	case p.Done == p.Total && !parent.Done():
		return tkm.markDoneOn(ctx, parent.ID, parent.IsDaily, "", day)
	case p.Done < p.Total && parent.Done():
		return tkm.reopen(ctx, parent)
	}
	return nil
}

// markDone records a completion today: the concluded flag of a one-off
// task, and a history row for every task.
func (tkm *TaskManager) markDone(ctx context.Context, id uuid.UUID, daily bool, notes string) error {
	return tkm.markDoneOn(ctx, id, daily, notes, time.Time{})
}

// markDoneOn is markDone for a completion made on day (zero: today).
func (tkm *TaskManager) markDoneOn(ctx context.Context, id uuid.UUID, daily bool, notes string, day time.Time) error {
	if !daily {
		if _, err := tkm.Queries.UpdateTodoStatusById(ctx, id); err != nil {
			return fmt.Errorf(i18n.T("tasks.error.updateTaskStatus"), id, err)
		}
	}

	var err error
	if day.IsZero() {
		_, err = tkm.Queries.AddTodoHistory(ctx, database.AddTodoHistoryParams{
			ID:          uuid.New(),
			TodoID:      id,
			Completed:   sql.NullBool{Bool: true, Valid: true},
			CompletedAt: sql.NullTime{Time: time.Now(), Valid: true},
			Notes:       sql.NullString{String: notes, Valid: true},
		})
	} else {
		_, err = tkm.Queries.AddTodoHistoryOn(ctx, database.AddTodoHistoryOnParams{
			ID:          uuid.New(),
			TodoID:      id,
			Date:        day.Format(time.DateOnly),
			Completed:   sql.NullBool{Bool: true, Valid: true},
			CompletedAt: sql.NullTime{Time: day, Valid: true},
			Notes:       sql.NullString{String: notes, Valid: true},
		})
	}
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("tasks.error.createTaskHistory"), err)
	}
	return nil
}

// reopen undoes a parent's completion: a one-off task is no longer
// concluded, a recurring one loses the history rows of its current period.
func (tkm *TaskManager) reopen(ctx context.Context, parent TodoView) error {
	if parent.IsDaily {
		return tkm.Queries.DeleteCurrentPeriodHistory(ctx, parent.ID)
	}
	return tkm.Queries.ReopenTodo(ctx, parent.ID)
}

// CompleteAll completes the task that best matches name and every pending
// subtask of it. It is what Complete asks the caller to confirm with
// ErrPendingSubtasks.
func (tkm *TaskManager) CompleteAll(ctx context.Context, name, notes string) (TodoView, error) {
	row, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, fmt.Errorf(i18n.T("tasks.error.updateTaskStatus"), name, err)
	}
	return tkm.complete(ctx, row, notes, true, time.Time{})
}

// pendingSubtasksError tells the caller how many subtasks completing the
// parent would also complete.
func pendingSubtasksError(parent TodoView) error {
	p := parent.SubtaskProgress()
	return localizedError{
		msg:      fmt.Sprintf(i18n.T("tasks.error.pendingSubtasks"), parent.Name, p.Total-p.Done),
		sentinel: ErrPendingSubtasks,
	}
}

// subtaskRules checks a new or edited subtask against its parent: it
// follows the parent's recurrence (kind and day) and is due no later than
// it.
func subtaskRules(parent TodoView, recurrence string, repeatOn sql.NullInt64, expiresAt time.Time) error {
	if recurrence != parent.Recurrence || int(repeatOn.Int64) != parent.RepeatOn {
		return localizedError{
			msg:      fmt.Sprintf(i18n.T("tasks.error.subtaskDaily"), parent.Name),
			sentinel: ErrInvalidInput,
		}
	}
	if !parent.IsDaily && dayOf(expiresAt).After(dayOf(parent.ExpiresAt)) {
		return localizedError{
			msg:      fmt.Sprintf(i18n.T("tasks.error.subtaskDate"), parent.Name, parent.ExpiresAt.Format("02/01/2006")),
			sentinel: ErrInvalidInput,
		}
	}
	return nil
}

func dayOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
