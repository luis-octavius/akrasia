package tasks

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	database "github.com/luis-octavius/akrasia/internal/db/out"
)

func TestParseRepeatOn(t *testing.T) {
	cases := []struct {
		recurrence, in string
		want           int
		wantErr        bool
	}{
		{RecurWeekly, "mon", 1, false},
		{RecurWeekly, "Friday", 5, false},
		{RecurWeekly, "sex", 5, false},
		{RecurWeekly, "sábado", 6, false},
		{RecurWeekly, "sabado", 6, false},
		{RecurWeekly, "7", 7, false},
		{RecurWeekly, "8", 0, true},
		{RecurWeekly, "someday", 0, true},
		{RecurWeekly, "", 0, false},
		{RecurMonthly, "15", 15, false},
		{RecurMonthly, "31", 31, false},
		{RecurMonthly, "32", 0, true},
		{RecurMonthly, "fri", 0, true},
		{RecurDaily, "3", 0, true},
		{RecurNone, "3", 0, true},
	}
	for _, c := range cases {
		got, err := ParseRepeatOn(c.recurrence, c.in)
		if got != c.want || (err != nil) != c.wantErr {
			t.Errorf("ParseRepeatOn(%q, %q) = %d, %v; want %d, error %v", c.recurrence, c.in, got, err, c.want, c.wantErr)
		}
	}
}

func TestParseRepeatField(t *testing.T) {
	cases := []struct {
		in      string
		rec     string
		on      int
		wantErr bool
	}{
		{"", RecurNone, 0, false},
		{"n", RecurNone, 0, false},
		{"d", RecurDaily, 0, false},
		{"y", RecurDaily, 0, false},
		{"s", RecurDaily, 0, false},
		{"w", RecurWeekly, 0, false},
		{"w fri", RecurWeekly, 5, false},
		{"W  Qua", RecurWeekly, 3, false},
		{"m 5", RecurMonthly, 5, false},
		{"m", RecurMonthly, 0, false},
		{"m fri", "", 0, true},
		{"d 3", "", 0, true},
		{"x", "", 0, true},
		{"w mon tue", "", 0, true},
	}
	for _, c := range cases {
		rec, on, err := ParseRepeatField(c.in)
		if rec != c.rec || on != c.on || (err != nil) != c.wantErr {
			t.Errorf("ParseRepeatField(%q) = %q, %d, %v; want %q, %d, error %v", c.in, rec, on, err, c.rec, c.on, c.wantErr)
		}
	}
}

func TestRepeatFieldRoundTrips(t *testing.T) {
	for _, v := range []TodoView{
		{Recurrence: RecurNone},
		{Recurrence: RecurDaily, IsDaily: true},
		{Recurrence: RecurWeekly, IsDaily: true, RepeatOn: 3},
		{Recurrence: RecurMonthly, IsDaily: true, RepeatOn: 31},
	} {
		rec, on, err := ParseRepeatField(v.RepeatField())
		if err != nil || rec != v.Recurrence || on != v.RepeatOn {
			t.Errorf("%+v -> %q -> %q, %d, %v", v, v.RepeatField(), rec, on, err)
		}
	}
}

