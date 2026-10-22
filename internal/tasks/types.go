package tasks

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	database "github.com/luis-octavius/akrasia/internal/db/out"
)

type TaskManager struct {
	Queries *database.Queries
}

// TodayView groups tasks into sections used by today/focus outputs.
type TodayView struct {
	Overdue      []TodoView
	DueToday     []TodoView
	Daily        []TodoView
	ExpiringSoon []TodoView
}

// Empty reports whether no section has any task.
func (v TodayView) Empty() bool {
	return len(v.Overdue) == 0 && len(v.DueToday) == 0 && len(v.Daily) == 0 && len(v.ExpiringSoon) == 0
}

// TodayOptions configures section filtering and limits.
type TodayOptions struct {
	Only     string
	Limit    int
	Priority string
}

type TodoView struct {
	ID          uuid.UUID
	Name        string
	Description sql.NullString
	Concluded   bool
	ExpiresAt   time.Time
	Priority    string
	IsDaily     bool
	DoneToday   bool
	// DueTime is the optional time of day, "HH:MM"; empty when not set
	DueTime string `json:",omitempty"`

	// ParentID is uuid.Nil for a top-level task; Parent is the parent's name.
	ParentID uuid.UUID `json:"-"`
	Parent   string    `json:",omitempty"`
	// Subtasks of a top-level task, nearest due date first.
	Subtasks []TodoView `json:",omitempty"`
}

// IsSubtask reports whether the task belongs to a parent.
func (v TodoView) IsSubtask() bool { return v.ParentID != uuid.Nil }

// Path is the name that finds this exact task: "Parent/Subtask" for a
// subtask, the name otherwise.
func (v TodoView) Path() string { return path(v.Parent, v.Name) }

// SubtaskProgress counts the subtasks done (today, for daily ones).
func (v TodoView) SubtaskProgress() Progress {
	p := Progress{Total: len(v.Subtasks)}
	for _, s := range v.Subtasks {
		if s.Done() {
			p.Done++
		}
	}
	return p
}

func path(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "/" + name
}

// Done reports whether the task counts as done right now: daily tasks are
// judged by today's history, the rest by their concluded flag.
func (v TodoView) Done() bool {
	if v.IsDaily {
		return v.DoneToday
	}
	return v.Concluded
}

// NewTodo holds the fields needed to create a task.
type NewTodo struct {
	Name        string
	Description string
	Priority    string
	IsDaily     bool
	// DueTime is an optional "HH:MM"; empty means no time of day
	DueTime string
	// ExpiresAt zero takes the parent's date for a subtask
	ExpiresAt time.Time
	// Parent is the name of the task this one is a subtask of, if any
	Parent string
}

// StreakPeriod is one completed run of consecutive days.
type StreakPeriod struct {
	StartDate string
	EndDate   string
	Days      int64
}

// HistoryEntry is one row of a task's completion history.
type HistoryEntry struct {
	Date      string
	Completed bool
	Notes     string
	InStreak  bool
}

// DebugReport is the raw data the streak queries operate on for one task.
type DebugReport struct {
	Todo         TodoView
	CreatedAt    time.Time
	HistorySince string // empty when the task has no streak window
	History      []HistoryEntry
	Streak       int64
}

// BackfillReport summarizes a backfill run.
type BackfillReport struct {
	Tasks []BackfillTaskResult
}

// Total returns how many days were recorded across all tasks.
func (r BackfillReport) Total() int {
	total := 0
	for _, t := range r.Tasks {
		total += t.Inserted
	}
	return total
}

// BackfillTaskResult is the outcome of a backfill for one daily task.
type BackfillTaskResult struct {
	Name     string
	Inserted int
	Failures []BackfillFailure
}

// BackfillFailure is a day that could not be recorded.
type BackfillFailure struct {
	Date string
	Err  error
}

// DumpEntry is one stored dump item.
type DumpEntry struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time // zero when unknown
}
