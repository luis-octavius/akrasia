package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/luis-octavius/akrasia/internal/database"
	_ "modernc.org/sqlite"
)

var cfg = Config{}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbPath := os.Getenv("DB_PATH")

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
