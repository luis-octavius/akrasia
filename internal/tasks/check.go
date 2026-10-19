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

	return viewsFromTodos(todos)
}

// Expiring returns pending tasks within ExpiringWindow of their deadline.
func (tkm *TaskManager) Expiring(ctx context.Context) ([]TodoView, error) {
	return tkm.listTodoViews(ctx, func(v TodoView) bool {
		return !v.Concluded && IsExpiringSoon(v.ExpiresAt)
	})
}

func (tkm *TaskManager) isDoneToday(ctx context.Context, id interface{}) (bool, error) {
	isDoneToday, err := tkm.Queries.IsDoneToday(ctx, id)
	if err != nil {
		return false, fmt.Errorf(i18n.T("tasks.error.findTask"), id, err)
	}

	return isDoneToday == 1, nil
}
