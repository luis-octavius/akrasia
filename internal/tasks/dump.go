package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// AddDump stores a named entry in the dump table and reports success.
func (tkm *TaskManager) AddDump(name string) error {
	_, err := tkm.Queries.AddDump(context.Background(), database.AddDumpParams{
		ID:   uuid.New(),
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("tasks.error.addDump"), err)
	}

	color.MsgSuccess(fmt.Sprintf(i18n.T("tasks.info.addedDump"), name))
	return nil
}

// GetAllDump prints every stored dump entry, oldest first.
func (tkm *TaskManager) GetAllDump() error {
	allDump, err := tkm.Queries.GetAllDump(context.Background())
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("tasks.error.getDump"), err)
	}

	if len(allDump) == 0 {
		fmt.Println(i18n.T("tasks.info.noDump"))
		return nil
	}

	for _, d := range allDump {
		createdAt := ""
		if d.CreatedAt.Valid {
			createdAt = d.CreatedAt.Time.Format(time.RFC822)
		}
		fmt.Printf(i18n.T("tasks.dump.row"), d.Name, createdAt)
	}

	return nil
}
