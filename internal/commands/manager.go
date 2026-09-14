package commands

import (
	"context"
	"fmt"

	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

type contextKey struct{}

var taskManagerKey = contextKey{}

// public API to main injection
func WithTaskManager(ctx context.Context, tkm *tasks.TaskManager) context.Context {
	return context.WithValue(ctx, taskManagerKey, tkm)
}

// taskManagerFromContext validates context that has to
// retrieve a TaskManager struct from it
// returns an error if:
// - the value at the context is nil
// - the value is not a TaskManager
func taskManagerFromContext(ctx context.Context) (*tasks.TaskManager, error) {
	v := ctx.Value(taskManagerKey)
	tkm, ok := v.(*tasks.TaskManager)
	if !ok || tkm == nil {
		return nil, fmt.Errorf(i18n.T("errorTkmNotFound"))
	}
	return tkm, nil
}
