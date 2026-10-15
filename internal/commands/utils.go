package commands

import (
	"fmt"
	"github.com/luis-octavius/akrasia/pkg/i18n"
	"log"
	"time"
)

func isDateBefore(date time.Time) (bool, error) {
	isBefore := date.Before(time.Now())
	if isBefore == true {
		return isBefore, fmt.Errorf(i18n.T("commands.error.dateBeforeNow"), date)
	}

	return isBefore, nil
}

func parseDate(expireDate []int) (time.Time, error) {
	lenDate := len(expireDate)
	if lenDate > 2 {
		log.Fatal(i18n.T("commands.error.dateNotEnoughArguments"))
	}

	actualTime := time.Now()
	year, month, _ := actualTime.Date()
	date := time.Time{}

	switch lenDate {
	case 0:
		date = time.Now().AddDate(0, 0, 1)
		_, err := isDateBefore(date)
		if err != nil {
			log.Fatalf(i18n.T("commands.error.error"), err)
		}
		return date, nil
	case 1:
		date := time.Date(year, month, expireDate[0], 0, 0, 0, 0, time.UTC)
		_, err := isDateBefore(date)
		if err != nil {
			log.Fatalf(i18n.T("commands.error.error"), err)
		}
		return date, nil
	case 2:
		month = getMonthByNum(expireDate[1])
		date := time.Date(year, month, expireDate[0], 0, 0, 0, 0, time.UTC)
		_, err := isDateBefore(date)
		if err != nil {
			log.Fatalf(i18n.T("commands.error.error"), err)
		}
		return date, nil
	}

	return date, nil
}

func getMonthByNum(num int) time.Month {
	if num <= 0 || num > 12 {
		log.Fatal(i18n.T("commands.error.invalidMonthNumber"))
	}

	if month, ok := monthCodes[num]; ok {
		return month
	}
	
	// fallback
	return time.Now().Month()
}

var monthCodes map[int]time.Month = map[int]time.Month{
	1: time.January, 2: time.February, 3: time.March, 
	4: time.April, 5: time.May, 6: time.June, 7: time.July,
	8: time.August, 9: time.September, 10: time.October, 
	11: time.November, 12: time.December,
}
	

