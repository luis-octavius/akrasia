package tasks

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
	result, err := ParseDate([]int{})
	dayResult := result.Day()
	if err != nil {
		t.Fatal("Error in parseTime: ", err)
	}

	if expected != dayResult {
		t.Errorf("expected %v, got %v", expected, dayResult)
	}
}

func TestIsDateBeforeFalse(t *testing.T) {
	date := time.Now().AddDate(0, 0, 7)

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

func TestParseDateTooManyArgsReturnsError(t *testing.T) {
	if _, err := ParseDate([]int{1, 2, 3}); err == nil {
		t.Error("expected error for more than two date parts")
	}
}

func TestGetMonthByNumInvalid(t *testing.T) {
	for _, num := range []int{0, 13, -1} {
		if _, err := getMonthByNum(num); err == nil {
			t.Errorf("expected error for month %d", num)
		}
	}
}
