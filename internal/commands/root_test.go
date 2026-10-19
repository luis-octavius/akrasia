package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootWithoutTerminalShowsHelp(t *testing.T) {
	orig := isInteractive
	isInteractive = func() bool { return false }
	t.Cleanup(func() { isInteractive = orig })

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetArgs([]string{})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	for _, want := range []string{"today", "tui"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("help output missing %q:\n%s", want, out.String())
		}
	}
}

func TestRootRejectsUnknownArgs(t *testing.T) {
	rootCmd.SetArgs([]string{"not-a-command"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})

	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected an error for an unknown command")
	}
}
