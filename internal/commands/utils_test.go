package commands

import (
	"testing"
	"time"
)

func TestParseDateWithoutArgs(t *testing.T) {
	oneDay, err := time.ParseDuration("24h")
	if err != nil {
		t.Fatal("Error parsing duration: ", err)
	}

	expected := time.Now().Add(oneDay).Day()
	result, err := parseDate([]int{})
	dayResult := result.Day()
	if err != nil {
		t.Fatal("Error in parseTime: ", err)
	}

	if expected != dayResult {
		t.Errorf("expected %v, got %v", expected, dayResult)
	}
}

func TestIsDateBeforeFalse(t *testing.T) {
	date := time.Date(2026, time.November, 03, 0, 0, 0, 0, time.UTC)

	isBefore, _ := isDateBefore(date)

	if isBefore {
		t.Errorf("expected %v to be %v", date, false)
	}
}

func TestIsDateBeforeTrue(t *testing.T) {
	date := time.Date(2025, time.September, 03, 0, 0, 0, 0, time.UTC)

	isBefore, _ := isDateBefore(date)

	if !isBefore {
		t.Errorf("expected %v to be %v", date, true)
	}

}