func TestIsScheduled(t *testing.T) {
	// 2026-10-21 is a Wednesday
	wed := time.Date(2026, time.October, 21, 9, 0, 0, 0, time.UTC)
	weekly := func(day int) TodoView { return TodoView{IsDaily: true, Recurrence: RecurWeekly, RepeatOn: day} }
	monthly := func(day int) TodoView { return TodoView{IsDaily: true, Recurrence: RecurMonthly, RepeatOn: day} }

	cases := []struct {
		name string
		todo TodoView
		now  time.Time
		want bool
	}{
		{"one-off", TodoView{}, wed, false},
		{"daily", TodoView{IsDaily: true, Recurrence: RecurDaily}, wed, true},
		{"weekly on Monday, from Monday on", weekly(1), wed, true},
		{"weekly on its day", weekly(3), wed, true},
		{"weekly on Friday, not yet", weekly(5), wed, false},
		{"weekly on Sunday, not yet", weekly(7), wed, false},
		{"weekly on Sunday, on Sunday", weekly(7), wed.AddDate(0, 0, 4), true},
		{"monthly, before its day", monthly(25), wed, false},
		{"monthly, on its day", monthly(21), wed, true},
		{"monthly, past its day", monthly(1), wed, true},
		// 31 falls on the last day of a shorter month
		{"monthly 31 in February, early", monthly(31), time.Date(2027, time.February, 27, 0, 0, 0, 0, time.UTC), false},
		{"monthly 31 in February, last day", monthly(31), time.Date(2027, time.February, 28, 0, 0, 0, 0, time.UTC), true},
		{"monthly 31 in a leap February", monthly(31), time.Date(2028, time.February, 29, 0, 0, 0, 0, time.UTC), true},
		{"monthly 31 in October, 30th", monthly(31), time.Date(2026, time.October, 30, 0, 0, 0, 0, time.UTC), false},
		// hand-built views without a recurrence still follow IsDaily
		{"daily without recurrence", TodoView{IsDaily: true}, wed, true},
	}
	for _, c := range cases {
		if got := IsScheduled(c.todo, c.now); got != c.want {
			t.Errorf("%s: IsScheduled = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestBuildTodayViewHidesRecurringTasksNotDue(t *testing.T) {
	wed := time.Date(2026, time.October, 21, 9, 0, 0, 0, time.UTC)
	todos := []TodoView{
		{Name: "daily", IsDaily: true, Recurrence: RecurDaily},
		{Name: "weekly-due", IsDaily: true, Recurrence: RecurWeekly, RepeatOn: 2},
		{Name: "weekly-later", IsDaily: true, Recurrence: RecurWeekly, RepeatOn: 6},
		{Name: "weekly-done", IsDaily: true, Recurrence: RecurWeekly, RepeatOn: 2, DonePeriod: true},
		{Name: "monthly-due", IsDaily: true, Recurrence: RecurMonthly, RepeatOn: 5},
		{Name: "monthly-later", IsDaily: true, Recurrence: RecurMonthly, RepeatOn: 28},
	}

	var got []string
	for _, todo := range buildTodayView(todos, wed).Daily {
		got = append(got, todo.Name)
	}
	if len(got) != 3 {
		t.Fatalf("Daily = %v, want daily, weekly-due and monthly-due", got)
	}
	for _, name := range got {
		if name == "weekly-later" || name == "weekly-done" || name == "monthly-later" {
			t.Errorf("%s should not be on today's list: %v", name, got)
		}
	}
}

func TestProgressCountsOnlyWhatIsDue(t *testing.T) {
	wed := time.Date(2026, time.October, 21, 9, 0, 0, 0, time.UTC)
	todos := []TodoView{
		{Name: "weekly-due", IsDaily: true, Recurrence: RecurWeekly, RepeatOn: 1},
		{Name: "weekly-later", IsDaily: true, Recurrence: RecurWeekly, RepeatOn: 6},
		// done on Monday: done for the week, but not part of today's work
		{Name: "weekly-done-monday", IsDaily: true, Recurrence: RecurWeekly, RepeatOn: 1, DonePeriod: true},
		{Name: "weekly-done-today", IsDaily: true, Recurrence: RecurWeekly, RepeatOn: 1, DonePeriod: true, DoneToday: true},
	}
	if got, want := todayProgress(todos, wed), (Progress{Done: 1, Total: 2}); got != want {
		t.Fatalf("todayProgress = %+v, want %+v", got, want)
	}
}

func TestAddWeeklyAndMonthly(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	got, err := tkm.AddTodo(ctx, NewTodo{Name: "review", Recurrence: RecurWeekly, RepeatOn: 5})
	if err != nil || got.Recurrence != RecurWeekly || got.RepeatOn != 5 || !got.IsDaily {
		t.Fatalf("weekly = %+v, %v", got, err)
	}

	got, err = tkm.AddTodo(ctx, NewTodo{Name: "rent", Recurrence: RecurMonthly, RepeatOn: 5})
	if err != nil || got.Recurrence != RecurMonthly || got.RepeatOn != 5 {
		t.Fatalf("monthly = %+v, %v", got, err)
	}

	// no day: today's weekday / day of the month
	got, err = tkm.AddTodo(ctx, NewTodo{Name: "no day", Recurrence: RecurWeekly})
	if err != nil || got.RepeatOn != ISOWeekday(time.Now()) {
		t.Fatalf("weekly without a day = %+v, %v; want today's weekday", got, err)
	}
	got, err = tkm.AddTodo(ctx, NewTodo{Name: "no month day", Recurrence: RecurMonthly})
	if err != nil || got.RepeatOn != time.Now().Day() {
		t.Fatalf("monthly without a day = %+v, %v; want today's day", got, err)
	}

	// IsDaily alone still means daily
	got, err = tkm.AddTodo(ctx, NewTodo{Name: "run", IsDaily: true})
	if err != nil || got.Recurrence != RecurDaily || got.RepeatOn != 0 {
		t.Fatalf("daily = %+v, %v", got, err)
	}

	found, err := tkm.FindTodo(ctx, "review")
	if err != nil || found.Recurrence != RecurWeekly || found.RepeatOn != 5 {
		t.Fatalf("FindTodo = %+v, %v", found, err)
	}

	for _, bad := range []NewTodo{
		{Name: "a", Recurrence: RecurWeekly, RepeatOn: 8},
		{Name: "b", Recurrence: RecurMonthly, RepeatOn: 32},
		{Name: "c", Recurrence: "yearly"},
		{Name: "d", Recurrence: RecurDaily, RepeatOn: 3},
	} {
		if _, err := tkm.AddTodo(ctx, bad); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("AddTodo(%+v) error = %v, want invalid input", bad, err)
		}
	}
}

// addHistory records a completion on a given day.
func addHistory(t *testing.T, tkm *TaskManager, id uuid.UUID, day time.Time) {
	t.Helper()
	_, err := tkm.Queries.BackfillDailyDone(context.Background(), database.BackfillDailyDoneParams{
		ID: uuid.New(), TodoID: id, Date: day.Format(time.DateOnly),
	})
	if err != nil {
		t.Fatalf("BackfillDailyDone() error = %v", err)
	}
}

func startOfWeek(now time.Time) time.Time {
	return dayOf(now).AddDate(0, 0, 1-ISOWeekday(now))
}

func TestWeeklyDoneForTheWholeWeek(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	weekly, err := tkm.AddTodo(ctx, NewTodo{Name: "review", Recurrence: RecurWeekly, RepeatOn: 1})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()

	if find(t, tkm, "review").Done() {
		t.Fatal("a new weekly task is not done")
	}

	// a completion last week does not count
	addHistory(t, tkm, weekly.ID, startOfWeek(now).AddDate(0, 0, -1))
	if find(t, tkm, "review").Done() {
		t.Fatal("last week's completion must not make this week's done")
	}

	// one on Monday of this week does, and it is not "done today" unless
	// today is Monday
	addHistory(t, tkm, weekly.ID, startOfWeek(now))
	got := find(t, tkm, "review")
	if !got.Done() || !got.DonePeriod {
		t.Fatalf("a completion this week should make it done, got %+v", got)
	}
	if want := ISOWeekday(now) == 1; got.DoneToday != want {
		t.Errorf("DoneToday = %v, want %v", got.DoneToday, want)
	}
}

func TestMonthlyDoneForTheWholeMonth(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	monthly, err := tkm.AddTodo(ctx, NewTodo{Name: "rent", Recurrence: RecurMonthly, RepeatOn: 1})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	addHistory(t, tkm, monthly.ID, first.AddDate(0, 0, -1))
	if find(t, tkm, "rent").Done() {
		t.Fatal("last month's completion must not make this month's done")
	}

	addHistory(t, tkm, monthly.ID, first)
	if !find(t, tkm, "rent").Done() {
		t.Fatal("a completion on the 1st should make the month done")
	}
}

func TestCompleteWeeklyThenUndoWithSubtask(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "review", Recurrence: RecurWeekly, RepeatOn: 3}); err != nil {
		t.Fatal(err)
	}
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "inbox", Parent: "review"}); err != nil {
		t.Fatal(err)
	}
	// the subtask follows its parent
	sub := find(t, tkm, "review/inbox")
	if sub.Recurrence != RecurWeekly || sub.RepeatOn != 3 {
		t.Fatalf("subtask = %+v, want it weekly on day 3", sub)
	}

	if _, err := tkm.Complete(ctx, "review/inbox", ""); err != nil {
		t.Fatal(err)
	}
	if !find(t, tkm, "review").Done() {
		t.Fatal("completing the only subtask should complete the weekly parent")
	}

	// a new pending subtask reopens the parent for the week
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "drafts", Parent: "review"}); err != nil {
		t.Fatal(err)
	}
	if find(t, tkm, "review").Done() {
		t.Fatal("a pending subtask should reopen the weekly parent")
	}
}

