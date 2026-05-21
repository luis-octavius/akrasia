package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed schema/*.sql
var embedMigrations embed.FS

var (
	dbPath string
)

func GetDBPath() string {
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

	if err := goose.Up(db, "schema"); err != nil {
		return err
	}

	return nil
}

func InitDB() (*sql.DB, error) {
	path := GetDBPath()

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

func GetQueries() (*database.Queries, error) {
	dbPath := GetDBPath()
	if dbPath == "" {
		return nil, fmt.Errorf("database path is not valid")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	queries := database.New(db)
	if err != nil {
		return nil, err
	}

	return queries, nil
}
