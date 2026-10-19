package tasks

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// AddDump stores a named entry in the dump table and returns it.
func (tkm *TaskManager) AddDump(ctx context.Context, name string) (DumpEntry, error) {
	d, err := tkm.Queries.AddDump(ctx, database.AddDumpParams{
		ID:   uuid.New(),
		Name: name,
	})
	if err != nil {
		return DumpEntry{}, fmt.Errorf("%s: %w", i18n.T("tasks.error.addDump"), err)
	}

	return dumpEntry(d)
}

// Dumps returns every stored dump entry, oldest first.
func (tkm *TaskManager) Dumps(ctx context.Context) ([]DumpEntry, error) {
	rows, err := tkm.Queries.GetAllDump(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("tasks.error.getDump"), err)
	}

	entries := make([]DumpEntry, 0, len(rows))
	for _, d := range rows {
		entry, err := dumpEntry(d)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func dumpEntry(d database.Dump) (DumpEntry, error) {
	id, err := toUUID(d.ID)
	if err != nil {
		return DumpEntry{}, err
	}

	entry := DumpEntry{ID: id, Name: d.Name}
	if d.CreatedAt.Valid {
		entry.CreatedAt = d.CreatedAt.Time
	}

	return entry, nil
}

// DumpAt returns the dump entry at position n (1-based) of Dumps, the
// number `akrasia dump` prints next to each entry.
func (tkm *TaskManager) DumpAt(ctx context.Context, n int) (DumpEntry, error) {
	entries, err := tkm.Dumps(ctx)
	if err != nil {
		return DumpEntry{}, err
	}
	if n < 1 || n > len(entries) {
		return DumpEntry{}, localizedError{
			msg:      fmt.Sprintf(i18n.T("tasks.error.dumpNotFound"), n, len(entries)),
			sentinel: ErrDumpNotFound,
		}
	}
	return entries[n-1], nil
}

// DeleteDump removes a dump entry.
func (tkm *TaskManager) DeleteDump(ctx context.Context, id uuid.UUID) error {
	if err := tkm.Queries.DeleteDump(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("tasks.error.deleteDump"), err)
	}
	return nil
}

// PromoteDump turns a dump entry into a task: it creates the task and then
// removes the entry from the dump. When the task can't be created (for
// example, the name is taken) the entry is kept.
func (tkm *TaskManager) PromoteDump(ctx context.Context, id uuid.UUID, in NewTodo) (TodoView, error) {
	todo, err := tkm.AddTodo(ctx, in)
	if err != nil {
		return TodoView{}, err
	}
	if err := tkm.DeleteDump(ctx, id); err != nil {
		return todo, err
	}
	return todo, nil
}