func TestSubtaskMustFollowParentRecurrence(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "review", Recurrence: RecurWeekly, RepeatOn: 3}); err != nil {
		t.Fatal(err)
	}
	_, err := tkm.AddTodo(ctx, NewTodo{Name: "x", Parent: "review", Recurrence: RecurMonthly, RepeatOn: 3})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("a monthly subtask under a weekly parent should be rejected, got %v", err)
	}

	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "inbox", Parent: "review"}); err != nil {
		t.Fatal(err)
	}
	_, err = tkm.EditTodo(ctx, "review/inbox", TodoPatch{RepeatOn: ptr(4)})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("a subtask can't move to another weekday than its parent, got %v", err)
	}
}

func TestEditRecurrence(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "review", IsDaily: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "inbox", Parent: "review"}); err != nil {
		t.Fatal(err)
	}

	got, err := tkm.EditTodo(ctx, "review", TodoPatch{Recurrence: ptr(RecurWeekly), RepeatOn: ptr(5)})
	if err != nil || got.Recurrence != RecurWeekly || got.RepeatOn != 5 || !got.IsDaily {
		t.Fatalf("to weekly = %+v, %v", got, err)
	}
	// the subtask came along
	if sub := find(t, tkm, "review/inbox"); sub.Recurrence != RecurWeekly || sub.RepeatOn != 5 {
		t.Fatalf("subtask = %+v, want weekly on 5", sub)
	}

	// only the day
	got, err = tkm.EditTodo(ctx, "review", TodoPatch{RepeatOn: ptr(2)})
	if err != nil || got.Recurrence != RecurWeekly || got.RepeatOn != 2 {
		t.Fatalf("day only = %+v, %v", got, err)
	}

	// another kind resets the day
	got, err = tkm.EditTodo(ctx, "review", TodoPatch{Recurrence: ptr(RecurMonthly)})
	if err != nil || got.Recurrence != RecurMonthly || got.RepeatOn != time.Now().Day() {
		t.Fatalf("to monthly = %+v, %v; want today's day", got, err)
	}

	// --daily wins over a weekly/monthly task
	got, err = tkm.EditTodo(ctx, "review", TodoPatch{IsDaily: ptr(true)})
	if err != nil || got.Recurrence != RecurDaily || got.RepeatOn != 0 {
		t.Fatalf("to daily = %+v, %v", got, err)
	}

	// a day means nothing for a daily task
	if _, err := tkm.EditTodo(ctx, "review", TodoPatch{RepeatOn: ptr(3)}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("a day on a daily task should be rejected, got %v", err)
	}

	got, err = tkm.EditTodo(ctx, "review", TodoPatch{IsDaily: ptr(false)})
	if err != nil || got.Recurrence != RecurNone || got.IsDaily {
		t.Fatalf("to one-off = %+v, %v", got, err)
	}
}

