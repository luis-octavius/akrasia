package main

import (
	"database/sql"
	"embed"
	"os"
	"path/filepath"

	"github.com/luis-octavius/akrasia/internal/database"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed sql/schema/*.sql
var embedMigrations embed.FS

var (
	dbPath string
)

type Config struct {
	Queries *database.Queries
}

func getDBPath() string {
	if dbPath != "" {
		return dbPath
	}

	if envPath := os.Getenv("DB_PATH"); envPath != "" {
		return envPath
	}

	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "akrasia", "akrasia.db")
}

func runMigrations(dbPath string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	goose.SetDialect("sqlite3")

	if err := goose.Up(db, "sql/schema"); err != nil {
		return err
	}

	return nil
}

func initDB() (*sql.DB, error) {
	path := getDBPath()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(path); err != nil {
		return nil, err
	}

	return db, nil
}
