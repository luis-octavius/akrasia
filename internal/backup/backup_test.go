package backup_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/luis-octavius/akrasia/internal/backup"
	"github.com/luis-octavius/akrasia/internal/db"
	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/internal/tasks"
)

// newTestDB creates a migrated database in a temp dir (DB_PATH points at
// it) with one daily task completed today, and returns its path.
func newTestDB(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "akrasia.db")
	t.Setenv("DB_PATH", path)

	conn, err := db.InitDB()
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	tkm := &tasks.TaskManager{Queries: database.New(conn)}
	if _, err := tkm.AddTodo(ctx, tasks.NewTodo{Name: "Morning run", IsDaily: true}); err != nil {
		t.Fatalf("AddTodo() error = %v", err)
	}
	if _, err := tkm.Complete(ctx, "Morning run", "5km"); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	return path
}

// counts returns how many tasks and history rows the database has.
func counts(t *testing.T, path string) (todos, history int) {
	t.Helper()

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := conn.QueryRow("SELECT count(*) FROM todos").Scan(&todos); err != nil {
		t.Fatal(err)
	}
	if err := conn.QueryRow("SELECT count(*) FROM todos_history").Scan(&history); err != nil {
		t.Fatal(err)
	}
	return todos, history
}

func TestSnapshotCopiesTasksAndHistory(t *testing.T) {
	dbPath := newTestDB(t)
	dir := t.TempDir()
	now := time.Date(2026, 10, 19, 9, 30, 0, 0, time.Local)

	path, err := backup.Snapshot(context.Background(), dbPath, dir, "home", now)
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if want := filepath.Join(dir, "akrasia-home-2026-10-19_093000.db"); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}

	todos, history := counts(t, path)
	if todos != 1 || history != 1 {
		t.Errorf("copy has %d tasks and %d history rows, want 1 and 1", todos, history)
	}
}

func TestSnapshotTwiceInOneSecondKeepsBoth(t *testing.T) {
	dbPath := newTestDB(t)
	dir := t.TempDir()
	now := time.Date(2026, 10, 19, 9, 30, 0, 0, time.Local)

	first, err := backup.Snapshot(context.Background(), dbPath, dir, "home", now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := backup.Snapshot(context.Background(), dbPath, dir, "home", now)
	if err != nil {
		t.Fatalf("second Snapshot() error = %v", err)
	}
	if first == second {
		t.Fatalf("both copies were written to %q", first)
	}

	copies, _ := backup.List(dir, "home")
	if len(copies) != 2 || copies[0].Path != second {
		t.Errorf("List() = %+v, want the second copy first", copies)
	}
}

func TestSnapshotWithoutDatabase(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "akrasia.db")

	_, err := backup.Snapshot(context.Background(), missing, t.TempDir(), "home", time.Now())
	if !errors.Is(err, backup.ErrNoDatabase) {
		t.Fatalf("Snapshot() error = %v, want ErrNoDatabase", err)
	}
	if _, err := os.Stat(missing); !errors.Is(err, os.ErrNotExist) {
		t.Error("Snapshot() created the missing database")
	}
}

func TestPruneKeepsNewestAndIgnoresOtherProfiles(t *testing.T) {
	dbPath := newTestDB(t)
	dir := t.TempDir()
	ctx := context.Background()
	start := time.Date(2026, 10, 1, 8, 0, 0, 0, time.Local)

	for i := range 5 {
		day := start.AddDate(0, 0, i)
		if _, err := backup.Snapshot(ctx, dbPath, dir, "home", day); err != nil {
			t.Fatal(err)
		}
		// "home-2" starts like "home"; its copies must not count as home's
		if _, err := backup.Snapshot(ctx, dbPath, dir, "home-2", day); err != nil {
			t.Fatal(err)
		}
	}

	removed, err := backup.Prune(dir, "home", 2)
	if err != nil {
		t.Fatalf("Prune() error = %v", err)
	}
	if len(removed) != 3 {
		t.Errorf("Prune() removed %d copies, want 3", len(removed))
	}

	home, _ := backup.List(dir, "home")
	if len(home) != 2 || home[0].Time.Day() != 5 || home[1].Time.Day() != 4 {
		t.Errorf("home copies = %+v, want the 5th and the 4th", home)
	}
	if other, _ := backup.List(dir, "home-2"); len(other) != 5 {
		t.Errorf("home-2 has %d copies, want all 5", len(other))
	}
}