func TestBackfillOnlyForDailyTasks(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "review", Recurrence: RecurWeekly, RepeatOn: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := tkm.AddTodo(ctx, NewTodo{Name: "run", IsDaily: true}); err != nil {
		t.Fatal(err)
	}

	if _, err := tkm.BackfillDailyHistory(ctx, 3, "review"); !errors.Is(err, ErrNotDaily) {
		t.Fatalf("backfilling a weekly task should fail with ErrNotDaily, got %v", err)
	}

	report, err := tkm.BackfillDailyHistory(ctx, 3, "")
	if err != nil || len(report.Tasks) != 1 || report.Tasks[0].Name != "run" {
		t.Fatalf("backfilling everything should only touch the daily task, got %+v, %v", report, err)
	}
}

func run(start time.Time, n int64, rec string) periodRun {
	last := start
	for i := int64(1); i < n; i++ {
		last = nextPeriod(last, rec)
	}
	return periodRun{first: start, last: last, length: n}
}

func TestPeriodStart(t *testing.T) {
	wed := time.Date(2026, time.October, 21, 15, 0, 0, 0, time.Local)
	if got := periodStart(wed, RecurWeekly).Format(time.DateOnly); got != "2026-10-19" {
		t.Errorf("week start = %s, want Monday 2026-10-19", got)
	}
	sun := time.Date(2026, time.October, 25, 23, 59, 0, 0, time.Local)
	if got := periodStart(sun, RecurWeekly).Format(time.DateOnly); got != "2026-10-19" {
		t.Errorf("Sunday's week start = %s, want 2026-10-19", got)
	}
	if got := periodStart(wed, RecurMonthly).Format(time.DateOnly); got != "2026-10-01" {
		t.Errorf("month start = %s, want 2026-10-01", got)
	}
}

