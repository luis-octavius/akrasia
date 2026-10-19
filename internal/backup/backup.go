// Package backup keeps dated copies of the database in a folder the user
// chooses, usually one a sync tool (Google Drive, Dropbox, Syncthing)
// already takes off the machine. Each machine writes under its own profile,
// so several can share one folder without touching each other's copies.
package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// DefaultKeep is how many copies of a profile are kept when the config
// doesn't say.
const DefaultKeep = 14

// stampLayout is the time in a copy's name: akrasia-<profile>-<stamp>.db.
// It sorts the same way as the time it stands for.
const stampLayout = "2006-01-02_150405"

var (
	// ErrNoDatabase is returned when there is no database to copy yet.
	ErrNoDatabase = errors.New("no database to back up")
	// ErrNotAkrasia is returned for a valid SQLite file that has no tasks table.
	ErrNotAkrasia = errors.New("not an Akrasia database")

	validProfile = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
)

// Copy is one backup file of a profile.
type Copy struct {
	Path    string
	Profile string
	Time    time.Time
	Size    int64
}

// ValidProfile reports whether name can be used in a copy's file name.
func ValidProfile(name string) bool {
	return validProfile.MatchString(name)
}

// Snapshot writes a consistent, compacted copy of the database at dbPath
// into dir and checks it. A copy that fails the check is deleted.
func Snapshot(ctx context.Context, dbPath, dir, profile string, now time.Time) (string, error) {
	if _, err := os.Stat(dbPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrNoDatabase
		}
		return "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// two copies in the same second (a manual backup right after the
	// daily one) move the second one along instead of failing
	path := copyPath(dir, profile, now)
	for exists(path) {
		now = now.Add(time.Second)
		path = copyPath(dir, profile, now)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// VACUUM INTO reads one snapshot of the database, so the copy is
	// consistent even while the app is writing
	if _, err := conn.ExecContext(ctx, "VACUUM INTO ?", path); err != nil {
		_ = os.Remove(path)
		return "", err
	}

	if err := Verify(ctx, path); err != nil {
		_ = os.Remove(path)
		return "", err
	}

	return path, nil
}

// Verify opens the copy at path and runs SQLite's integrity check. It also
// fails for a valid SQLite file that isn't an Akrasia database.
func Verify(ctx context.Context, path string) error {
	// sql.Open would create a missing file
	if _, err := os.Stat(path); err != nil {
		return err
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer conn.Close()

	var result string
	if err := conn.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result); err != nil {
		return fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	if result != "ok" {
		return fmt.Errorf("%s: integrity check failed: %s", filepath.Base(path), result)
	}

	var tables int
	err = conn.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'todos'").Scan(&tables)
	if err != nil {
		return err
	}
	if tables == 0 {
		return fmt.Errorf("%s: %w", filepath.Base(path), ErrNotAkrasia)
	}

	return nil
}

// List returns the copies of profile in dir, newest first. A folder that
// doesn't exist yet has no copies.
func List(dir, profile string) ([]Copy, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var copies []Copy
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		t, ok := parseName(e.Name(), profile)
		if !ok {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		copies = append(copies, Copy{
			Path:    filepath.Join(dir, e.Name()),
			Profile: profile,
			Time:    t,
			Size:    info.Size(),
		})
	}

	slices.SortFunc(copies, func(a, b Copy) int { return b.Time.Compare(a.Time) })
	return copies, nil
}

// Latest returns the newest copy of profile in dir.
func Latest(dir, profile string) (Copy, bool, error) {
	copies, err := List(dir, profile)
	if err != nil || len(copies) == 0 {
		return Copy{}, false, err
	}
	return copies[0], true, nil
}

// Prune deletes the copies of profile beyond the newest keep and returns
// their paths. Other profiles' copies are never touched.
func Prune(dir, profile string, keep int) ([]string, error) {
	copies, err := List(dir, profile)
	if err != nil {
		return nil, err
	}

	var removed []string
	for _, c := range copies[min(keep, len(copies)):] {
		if err := os.Remove(c.Path); err != nil {
			return removed, err
		}
		removed = append(removed, c.Path)
	}
	return removed, nil
}

// Daily makes a copy and prunes old ones when backups are set up and this
// profile has no copy from today. It returns the new copy's path, or ""
// when nothing was needed.
func Daily(ctx context.Context, dbPath string, cfg Config, now time.Time) (string, error) {
	if !cfg.Enabled() {
		return "", nil
	}

	profile := cfg.ProfileName()
	latest, ok, err := Latest(cfg.Dir, profile)
	if err != nil {
		return "", err
	}
	if ok && sameDay(latest.Time, now) {
		return "", nil
	}

	path, err := Snapshot(ctx, dbPath, cfg.Dir, profile, now)
	if errors.Is(err, ErrNoDatabase) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	if _, err := Prune(cfg.Dir, profile, cfg.KeepN()); err != nil {
		return path, err
	}
	return path, nil
}

// Restore replaces the database at dbPath with the copy at src. The copy is
// checked first, and the current database is saved before it is replaced;
// the returned path is that safety copy ("" when there was no database).
// The safety copy goes to the backup folder, or next to the database when
// backups aren't set up.
func Restore(ctx context.Context, dbPath, src string, cfg Config, now time.Time) (string, error) {
	if err := Verify(ctx, src); err != nil {
		return "", err
	}

	dir := cfg.Dir
	if dir == "" {
		dir = filepath.Dir(dbPath)
	}

	safety, err := Snapshot(ctx, dbPath, dir, cfg.ProfileName(), now)
	if err != nil && !errors.Is(err, ErrNoDatabase) {
		return "", fmt.Errorf("saving the current database first: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return safety, err
	}

	// copy next to the database and rename over it, so a failure halfway
	// never leaves a half-written database behind
	tmp, err := os.CreateTemp(filepath.Dir(dbPath), ".akrasia-restore-*")
	if err != nil {
		return safety, err
	}
	defer os.Remove(tmp.Name())

	if err := copyInto(tmp, src); err != nil {
		tmp.Close()
		return safety, err
	}
	if err := tmp.Close(); err != nil {
		return safety, err
	}

	// a journal left by the old database must not be replayed on the new one
	for _, suffix := range []string{"-journal", "-wal", "-shm"} {
		if err := os.Remove(dbPath + suffix); err != nil && !errors.Is(err, os.ErrNotExist) {
			return safety, err
		}
	}

	if err := os.Rename(tmp.Name(), dbPath); err != nil {
		return safety, err
	}
	return safety, nil
}

func copyInto(dst *os.File, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if _, err := io.Copy(dst, in); err != nil {
		return err
	}
	return dst.Sync()
}

func copyPath(dir, profile string, t time.Time) string {
	return filepath.Join(dir, "akrasia-"+profile+"-"+t.Format(stampLayout)+".db")
}

// parseName returns the time of a copy named for profile. Names of other
// profiles don't parse, even when one profile's name starts with another's.
func parseName(name, profile string) (time.Time, bool) {
	stamp, ok := strings.CutPrefix(name, "akrasia-"+profile+"-")
	if !ok {
		return time.Time{}, false
	}
	stamp, ok = strings.CutSuffix(stamp, ".db")
	if !ok {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation(stampLayout, stamp, time.Local)
	return t, err == nil
}

func sameDay(a, b time.Time) bool {
	a, b = a.In(time.Local), b.In(time.Local)
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
