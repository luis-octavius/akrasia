package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
)

// The subtasks migration rebuilds the todos table; the tasks, their ids
// (and so their history) and the name rules must survive it.
func TestSubtasksMigrationKeepsData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "akrasia.db")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}
	if err := goose.UpTo(conn, "schema", 8); err != nil {
		t.Fatalf("migrating to 8: %v", err)
	}

	_, err = conn.Exec(`
		INSERT INTO todos (id, name, description, created_at, updated_at, concluded, expires_at, priority, is_daily, history_since)
		VALUES ('11111111-1111-1111-1111-111111111111', 'Run', 'Park', '2026-10-01', '2026-10-01 08:00:00', 0, '2026-10-02 00:00:00', 'high', 1, '2026-10-01');
		INSERT INTO todos_history (id, todo_id, date, completed, notes)
		VALUES ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', '2026-10-01', 1, '5km');`)
	if err != nil {
		t.Fatal(err)
	}

	if err := runMigrations(path, true); err != nil {
		t.Fatalf("migrating to 9: %v", err)
	}

	var name, desc, priority, since string
	var daily bool
	err = conn.QueryRow(`SELECT name, description, priority, is_daily, CAST(history_since AS TEXT) FROM todos
		WHERE id = '11111111-1111-1111-1111-111111111111' AND parent_id IS NULL`).Scan(&name, &desc, &priority, &daily, &since)
	if err != nil {
		t.Fatalf("task after the migration: %v", err)
	}
	if name != "Run" || desc != "Park" || priority != "high" || !daily || since != "2026-10-01" {
		t.Errorf("task = %q %q %q %v %q", name, desc, priority, daily, since)
	}

	var joined int
	err = conn.QueryRow(`SELECT count(*) FROM todos_history h JOIN todos t ON t.id = h.todo_id`).Scan(&joined)
	if err != nil || joined != 1 {
		t.Errorf("history rows still linked = %d, %v; want 1", joined, err)
	}

	insert := `INSERT INTO todos (id, name, created_at, updated_at, concluded, expires_at, parent_id)
		VALUES (?, ?, '2026-10-01', '2026-10-01', 0, '2026-10-02', ?)`
	if _, err := conn.Exec(insert, "a", "Run", nil); err == nil {
		t.Error("a second top-level task named Run was accepted")
	}
	if _, err := conn.Exec(insert, "b", "Run", "11111111-1111-1111-1111-111111111111"); err != nil {
		t.Errorf("a subtask named like another task was rejected: %v", err)
	}
	if _, err := conn.Exec(insert, "c", "Run", "11111111-1111-1111-1111-111111111111"); err == nil {
		t.Error("two subtasks with the same name under one parent were accepted")
	}
}
