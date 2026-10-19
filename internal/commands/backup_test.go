package commands

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/luis-octavius/akrasia/internal/backup"
)

func TestDailyBackupSkipsBackupCommands(t *testing.T) {
	var runs int
	orig := dailyBackup
	dailyBackup = func(context.Context) error { runs++; return nil }
	t.Cleanup(func() { dailyBackup = orig })

	cases := []struct {
		args []string
		want int
	}{
		{[]string{"backup"}, 0},
		{[]string{"backup", "list"}, 0},
		{[]string{"restore", "x"}, 0},
		{[]string{"config", "backup", "show"}, 0},
		{[]string{"today"}, 1},
		{[]string{"dump"}, 1},
	}

	for _, c := range cases {
		cmd, _, err := rootCmd.Find(c.args)
		if err != nil {
			t.Fatalf("Find(%v) error = %v", c.args, err)
		}
		runs = 0
		runDailyBackup(cmd)
		if runs != c.want {
			t.Errorf("%v ran the daily backup %d time(s), want %d", c.args, runs, c.want)
		}
	}
}

func TestConfigBackupSavesSettings(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir := filepath.Join(t.TempDir(), "sync", "akrasia")

	for _, args := range [][]string{
		{"config", "backup", "dir", dir},
		{"config", "backup", "keep", "30"},
		{"config", "backup", "profile", "work-pc"},
	} {
		rootCmd.SetArgs(args)
		rootCmd.SetOut(&bytes.Buffer{})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("%v error = %v", args, err)
		}
	}
	t.Cleanup(func() { rootCmd.SetArgs(nil); rootCmd.SetOut(nil) })

	cfg, err := backup.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if want := (backup.Config{Dir: dir, Keep: 30, Profile: "work-pc"}); cfg != want {
		t.Errorf("config = %+v, want %+v", cfg, want)
	}

	for _, args := range [][]string{
		{"config", "backup", "keep", "0"},
		{"config", "backup", "profile", "my pc"},
		{"config", "backup", "dir"},
		{"config", "backup", "folder", "x"},
	} {
		rootCmd.SetArgs(args)
		rootCmd.SetErr(&bytes.Buffer{})
		if err := rootCmd.Execute(); err == nil {
			t.Errorf("%v: expected an error", args)
		}
	}
	rootCmd.SetErr(nil)
}

func TestRestoreLatestWithoutBackups(t *testing.T) {
	cfg := backup.Config{Dir: t.TempDir(), Profile: "home"}
	if _, err := restoreSource(cfg, "latest"); err == nil {
		t.Error("restoreSource(latest) with no copies: expected an error")
	}
	if _, err := restoreSource(backup.Config{}, "latest"); err == nil {
		t.Error("restoreSource(latest) with backups off: expected an error")
	}
}
