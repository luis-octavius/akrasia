package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/luis-octavius/akrasia/internal/commands"
	"github.com/luis-octavius/akrasia/internal/db"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/i18n"
	_ "modernc.org/sqlite"
)

var tkm = tasks.TaskManager{}

func main() {
	// the app still works without its translations (messages show as ids),
	// but say so instead of leaving it a mystery
	if err := i18n.LoadError(); err != nil {
		fmt.Fprintf(os.Stderr, "akrasia: could not load the translations: %v\n", err)
	}

	queries, err := db.GetQueries()
	if err != nil {
		log.Fatalf("Error initializing the database")
	}

	tkm.Queries = queries

	ctx := commands.WithTaskManager(context.Background(), &tkm)

	// initialize cobra root command with context and a TaskManager within
	commands.ExecuteWithContext(ctx)
}
