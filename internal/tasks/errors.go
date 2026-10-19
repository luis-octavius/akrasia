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
)

// localizedError shows a translated message to the user while still
// matching its sentinel through errors.Is.
type localizedError struct {
	msg      string
	sentinel error
}

func (e localizedError) Error() string { return e.msg }
func (e localizedError) Unwrap() error { return e.sentinel }
