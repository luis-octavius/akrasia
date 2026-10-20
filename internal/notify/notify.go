// Package notify sends reminders of tasks that are due soon to a Discord
// channel, at fixed times of the day. Each time is sent at most once a day,
// and a time missed while the machine was off is caught up on the next run.
package notify

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// maxMessage is Discord's limit for a message's content.
const maxMessage = 2000

// Service is the part of tasks.TaskManager reminders use.
type Service interface {
	Reminders(ctx context.Context, now time.Time, days int, withDaily bool) (tasks.Reminders, error)
	SentSlots(ctx context.Context, day string) ([]string, error)
	MarkSlotsSent(ctx context.Context, day string, slots []string) error
}

// Sender posts a message.
type Sender func(ctx context.Context, content string) error

// Run sends the reminder of every time that has passed today and wasn't
// handled yet. Several pending times (the machine was off) become one
// message; it lists the daily tasks when one of them is the daily time.
// A time with nothing to remind is marked as handled without a message.
// It reports whether a message was sent.
func Run(ctx context.Context, svc Service, cfg Config, now time.Time, send Sender) (bool, error) {
	if !cfg.Enabled() {
		return false, nil
	}

	day := now.Format(time.DateOnly)
	sent, err := svc.SentSlots(ctx, day)
	if err != nil {
		return false, err
	}

	pending := PendingSlots(cfg.Slots(), sent, now)
	if len(pending) == 0 {
		return false, nil
	}

	withDaily := slices.Contains(pending, cfg.DailySlot())
	r, err := svc.Reminders(ctx, now, cfg.DaysAhead(), withDaily)
	if err != nil {
		return false, err
	}

	if !r.Empty() {
		// nothing is marked when sending fails, so the next run retries
		if err := send(ctx, Message(r, cfg, now)); err != nil {
			return false, err
		}
	}

	if err := svc.MarkSlotsSent(ctx, day, pending); err != nil {
		return !r.Empty(), err
	}
	return !r.Empty(), nil
}

// Test sends what the daily reminder would say right now, or a short note
// when there is nothing to remind, without marking any time as handled.
func Test(ctx context.Context, svc Service, cfg Config, now time.Time, send Sender) error {
	r, err := svc.Reminders(ctx, now, cfg.DaysAhead(), true)
	if err != nil {
		return err
	}

	msg := Message(r, cfg, now)
	if r.Empty() {
		msg = mention(cfg) + i18n.T("notify.test")
	}
	return send(ctx, msg)
}

// PendingSlots returns the times (HH:MM) that have passed today and aren't
// in sent.
func PendingSlots(slots, sent []string, now time.Time) []string {
	clock := now.Format("15:04")

	var pending []string
	for _, slot := range slots {
		// HH:MM strings compare like the times they stand for
		if slot <= clock && !slices.Contains(sent, slot) {
			pending = append(pending, slot)
		}
	}
	return pending
}

// Message renders the reminder, cut to fit Discord's limit.
func Message(r tasks.Reminders, cfg Config, now time.Time) string {
	lines := []string{mention(cfg) + "**" + i18n.T("notify.title") + "**"}

	section := func(title string, todos []tasks.TodoView, detail func(tasks.TodoView) string) {
		if len(todos) == 0 {
			return
		}
		lines = append(lines, "", fmt.Sprintf("**%s (%d)**", title, len(todos)))
		for _, t := range todos {
			line := "• " + t.Name
			if len(t.Subtasks) > 0 {
				p := t.SubtaskProgress()
				line += fmt.Sprintf(" (%d/%d)", p.Done, p.Total)
			}
			if d := detail(t); d != "" {
				line += " — " + d
			}
			lines = append(lines, line)
		}
	}

	section(i18n.T("notify.overdue"), r.Overdue, func(t tasks.TodoView) string {
		return fmt.Sprintf(i18n.T("notify.wasDue"), t.ExpiresAt.Format("02/01"))
	})
	section(i18n.T("notify.dueSoon"), r.DueSoon, func(t tasks.TodoView) string {
		return dueIn(t.ExpiresAt, now)
	})
	section(i18n.T("notify.daily"), r.Daily, func(tasks.TodoView) string { return "" })

	return fit(lines, maxMessage)
}

// dueIn says when a due date is: today, tomorrow or in N days.
func dueIn(due, now time.Time) string {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	day := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, now.Location())

	switch days := int(day.Sub(today).Hours() / 24); days {
	case 0:
		return i18n.T("notify.today")
	case 1:
		return i18n.T("notify.tomorrow")
	default:
		return fmt.Sprintf(i18n.T("notify.inDays"), days, due.Format("02/01"))
	}
}

func mention(cfg Config) string {
	if cfg.Mention == "" {
		return ""
	}
	return "<@" + cfg.Mention + "> "
}

// fit joins lines, dropping the last ones (with a "+N more" line) when the
// message would be longer than limit characters.
func fit(lines []string, limit int) string {
	msg := strings.Join(lines, "\n")
	if utf8.RuneCountInString(msg) <= limit {
		return msg
	}

	for n := len(lines) - 1; n > 0; n-- {
		more := fmt.Sprintf(i18n.T("notify.more"), countItems(lines[n:]))
		msg = strings.Join(lines[:n], "\n") + "\n" + more
		if utf8.RuneCountInString(msg) <= limit {
			return msg
		}
	}
	return string([]rune(msg)[:limit])
}

func countItems(lines []string) int {
	n := 0
	for _, l := range lines {
		if strings.HasPrefix(l, "• ") {
			n++
		}
	}
	return n
}
