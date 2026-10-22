package tasks

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// How a task recurs. A recurring task is never concluded: it is done for
// the current period (today, this week, this month) by a history row.
const (
	RecurNone    = "none"
	RecurDaily   = "daily"
	RecurWeekly  = "weekly"
	RecurMonthly = "monthly"
)

// weekdayNames are the names accepted for a weekly task's day, in English
// and Portuguese, with and without accents. Monday is 1 and Sunday 7.
var weekdayNames = map[string]int{
	"mon": 1, "monday": 1, "seg": 1, "segunda": 1,
	"tue": 2, "tuesday": 2, "ter": 2, "terca": 2, "terça": 2,
	"wed": 3, "wednesday": 3, "qua": 3, "quarta": 3,
	"thu": 4, "thursday": 4, "qui": 4, "quinta": 4,
	"fri": 5, "friday": 5, "sex": 5, "sexta": 5,
	"sat": 6, "saturday": 6, "sab": 6, "sabado": 6, "sábado": 6,
	"sun": 7, "sunday": 7, "dom": 7, "domingo": 7,
}

var weekdayKeys = [...]string{"", "mon", "tue", "wed", "thu", "fri", "sat", "sun"}

// ISOWeekday is time.Weekday with Monday = 1 and Sunday = 7.
func ISOWeekday(t time.Time) int {
	if t.Weekday() == time.Sunday {
		return 7
	}
	return int(t.Weekday())
}

// normalizeRecurrence validates a task's recurrence and returns it with
// its repeat_on column. isDaily is the old boolean: with no recurrence
// given it means daily. A weekly or monthly task without a day takes
// today's (the weekday or the day of the month of ref).
func normalizeRecurrence(isDaily bool, recurrence string, repeatOn int, ref time.Time) (string, sql.NullInt64, error) {
	rec := strings.ToLower(strings.TrimSpace(recurrence))
	if rec == "" {
		rec = RecurNone
		if isDaily {
			rec = RecurDaily
		}
	}

	switch rec {
	case RecurNone, RecurDaily:
		if repeatOn != 0 {
			return "", sql.NullInt64{}, invalidRepeat(i18n.T("tasks.error.repeatOnNeedsWeeklyOrMonthly"))
		}
		return rec, sql.NullInt64{}, nil
	case RecurWeekly:
		if repeatOn == 0 {
			repeatOn = ISOWeekday(ref)
		}
		if repeatOn < 1 || repeatOn > 7 {
			return "", sql.NullInt64{}, invalidRepeat(fmt.Sprintf(i18n.T("tasks.error.invalidWeekday"), strconv.Itoa(repeatOn)))
		}
	case RecurMonthly:
		if repeatOn == 0 {
			repeatOn = ref.Day()
		}
		if repeatOn < 1 || repeatOn > 31 {
			return "", sql.NullInt64{}, invalidRepeat(fmt.Sprintf(i18n.T("tasks.error.invalidMonthDay"), strconv.Itoa(repeatOn)))
		}
	default:
		return "", sql.NullInt64{}, invalidRepeat(fmt.Sprintf(i18n.T("tasks.error.invalidRecurrence"), recurrence))
	}
	return rec, sql.NullInt64{Int64: int64(repeatOn), Valid: true}, nil
}

func invalidRepeat(msg string) error {
	return localizedError{msg: msg, sentinel: ErrInvalidInput}
}

// recurrenceOf is the recurrence of a stored task. Rows written before the
// column existed (or by hand) may have none: is_daily then means daily.
func recurrenceOf(isDaily bool, recurrence string) string {
	switch recurrence {
	case RecurDaily, RecurWeekly, RecurMonthly:
		return recurrence
	}
	if isDaily {
		return RecurDaily
	}
	return RecurNone
}

// ParseRepeatOn reads the day a weekly task falls on (a weekday name in
// English or Portuguese, or 1-7 with Monday = 1) or a monthly task's day of
// the month (1-31).
func ParseRepeatOn(recurrence, s string) (int, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return 0, nil
	}

	switch recurrence {
	case RecurWeekly:
		if n, ok := weekdayNames[s]; ok {
			return n, nil
		}
		if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= 7 {
			return n, nil
		}
		return 0, invalidRepeat(fmt.Sprintf(i18n.T("tasks.error.invalidWeekday"), s))
	case RecurMonthly:
		if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= 31 {
			return n, nil
		}
		return 0, invalidRepeat(fmt.Sprintf(i18n.T("tasks.error.invalidMonthDay"), s))
	}
	return 0, invalidRepeat(i18n.T("tasks.error.repeatOnNeedsWeeklyOrMonthly"))
}

