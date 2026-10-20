package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/luis-octavius/akrasia/internal/web"
)

func TestServiceUnit(t *testing.T) {
	t.Setenv("DB_PATH", `/home/me/my "data"/100%.db`)
	t.Setenv("XDG_CONFIG_HOME", "/home/me/.config")

	unit := serviceUnit("/usr/local/bin/akrasia")
	for _, want := range []string{
		`ExecStart="/usr/local/bin/akrasia" serve`,
		`Environment="DB_PATH=/home/me/my \"data\"/100%%.db"`,
		`Environment="XDG_CONFIG_HOME=/home/me/.config"`,
		"Restart=on-failure",
		"WantedBy=default.target",
	} {
		if !strings.Contains(unit, want) {
			t.Errorf("unit misses %q:\n%s", want, unit)
		}
	}
}

func TestServeInstallAndUninstall(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir()) // no crontab: nothing to remove there

	var calls []string
	orig := systemctl
	systemctl = func(args ...string) error { calls = append(calls, strings.Join(args, " ")); return nil }
	t.Cleanup(func() { systemctl = orig; rootCmd.SetArgs(nil); rootCmd.SetOut(nil) })
	rootCmd.SetOut(&bytes.Buffer{})

	rootCmd.SetArgs([]string{"serve", "install"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("serve install error = %v", err)
	}
	if _, err := os.Stat(unitFile()); err != nil {
		t.Fatalf("unit file not written: %v", err)
	}
	if !slices.Equal(calls, []string{"daemon-reload", "enable --now akrasia.service"}) {
		t.Errorf("systemctl calls = %q", calls)
	}

	calls = nil
	rootCmd.SetArgs([]string{"serve", "uninstall"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("serve uninstall error = %v", err)
	}
	if _, err := os.Stat(unitFile()); !os.IsNotExist(err) {
		t.Errorf("unit file still there: %v", err)
	}
	if !slices.Equal(calls, []string{"disable --now akrasia.service", "daemon-reload"}) {
		t.Errorf("systemctl calls = %q", calls)
	}
}

func TestRestoreRefusesWhileServing(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "akrasia.db")
	t.Setenv("DB_PATH", dbPath)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	release, err := web.Lock(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	rootCmd.SetArgs([]string{"restore", "latest", "--yes"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	t.Cleanup(func() { rootCmd.SetArgs(nil); rootCmd.SetOut(nil); rootCmd.SetErr(nil) })

	err = rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "systemctl --user stop akrasia") {
		t.Fatalf("restore while serving: error = %v", err)
	}
}
