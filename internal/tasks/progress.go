package tasks

import (
	"context"
	"time"
)

// Progress counts how many of today's tasks are done.
type Progress struct {
	Done  int
	Total int
}

// Ratio is Done/Total in [0, 1]. A day with nothing to do counts as done.
func (p Progress) Ratio() float64 {
	if p.Total == 0 {
		return 1
	}
	return float64(p.Done) / float64(p.Total)
}

// TodayProgress measures today's tasks: every daily task, every weekly or
// monthly one that is due and pending (or was completed today), plus every
// one-off task that is overdue, due today or was completed today.
func (tkm *TaskManager) TodayProgress(ctx context.Context) (Progress, error) {
	todos, err := tkm.ListTodos(ctx, "")
	if err != nil {
		return Progress{}, err
	}
	return todayProgress(todos, time.Now()), nil
}

func todayProgress(todos []TodoView, now time.Time) Progress {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	var p Progress
	for _, todo := range todos {
		// a one-off task concluded on another day, or due after today, is
		// not part of today
		pendingToday := !todo.Concluded && todo.ExpiresAt.Before(dayEnd)
		if todo.IsDaily {
			pendingToday = IsScheduled(todo, now) && !todo.DonePeriod
		}
		if !todo.DoneToday && !pendingToday {
			continue
		}

		p.Total++
		if todo.DoneToday {
			p.Done++
		}
	}
	return p
}
