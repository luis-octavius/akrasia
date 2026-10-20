package tasks

import (
	"context"
	"slices"
	"testing"
)

func TestSlotsAreMarkedOncePerDay(t *testing.T) {
	tkm, _ := newTestTaskManager(t)
	ctx := context.Background()

	for range 2 {
		if err := tkm.MarkSlotsSent(ctx, "2026-10-19", []string{"10:00", "19:00"}); err != nil {
			t.Fatalf("MarkSlotsSent() error = %v", err)
		}
	}

	got, err := tkm.SentSlots(ctx, "2026-10-19")
	slices.Sort(got)
	if err != nil || !slices.Equal(got, []string{"10:00", "19:00"}) {
		t.Errorf("SentSlots() = %v, %v", got, err)
	}
	if other, _ := tkm.SentSlots(ctx, "2026-10-20"); len(other) != 0 {
		t.Errorf("next day already has %v", other)
	}
}
