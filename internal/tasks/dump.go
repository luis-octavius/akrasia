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
