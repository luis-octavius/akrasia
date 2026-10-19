package tasks

import (
	"testing"
	"time"
)

// now is a fixed "today" in a non-UTC zone, to catch dates built in UTC.
var dateTestNow = time.Date(2026, time.December, 10, 15, 30, 0, 0, time.FixedZone("BRT", -3*60*60))

func TestParseDateAt(t *testing.T) {
	loc := dateTestNow.Location()
	cases := []struct {
		parts []int
		want  time.Time
	}{
		{[]int{20}, time.Date(2026, time.December, 20, 0, 0, 0, 0, loc)},
		{[]int{10}, time.Date(2026, time.December, 10, 0, 0, 0, 0, loc)}, // today is allowed
		{[]int{31, 12}, time.Date(2026, time.December, 31, 0, 0, 0, 0, loc)},
		{[]int{5, 1, 2027}, time.Date(2027, time.January, 5, 0, 0, 0, 0, loc)},
		{[]int{5, 1, 27}, time.Date(2027, time.January, 5, 0, 0, 0, 0, loc)}, // two-digit year
		{[]int{29, 2, 2028}, time.Date(2028, time.February, 29, 0, 0, 0, 0, loc)},
	}
	for _, c := range cases {
		got, err := ParseDateAt(c.parts, dateTestNow)
		if err != nil {
			t.Errorf("ParseDateAt(%v) error = %v", c.parts, err)
			continue
		}
		if !got.Equal(c.want) || got.Location() != loc {
			t.Errorf("ParseDateAt(%v) = %v, want %v (local midnight)", c.parts, got, c.want)
		}
	}
}

func TestParseDateAtRejects(t *testing.T) {
	cases := map[string][]int{
		"past day":         {9},
		"past month":       {5, 1},
		"past year":        {5, 1, 2026},
		"non-existent day": {31, 2, 2027},
		"not a leap year":  {29, 2, 2027},
		"day zero":         {0, 12},
		"bad month":        {1, 13},
		"too many parts":   {1, 2, 2027, 4},
	}
	for name, parts := range cases {
		if _, err := ParseDateAt(parts, dateTestNow); err == nil {
			t.Errorf("%s: ParseDateAt(%v) should fail", name, parts)
		}
	}
}

func TestParseDateWithoutPartsIsTomorrow(t *testing.T) {
	got, err := ParseDateAt(nil, dateTestNow)
	if err != nil {
		t.Fatalf("ParseDateAt(nil) error = %v", err)
	}
	if !got.Equal(dateTestNow.AddDate(0, 0, 1)) {
		t.Errorf("no date should mean tomorrow, got %v", got)
	}
}

func TestGetMonthByNumInvalid(t *testing.T) {
	for _, num := range []int{0, 13, -1} {
		if _, err := getMonthByNum(num); err == nil {
			t.Errorf("expected error for month %d", num)
		}
	}
}
