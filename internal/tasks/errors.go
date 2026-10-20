package tasks

import "errors"

// Sentinel errors let callers (CLI or TUI) react to a failure with
// errors.Is instead of matching on translated messages.
var (
	ErrTodoNotFound = errors.New("todo not found")
	ErrNotDaily     = errors.New("todo is not daily")
	ErrNameTaken    = errors.New("todo name already taken")
	ErrInvalidInput = errors.New("invalid todo field")
	ErrDumpNotFound = errors.New("dump entry not found")
	// ErrAmbiguousName: several tasks have exactly that name; use
	// "Parent/Subtask" to pick one.
	ErrAmbiguousName = errors.New("todo name is ambiguous")
	// ErrPendingSubtasks: completing the task needs confirmation, since it
	// also completes its pending subtasks (see CompleteAll).
	ErrPendingSubtasks = errors.New("todo has pending subtasks")
	ErrNestedSubtask   = errors.New("subtasks can't have subtasks")
)

// localizedError shows a translated message to the user while still
// matching its sentinel through errors.Is.
type localizedError struct {
	msg      string
	sentinel error
}

func (e localizedError) Error() string { return e.msg }
func (e localizedError) Unwrap() error { return e.sentinel }
