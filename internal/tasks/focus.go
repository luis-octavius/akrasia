package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// getTodayFocus builds and prints the daily dashboard, or JSON when requested.
func (tkm *TaskManager) GetTodayFocus(ctx context.Context, opts TodayOptions) error {
	todos, err := tkm.Queries.GetTodos(ctx)
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTaskDatabase"), err)
	}

	todos = filterTodosByPriority(todos, opts.Priority)
	doneToday, err := tkm.doneTodaySet(ctx)
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTaskDatabase"), err)
	}
	view := buildTodayView(todos, time.Now(), doneToday)
	view = applyTodayOptions(view, opts)

	if opts.JSON {
		payload, err := json.MarshalIndent(view, "", "  ")
		if err != nil {
			return fmt.Errorf(i18n.T("tasks.error.encodeTodayJSON"), err)
		}

		fmt.Println(string(payload))
		return nil
	}

	if len(view.Overdue) == 0 && len(view.DueToday) == 0 && len(view.Daily) == 0 && len(view.ExpiringSoon) == 0 {
		color.MsgSuccess(i18n.T("tasks.info.noPendingTasks"))
		return nil
	}

	fmt.Println(i18n.T("tasks.tasks.todayFocus"))

	printTodaySection(tkm, i18n.T("tasks.tasks.overdue"), view.Overdue)
	printTodaySection(tkm, i18n.T("tasks.tasks.dueToday"), view.DueToday)
	printTodaySection(tkm, i18n.T("tasks.tasks.dailyPending"), view.Daily)
	printTodaySection(tkm, i18n.T("tasks.tasks.expiringSoon"), view.ExpiringSoon)

	return nil
}

// getFocus returns the top actionable items across today sections.
func (tkm *TaskManager) GetFocus(ctx context.Context, limit int, priorityFilter string) error {
	todos, err := tkm.Queries.GetTodos(ctx)
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTaskDatabase"), err)
	}

	todos = filterTodosByPriority(todos, priorityFilter)
	doneToday, err := tkm.doneTodaySet(ctx)
	if err != nil {
		return fmt.Errorf(i18n.T("tasks.error.getTaskDatabase"), err)
	}
	view := buildTodayView(todos, time.Now(), doneToday)

	focus := make([]database.Todo, 0, limit)
	for _, section := range [][]database.Todo{view.Overdue, view.DueToday, view.Daily, view.ExpiringSoon} {
		for _, todo := range section {
			if len(focus) >= limit {
				break
			}
			focus = append(focus, todo)
		}
		if len(focus) >= limit {
			break
		}
	}

	if len(focus) == 0 {
		color.MsgSuccess(i18n.T("tasks.info.noFocusTasks"))
		return nil
	}

	fmt.Printf(i18n.T("tasks.tasks.focus"), len(focus))
	for _, todo := range focus {
		printTodo(todo)
	}

	return nil
}
