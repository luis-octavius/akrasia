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

// todayView groups tasks into sections used by today/focus outputs.
type TodayView struct {
	Overdue      []database.Todo
	DueToday     []database.Todo
	Daily        []database.Todo
	ExpiringSoon []database.Todo
}

// todayOptions configures section filtering, limits, and output format.
type TodayOptions struct {
	Only     string
	Limit    int
	JSON     bool
	Priority string
}

type TodoView struct {
	ID					uuid.UUID
	Name 				string 
	Description sql.NullString
	Concluded 	bool 
	ExpiresAt 	time.Time 
	Priority 		string 
	IsDaily 		bool
	DoneToday   bool
}
