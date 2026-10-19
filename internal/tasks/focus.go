package tasks

import (
	"context"
	"time"
)

// Today builds the daily dashboard: overdue, due today, pending daily and
// expiring-soon tasks, filtered and limited by opts.
func (tkm *TaskManager) Today(ctx context.Context, opts TodayOptions) (TodayView, error) {
	todos, err := tkm.ListTodos(ctx, opts.Priority)
	if err != nil {
		return TodayView{}, err
	}

	view := buildTodayView(todos, time.Now())
	return applyTodayOptions(view, opts), nil
}

// Focus returns up to limit actionable tasks, taken from the today
// sections in order of urgency.
func (tkm *TaskManager) Focus(ctx context.Context, limit int, priorityFilter string) ([]TodoView, error) {
	todos, err := tkm.ListTodos(ctx, priorityFilter)
	if err != nil {
		return nil, err
	}

	view := buildTodayView(todos, time.Now())

	focus := make([]TodoView, 0, limit)
	for _, section := range [][]TodoView{view.Overdue, view.DueToday, view.Daily, view.ExpiringSoon} {
		for _, todo := range section {
			if len(focus) >= limit {
				return focus, nil
			}
			focus = append(focus, todo)
		}
	}

	return focus, nil
}
