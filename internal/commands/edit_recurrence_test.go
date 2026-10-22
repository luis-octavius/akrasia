package commands

import (
	"testing"

	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/spf13/cobra"
)

// patchFor parses edit's recurrence flags the way the real command binds
// them and returns the patch they make.
func patchFor(t *testing.T, args ...string) (tasks.TodoPatch, error) {
	t.Helper()
	editDaily, editWeekly, editMonthly, editOn = false, false, false, ""

	cmd := &cobra.Command{}
	cmd.Flags().BoolVar(&editDaily, "daily", false, "")
	cmd.Flags().BoolVar(&editWeekly, "weekly", false, "")
	cmd.Flags().BoolVar(&editMonthly, "monthly", false, "")
	cmd.Flags().StringVar(&editOn, "on", "", "")
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}

	var patch tasks.TodoPatch
	err := editRecurrence(cmd, &patch)
	return patch, err
}

func TestEditRecurrenceFlags(t *testing.T) {
	patch, err := patchFor(t)
	if err != nil || !patch.Empty() {
		t.Fatalf("no flags = %+v, %v; want an empty patch", patch, err)
	}

	patch, err = patchFor(t, "--weekly", "--on", "fri")
	if err != nil || patch.Recurrence == nil || *patch.Recurrence != tasks.RecurWeekly || patch.RepeatOn == nil || *patch.RepeatOn != 5 {
		t.Fatalf("--weekly --on fri = %+v, %v", patch, err)
	}

	patch, err = patchFor(t, "--monthly")
	if err != nil || patch.Recurrence == nil || *patch.Recurrence != tasks.RecurMonthly || patch.RepeatOn != nil {
		t.Fatalf("--monthly = %+v, %v", patch, err)
	}

	patch, err = patchFor(t, "--daily")
	if err != nil || patch.Recurrence == nil || *patch.Recurrence != tasks.RecurDaily {
		t.Fatalf("--daily = %+v, %v", patch, err)
	}

	// the day alone moves a weekly or monthly task
	patch, err = patchFor(t, "--on", "mon")
	if err != nil || patch.Recurrence != nil || patch.RepeatOn == nil || *patch.RepeatOn != 1 {
		t.Fatalf("--on mon = %+v, %v", patch, err)
	}

	// turning it off makes a one-off task
	for _, flag := range []string{"--daily=false", "--weekly=false", "--monthly=false"} {
		patch, err = patchFor(t, flag)
		if err != nil || patch.IsDaily == nil || *patch.IsDaily || patch.Recurrence != nil {
			t.Fatalf("%s = %+v, %v; want a one-off task", flag, patch, err)
		}
	}

	// a false flag next to a true one is just the kind of the true one
	patch, err = patchFor(t, "--weekly", "--daily=false")
	if err != nil || patch.Recurrence == nil || *patch.Recurrence != tasks.RecurWeekly {
		t.Fatalf("--weekly --daily=false = %+v, %v", patch, err)
	}

	if _, err = patchFor(t, "--weekly", "--monthly"); err == nil {
		t.Fatal("--weekly --monthly should be rejected")
	}
	if _, err = patchFor(t, "--on", "someday"); err == nil {
		t.Fatal("--on someday should be rejected")
	}
}
