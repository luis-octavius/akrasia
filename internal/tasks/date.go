package tasks

import (
	"errors"
	"fmt"
	"time"

	"github.com/luis-octavius/akrasia/pkg/i18n"
)

func isDateBefore(date time.Time) (bool, error) {
	isBefore := date.Before(time.Now())
	if isBefore == true {
		return isBefore, fmt.Errorf(i18n.T("commands.error.dateBeforeNow"), date)
	}

	return isBefore, nil
}

// ParseDate turns the CLI date parts (day, or day and month of the current
// year) into a due date. No parts means tomorrow; dates in the past are
// rejected.
func ParseDate(expireDate []int) (time.Time, error) {
	lenDate := len(expireDate)
	if lenDate > 2 {
		return time.Time{}, errors.New(i18n.T("commands.error.dateNotEnoughArguments"))
	}

	year, month, _ := time.Now().Date()
	var date time.Time

	switch lenDate {
	case 0:
		date = time.Now().AddDate(0, 0, 1)
	case 1:
		date = time.Date(year, month, expireDate[0], 0, 0, 0, 0, time.UTC)
	case 2:
		m, err := getMonthByNum(expireDate[1])
		if err != nil {
			return time.Time{}, err
		}
		date = time.Date(year, m, expireDate[0], 0, 0, 0, 0, time.UTC)
	}

	if _, err := isDateBefore(date); err != nil {
		return time.Time{}, err
	}

	return date, nil
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
