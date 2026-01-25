package main

import (
	"database/sql"
	"log"

	"github.com/luis-octavius/akrasia/internal/database"
	_ "modernc.org/sqlite"
)

var cfg = Config{}

func main() {
	dbPath := getDBPath()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("Error opening the database: ", err)
	}

	defer db.Close()

	queries := database.New(db)
	cfg.Queries = queries

	// initialize cobra root command
	Execute()
}
