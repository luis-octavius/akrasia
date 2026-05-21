package main

import (
	"context"
	"log"

	"github.com/luis-octavius/akrasia/internal/commands"
	"github.com/luis-octavius/akrasia/internal/db"
	"github.com/luis-octavius/akrasia/internal/tasks"
	_ "modernc.org/sqlite"
)

var tkm = tasks.TaskManager{}

func main() {
	queries, err := db.GetQueries()
	if err != nil {
		log.Fatalf("Error initializing the database")
	}

	tkm.Queries = queries

	ctx := commands.WithTaskManager(context.Background(), &tkm)

	// initialize cobra root command with context and a TaskManager within
	commands.ExecuteWithContext(ctx)
}