func TestCurrentPeriodStreak(t *testing.T) {
	wed := time.Date(2026, time.October, 21, 9, 0, 0, 0, time.UTC)
	thisWeek := time.Date(2026, time.October, 19, 0, 0, 0, 0, time.UTC)

	// three weeks ending last week: the week in progress does not break it
	runs := []periodRun{run(thisWeek.AddDate(0, 0, -21), 3, RecurWeekly)}
	if got := currentPeriodStreak(runs, RecurWeekly, wed); got != 3 {
		t.Errorf("streak ending last week = %d, want 3", got)
	}

	// done this week as well
	runs = []periodRun{run(thisWeek.AddDate(0, 0, -14), 3, RecurWeekly)}
	if got := currentPeriodStreak(runs, RecurWeekly, wed); got != 3 {
		t.Errorf("streak ending this week = %d, want 3", got)
	}

	// last week missed: broken
	runs = []periodRun{run(thisWeek.AddDate(0, 0, -28), 2, RecurWeekly)}
	if got := currentPeriodStreak(runs, RecurWeekly, wed); got != 0 {
		t.Errorf("broken streak = %d, want 0", got)
	}

	if got := currentPeriodStreak(nil, RecurWeekly, wed); got != 0 {
		t.Errorf("no history = %d, want 0", got)
	}

	// months, across a year end
	jan := time.Date(2027, time.January, 15, 0, 0, 0, 0, time.UTC)
	runs = []periodRun{run(time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC), 2, RecurMonthly)}
	if got := currentPeriodStreak(runs, RecurMonthly, jan); got != 2 {
		t.Errorf("monthly streak ending in December seen in January = %d, want 2", got)
	}
	if got := currentPeriodStreak(runs, RecurMonthly, jan.AddDate(0, 1, 0)); got != 0 {
		t.Errorf("monthly streak ending in December seen in February = %d, want 0", got)
	}
}

func TestWeeklyStreakFromHistory(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	weekly, err := tkm.AddTodo(ctx, NewTodo{Name: "review", Recurrence: RecurWeekly, RepeatOn: 1})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	thisWeek := startOfWeek(now)

	// weeks -5,-4 (a run of 2), a gap at -3, then -2,-1 and this one
	for _, back := range []int{5, 4, 2, 1, 0} {
		day := thisWeek.AddDate(0, 0, -7*back+2) // a Wednesday
		if day.After(now) {
			day = thisWeek
		}
		addHistory(t, tkm, weekly.ID, day)
	}
	if err := tkm.Queries.ExtendHistorySince(ctx, database.ExtendHistorySinceParams{
		Earliest: sqlString(thisWeek.AddDate(0, 0, -40).Format(time.DateOnly)),
		ID:       weekly.ID,
	}); err != nil {
		t.Fatal(err)
	}

	_, streak, err := tkm.CurrentStreak(ctx, "review")
	if err != nil || streak != 3 {
		t.Fatalf("CurrentStreak = %d, %v; want 3 weeks", streak, err)
	}

	_, periods, err := tkm.StreakHistory(ctx, "review")
	if err != nil || len(periods) != 2 || periods[0].Days != 3 || periods[1].Days != 2 {
		t.Fatalf("StreakHistory = %+v, %v; want a run of 3 and one of 2", periods, err)
	}
	if want := thisWeek.AddDate(0, 0, -14).Format(time.DateOnly); periods[0].StartDate != want {
		t.Errorf("longest run starts %s, want %s", periods[0].StartDate, want)
	}

	report, err := tkm.Debug(ctx, "review")
	if err != nil || report.Streak != 3 {
		t.Fatalf("Debug streak = %d, %v; want 3", report.Streak, err)
	}
}

func sqlString(s string) sql.NullString { return sql.NullString{String: s, Valid: true} }
