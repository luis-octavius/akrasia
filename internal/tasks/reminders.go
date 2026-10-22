package tasks

import (
	"context"
	"slices"
	"time"

	database "github.com/luis-octavius/akrasia/internal/db/out"
)

// Reminders is what a reminder message lists.
type Reminders struct {
	Overdue []TodoView
	// DueSoon runs from today to the end of the last day of the window
	DueSoon []TodoView
	// Daily is only filled when habits are part of the reminder
	Daily []TodoView
}

// Empty reports whether there is nothing to remind.
func (r Reminders) Empty() bool {
	return len(r.Overdue) == 0 && len(r.DueSoon) == 0 && len(r.Daily) == 0
}

// Reminders returns the pending tasks that are overdue or due within days
// (0 = today only) and, when withDaily is set, the daily tasks not done today.
func (tkm *TaskManager) Reminders(ctx context.Context, now time.Time, days int, withDaily bool) (Reminders, error) {
	todos, err := tkm.ListTodos(ctx, "")
	if err != nil {
		return Reminders{}, err
	}
	return reminders(todos, now, days, withDaily), nil
}

func reminders(todos []TodoView, now time.Time, days int, withDaily bool) Reminders {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	windowEnd := dayStart.AddDate(0, 0, days+1)

	var r Reminders
	for _, todo := range todos {
		switch {
		case todo.IsDaily:
			if withDaily && IsScheduled(todo, now) && !todo.DonePeriod {
				r.Daily = append(r.Daily, todo)
			}
		case todo.Concluded:
		case todo.ExpiresAt.Before(dayStart):
			r.Overdue = append(r.Overdue, todo)
		case todo.ExpiresAt.Before(windowEnd):
			r.DueSoon = append(r.DueSoon, todo)
		}
	}

	sortTodayTodos(r.Overdue)
	// nearest date first: that's what the reminder is about
	slices.SortStableFunc(r.DueSoon, func(a, b TodoView) int {
		if c := a.ExpiresAt.Compare(b.ExpiresAt); c != 0 {
			return c
		}
		return priorityRank(a.Priority) - priorityRank(b.Priority)
	})
	sortTodayTodos(r.Daily)
	return r
}

// SentSlots returns the reminder times already handled on day (yyyy-mm-dd).
func (tkm *TaskManager) SentSlots(ctx context.Context, day string) ([]string, error) {
	return tkm.Queries.GetNotifiedSlots(ctx, day)
}

// MarkSlotsSent records reminder times as handled on day, so they are
// never sent twice.
func (tkm *TaskManager) MarkSlotsSent(ctx context.Context, day string, slots []string) error {
	for _, slot := range slots {
		err := tkm.Queries.AddNotifiedSlot(ctx, database.AddNotifiedSlotParams{Day: day, Slot: slot})
		if err != nil {
			return err
		}
	}
	return nil
}