// RecurringTitle is the heading of the section with the recurring tasks
// that are due: "DAILY PENDING", or "RECURRING PENDING" once it holds a
// weekly or monthly task.
func RecurringTitle(todos []TodoView) string {
	for _, todo := range todos {
		if todo.Recurrence == RecurWeekly || todo.Recurrence == RecurMonthly {
			return i18n.T("tasks.tasks.recurringPending")
		}
	}
	return i18n.T("tasks.tasks.dailyPending")
}

// ParseDay reads a day without knowing what kind of task it is for: a
// weekday name (a weekly task's day) or a number. Whether the number fits
// a weekly (1-7) or a monthly (1-31) task is checked when it is applied.
func ParseDay(s string) (int, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if n, ok := weekdayNames[s]; ok {
		return n, nil
	}
	if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= 31 {
		return n, nil
	}
	return 0, invalidRepeat(fmt.Sprintf(i18n.T("tasks.error.invalidDay"), s))
}

// monthlyDay is the day a monthly task falls on in the month containing
// now: a monthly task due on the 31st falls on the last day of a shorter
// month.
func monthlyDay(repeatOn int, now time.Time) int {
	last := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
	return min(repeatOn, last)
}

// IsScheduled reports whether a recurring task is due at now: a daily one
// always, a weekly one from its weekday to the end of the week, a monthly
// one from its day to the end of the month. Whether it was already done in
// that period is DonePeriod's business. A one-off task is never scheduled.
func IsScheduled(todo TodoView, now time.Time) bool {
	switch todo.Recurrence {
	case RecurDaily:
		return true
	case RecurWeekly:
		return ISOWeekday(now) >= max(todo.RepeatOn, 1)
	case RecurMonthly:
		return now.Day() >= monthlyDay(max(todo.RepeatOn, 1), now)
	}
	return todo.IsDaily
}

// Repeat describes how a recurring task repeats ("weekly on Fri",
// "monthly on day 5"); it is empty for a one-off task.
func (v TodoView) Repeat() string {
	switch v.Recurrence {
	case RecurDaily:
		return i18n.T("tasks.tasks.repeatDaily")
	case RecurWeekly:
		if v.RepeatOn >= 1 && v.RepeatOn <= 7 {
			return fmt.Sprintf(i18n.T("tasks.tasks.repeatWeekly"), i18n.T("tasks.tasks.weekday_"+weekdayKeys[v.RepeatOn]))
		}
		return i18n.T("tasks.tasks.repeatWeeklyAny")
	case RecurMonthly:
		return fmt.Sprintf(i18n.T("tasks.tasks.repeatMonthly"), v.RepeatOn)
	}
	if v.IsDaily {
		return i18n.T("tasks.tasks.repeatDaily")
	}
	return ""
}

// RepeatField is the recurrence as the TUI form writes it: "n", "d",
// "w fri" or "m 5".
func (v TodoView) RepeatField() string {
	switch v.Recurrence {
	case RecurWeekly:
		if v.RepeatOn >= 1 && v.RepeatOn <= 7 {
			return "w " + weekdayKeys[v.RepeatOn]
		}
		return "w"
	case RecurMonthly:
		return fmt.Sprintf("m %d", v.RepeatOn)
	case RecurDaily:
		return "d"
	}
	if v.IsDaily {
		return "d"
	}
	return "n"
}

// ParseRepeatField reads what RepeatField writes, plus y/s as daily (the
// old yes/no field): the kind, then the day for a weekly or monthly task.
func ParseRepeatField(s string) (recurrence string, repeatOn int, err error) {
	fields := strings.Fields(strings.ToLower(s))
	if len(fields) == 0 {
		return RecurNone, 0, nil
	}

	switch fields[0] {
	case "n", "no", "none":
		recurrence = RecurNone
	case "d", "y", "s", "daily":
		recurrence = RecurDaily
	case "w", "weekly":
		recurrence = RecurWeekly
	case "m", "monthly":
		recurrence = RecurMonthly
	default:
		return "", 0, invalidRepeat(i18n.T("tui.form.invalidDaily"))
	}
	if len(fields) > 2 || (len(fields) == 2 && (recurrence == RecurNone || recurrence == RecurDaily)) {
		return "", 0, invalidRepeat(i18n.T("tui.form.invalidDaily"))
	}
	if len(fields) == 2 {
		if repeatOn, err = ParseRepeatOn(recurrence, fields[1]); err != nil {
			return "", 0, err
		}
	}
	return recurrence, repeatOn, nil
}
