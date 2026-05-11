package main

import (
	"log"

	"github.com/luis-octavius/akrasia/internal/commands"
	"github.com/luis-octavius/akrasia/internal/db"
	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/internal/tasks"
	_ "modernc.org/sqlite"
)

var tkm = tasks.TaskManager{}

func main() {
	db, err := db.InitDB()
	if err != nil {
		log.Fatalf("Error initializing the database")
	}

	queries := database.New(db)
	tkm.Queries = queries

	// initialize cobra root command
	commands.Execute()
}
