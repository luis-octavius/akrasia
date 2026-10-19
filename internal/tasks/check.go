package tasks

import (
	"context"
	"fmt"

	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// checkExpired prints expired non-daily tasks.
func (tkm *TaskManager) CheckExpired(ctx context.Context) error {
	todos, err := tkm.Queries.CheckExpired(ctx)
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.checkExpiredTasks"), err)
	}

	if len(todos) == 0 {
		fmt.Println(NoExpiring)
		return nil
	}

	fmt.Println(i18n.T("tasks.tasks.expired"))
	for _, todo := range todos {
		printTodo(todo)
	}

	return nil
}

// checkExpiring prints tasks that expire within the configured warning window.
func (tkm *TaskManager) CheckExpiring(ctx context.Context) error {
	todos, err := tkm.Queries.GetTodos(ctx)
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.checkExpiringTasks"), err)
	}

	var countExpiring int

	for _, todo := range todos {
		isTodoExpiring := checkIfTodoExpires(todo.ExpiresAt)
		if isTodoExpiring && !todo.Concluded {
			fmt.Println(i18n.T("tasks.tasks.expiring"))
			printTodo(todo)
			countExpiring++
		}
	}

	if countExpiring == 0 {
		color.MsgError(NoExpiring)
	}

	return nil
}

func (tkm *TaskManager) CheckIfTodoIsDoneToday(ctx context.Context, id interface{}) (bool, error) {
	isDoneToday, err := tkm.Queries.IsDoneToday(ctx, id)
	if err != nil {
		return false, fmt.Errorf(i18n.T("tasks.error.findTask"), id, err)
	}

	if isDoneToday == 1 {
		return true, nil
	}

	return false, nil

}
