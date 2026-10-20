package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/pkg/i18n"
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

// runMigrations applies the pending migrations. quiet hides goose's log,
// for the check every command makes on startup.
func runMigrations(dbPath string, quiet bool) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	defer db.Close()

	goose.SetBaseFS(embedMigrations)
	if quiet {
		goose.SetLogger(goose.NopLogger())
		defer goose.SetLogger(log.Default())
	}

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

	if err := runMigrations(path, false); err != nil {
		return nil, err
	}

	return db, nil
}

func GetQueries() (*database.Queries, error) {
	dbPath := GetDBPath()
	if dbPath == "" {
		return nil, fmt.Errorf("%s", i18n.T("db.error.invalidDatabasePath"))
	}

	// an existing database gets the migrations of a newer version without
	// running `init` again; a missing one is still created by `init`
	if _, err := os.Stat(dbPath); err == nil {
		if err := runMigrations(dbPath, true); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	queries := database.New(db)

	return queries, nil
}
