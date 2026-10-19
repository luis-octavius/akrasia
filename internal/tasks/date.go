package tasks

import (
	"errors"
	"fmt"
	"time"

	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// ParseDate turns the CLI date parts into a due date at local midnight:
//   - no parts: tomorrow (the current time, one day ahead)
//   - dd: that day of the current month
//   - dd,mm: that day of the current year
//   - dd,mm,yyyy: that exact day (a two-digit year means 20yy)
//
// Days before today are rejected; today itself is accepted.
func ParseDate(parts []int) (time.Time, error) {
	return ParseDateAt(parts, time.Now())
}

// ParseDateAt is ParseDate with "now" given, for tests.
func ParseDateAt(parts []int, now time.Time) (time.Time, error) {
	if len(parts) == 0 {
		return now.AddDate(0, 0, 1), nil
	}
	if len(parts) > 3 {
		return time.Time{}, errors.New(i18n.T("commands.error.dateNotEnoughArguments"))
	}

	day := parts[0]
	year, month, _ := now.Date()

	if len(parts) >= 2 {
		m, err := getMonthByNum(parts[1])
		if err != nil {
			return time.Time{}, err
		}
		month = m
	}
	if len(parts) == 3 {
		year = parts[2]
		if year < 100 {
			year += 2000
		}
	}

	date := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	// time.Date normalizes overflow (31/02 becomes 03/03); reject it instead
	if day < 1 || date.Day() != day || date.Month() != month {
		return time.Time{}, fmt.Errorf(i18n.T("commands.error.invalidDay"), day, int(month), year)
	}

	if _, err := isDateBefore(date, now); err != nil {
		return time.Time{}, err
	}
	return date, nil
}

// isDateBefore reports whether date falls on a day before now's day.
func isDateBefore(date, now time.Time) (bool, error) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if date.Before(today) {
		return true, fmt.Errorf(i18n.T("commands.error.dateBeforeNow"), date.Format("02/01/2006"))
	}
	return false, nil
}

func getMonthByNum(num int) (time.Month, error) {
	month, ok := monthCodes[num]
	if !ok {
		return 0, errors.New(i18n.T("commands.error.invalidMonthNumber"))
	}

	return month, nil
}

var monthCodes map[int]time.Month = map[int]time.Month{
	1: time.January, 2: time.February, 3: time.March,
	4: time.April, 5: time.May, 6: time.June, 7: time.July,
	8: time.August, 9: time.September, 10: time.October,
	11: time.November, 12: time.December,
}