func TestVerifyRejectsBadCopies(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	corrupted := filepath.Join(dir, "corrupted.db")
	if err := os.WriteFile(corrupted, []byte("not a database at all, just text"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := backup.Verify(ctx, corrupted); err == nil {
		t.Error("Verify() accepted a corrupted file")
	}

	// a valid SQLite file that isn't Akrasia's
	other := filepath.Join(dir, "other.db")
	conn, err := sql.Open("sqlite", other)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec("CREATE TABLE notes (id INTEGER)"); err != nil {
		t.Fatal(err)
	}
	conn.Close()
	if err := backup.Verify(ctx, other); !errors.Is(err, backup.ErrNotAkrasia) {
		t.Errorf("Verify() error = %v, want ErrNotAkrasia", err)
	}

	if err := backup.Verify(ctx, filepath.Join(dir, "missing.db")); err == nil {
		t.Error("Verify() accepted a missing file")
	}
}

func TestDailyRunsOncePerDay(t *testing.T) {
	dbPath := newTestDB(t)
	cfg := backup.Config{Dir: t.TempDir(), Profile: "home", Keep: 3}
	ctx := context.Background()
	morning := time.Date(2026, 10, 19, 9, 0, 0, 0, time.Local)

	first, err := backup.Daily(ctx, dbPath, cfg, morning)
	if err != nil || first == "" {
		t.Fatalf("Daily() = %q, %v; want a new copy", first, err)
	}

	again, err := backup.Daily(ctx, dbPath, cfg, morning.Add(10*time.Hour))
	if err != nil || again != "" {
		t.Fatalf("Daily() later the same day = %q, %v; want nothing", again, err)
	}

	next, err := backup.Daily(ctx, dbPath, cfg, morning.AddDate(0, 0, 1))
	if err != nil || next == "" {
		t.Fatalf("Daily() the next day = %q, %v; want a new copy", next, err)
	}

	for i := 2; i <= 5; i++ {
		if _, err := backup.Daily(ctx, dbPath, cfg, morning.AddDate(0, 0, i)); err != nil {
			t.Fatal(err)
		}
	}
	if copies, _ := backup.List(cfg.Dir, "home"); len(copies) != 3 {
		t.Errorf("%d copies after 6 days with keep 3, want 3", len(copies))
	}
}

func TestDailyDoesNothingWhenOff(t *testing.T) {
	dbPath := newTestDB(t)

	path, err := backup.Daily(context.Background(), dbPath, backup.Config{}, time.Now())
	if err != nil || path != "" {
		t.Fatalf("Daily() without a folder = %q, %v; want nothing", path, err)
	}
}

func TestRestoreSavesCurrentDatabaseFirst(t *testing.T) {
	dbPath := newTestDB(t)
	cfg := backup.Config{Dir: t.TempDir(), Profile: "home"}
	ctx := context.Background()
	now := time.Date(2026, 10, 19, 9, 0, 0, 0, time.Local)

	old, err := backup.Snapshot(ctx, dbPath, cfg.Dir, "home", now)
	if err != nil {
		t.Fatal(err)
	}

	// a task added after the backup
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	tkm := &tasks.TaskManager{Queries: database.New(conn)}
	if _, err := tkm.AddTodo(ctx, tasks.NewTodo{Name: "Read"}); err != nil {
		t.Fatal(err)
	}
	conn.Close()

	safety, err := backup.Restore(ctx, dbPath, old, cfg, now.Add(time.Hour))
	if err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	if todos, history := counts(t, dbPath); todos != 1 || history != 1 {
		t.Errorf("restored database has %d tasks and %d history rows, want 1 and 1", todos, history)
	}
	if todos, _ := counts(t, safety); todos != 2 {
		t.Errorf("safety copy has %d tasks, want the 2 from before the restore", todos)
	}
}

func TestRestoreRejectsCorruptedCopy(t *testing.T) {
	dbPath := newTestDB(t)
	bad := filepath.Join(t.TempDir(), "bad.db")
	if err := os.WriteFile(bad, []byte("garbage"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := backup.Restore(context.Background(), dbPath, bad, backup.Config{}, time.Now()); err == nil {
		t.Fatal("Restore() accepted a corrupted copy")
	}
	if todos, _ := counts(t, dbPath); todos != 1 {
		t.Errorf("database has %d tasks after a refused restore, want 1", todos)
	}
}

func TestRestoreOnNewMachine(t *testing.T) {
	dbPath := newTestDB(t)
	ctx := context.Background()
	copyPath, err := backup.Snapshot(ctx, dbPath, t.TempDir(), "home", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	fresh := filepath.Join(t.TempDir(), "new", "akrasia.db")
	safety, err := backup.Restore(ctx, fresh, copyPath, backup.Config{}, time.Now())
	if err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	if safety != "" {
		t.Errorf("safety copy = %q, want none when there was no database", safety)
	}
	if todos, _ := counts(t, fresh); todos != 1 {
		t.Errorf("restored database has %d tasks, want 1", todos)
	}
}

func TestConfigRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := backup.LoadConfig()
	if err != nil || cfg.Enabled() {
		t.Fatalf("LoadConfig() without a file = %+v, %v; want backups off", cfg, err)
	}
	if cfg.KeepN() != backup.DefaultKeep || cfg.ProfileName() == "" {
		t.Errorf("defaults: keep %d, profile %q", cfg.KeepN(), cfg.ProfileName())
	}

	want := backup.Config{Dir: "/tmp/backups", Keep: 7, Profile: "work"}
	if err := backup.SaveConfig(want); err != nil {
		t.Fatal(err)
	}
	if got, err := backup.LoadConfig(); err != nil || got != want {
		t.Errorf("LoadConfig() = %+v, %v; want %+v", got, err, want)
	}
}
