package tasks

import (
	"context"
	"fmt"

	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// Expired returns non-daily tasks whose deadline has passed.
func (tkm *TaskManager) Expired(ctx context.Context) ([]TodoView, error) {
	todos, err := tkm.Queries.CheckExpired(ctx)
	if err != nil {
		return nil, fmt.Errorf(i18n.T("tasks.error.checkExpiredTasks"), err)
	}

	views, err := viewsFromTodos(todos)
	if err != nil {
		return nil, err
	}

	// a subtask shows under its parent, not on its own
	var top []TodoView
	for _, v := range views {
		if !v.IsSubtask() {
			top = append(top, v)
		}
	}
	return top, nil
}

// Expiring returns pending tasks within ExpiringWindow of their deadline.
func (tkm *TaskManager) Expiring(ctx context.Context) ([]TodoView, error) {
	return tkm.listTodoViews(ctx, func(v TodoView) bool {
		return !v.Concluded && IsExpiringSoon(v.ExpiresAt)
	})
}

// doneStatus tells whether the task was completed today and whether it is
// done for its current period (the same, but for a weekly or monthly task).
func (tkm *TaskManager) doneStatus(ctx context.Context, id interface{}) (today, period bool, err error) {
	if today, err = tkm.isDoneToday(ctx, id); err != nil {
		return false, false, err
	}
	n, err := tkm.Queries.IsDoneThisPeriod(ctx, id)
	if err != nil {
		return false, false, fmt.Errorf(i18n.T("tasks.error.findTask"), id, err)
	}
	return today, n == 1, nil
}

func (tkm *TaskManager) isDoneToday(ctx context.Context, id interface{}) (bool, error) {
	isDoneToday, err := tkm.Queries.IsDoneToday(ctx, id)
	if err != nil {
		return false, fmt.Errorf(i18n.T("tasks.error.findTask"), id, err)
	}

	return isDoneToday == 1, nil
}
