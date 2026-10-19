package tasks

import (
	"context"
	"fmt"

	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// backfilledNote is what backfill writes in the notes of each day it
// records; it is a marker, not something the user wrote.
const backfilledNote = "backfilled"

// NoteEntry is one completed day of a task and what was noted on it.
type NoteEntry struct {
	Date       string
	Notes      string // empty when nothing was written that day
	Backfilled bool   // recorded by backfill-history, not by the user
}

// Notes returns the completed days of the task that best matches name,
// newest first, with their notes. Days without notes are included so the
// whole streak shows. limit <= 0 returns every day.
func (tkm *TaskManager) Notes(ctx context.Context, name string, limit int) (TodoView, []NoteEntry, error) {
	row, err := findTodoByName(ctx, tkm, name)
	if err != nil {
		return TodoView{}, nil, err
	}

	history, err := tkm.Queries.GetHistoryByTodo(ctx, row.ID)
	if err != nil {
		return TodoView{}, nil, fmt.Errorf("%s: %w", i18n.T("tasks.error.getNotes"), err)
	}

	// the query is oldest first; walk it backwards for newest first
	var entries []NoteEntry
	for i := len(history) - 1; i >= 0; i-- {
		h := history[i]
		if !h.Completed.Bool {
			continue
		}

		entry := NoteEntry{Date: h.Date, Notes: h.Notes.String}
		if entry.Notes == backfilledNote {
			entry.Notes, entry.Backfilled = "", true
		}
		entries = append(entries, entry)

		if limit > 0 && len(entries) == limit {
			break
		}
	}

	view, err := viewFromNameRow(row)
	return view, entries, err
}
