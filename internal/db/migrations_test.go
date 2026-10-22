package db

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

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

// Two connections writing at once (the TUI and `serve`, say): the second
// waits for the first instead of failing with "database is locked".
func TestConcurrentWritersWait(t *testing.T) {
	path := filepath.Join(t.TempDir(), "akrasia.db")
	if err := runMigrations(path, true); err != nil {
		t.Fatal(err)
	}

	open := func() *sql.DB {
		conn, err := sql.Open("sqlite", DSN(path))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.Close() })
		return conn
	}
	first, second := open(), open()

	var mode string
	if err := first.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil || mode != "wal" {
		t.Fatalf("journal_mode = %q, %v; want wal", mode, err)
	}

	tx, err := first.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO dump (id, name) VALUES ('a', 'first')`); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := second.Exec(`INSERT INTO dump (id, name) VALUES ('b', 'second')`)
		done <- err
	}()

	// the second write is waiting on the first one's lock
	select {
	case err := <-done:
		t.Fatalf("second write finished while the first held the lock: %v", err)
	case <-time.After(200 * time.Millisecond):
	}

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatalf("second write failed instead of waiting: %v", err)
	}

	var n int
	if err := first.QueryRow("SELECT count(*) FROM dump").Scan(&n); err != nil || n != 2 {
		t.Errorf("dump has %d rows, %v; want both writes", n, err)
	}
}

// The due_time migration only adds a nullable column: tasks keep their data
// and have no time until one is set.
func TestDueTimeMigrationKeepsData(t *testing.T) {
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
	if err := goose.UpTo(conn, "schema", 9); err != nil {
		t.Fatalf("migrating to 9: %v", err)
	}
	_, err = conn.Exec(`
		INSERT INTO todos (id, name, created_at, updated_at, concluded, expires_at, priority, is_daily)
		VALUES ('11111111-1111-1111-1111-111111111111', 'Run', '2026-10-01', '2026-10-01 08:00:00', 0, '2026-10-02 00:00:00', 'high', 1)`)
	if err != nil {
		t.Fatal(err)
	}

	if err := runMigrations(path, true); err != nil {
		t.Fatalf("migrating to 10: %v", err)
	}

	var name string
	var dueTime sql.NullString
	err = conn.QueryRow(`SELECT name, due_time FROM todos`).Scan(&name, &dueTime)
	if err != nil || name != "Run" || dueTime.Valid {
		t.Fatalf("task after the migration = %q, %v, %v; want Run without a time", name, dueTime, err)
	}
}

// The recurrence migration turns the old is_daily flag into recurrence =
// 'daily' and leaves one-off tasks alone.
func TestRecurrenceMigration(t *testing.T) {
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
	if err := goose.UpTo(conn, "schema", 10); err != nil {
		t.Fatalf("migrating to 10: %v", err)
	}
	_, err = conn.Exec(`
		INSERT INTO todos (id, name, created_at, updated_at, concluded, expires_at, priority, is_daily)
		VALUES ('11111111-1111-1111-1111-111111111111', 'Run', '2026-10-01', '2026-10-01 08:00:00', 0, '2026-10-02 00:00:00', 'high', 1),
		       ('22222222-2222-2222-2222-222222222222', 'Taxes', '2026-10-01', '2026-10-01 08:00:00', 0, '2026-10-30 00:00:00', 'high', 0)`)
	if err != nil {
		t.Fatal(err)
	}

	if err := runMigrations(path, true); err != nil {
		t.Fatalf("migrating to 11: %v", err)
	}

	rows, err := conn.Query(`SELECT name, recurrence, repeat_on FROM todos ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	want := map[string]string{"Run": "daily", "Taxes": "none"}
	for rows.Next() {
		var name, recurrence string
		var repeatOn sql.NullInt64
		if err := rows.Scan(&name, &recurrence, &repeatOn); err != nil {
			t.Fatal(err)
		}
		if recurrence != want[name] || repeatOn.Valid {
			t.Errorf("%s = %q, %v; want %q without a day", name, recurrence, repeatOn, want[name])
		}
		delete(want, name)
	}
	if len(want) != 0 {
		t.Errorf("tasks missing after the migration: %v", want)
	}
}
